package unicode

import (
	"strings"
	"unicode/utf8"
)

// TrimUTF8 按照 maxRunes 和 maxBytes 值限制字符串的长度，如果字符串长度超过 maxRunes 或
// maxBytes, 则截取字符串的末尾, 并添加 tail 字符串。其中 maxXXX 取 <= 0 表示不限制, tail
// 取 "" 表示不添加尾部。本函数确保加了 tail 之后也不超 maxXXX 值。
//
// 请确保传入的 orig 和 tail 均为 UTF8 字符串, 否则逻辑会出现意想不到的错误
//
// Options 暂时只支持 WithDebugger
func TrimUTF8(orig string, tail string, maxRunes, maxBytes int, opts ...Option) string {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}
	if maxRunes <= 0 && maxBytes <= 0 {
		return orig // 全无限制, 直接返回
	}

	if maxRunes > 0 {
		if count := utf8.RuneCountInString(tail); count > maxRunes {
			opt.debug("tail 字符串 '%s' 拥有 %d 个字符, 超过了限制的 %d, 只能设置为空", tail, maxRunes)
			tail = ""
		}
	}
	if maxBytes > 0 {
		if count := len(tail); count > maxBytes {
			opt.debug("tail 字符串 '%s' 拥有 %d 个字节, 超过了限制的 %d, 只能设置为空", tail, maxBytes)
			tail = ""
		}
	}

	tailByteLen := len(tail)
	tailRuneLen := utf8.RuneCountInString(tail)

	var possibleRunes []rune
	var possibleBytes int
	needTrim := false

	opt.debug(
		"tail bytes %d, runes %d, orig bytes %d, rune limit %d, bytes limit %d",
		tailByteLen, tailRuneLen, len(orig), maxRunes, maxBytes,
	)

	// 首先从前往后找, 找到限制值
	for _, r := range orig {
		possibleBytes += len(string(r))
		possibleRunes = append(possibleRunes, r)
		if maxRunes > 0 && len(possibleRunes) > maxRunes {
			needTrim = true
			break
		}
		if maxBytes > 0 && possibleBytes > maxBytes {
			needTrim = true
			break
		}
	}

	// 然后从挑选出来的从后往前找, 找到需要截断的点
	if !needTrim {
		opt.debug("无需截断, runes %d, bytes %d", len(possibleRunes), possibleBytes)
		return string(possibleRunes)
	}
	opt.debug("需截断, runes %d, bytes %d", len(possibleRunes), possibleBytes)

	for len(possibleRunes) > 0 {
		le := len(possibleRunes)
		r := possibleRunes[le-1]

		// rune 超限
		if maxRunes > 0 && len(possibleRunes)+tailRuneLen > maxRunes {
			possibleRunes = possibleRunes[:le-1]
			possibleBytes -= len(string(r))
			continue
		}
		// bytes 超限
		if maxBytes > 0 && possibleBytes+tailByteLen > maxBytes {
			possibleRunes = possibleRunes[:le-1]
			possibleBytes -= len(string(r))
			continue
		}
		// else
		break
	}

	// 返回
	bdr := strings.Builder{}
	for _, r := range possibleRunes {
		bdr.WriteRune(r)
	}
	bdr.WriteString(tail)
	return bdr.String()
}
