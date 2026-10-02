package practice

// WordPractice 单词练习
func WordPractice(fileName string) error {
	return runPractice(Words, fileName, "单词", func(line string) (string, string, []Field) {
		entry := ParseWordEntry(line)
		return entry.Word, PromptFor(Words, line), entry.DisplayFields()
	})
}

// ListWordFiles 列出单词文件
func ListWordFiles() ([]string, error) {
	return GetResourceFiles(Words)
}
