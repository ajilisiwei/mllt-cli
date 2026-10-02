package practice

import "math/rand"

// PhrasePractice 短语练习
func PhrasePractice(fileName string) error {
	return runPractice(Phrases, fileName, "短语", func(line string) (string, string, []Field) {
		entry := ParseEntryLine(line)
		textLabel, noteLabel := ExampleLabels(Phrases)
		return entry.Text, PromptFor(Phrases, line), entry.DisplayFields(textLabel, noteLabel)
	})
}

// ListPhraseFiles 列出短语文件
func ListPhraseFiles() ([]string, error) {
	return GetResourceFiles(Phrases)
}

// orderAndExpand 先按 order 排好条目顺序，再把每个条目展开成练习块。
// 展开后必须顺序出题，否则例句会和它的短语被拆散。
func orderAndExpand(items []string, order string) []string {
	indexes := make([]int, len(items))
	for i := range indexes {
		indexes[i] = i
	}

	if order != "sequential" && len(indexes) > 1 {
		rand.Shuffle(len(indexes), func(i, j int) {
			indexes[i], indexes[j] = indexes[j], indexes[i]
		})
	}

	blocks := ExpandEntryBlocks(items)

	expanded := make([]string, 0, len(items)*2)
	for _, idx := range indexes {
		expanded = append(expanded, blocks[idx]...)
	}

	return expanded
}
