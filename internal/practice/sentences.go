package practice

// SentencePractice 句子练习
func SentencePractice(fileName string) error {
	return runPractice(Sentences, fileName, "句子", func(line string) (string, string, []Field) {
		entry := ParseEntryLine(line)
		textLabel, noteLabel := ExampleLabels(Sentences)
		return entry.Text, PromptFor(Sentences, line), entry.DisplayFields(textLabel, noteLabel)
	})
}

// ListSentenceFiles 列出句子文件
func ListSentenceFiles() ([]string, error) {
	return GetResourceFiles(Sentences)
}
