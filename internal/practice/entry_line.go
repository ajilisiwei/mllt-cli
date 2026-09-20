package practice

import (
	"strings"

	"github.com/ajilisiwei/mllt-cli/internal/config"
)

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

// ExpandEntryBlocks 把每个条目展开成一个练习块：条目本身，以及（如果有）它的例句。
//
// 同一个块里的项必须连续出题——先打短语、紧接着打用到它的那句话，例句才有
// 上下文。所以打乱顺序要在块这一层做，不能打乱展开之后的练习项。
//
// 例句项本身就是一条合法的「正文 ->> 翻译」条目，所以打字判定、标记收藏都能
// 按普通条目处理，不需要额外的特殊分支。
func ExpandEntryBlocks(items []string) [][]string {
	blocks := make([][]string, 0, len(items))

	for _, item := range items {
		block := []string{item}

		if example := exampleItemOf(item); example != "" {
			block = append(block, example)
		}

		blocks = append(blocks, block)
	}

	return blocks
}

// ExpandEntries 按条目顺序展平所有练习块。
func ExpandEntries(items []string) []string {
	expanded := make([]string, 0, len(items)*2)
	for _, block := range ExpandEntryBlocks(items) {
		expanded = append(expanded, block...)
	}
	return expanded
}

// exampleItemOf 返回一条记录对应的例句练习项，没有例句时返回空串。
//
// 例句项沿用原记录的分隔风格：短语和句子用 " ->> "，单词表用制表符分列，
// 这样它在各自的资源类型里都是一条合法记录。
func exampleItemOf(line string) string {
	if strings.Contains(line, Separator) {
		entry := ParseEntryLine(line)
		if entry.Example == "" {
			return ""
		}
		if entry.ExampleNote == "" {
			return entry.Example
		}
		return entry.Example + Separator + entry.ExampleNote
	}

	word := ParseWordEntry(line)
	if word.Example == "" {
		return ""
	}
	// 单词表是制表符分列的，例句项也保持同样的列结构，音标列留空
	return word.Example + "\t\t" + word.ExampleNote
}

// 练习方向
const (
	DirectionCopy      = "copy"      // 看英文抄写
	DirectionTranslate = "translate" // 看中文译写
)

// PromptFor 返回练习时要显示的题面。
//
// 译写模式下给中文提示（单词给音标加释义，其余给译文），条目没有译文时退回英文原文——
// 给一个空白提示比给原文更糟。抄写模式下始终返回英文原文。
func PromptFor(resourceType, line string) string {
	text := WordPrimaryText(line)
	if resourceType != Words {
		if primary, _ := ParseLine(line); primary != "" {
			text = primary
		}
	}

	if !strings.EqualFold(config.AppConfig.PracticeDirection, DirectionTranslate) {
		return text
	}

	if resourceType == Words {
		entry := ParseWordEntry(line)

		hints := make([]string, 0, 2)
		if entry.Phonetic != "" {
			hints = append(hints, entry.Phonetic)
		}
		if entry.Meaning != "" {
			hints = append(hints, entry.Meaning)
		}
		if len(hints) > 0 {
			return strings.Join(hints, "  ")
		}
		return text
	}

	if meaning := ParseEntryLine(line).Meaning; meaning != "" {
		return meaning
	}
	return text
}
