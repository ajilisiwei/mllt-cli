package practice

import (
	"regexp"
	"strings"
)

// trailingPhonetic 匹配紧跟在单词后面的音标，例如 "fool/fuːl/" 或 "gay /ɡeɪ/"。
// 必须以 "/" 结尾才算音标，因此 "and/or"、"km/h" 之类的词条不会被误拆。
var trailingPhonetic = regexp.MustCompile(`^(.*?)\s*(/[^/\t]*/)\s*$`)

// WordEntry 表示一条单词记录。列按使用频率排序，靠后的列可以整列省略：
//
//	单词 \t 音标 \t 释义 \t 例句 \t 例句翻译 \t 搭配 \t 词族
//
// 高频简单词只写前三列即可；只有"认识≠会用"的词才值得补上例句、搭配和词族。
type WordEntry struct {
	Word        string
	Phonetic    string
	Meaning     string
	Example     string
	ExampleNote string
	Collocation string
	Family      string
}

// ParseWordEntry 解析单词行。四列以上按上面的列定义读取，三列及以下走兼容解析，
// 因此现有词库和历史脏数据都不受影响。
func ParseWordEntry(line string) WordEntry {
	cols := strings.Split(line, "\t")
	if len(cols) < 4 {
		word, phonetic, meaning := parseLegacyWordLine(line)
		return WordEntry{Word: word, Phonetic: phonetic, Meaning: meaning}
	}

	column := func(i int) string {
		if i < len(cols) {
			return strings.TrimSpace(cols[i])
		}
		return ""
	}

	// 单词列理论上是干净的，但导入的数据未必，所以仍然尝试剥离黏在后面的音标
	word, phonetic := splitTrailingPhonetic(column(0))
	if written := column(1); written != "" {
		phonetic = normalizePhonetic(written)
	}

	return WordEntry{
		Word:        word,
		Phonetic:    phonetic,
		Meaning:     column(2),
		Example:     column(3),
		ExampleNote: column(4),
		Collocation: column(5),
		Family:      column(6),
	}
}

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
	entry := ParseWordEntry(line)
	return entry.Word, entry.Phonetic, entry.Meaning
}

// parseLegacyWordLine 解析三列及以下的历史写法。
func parseLegacyWordLine(line string) (word string, phonetic string, meaning string) {
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
