package openai

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/mark3labs/mcp-go/mcp"
)

// SkillMetadata 表示一个 SKILL.md 的 metadata 部分
type SkillMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// SkillInfo 表示一个包含 SKILL.md skill
type SkillInfo struct {
	Metadata SkillMetadata
	Dir      string
}

// ListAllSkills 列出指定目录下的所有 SKILL.md 文件
func ListAllSkills(dir string) ([]SkillInfo, error) {
	var skills []SkillInfo

	err := filepath.WalkDir(dir, walkSkillDir(&skills))
	if err != nil {
		return nil, err
	}

	return skills, nil
}

func walkSkillDir(skills *[]SkillInfo) func(string, os.DirEntry, error) error {
	return func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.EqualFold(d.Name(), "SKILL.md") {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		var meta SkillMetadata
		if _, err = frontmatter.MustParse(f, &meta); err != nil {
			return err
		}
		meta.Name = strings.TrimSpace(meta.Name)
		meta.Description = strings.TrimSpace(meta.Description)

		*skills = append(*skills, SkillInfo{
			Metadata: meta,
			Dir:      filepath.Dir(path),
		})
		return nil
	}
}

// ToolsForSkill 返回两个可用于 SKILL 能力的工具: Read 和 Bash。请注意: 这仅适用于类 UNIX
func ToolsForSkill() ToolManager {
	return readAndBashTools{}
}

var skillTools = []mcp.Tool{
	mcp.NewTool("Read",
		mcp.WithDescription("Reads a file from the local filesystem. "+
			"Lines in the output are numbered: \"LINE_NUMBER|LINE_CONTENT\"."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("The absolute path of the file to read."),
		),
		mcp.WithNumber("offset",
			mcp.Description("The line number to start reading from. "+
				"Positive values are 1-indexed from the start; "+
				"negative values count backwards from the end (e.g. -1 is the last line)."),
		),
		mcp.WithNumber("limit",
			mcp.Description("The maximum number of lines to read."),
		),
	),
	mcp.NewTool("Bash",
		mcp.WithDescription("Executes a given command in a bash shell session."),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("The shell command to execute."),
		),
		mcp.WithString("description",
			mcp.Description("A short description of what the command does in 5-10 words."),
		),
		mcp.WithString("working_directory",
			mcp.Description("The absolute path to the working directory to execute the command in."),
		),
		mcp.WithNumber("block_until_ms",
			mcp.Description("How long to block and wait for the command to complete (in milliseconds). "+
				"Defaults to 30000ms (30 seconds)."),
		),
	),
}

type readAndBashTools struct{}

func (readAndBashTools) ListTools(_ context.Context, _ mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return &mcp.ListToolsResult{Tools: skillTools}, nil
}

func (readAndBashTools) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	switch req.Params.Name {
	case "Read":
		return callSkillRead(args)
	case "Bash":
		return callSkillBash(ctx, args)
	default:
		return mcp.NewToolResultError("unknown tool: " + req.Params.Name), nil
	}
}

func callSkillRead(args map[string]any) (*mcp.CallToolResult, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	lines := strings.Split(string(data), "\n")
	total := uint64(len(lines))

	offset, ok := anyToInt(args["offset"])
	if !ok {
		return mcp.NewToolResultError("illegal 'offset' parameter"), nil
	}

	limit, ok := anyToInt(args["limit"])
	if !ok {
		return mcp.NewToolResultError("illegal 'limit' parameter"), nil
	}
	if limit == 0 {
		limit = total
	}
	if offset >= total {
		return mcp.NewToolResultText(""), nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	var sb strings.Builder
	for i, line := range lines[offset:end] {
		fmt.Fprintf(&sb, "%6d|%s\n", offset+uint64(i)+1, line)
	}
	return mcp.NewToolResultText(sb.String()), nil
}

func callSkillBash(_ context.Context, args map[string]any) (*mcp.CallToolResult, error) {
	command, _ := args["command"].(string)
	if command == "" {
		return mcp.NewToolResultError("command is required"), nil
	}

	workDir, _ := args["working_directory"].(string)

	cmd := exec.Command("bash", "-c", command)
	if workDir != "" {
		cmd.Dir = workDir
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()

	var sb strings.Builder
	if outBuf.Len() > 0 {
		sb.WriteString(outBuf.String())
	}
	if errBuf.Len() > 0 {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString("[stderr]\n")
		sb.WriteString(errBuf.String())
	}
	if runErr != nil {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString("[error] ")
		sb.WriteString(runErr.Error())
		return mcp.NewToolResultError(sb.String()), nil
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func anyToInt(v any) (uint64, bool) {
	if v == nil {
		return 0, true
	}
	val := reflect.ValueOf(v)

	switch val.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := val.Int()
		return uint64(i), i >= 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint(), true

	case reflect.String:
		u, err := strconv.ParseUint(val.String(), 10, 64)
		return u, (err == nil)

	default:
		return 0, false
	}
}
