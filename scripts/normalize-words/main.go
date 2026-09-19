// Command normalize-words 校验并规范化单词资源文件。
//
// 规范格式为三列制表符分隔：单词 \t /音标/ \t 释义（音标、释义均可为空）。
// 历史数据里存在 "fool/fuːl/\t释义"（缺少第一个制表符）等写法，会让打字练习
// 把音标也当成答案，本工具负责把它们统一回规范格式。
//
//	go run ./scripts/normalize-words -check   # 只校验，发现问题返回非零退出码
//	go run ./scripts/normalize-words -write   # 就地规范化
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

	"github.com/ajilisiwei/mllt-cli/internal/practice"
)

const wordsGlob = "resources/*/words/*/*.txt"

type issue struct {
	file string
	line int
	raw  string
	want string
}

type suspect struct {
	line   int
	raw    string
	want   string
	reason string
}

// legacySeparator 匹配早期空格分隔格式里的 " / "。规范化后释义中若仍出现它，
// 说明这一行的音标没有被正确拆出（常见于词条本身含空格的情况）。
var legacySeparator = regexp.MustCompile(`\s/\s`)

// cjkInWord 匹配单词列里的中日韩文字与全角标点。英文词表的单词列不该出现这些，
// 一旦出现基本都是上一行的释义被截断成了独立的一行。
// cyrillic 匹配西里尔字母。英文列里出现它基本都是输入法串行打出来的，
// 肉眼几乎分不出来（objectивity vs objectivity），必须靠工具挡住。
// 音标列不查——IPA 的 θ 属于希腊字母块，是合法字符。
var cyrillic = regexp.MustCompile(`[\p{Cyrillic}]`)

var cjkInWord = regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}\x{3000}-\x{303F}\x{FF00}-\x{FFEF}]`)

func main() {
	check := flag.Bool("check", false, "只校验，不修改文件")
	write := flag.Bool("write", false, "就地规范化文件")
	root := flag.String("root", ".", "仓库根目录")
	flag.Parse()

	if *check == *write {
		fmt.Fprintln(os.Stderr, "请且只请指定 -check 或 -write 其中之一")
		os.Exit(2)
	}

	files, err := filepath.Glob(filepath.Join(*root, wordsGlob))
	if err != nil {
		fmt.Fprintf(os.Stderr, "扫描单词文件失败: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "未找到任何单词文件: %s\n", wordsGlob)
		os.Exit(1)
	}
	sort.Strings(files)

	total := 0
	for _, file := range files {
		issues, suspects, normalized, err := inspect(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "处理 %s 失败: %v\n", file, err)
			os.Exit(1)
		}

		total += len(issues)
		report(file, issues)

		if len(suspects) > 0 {
			reportSuspects(file, suspects)
			fmt.Fprintf(os.Stderr, "%s 存在 %d 行无法可靠解析，已中止，请先人工修正。\n", file, len(suspects))
			os.Exit(1)
		}

		if *write && len(issues) > 0 {
			if err := os.WriteFile(file, []byte(strings.Join(normalized, "\n")+"\n"), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "写入 %s 失败: %v\n", file, err)
				os.Exit(1)
			}
			fmt.Printf("  已规范化 %d 行\n", len(issues))
		}
	}

	if total == 0 {
		fmt.Println("所有单词文件均符合规范格式。")
		return
	}
	if *check {
		fmt.Fprintf(os.Stderr, "\n共 %d 行不符合规范，请运行 `go run ./scripts/normalize-words -write` 修复。\n", total)
		os.Exit(1)
	}
}

// inspect 读取单词文件，返回不规范的行以及规范化后的完整内容。
func inspect(path string) ([]issue, []suspect, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, err
	}
	defer file.Close()

	var issues []issue
	var suspects []suspect
	var normalized []string

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		if strings.TrimSpace(raw) == "" {
			continue
		}

		entry := practice.ParseWordEntry(raw)
		want := canonical(entry)

		normalized = append(normalized, want)
		if want != raw {
			issues = append(issues, issue{file: path, line: lineNo, raw: raw, want: want})
		}
		if reason := suspicious(entry); reason != "" {
			suspects = append(suspects, suspect{line: lineNo, raw: raw, want: want, reason: reason})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, nil, err
	}

	return issues, suspects, normalized, nil
}

// suspicious 判断规范化结果是否可信，返回非空字符串表示这一行需要人工确认。
func suspicious(entry practice.WordEntry) string {
	word, phonetic, meaning := entry.Word, entry.Phonetic, entry.Meaning
	switch {
	case word == "":
		return "单词为空"
	case strings.Contains(word, "\t"):
		return "单词列仍含制表符"
	case cjkInWord.MatchString(word):
		return "单词列出现中文或全角标点，疑似上一行的断行残片"
	case cyrillic.MatchString(word + entry.Example + entry.Collocation + entry.Family):
		return "英文列里混入了西里尔字母"
	case strings.Contains(meaning, "\t"):
		return "释义列仍含制表符，疑似音标缺少右斜杠"
	case strings.HasSuffix(word, "/"):
		// "CI/CD"、"and/or" 这类词条本身含斜杠是合法的，只有以斜杠结尾才说明音标没拆干净
		return "单词列仍以斜杠结尾，音标未拆出"
	case phonetic == "" && legacySeparator.MatchString(meaning):
		return "释义中仍残留 \" / \" 分隔符，音标未拆出"
	case phonetic == "" && meaning == "":
		return "音标与释义均为空，疑似上一行的断行残片"
	}
	return ""
}

// canonical 把解析结果拼回规范的制表符分列形式：
//
//	单词 \t 音标 \t 释义 \t 例句 \t 例句翻译 \t 搭配 \t 词族
//
// 尾部的空列整列省略，中间的空列必须保留，否则后面的列会整体前移串位。
func canonical(entry practice.WordEntry) string {
	cols := []string{
		entry.Word,
		entry.Phonetic,
		entry.Meaning,
		entry.Example,
		entry.ExampleNote,
		entry.Collocation,
		entry.Family,
	}

	for len(cols) > 1 && cols[len(cols)-1] == "" {
		cols = cols[:len(cols)-1]
	}

	return strings.Join(cols, "\t")
}

func reportSuspects(path string, suspects []suspect) {
	fmt.Printf("%s: %d 行无法可靠解析\n", path, len(suspects))
	for i, sp := range suspects {
		if i >= 10 {
			fmt.Printf("  ... 其余 %d 行省略\n", len(suspects)-10)
			break
		}
		fmt.Printf("  第 %d 行（%s）\n    现有: %s\n    结果: %s\n", sp.line, sp.reason, visualize(sp.raw), visualize(sp.want))
	}
}

func report(path string, issues []issue) {
	if len(issues) == 0 {
		return
	}

	fmt.Printf("%s: %d 行需要规范化\n", path, len(issues))
	for i, it := range issues {
		if i >= 3 {
			fmt.Printf("  ... 其余 %d 行省略\n", len(issues)-3)
			break
		}
		fmt.Printf("  第 %d 行\n    现有: %s\n    规范: %s\n", it.line, visualize(it.raw), visualize(it.want))
	}
}

// visualize 把制表符显示成 \t，便于肉眼比对，并截断过长的释义。
func visualize(line string) string {
	line = strings.ReplaceAll(line, "\t", "\\t")
	runes := []rune(line)
	if len(runes) > 60 {
		return string(runes[:60]) + "…"
	}
	return line
}
