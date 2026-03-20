package openai_test

import (
	"path/filepath"
	"sort"
	"testing"

	utils "github.com/Andrew-M-C/go.util/openai"
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
			so(got.Metadata.Description, eq, want.Description)
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
