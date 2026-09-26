package practice

import (
	"fmt"
	"regexp"
	"strings"
)

// trailingPhonetic 匹配紧跟在单词后面的音标，例如 "fool/fuːl/" 或 "gay /ɡeɪ/"。
// 必须以 "/" 结尾才算音标，因此 "and/or"、"km/h" 之类的词条不会被误拆。
var trailingPhonetic = regexp.MustCompile(`^(.*?)\s*(/[^/\t]*/)\s*$`)

// WordEntry 表示一条单词记录。定长列在前，例句在后按「例句 + 译文」成对排列，
// 数量不限：
//
//	单词 \t 音标 \t 释义 \t 搭配 \t 词族 \t 例句1 \t 译文1 \t 例句2 \t 译文2 \t …
//
// 高频简单词只写前三列即可；只有"认识≠会用"的词才值得补上搭配、词族和例句。
// 一个词配多个例句是有意的——单个例句只教会一种搭配，看过它在几种语境里怎么用，
// 才谈得上掌握。
type WordEntry struct {
	Word        string
	Phonetic    string
	Meaning     string
	Collocation string
	Family      string
	// Examples 是这个词的全部例句，按出现顺序排列。
	Examples []Contrast
	// Example / ExampleNote 指向第一个例句，保留给既有调用方。
	Example     string
	ExampleNote string
}

// ParseWordEntry 解析单词行。三列以上按上面的列定义读取，更少的列走兼容解析，
// 因此现有词库和历史脏数据都不受影响。
func ParseWordEntry(line string) WordEntry {
	cols := strings.Split(line, "\t")
	// 三列（单词 / 音标 / 释义）已经是标准格式，按列解析；
	// 更少的列说明是缺制表符之类的历史写法，交给兼容解析去兜。
	if len(cols) < 3 {
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

	entry := WordEntry{
		Word:        word,
		Phonetic:    phonetic,
		Meaning:     column(2),
		Collocation: column(3),
		Family:      column(4),
	}

	// 第六列起每两列一个例句；最后一句缺译文时译文留空
	for i := 5; i < len(cols); i += 2 {
		example := Contrast{Text: column(i), Note: column(i + 1)}
		if example.Text == "" {
			continue
		}
		entry.Examples = append(entry.Examples, example)
	}

	if len(entry.Examples) > 0 {
		entry.Example = entry.Examples[0].Text
		entry.ExampleNote = entry.Examples[0].Note
	}

	return entry
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

// parseLegacyWordLine 解析两列及以下的历史写法：缺制表符、空格分隔、箭头分隔等。
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

// DisplayFields 返回单词正文之外要展示的内容。
//
// 顺序是先搭配词族（怎么用），再例句（在哪用）。挂多个例句时编号，
// 一个词配三句是有意的：单句只教一种搭配，几种语境摆在一起才看得出词的用法范围。
func (e WordEntry) DisplayFields() []Field {
	fields := make([]Field, 0, 4+len(e.Examples)*2)

	for _, f := range []Field{
		{Label: "音标", Value: e.Phonetic},
		{Label: "翻译", Value: e.Meaning},
		{Label: "搭配", Value: e.Collocation},
		{Label: "词族", Value: e.Family},
	} {
		if f.Value != "" {
			fields = append(fields, f)
		}
	}

	numbered := len(e.Examples) > 1
	for i, example := range e.Examples {
		textLabel, noteLabel := "例句", "译文"
		if numbered {
			textLabel = fmt.Sprintf("例句%d", i+1)
			noteLabel = fmt.Sprintf("译文%d", i+1)
		}
		if example.Text != "" {
			fields = append(fields, Field{Label: textLabel, Value: example.Text})
		}
		if example.Note != "" {
			fields = append(fields, Field{Label: noteLabel, Value: example.Note})
		}
	}

	return fields
}
