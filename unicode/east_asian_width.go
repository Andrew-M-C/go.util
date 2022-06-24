package unicode

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	emoji "github.com/Andrew-M-C/go.emoji"
)

// Align 表示对齐方式
type Align int

const (
	AlignRight Align = iota
	AlignLeft
	AlignCenter
)

// ActualEastAsianWidth 返回一个 Formatter 用于按照东亚字符真正的字符宽度进行展示
//
// Reference:
//
// - golang获取字符的宽度(East_Asian_Width) - http://www.nbtuan.vip/2017/05/10/golang-char-width/
func ActualEastAsianWidth(v interface{}, asciiWidth int, opts ...Option) Formatter {
	f := eastAsianWidthFmt{
		v:     v,
		width: asciiWidth,
		opt:   defaultOption(),
	}
	for _, o := range opts {
		o(f.opt)
	}
	return f
}

type eastAsianWidthFmt struct {
	v     interface{}
	width int
	opt   *option
}

func (f eastAsianWidthFmt) String() string {
	orig := fmt.Sprint(f.v)
	s := orig

	actualWidth := eastAsianStringWidth(s)
	if actualWidth >= f.width {
		return s
	}

	spaceWidth := f.width - actualWidth
	switch f.opt.align {
	default:
		fallthrough
	case AlignRight:
		return f.opt.blanks(spaceWidth) + orig
	case AlignLeft:
		return orig + f.opt.blanks(spaceWidth)
	case AlignCenter:
		leftWidth := spaceWidth / 2
		rightWidth := spaceWidth - leftWidth
		return f.opt.blanks(leftWidth) + orig + f.opt.blanks(rightWidth)
	}
}

func eastAsianStringWidth(s string) int {
	width := 0

	s = emoji.ReplaceAllEmojiFunc(s, func(_ string) string {
		width += 2
		return ""
	})

	for _, r := range s {
		if w, exist := internal.eastAsianWidth[r]; exist {
			width += w
		} else {
			width++
		}
	}

	return width
}

//go:generate rm -f EastAsianWidth.txt
//go:generate wget http://www.unicode.org/Public/UCD/latest/ucd/EastAsianWidth.txt

//go:embed EastAsianWidth.txt
var standardFile string

func init() {
	// 解析 unicode 的 EastAsianWidth.txt 文件
	parseStandardFile()
}

func parseStandardFile() {
	internal.eastAsianWidth = make(map[rune]int)
	lines := strings.Split(standardFile, "\n")
	for _, line := range lines {
		parseStandardLine(line)
	}
}

func parseStandardLine(line string) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "#") {
		return
	}

	tail := strings.Index(line, "#")
	if tail <= 0 {
		return
	}
	line = line[:tail]
	line = strings.TrimSpace(line)
	parts := strings.Split(line, ";")
	if len(parts) < 2 {
		fmt.Println(fmt.Errorf("illegal line: '%s'", line))
		return
	}

	property := parts[1]
	start, end, err := parseRunes(parts[0])
	if err != nil {
		fmt.Println(fmt.Errorf("illegal line: '%s'", line))
		return
	}

	width := 1
	switch property {
	default:
		fallthrough
	case "A", "H", "N", "Na":
		width = 1
	case "F", "W":
		width = 2
	}

	for r := start; r <= end; r++ {
		internal.eastAsianWidth[r] = width
	}
}

func parseRunes(s string) (start, end rune, err error) {
	parts := strings.Split(s, "..")
	if len(parts) == 1 {
		i, err := strconv.ParseInt(s, 16, 32)
		return rune(i), rune(i), err
	}

	startI, err := strconv.ParseInt(parts[0], 16, 32)
	if err != nil {
		return 0, 0, err
	}
	endI, err := strconv.ParseInt(parts[1], 16, 32)
	return rune(startI), rune(endI), err
}
