// Package syllables 提供一个简单的音节处理逻辑, 目前仅支持中英文, 不支持其他语言
package syllables

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Andrew-M-C/go.util/maps"
	"github.com/huichen/sego"
	"github.com/mtso/syllables"
)

//go:generate rm -f dictionary.txt
//go:generate wget https://raw.githubusercontent.com/huichen/sego/master/data/dictionary.txt

//go:embed dictionary.txt
var dictionary []byte

var internal = struct {
	symbols maps.Set[rune]
	sego    *sego.Segmenter
}{}

func init() {
	// reference: [2500个常用中文字符 + 130常用中英文字符](https://blog.csdn.net/qq285744011/article/details/125621736)
	const englishSymbols = `~!@#$%^&*()-_=+[{}]\|;:'",<.>/?/*` + "`"
	const chineseSymbols = `~·！@#￥%……&*（）——++-=、|【{}】；：‘“，《。》/？*`

	internal.symbols = maps.NewSet[rune]()
	for _, r := range englishSymbols + chineseSymbols {
		internal.symbols.Add(r)
	}

	// sego init
	initSego()
}

func initSego() {

	var tmpDictFile string
	switch strings.ToLower(runtime.GOOS) {
	case "windows":
		tmpDictFile = fmt.Sprintf("./tmp_dict_%d.txt", time.Now().UnixMilli())
	case "linux", "darwin", "freebsd":
		tmpDictFile = fmt.Sprintf("/tmp/tmp_dict_%d.txt", time.Now().UnixMilli())
	default: // 未知类型, 尝试当前文件
		tmpDictFile = fmt.Sprintf("./tmp_dict_%d.txt", time.Now().UnixMilli())
	}

	_ = os.WriteFile(tmpDictFile, dictionary, 0644)
	defer os.Remove(tmpDictFile)

	s := &sego.Segmenter{}
	s.LoadDictionary(tmpDictFile)
	internal.sego = s
}

func isEnglishChar(r rune) bool {
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= 'a' && r <= 'z' {
		return true
	}
	return false
}

func isNumericChar(r rune) bool {
	return r >= '0' && r <= '9'
}

// 计算英语单词音节数
func estimateEnglishSyllables(word string) int {
	return syllables.In(word)
}

// 拆分中文语句
func splitChineseSentences(text string) []string {
	cut := internal.sego.Segment([]byte(text))
	return sego.SegmentsToSlice(cut, false)
}

func utf8Len(s string) (length int) {
	for range s {
		length++
	}
	return
}
