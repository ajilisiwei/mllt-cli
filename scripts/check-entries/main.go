// Command check-entries 校验短语与句子资源文件。
//
// 规范格式（用 " ->> " 分段，正文之外的字段可省略）：
//
//	正文 ->> 释义
//	正文 ->> 释义 ->> 例句
//	正文 ->> 释义 ->> 例句 ->> 例句翻译
//
// 报错项（退出码非零）：正文为空、段数超过四段、整行完全重复。
// 提示项：正文相同但释义/例句不同，通常是一词多义，属于有意保留。
//
//	go run ./scripts/check-entries
//	go run ./scripts/check-entries -root ~/.mllt-cli
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ajilisiwei/mllt-cli/internal/practice"
)

const maxFields = 4

type record struct {
	file string
	line int
	raw  string
}

func main() {
	root := flag.String("root", ".", "仓库根目录")
	flag.Parse()

	types := []string{practice.Phrases, practice.Sentences}

	failed := false
	for _, resourceType := range types {
		ok, err := checkType(*root, resourceType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "校验 %s 失败: %v\n", resourceType, err)
			os.Exit(1)
		}
		if !ok {
			failed = true
		}
	}

	if failed {
		os.Exit(1)
	}
}

func checkType(root, resourceType string) (bool, error) {
	pattern := filepath.Join(root, "resources", "*", resourceType, "*", "*.txt")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return false, err
	}
	sort.Strings(files)

	if len(files) == 0 {
		fmt.Printf("%s: 未找到资源文件\n", resourceType)
		return true, nil
	}

	var problems []string
	byText := map[string][]record{}
	byLine := map[string][]record{}
	total := 0

	for _, file := range files {
		records, err := readRecords(file)
		if err != nil {
			return false, err
		}
		total += len(records)

		for _, rec := range records {
			entry := practice.ParseEntryLine(rec.raw)
			label := fmt.Sprintf("%s:%d", shortPath(root, rec.file), rec.line)

			if entry.Text == "" {
				problems = append(problems, label+" 正文为空")
				continue
			}
			if n := len(strings.Split(rec.raw, practice.Separator)); n > maxFields {
				problems = append(problems, fmt.Sprintf("%s 分成了 %d 段，最多 %d 段", label, n, maxFields))
			}

			byText[normalize(entry.Text)] = append(byText[normalize(entry.Text)], rec)
			byLine[strings.TrimSpace(rec.raw)] = append(byLine[strings.TrimSpace(rec.raw)], rec)
		}
	}

	for text, recs := range byLine {
		if len(recs) > 1 {
			problems = append(problems, fmt.Sprintf("整行重复 %d 次: %s", len(recs), truncate(text)))
		}
	}

	var notes []string
	for _, recs := range byText {
		if len(recs) > 1 && !sameLine(recs) {
			notes = append(notes, fmt.Sprintf("%s（%s）", truncate(firstText(recs)), joinFiles(root, recs)))
		}
	}

	sort.Strings(problems)
	sort.Strings(notes)

	fmt.Printf("%s: %d 册 / %d 条\n", resourceType, len(files), total)
	for _, note := range notes {
		fmt.Printf("  提示 正文重复（一词多义）: %s\n", note)
	}
	for _, p := range problems {
		fmt.Printf("  错误 %s\n", p)
	}

	return len(problems) == 0, nil
}

func readRecords(path string) ([]record, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []record
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		if strings.TrimSpace(raw) == "" {
			continue
		}
		records = append(records, record{file: path, line: lineNo, raw: raw})
	}

	return records, scanner.Err()
}

func normalize(text string) string {
	return strings.TrimRight(strings.ToLower(strings.TrimSpace(text)), ".!?")
}

func sameLine(recs []record) bool {
	first := strings.TrimSpace(recs[0].raw)
	for _, rec := range recs[1:] {
		if strings.TrimSpace(rec.raw) != first {
			return false
		}
	}
	return true
}

func firstText(recs []record) string {
	return practice.ParseEntryLine(recs[0].raw).Text
}

func joinFiles(root string, recs []record) string {
	names := make([]string, 0, len(recs))
	for _, rec := range recs {
		names = append(names, filepath.Base(filepath.Dir(rec.file))+"/"+filepath.Base(rec.file))
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func shortPath(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return rel
	}
	return path
}

func truncate(text string) string {
	runes := []rune(text)
	if len(runes) > 50 {
		return string(runes[:50]) + "…"
	}
	return text
}
