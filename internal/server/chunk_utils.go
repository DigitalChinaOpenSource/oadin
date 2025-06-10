package server

import (
	"unicode"
)

// 尝试在一段文本中找到句子的边界
// 返回找到的句子边界的索引，如果没找到则返回end
// start: 起始位置，end: 结束位置的上限
func findSentenceBoundary(text []rune, start, end int) int {
	// 句子结束标记
	sentenceEnders := []rune{'.', '!', '?', '。', '！', '？', '；', ';', '\n', '\r'}

	// 从后往前查找，尝试在句子边界处分割
	for i := end - 1; i > start; i-- {
		// 如果找到句子结束标记
		for _, ender := range sentenceEnders {
			if text[i] == ender {
				// 向后移动一位，包含结束标记
				if i+1 < end {
					return i + 1
				}
				return i
			}
		}

		// 找到段落分隔
		if i+1 < end && text[i] == '\n' && text[i+1] == '\n' {
			return i
		}
	}

	// 如果没找到句子边界，尝试在单词边界处分割
	for i := end - 1; i > start+10; i-- { // 至少向前查找10个字符
		if unicode.IsSpace(text[i]) && !unicode.IsSpace(text[i+1]) {
			return i + 1
		}
	}

	// 没有找到合适的边界，返回end
	return end
}

// 检测一个字符是否为标点符号
func isPunctuation(r rune) bool {
	return unicode.IsPunct(r) ||
		r == '。' || r == '，' || r == '、' || r == '：' ||
		r == '；' || r == '！' || r == '？' ||
		r == '"' || r == '（' || r == '）' || r == '【' ||
		r == '】' || r == '《' || r == '》'
}
