package practice

import "strings"

// Entry 表示一条短语/句子条目拆分后的各个字段。
type Entry struct {
	Text        string // 需要输入的正文
	Meaning     string // 中文释义
	Example     string // 英文例句
	ExampleNote string // 例句的中文翻译
}

// ParseEntryLine 解析短语或句子行，支持用 " ->> " 分成最多四段：
//
//	正文 ->> 释义
//	正文 ->> 释义 ->> 例句
//	正文 ->> 释义 ->> 例句 ->> 例句翻译
//
// 两段的旧格式继续按「正文 + 释义」解析；不含 " ->> " 时回退到通用分隔符规则。
// 正文始终与 ParseLine 的结果一致，因此打字答案、记忆计划的键都不受影响。
func ParseEntryLine(line string) Entry {
	if !strings.Contains(line, Separator) {
		text, meaning := ParseLine(line)
		return Entry{Text: strings.TrimSpace(text), Meaning: strings.TrimSpace(meaning)}
	}

	parts := strings.Split(line, Separator)

	entry := Entry{Text: strings.TrimSpace(parts[0])}
	if len(parts) > 1 {
		entry.Meaning = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 {
		entry.Example = strings.TrimSpace(parts[2])
	}
	if len(parts) > 3 {
		// 例句翻译里若再出现分隔符，原样拼回，避免内容被截断
		entry.ExampleNote = strings.TrimSpace(strings.Join(parts[3:], Separator))
	}

	return entry
}

// ExpandEntries 把带例句的条目展开成两个练习项：条目本身，以及它的例句。
//
// 例句项本身就是一条合法的「正文 ->> 翻译」条目，所以打字判定、标记收藏、
// 记忆计划都能按普通条目处理，不需要额外的特殊分支。
func ExpandEntries(items []string) []string {
	expanded := make([]string, 0, len(items)*2)

	for _, item := range items {
		expanded = append(expanded, item)

		entry := ParseEntryLine(item)
		if entry.Example == "" {
			continue
		}

		example := entry.Example
		if entry.ExampleNote != "" {
			example += Separator + entry.ExampleNote
		}
		expanded = append(expanded, example)
	}

	return expanded
}
