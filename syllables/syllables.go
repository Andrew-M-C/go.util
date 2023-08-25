// Package syllables 提供一个简单的音节处理逻辑, 目前仅支持中英文, 不支持其他语言
package syllables

import (
	"bytes"

	emoji "github.com/Andrew-M-C/go.emoji"
)

// Word 表示一个单词
type Word struct {
	// Word 表示单词内容
	Word string `json:"w,omitempty"`
	// SyllableCount 表示预估音节数量。-1 表示无法判断
	SyllableCount int `json:"c,omitempty"`
}

// SplitAndCount 拆分并计算各个音节长, 并按照拆分之后的结果返回单词 (中文是按照分词结果拆分)。
// 目前只支持中英文和颜文字。标点符号、空格和颜文字也会返回, 只是音节为零。
func SplitAndCount(text string) (total int, words []Word) {
	defer func() {
		for _, w := range words {
			total += w.SyllableCount
		}
	}()

	// 提取符号和 emoji
	textsWithSymbolsExtracted := splitSymbolsAndEmojis(text)

	// 提取英文
	var textWithEnglishExtracted []Word
	for _, item := range textsWithSymbolsExtracted {
		if item.SyllableCount >= 0 {
			textWithEnglishExtracted = append(textWithEnglishExtracted, item)
			continue
		}
		items := extractEnglishWords(item)
		textWithEnglishExtracted = append(textWithEnglishExtracted, items...)
	}

	// 提取剩余的, 统一按照 unicode 长度处理
	for _, item := range textWithEnglishExtracted {
		if item.SyllableCount >= 0 {
			words = append(words, item)
			continue
		}
		items := extractChineseWords(item.Word)
		words = append(words, items...)
	}

	return 0, words
}

func extractChineseWords(text string) (res []Word) {
	// 调用 gse 进行分词
	words := splitChineseSentences(text)
	for _, w := range words {
		res = append(res, Word{
			Word:          w,
			SyllableCount: utf8Len(w),
		})
	}
	return res
}

// 提取出英文单词并返回字数。纯阿拉伯数字按照中文处理
func extractEnglishWords(word Word) (res []Word) {
	add := func(w string, count int) {
		word := Word{
			Word:          w,
			SyllableCount: count,
		}
		res = append(res, word)
	}

	unEnglishChars := bytes.NewBuffer(make([]byte, 0, len(word.Word)))
	englishChars := bytes.NewBuffer(make([]byte, 0, len(word.Word)))

	for _, r := range word.Word {
		if isEnglishChar(r) {
			if unEnglishChars.Len() > 0 {
				add(unEnglishChars.String(), -1)
				unEnglishChars.Reset()
			}
			englishChars.WriteRune(r)
		} else {
			if englishChars.Len() > 0 {
				s := englishChars.String()
				add(s, estimateEnglishSyllables(s))
				englishChars.Reset()
			}
			unEnglishChars.WriteRune(r)
		}
	}

	if unEnglishChars.Len() > 0 {
		add(unEnglishChars.String(), -1)
	}
	if englishChars.Len() > 0 {
		s := englishChars.String()
		add(s, estimateEnglishSyllables(s))
	}
	return
}

// 初步拆分, 先把 emoji 和标点符号分离出来
func splitSymbolsAndEmojis(text string) (res []Word) {
	currValidText := bytes.NewBuffer(make([]byte, 0, len(text)))

	add := func(w string, count int) {
		word := Word{
			Word:          w,
			SyllableCount: count,
		}
		res = append(res, word)
	}

	strToRune := func(s string) rune {
		for _, r := range s {
			return r
		}
		return ' '
	}

	for it := emoji.IterateChars(text); it.Next(); {
		s := it.Current()
		if it.CurrentIsEmoji() ||
			s == " " ||
			internal.symbols.Has(strToRune(s)) {
			if currValidText.Len() != 0 {
				add(currValidText.String(), -1)
				currValidText.Reset()
			}
			add(s, 0)

		} else {
			currValidText.WriteString(s)
		}
	}

	if currValidText.Len() > 0 {
		add(currValidText.String(), -1)
	}
	return
}
