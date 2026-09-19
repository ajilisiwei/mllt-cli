// Command clean-words 清洗单词释义，把词典 dump 变成能用来背的材料。
//
// 三件事：
//  1. 去掉人名/地名义项组——"able" 的释义里不需要"(伊朗)阿布勒"
//  2. 补上被吞掉的词性分隔——"…陷入n. 放任" 变成 "…陷入 n. 放任"
//  3. 截断义项——每个词性最多 3 个义项，每条释义最多 3 个词性
//
// 词典把最常用的义项排在最前，所以截断天然保留核心义；语义层面的挑选与改写
// 需要人工，不在本工具范围内。
//
//	go run ./scripts/clean-words -check
//	go run ./scripts/clean-words -write
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	wordsGlob     = "resources/*/words/*/*.txt"
	maxSenses     = 3
	maxPOSGroups  = 3
	meaningColumn = 2
)

// posMarker 匹配释义里的词性标记。长的写在前面，避免 "vt." 被拆成 "v" + "t."。
// 标记前必须是非字母数字，但这个条件在匹配后单独校验——写进正则会把分隔符
// 本身吃掉，导致 "adv.adj." 这样连续的两个标记只能切出第一个。
var posMarker = regexp.MustCompile(`(adj|adv|abbr|aux|art|prep|pron|conj|vbl|num|int|vt|vi|ad|n|v|a)\.`)

// isWordChar 判断是不是英文字母或数字。
func isWordChar(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

// nameNoise 匹配人名/地名义项。
var nameNoise = regexp.MustCompile(`人名|地名`)

type posGroup struct {
	pos  string // 词性标记，如 "n."；整条释义没有词性时为空
	body string
}

type change struct {
	line   int
	word   string
	before string
	after  string
}

func main() {
	check := flag.Bool("check", false, "只统计，不修改文件")
	write := flag.Bool("write", false, "就地清洗")
	root := flag.String("root", ".", "仓库根目录")
	flag.Parse()

	if *check == *write {
		fmt.Fprintln(os.Stderr, "请且只请指定 -check 或 -write 其中之一")
		os.Exit(2)
	}

	files, err := filepath.Glob(filepath.Join(*root, wordsGlob))
	if err != nil || len(files) == 0 {
		fmt.Fprintf(os.Stderr, "未找到单词文件: %v\n", err)
		os.Exit(1)
	}
	sort.Strings(files)

	total := 0
	for _, file := range files {
		changes, cleaned, err := inspect(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "处理 %s 失败: %v\n", file, err)
			os.Exit(1)
		}

		total += len(changes)
		report(file, changes)

		if *write && len(changes) > 0 {
			content := strings.Join(cleaned, "\n") + "\n"
			if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "写入 %s 失败: %v\n", file, err)
				os.Exit(1)
			}
			fmt.Printf("  已清洗 %d 行\n", len(changes))
		}
	}

	if total == 0 {
		fmt.Println("所有释义均已清洗。")
	}
}

func inspect(path string) ([]change, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var changes []change
	var cleaned []string

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		if strings.TrimSpace(raw) == "" {
			continue
		}

		cols := strings.Split(raw, "\t")
		if len(cols) <= meaningColumn {
			cleaned = append(cleaned, raw)
			continue
		}

		before := cols[meaningColumn]
		after := cleanMeaning(before)
		if after != before {
			changes = append(changes, change{line: lineNo, word: cols[0], before: before, after: after})
			cols[meaningColumn] = after
		}

		cleaned = append(cleaned, strings.Join(cols, "\t"))
	}

	return changes, cleaned, scanner.Err()
}

// cleanMeaning 清洗单条释义。清洗后为空时保留原文，宁可留噪音也不丢内容。
func cleanMeaning(meaning string) string {
	groups := splitPOSGroups(meaning)
	if len(groups) == 0 {
		return meaning
	}

	kept := make([]posGroup, 0, len(groups))

	// "adv.adj.向后地" 这种连写的标记之间没有内容，标记要并到下一组，
	// 而不是当成空组丢掉——那样会丢失一个词性。
	var pending string

	for _, group := range groups {
		if strings.TrimSpace(group.body) == "" {
			if group.pos != "" {
				pending += group.pos + " "
			}
			continue
		}

		body := filterSenses(group.body)
		if body == "" {
			continue // 整组是人名噪音
		}

		kept = append(kept, posGroup{pos: pending + group.pos, body: body})
		pending = ""

		if len(kept) == maxPOSGroups {
			break
		}
	}

	if len(kept) == 0 {
		return meaning
	}

	parts := make([]string, 0, len(kept))
	for _, group := range kept {
		if group.pos == "" {
			parts = append(parts, group.body)
			continue
		}
		parts = append(parts, group.pos+" "+group.body)
	}

	return strings.Join(parts, " ")
}

// splitPOSGroups 按词性标记把释义切成若干组。
func splitPOSGroups(meaning string) []posGroup {
	var marks [][]int
	for _, loc := range posMarker.FindAllStringIndex(meaning, -1) {
		if loc[0] > 0 && isWordChar(meaning[loc[0]-1]) {
			continue // 是某个英文单词的一部分，不是词性标记
		}
		marks = append(marks, loc)
	}

	if len(marks) == 0 {
		return []posGroup{{body: meaning}}
	}

	var groups []posGroup

	// 第一个词性标记之前的内容没有词性，单独成组
	if lead := strings.TrimSpace(meaning[:marks[0][0]]); lead != "" {
		groups = append(groups, posGroup{body: lead})
	}

	for i, loc := range marks {
		end := len(meaning)
		if i+1 < len(marks) {
			end = marks[i+1][0]
		}
		groups = append(groups, posGroup{
			pos:  meaning[loc[0]:loc[1]],
			body: strings.TrimSpace(meaning[loc[1]:end]),
		})
	}

	return groups
}

// filterSenses 丢掉人名/地名义项并截断义项数量。
// 首个义项就是人名时整组丢弃——那是一整组人名条目，后面跟的都是音译。
func filterSenses(body string) string {
	senses := strings.FieldsFunc(body, func(r rune) bool {
		return r == '；' || r == ';'
	})

	trimmed := make([]string, 0, len(senses))
	for _, sense := range senses {
		if sense = strings.TrimSpace(sense); sense != "" {
			trimmed = append(trimmed, sense)
		}
	}
	if len(trimmed) == 0 {
		return ""
	}
	if nameNoise.MatchString(trimmed[0]) {
		return ""
	}

	kept := make([]string, 0, maxSenses)
	for _, sense := range trimmed {
		if nameNoise.MatchString(sense) {
			continue
		}
		kept = append(kept, sense)
		if len(kept) == maxSenses {
			break
		}
	}

	return strings.Join(kept, "；")
}

func report(path string, changes []change) {
	if len(changes) == 0 {
		return
	}

	fmt.Printf("%s: %d 行需要清洗\n", filepath.Base(path), len(changes))
	for i, c := range changes {
		if i >= 4 {
			break
		}
		fmt.Printf("  %s\n    原: %s\n    新: %s\n", c.word, truncate(c.before), truncate(c.after))
	}
}

func truncate(text string) string {
	runes := []rune(text)
	if len(runes) > 64 {
		return string(runes[:64]) + "…"
	}
	return text
}
