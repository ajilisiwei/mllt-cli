package practice

import (
	"regexp"
	"strings"
)

// trailingPhonetic 匹配紧跟在单词后面的音标，例如 "fool/fuːl/" 或 "gay /ɡeɪ/"。
// 必须以 "/" 结尾才算音标，因此 "and/or"、"km/h" 之类的词条不会被误拆。
var trailingPhonetic = regexp.MustCompile(`^(.*?)\s*(/[^/\t]*/)\s*$`)

// ParseWordLine 解析单词行，返回单词、音标和释义。
//
// 兼容仓库里出现过的全部写法：
//   - "word\t/音标/\t释义"    标准三列
//   - "word/音标/\t释义"      缺少第一个制表符（历史脏数据）
//   - "word /音标/\t释义"     用空格代替制表符
//   - "word / 音标 / 释义"    早期空格分隔格式
//   - "word ->> /音标/ 释义"  箭头分隔格式
//
// 音标缺失时 phonetic 返回空字符串。
func ParseWordLine(line string) (word string, phonetic string, meaning string) {
	primary, rest := ParseLine(line)

	word, phonetic = splitTrailingPhonetic(primary)
	if phonetic == "" {
		phonetic, rest = splitLeadingPhonetic(rest)
	}

	return word, phonetic, strings.TrimSpace(rest)
}

// WordPrimaryText 返回单词行中需要用户输入的正文部分。
func WordPrimaryText(line string) string {
	word, _, _ := ParseWordLine(line)
	if word != "" {
		return word
	}
	return line
}

// splitTrailingPhonetic 从 "word/音标/" 形式中拆出单词和音标。
func splitTrailingPhonetic(text string) (string, string) {
	text = strings.TrimSpace(text)

	matches := trailingPhonetic.FindStringSubmatch(text)
	if matches == nil {
		return text, ""
	}

	word := strings.TrimSpace(matches[1])
	if word == "" {
		// 整段就是音标，没有单词可拆，保持原样。
		return text, ""
	}

	return word, normalizePhonetic(matches[2])
}

// splitLeadingPhonetic 从 "/音标/ 释义" 形式中拆出音标和剩余释义。
func splitLeadingPhonetic(text string) (string, string) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "/") {
		return "", text
	}

	end := strings.IndexByte(trimmed[1:], '/')
	if end < 0 {
		return "", text
	}
	end++

	body := trimmed[1:end]
	if strings.ContainsAny(body, "\t\n") {
		return "", text
	}

	return normalizePhonetic(trimmed[:end+1]), trimmed[end+1:]
}

// normalizePhonetic 去掉音标内侧多余的空白，"/ əˈbændən /" -> "/əˈbændən/"。
func normalizePhonetic(phonetic string) string {
	body := strings.TrimSpace(strings.Trim(phonetic, "/"))
	if body == "" {
		return ""
	}
	return "/" + body + "/"
}
