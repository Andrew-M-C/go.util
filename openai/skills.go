package openai

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
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
