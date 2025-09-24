package openai

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"

	hutil "github.com/Andrew-M-C/go.util/net/http"
	"github.com/Andrew-M-C/go.util/unsafe"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
)

const mcpClientNameSeparator = "-"

type processor struct {
	// 入参
	Conf     ModelConfig
	Opts     *options
	Messages []openai.ChatCompletionMessage // 同时也是出参

	// 中间参数
	deferFuncs    []func(context.Context)
	mcpClientByID map[string]InitializedMCPClient
	mcpTools      []openai.Tool

	lastFinishReason openai.FinishReason
}

func (p *processor) do(ctx context.Context) (ProcessResponse, error) {
	defer p.callDefers(ctx)

	// 一些初始化工作
	p.mcpClientByID = make(map[string]InitializedMCPClient,
		len(p.Opts.mcpURL)+len(p.Opts.customizeMCPs),
	)
	// 主流程
	procedures := []func(context.Context) error{
		p.copyMessages,      // 浅拷贝入参
		p.addCustomizedMCPs, // 自定义 MCP 或者是初始化好了的 MCP
		p.connectRemoteMCP,  // 连接远程 MCP
		p.packMCPTools,      // 打包 MCP 工具作为后续请求的参数
		p.iteration,         // 开始迭代
	}
	for _, proc := range procedures {
		if err := proc(ctx); err != nil {
			return ProcessResponse{}, err
		}
	}
	// 打包返回
	rsp := ProcessResponse{
		Messages:     p.Messages,
		FinishReason: p.lastFinishReason,
	}
	return rsp, nil
}

func (p *processor) callDefers(ctx context.Context) {
	slices.Reverse(p.deferFuncs)
	for _, fn := range p.deferFuncs {
		fn(ctx)
	}
}

func (p *processor) copyMessages(ctx context.Context) error {
	if len(p.Messages) == 0 {
		return errors.New("没有待请求的消息")
	}
	p.Messages = slices.Clone(p.Messages)
	return nil
}

func (p *processor) connectRemoteMCP(ctx context.Context) error {
	iterateURL := func(index int, url string) error {
		cli, err := mcpclient.NewSSEMCPClient(url)
		if err != nil {
			return fmt.Errorf("连接 MCP '%s' 失败 (%w)", url, err)
		}

		// 添加关闭操作
		p.deferFuncs = append(p.deferFuncs, func(ctx context.Context) {
			cli.Close()
		})

		// 启动客户端，获取endpoint
		if err := cli.Start(ctx); err != nil {
			return fmt.Errorf("启动 MCP 客户端 '%s' 失败 (%w)", url, err)
		}

		// 初始化
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "llm-client",
			Version: "1.0.0",
		}
		initResult, err := cli.Initialize(ctx, initRequest)
		if err != nil {
			return fmt.Errorf("初始化 MCP '%s' 失败 (%w)", url, err)
		}

		id := p.mcpID(cli)
		p.mcpClientByID[id] = cli

		p.Opts.debugf(
			"连接 MCP '%s' 并初始化成功, name '%s', id: %s, version '%s'",
			url, initResult.ServerInfo.Name, id, initResult.ServerInfo.Version,
		)
		return nil
	}

	for i, u := range p.Opts.mcpURL {
		if err := iterateURL(i, u); err != nil {
			return err
		}
	}
	return nil
}

func (p *processor) addCustomizedMCPs(ctx context.Context) error {
	for _, cli := range p.Opts.customizeMCPs {
		p.mcpClientByID[p.mcpID(cli)] = cli
	}
	return nil
}

func (p *processor) packMCPTools(ctx context.Context) error {
	if len(p.mcpClientByID) == 0 {
		p.Opts.debugf("没有配置远程 MCP 工具")
		return nil
	}

	for id, cli := range p.mcpClientByID {
		t, err := cli.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			return fmt.Errorf("获取 MCP '%s' 工具列表失败 (%w)", id, err)
		}
		if t == nil {
			continue
		}
		for _, t := range t.Tools {
			fu := &openai.FunctionDefinition{
				Name:        fmt.Sprintf("%s%s%s", id, mcpClientNameSeparator, t.Name),
				Description: t.Description,
				Parameters:  t.InputSchema,
			}
			p.mcpTools = append(p.mcpTools, openai.Tool{
				Type:     openai.ToolTypeFunction,
				Function: fu,
			})
		}
	}
	p.Opts.debugf("打包 MCP 工具成功: %v", toJSON{p.mcpTools})
	return nil
}

