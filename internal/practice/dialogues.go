package practice

import "fmt"

// DialoguePractice 对话练习
func DialoguePractice(fileName string) error {
	return runPractice(Dialogues, fileName, "对话", func(line string) (string, string, []Field) {
		entry := ParseEntryLine(line)
		return entry.Text, PromptFor(Dialogues, line), dialogueFields(entry)
	})
}

// ListDialogueFiles 列出对话文件
func ListDialogueFiles() ([]string, error) {
	return GetResourceFiles(Dialogues)
}

// dialogueFields 按轮次展示整段对话。正文是第一轮，后面依次编号。
// 只有一轮时不编号，避免孤零零一个「译文1」。
func dialogueFields(entry Entry) []Field {
	if len(entry.Contrasts) == 0 {
		if entry.Meaning == "" {
			return nil
		}
		return []Field{{Label: "译文", Value: entry.Meaning}}
	}

	fields := make([]Field, 0, 1+len(entry.Contrasts)*2)
	if entry.Meaning != "" {
		fields = append(fields, Field{Label: "译文1", Value: entry.Meaning})
	}
	for i, turn := range entry.Contrasts {
		if turn.Text != "" {
			fields = append(fields, Field{Label: fmt.Sprintf("第%d句", i+2), Value: turn.Text})
		}
		if turn.Note != "" {
			fields = append(fields, Field{Label: fmt.Sprintf("译文%d", i+2), Value: turn.Note})
		}
	}
	return fields
}
