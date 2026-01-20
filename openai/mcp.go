package openai

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

type mcpClientWithSpecifiedTools struct {
	client       InitializedMCPClient
	includeTools map[string]struct{}
}

func (c *mcpClientWithSpecifiedTools) ListTools(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	orig, err := c.client.ListTools(ctx, request)
	if err != nil {
		return nil, err
	}
	res := &mcp.ListToolsResult{
		PaginatedResult: orig.PaginatedResult,
		Tools:           make([]mcp.Tool, 0, len(c.includeTools)),
	}
	for _, tool := range orig.Tools {
		if _, exist := c.includeTools[tool.Name]; !exist {
			continue
		}
		res.Tools = append(res.Tools, tool)
	}
	return res, nil
}

func (c *mcpClientWithSpecifiedTools) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return c.client.CallTool(ctx, request)
}
