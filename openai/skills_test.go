package openai_test

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	utils "github.com/Andrew-M-C/go.util/openai"
	"github.com/sashabaranov/go-openai"
)

// 测试目录结构（test_dir/）：
//
//	test_dir/
//	  skill-a/          SKILL.md  (同层级 1)
//	  skill-a/skill-b/  SKILL.md  (二级嵌套)
//	  skill-a/skill-b/skill-c/  SKILL.md  (三级嵌套)
//	  skill-d/          SKILL.md  (同层级 2)
//	  no-skill/         README.md，无 SKILL.md，应被忽略
//	  empty/            .gitkeep，无 SKILL.md

func TestListAllSkills(t *testing.T) {
	cv("多层级、同层级 SKILL.md 场景", t, func() {
		root, err := filepath.Abs("test_dir")
		so(err, isNil)

		wantSkills := map[string]utils.SkillMetadata{
			"skill-a":                 {Name: "skill-a", Description: "技能 A"},
			"skill-d":                 {Name: "skill-d", Description: "技能 D"},
			"skill-a/skill-b":         {Name: "skill-b", Description: "技能 B（二级）"},
			"skill-a/skill-b/skill-c": {Name: "skill-c", Description: "技能 C（三级）"},
			"skill-date":              {Name: "date-query", Description: ""},
		}

		skills, err := utils.ListAllSkills(root)
		so(err, isNil)
		so(len(skills), eq, len(wantSkills))

		sort.Slice(skills, func(i, j int) bool {
			return skills[i].Dir < skills[j].Dir
		})

		gotByRel := make(map[string]utils.SkillInfo, len(skills))
		for _, s := range skills {
			rel, relErr := filepath.Rel(root, s.Dir)
			so(relErr, isNil)
			gotByRel[rel] = s
		}

		for rel, want := range wantSkills {
			got, ok := gotByRel[rel]
			so(ok, eq, true)
			so(got.Metadata.Name, eq, want.Name)

			if want.Description != "" {
				so(got.Metadata.Description, eq, want.Description)
			}

			so(got.Dir, eq, filepath.Join(root, rel))
		}
	})

	cv("无 SKILL.md 的目录应返回空列表", t, func() {
		root, err := filepath.Abs("test_dir/empty")
		so(err, isNil)

		skills, err := utils.ListAllSkills(root)
		so(err, isNil)
		so(len(skills), eq, 0)
	})

	cv("不存在的目录应返回错误", t, func() {
		_, err := utils.ListAllSkills("/nonexistent/path/that/does/not/exist")
		so(err, notNil)
	})
}

func TestToolsForSkill(t *testing.T) {
	cv("通过 SKILL 调用 date 命令询问当前时间", t, func() {
		ctx := context.Background()

		skillDir, err := filepath.Abs("test_dir/skill-date")
		so(err, isNil)
		skills, err := utils.ListAllSkills(skillDir)
		so(err, isNil)
		so(len(skills), eq, 1)

		systemPrompt := buildSkillSystemPrompt(skills)

		config := utils.ModelConfig{
			Model:   deepseekModel,
			BaseURL: deepseekBaseURL,
			APIKey:  deepseekAPIKey,
		}
		messages := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: "现在几点了？"},
		}

		content := func(c string) { fmt.Printf("%s", c) }
		tcReq := func(tc openai.ToolCall) { printf("工具调用: %s %s", tc.Function.Name, tc.Function.Arguments) }
		tcRsp := func(m openai.ChatCompletionMessage) { printf("工具结果: %s", m.Content) }

		rsp, err := utils.Process(ctx, config, messages,
			utils.WithContentCallback(content),
			utils.WithToolCallRequestCallback(tcReq),
			utils.WithToolCallResponseCallback(tcRsp),
			utils.WithTools(utils.ToolsForSkill(), "skill-tools"),
		)
		so(err, isNil)
		so(rsp, notNil)
		so(len(rsp.Messages), ge, 3) // 至少: 系统+问题、工具调用轮、最终回答

		answer := rsp.Messages[len(rsp.Messages)-1].Content
		t.Logf("\n最终回答: %s", answer)
		so(len(answer), ge, 1)

		// 验证回答中包含类似 HH:MM 的时间格式
		so(regexp.MustCompile(`\d{1,2}:\d{2}`).MatchString(answer), eq, true)
	})
}

func buildSkillSystemPrompt(skills []utils.SkillInfo) string {
	var sb strings.Builder
	sb.WriteString("You are a helpful assistant.\n")
	sb.WriteString("When a skill is relevant, read and follow it IMMEDIATELY as your first action using the Read tool.\n\n")
	sb.WriteString("<available_skills>\n")
	for _, s := range skills {
		skillPath := filepath.Join(s.Dir, "SKILL.md")
		fmt.Fprintf(&sb, "<agent_skill fullPath=%q>\n  %s\n</agent_skill>\n", skillPath, s.Metadata.Description)
	}
	sb.WriteString("</available_skills>")
	return sb.String()
}
