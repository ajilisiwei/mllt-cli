package practice

import (
	"fmt"
	"strings"

	"github.com/ajilisiwei/mllt-cli/internal/config"
)

// Contrast 表示一条对比句：英文加它的中文注释。
type Contrast struct {
	Text string
	Note string
}

// Entry 表示一条短语/句子条目拆分后的各个字段。
type Entry struct {
	Text        string // 需要输入的正文
	Meaning     string // 中文释义
	Example     string // 第一条对比句，等同于 Contrasts[0].Text（保留给既有调用方）
	ExampleNote string // 第一条对比句的注释
	// Contrasts 是正文之后的全部对比句。短语只有一条例句，语法册可以挂一整套时态。
	Contrasts []Contrast
}

// ParseEntryLine 解析短语或句子行。各段按「文本, 注释」两两成对：
//
//	正文 ->> 释义
//	正文 ->> 释义 ->> 例句
//	正文 ->> 释义 ->> 例句 ->> 例句翻译
//	正文 ->> 注释 ->> 对比句1 ->> 注释1 ->> 对比句2 ->> 注释2 ->> …
//
// 成对之后就不再有段数上限：短语挂一条例句，语法条目可以挂一整套时态，
// 它们会被展开成连续的练习项，对比才不会被拆散。
//
// 两段的旧格式继续按「正文 + 释义」解析；不含 " ->> " 时回退到通用分隔符规则。
// 正文始终与 ParseLine 的结果一致，因此打字答案、记忆计划的键都不受影响。
func ParseEntryLine(line string) Entry {
	if !strings.Contains(line, Separator) {
		text, meaning := ParseLine(line)
		return Entry{Text: strings.TrimSpace(text), Meaning: strings.TrimSpace(meaning)}
	}

	parts := strings.Split(line, Separator)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	entry := Entry{Text: parts[0]}
	if len(parts) > 1 {
		entry.Meaning = parts[1]
	}

	// 从第三段起每两段构成一条对比句；最后一句缺注释时注释留空
	for i := 2; i < len(parts); i += 2 {
		contrast := Contrast{Text: parts[i]}
		if i+1 < len(parts) {
			contrast.Note = parts[i+1]
		}
		if contrast.Text == "" {
			continue
		}
		entry.Contrasts = append(entry.Contrasts, contrast)
	}

	if len(entry.Contrasts) > 0 {
		entry.Example = entry.Contrasts[0].Text
		entry.ExampleNote = entry.Contrasts[0].Note
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

		block = append(block, contrastItemsOf(item)...)

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

// contrastItemsOf 返回一条记录挂着的全部对比句练习项，没有则返回空切片。
//
// 练习项沿用原记录的分隔风格：短语和句子用 " ->> "，单词表用制表符分列，
// 这样它在各自的资源类型里都是一条合法记录。
func contrastItemsOf(line string) []string {
	if strings.Contains(line, Separator) {
		entry := ParseEntryLine(line)

		items := make([]string, 0, len(entry.Contrasts))
		for _, contrast := range entry.Contrasts {
			if contrast.Note == "" {
				items = append(items, contrast.Text)
				continue
			}
			items = append(items, contrast.Text+Separator+contrast.Note)
		}
		return items
	}

	word := ParseWordEntry(line)
	if word.Example == "" {
		return nil
	}
	// 单词表是制表符分列的，例句项也保持同样的列结构，音标列留空
	return []string{word.Example + "\t\t" + word.ExampleNote}
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

// Field 是一条记录在正文之外要展示的字段。
type Field struct {
	Label string
	Value string
}

// DisplayFields 返回正文之外要展示的内容。
//
// 只挂一条对比句时按「例句 / 译文」显示，这是短语册的样子；挂一整套时
// （比如同一场景铺开各种时态）改成带编号的「对比N / 注释N」，
// 否则一屏全是同名标签，根本没法对照着看。
func (e Entry) DisplayFields() []Field {
	fields := make([]Field, 0, 1+len(e.Contrasts)*2)

	if e.Meaning != "" {
		fields = append(fields, Field{Label: "翻译", Value: e.Meaning})
	}

	numbered := len(e.Contrasts) > 1
	for i, contrast := range e.Contrasts {
		textLabel, noteLabel := "例句", "译文"
		if numbered {
			textLabel = fmt.Sprintf("对比%d", i+1)
			noteLabel = fmt.Sprintf("注释%d", i+1)
		}
		if contrast.Text != "" {
			fields = append(fields, Field{Label: textLabel, Value: contrast.Text})
		}
		if contrast.Note != "" {
			fields = append(fields, Field{Label: noteLabel, Value: contrast.Note})
		}
	}

	return fields
}