func (p *processor) iteration(ctx context.Context) error {
	iterate := func() error {
		// 先只单次迭代一次
		oneTimeProcessor := &oneTimeProcessor{
			processor: p,
		}
		rsp, err := oneTimeProcessor.do(ctx)
		if err != nil {
			return fmt.Errorf("问答错误 (%w)", err)
		}

		p.Messages = append(p.Messages, openai.ChatCompletionMessage{
			Role:             openai.ChatMessageRoleAssistant,
			ReasoningContent: rsp.Choices[0].Delta.ReasoningContent,
			Content:          rsp.Choices[0].Delta.Content,
			ToolCalls:        rsp.Choices[0].Delta.ToolCalls,
		})
		p.lastFinishReason = rsp.Choices[0].FinishReason
		return nil
	}

	for {
		if err := iterate(); err != nil {
			return err
		}

		// 检查一下最后一个消息, 看看是不是已经结束了
		switch p.lastFinishReason {
		case openai.FinishReasonStop, openai.FinishReasonLength:
			return nil

		case openai.FinishReasonToolCalls:
			if len(p.lastMessage().ToolCalls) == 0 {
				p.Opts.debugf("大模型返回要求工具调用, 但没有返回工具调用列表, 视为异常, 直接结束")
				return nil
			}
			tcP := &toolProcessor{processor: p}
			if err := tcP.do(ctx); err != nil {
				return fmt.Errorf("工具调用错误 (%w)", err)
			}
			// continue to next iteration

		case openai.FinishReasonFunctionCall:
			// TODO: 暂不支持, 要找一个使用 function call 的模型试试看
			return nil
		default:
			return nil
		}
	}
}

func (p *processor) lastMessage() openai.ChatCompletionMessage {
	return p.Messages[len(p.Messages)-1]
}

func (p *processor) mcpID(cli InitializedMCPClient) string {
	desc := fmt.Sprintf("%p-%v", cli, reflect.TypeOf(cli)) // 不是 pointer 的话, 自然会把 type 打印出来

	// MD5 hash
	hash := md5.Sum(unsafe.StoB(desc))
	return hex.EncodeToString(hash[:])
}

// -------- 单次迭代 --------

type oneTimeProcessor struct {
	*processor
}

func (p *oneTimeProcessor) do(ctx context.Context) (openai.ChatCompletionStreamResponse, error) {
	emptyRsp := openai.ChatCompletionStreamResponse{}

	// 首先发起请求, 获取响应
	rsp, err := connect(ctx, p.Conf, p.Messages, p.mcpTools, p.Opts)
	if err != nil {
		return emptyRsp, fmt.Errorf("发起请求失败 (%w)", err)
	}
	defer rsp.Body.Close()

	// 逐步接收响应
	builder := &streamBuilder{opts: p.Opts}
	if err := hutil.ReadSSEJsonData(ctx, rsp.Body, builder.AddResponse, ignoreNonJSON()); err != nil {
		return emptyRsp, fmt.Errorf("读取 SSE 数据失败 (%w)", err)
	}
	return builder.Done(), nil
}

// -------- 工具调用 --------

type toolProcessor struct {
	*processor
}

func (p *toolProcessor) do(ctx context.Context) error {
	tcList := p.lastMessage().ToolCalls

	var err error
	lck := sync.Mutex{}
	wg := sync.WaitGroup{}

	// 并发调用
	for i, tc := range tcList {
		p.Opts.debugf("需要调用工具: %v", tc.Function.Name)
		p.Opts.toolCallRequestCallback(tc)

		wg.Add(1)
		go func(i int, tc openai.ToolCall) {
			defer wg.Done()

			res, e := p.doToolCall(ctx, tc)

			lck.Lock()
			defer lck.Unlock()

			if e != nil {
				err = e
				return
			}
			m := openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    res,
				ToolCallID: tc.ID,
			}
			p.Opts.debugf("工具调用返回: %v", toJSON{m})
			p.Messages = append(p.Messages, m)

			p.Opts.toolCallResponseCallback(m)
		}(i, tc)
	}

	wg.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (p *toolProcessor) doToolCall(ctx context.Context, tc openai.ToolCall) (string, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("解析工具调用参数失败 (%w)", err)
	}

	parts := strings.SplitN(tc.Function.Name, mcpClientNameSeparator, 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("工具调用名称格式错误 (%s)", tc.Function.Name)
	}

	clientID, toolName := parts[0], parts[1]
	client, exist := p.mcpClientByID[clientID]
	if !exist {
		return "", fmt.Errorf("未找到 MCP 客户端 (%s)", clientID)
	}

	req := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}

	rsp, err := client.CallTool(ctx, req)
	if err != nil {
		return "", fmt.Errorf("调用 MCP 工具 '%s' 失败 (%w)", toolName, err)
	}
	if len(rsp.Content) == 0 {
		return "", fmt.Errorf("调用 MCP 工具 '%s' 但没有返回", toolName)
	}
	text, ok := rsp.Content[0].(mcp.TextContent)
	if !ok {
		return "", fmt.Errorf("content is not a text content (%v)", reflect.TypeOf(rsp.Content[0]))
	}

	p.Opts.debugf("调用工具 '%s', 返回 '%s'", toolName, text.Text)
	return text.Text, nil
}

// -------- 内部函数 --------

type toJSON struct {
	v any
}

func (t toJSON) String() string {
	b, _ := json.Marshal(t.v)
	return string(b)
}

func ignoreNonJSON() hutil.RequestOption {
	return hutil.WithSSEUnmarshalErrorCallback(func(err error, data string) {})
}
