package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/inhies/go-bytesize"
	bs "github.com/inhies/go-bytesize"
	"gopkg.in/yaml.v3"
)

type config struct {
	FileName string      `yaml:"file_name"`
	FileSize byteSizeStr `yaml:"file_size"`
	Database struct {
		Username  string `yaml:"username"`
		Password  string `yaml:"password"`
		Host      string `yaml:"host"`
		Port      int32  `yaml:"port"`
		DB        string `yaml:"db"`
		TableName string `yaml:"table_name"`
	} `yaml:"database"`
}

type byteSizeStr bs.ByteSize

func (b *byteSizeStr) UnmarshalYAML(in *yaml.Node) error {
	res, err := bytesize.Parse(in.Value)
	if err != nil {
		return fmt.Errorf("解析失败 (%w)", err)
	}
	*b = byteSizeStr(res)
	return nil
}

func parseConfig() (c config, err error) {
	if len(os.Args) < 2 {
		return c, errors.New("请指定配置文件路径")
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		return c, fmt.Errorf("读取配置文件 (%v) 失败 (%w)", os.Args[1], err)
	}
	if err := yaml.Unmarshal(b, &c); err != nil {
		return c, fmt.Errorf("解析配置文件 (%v) 失败 (%w)", os.Args[1], err)
	}
	if c.FileSize <= 0 {
		c.FileSize = 10 * 1024 * 1024
	}
	if c.Database.Port <= 0 {
		c.Database.Port = 3306
	}
	if c.Database.DB == "" {
		return c, errors.New("请指定 database 名称")
	}
	if c.Database.TableName == "" {
		c.Database.TableName = "t_china_admin_districts"
	}
	return c, nil
}
