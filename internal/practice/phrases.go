package practice

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/ajilisiwei/mllt-cli/internal/config"
)

// PhrasePractice 短语练习
func PhrasePractice(fileName string) error {
	// 读取短语列表
	phrases, err := ReadResourceFile(Phrases, fileName)
	if err != nil {
		return err
	}

	// 检查短语列表是否为空
	if len(phrases) == 0 {
		fmt.Println("短语列表为空，请先添加短语。")
		return nil
	}

	fmt.Printf("开始练习短语列表: %s\n", fileName)
	fmt.Println("输入 'q' 退出练习。")

	// 获取配置
	nextOneOrder := config.AppConfig.NextOneOrder
	showTranslation := config.AppConfig.ShowTranslation

	// 例句也作为练习项：先在条目层定好顺序，再展开，保证例句紧跟它的短语
	if config.AppConfig.PracticeExamples {
		phrases = orderAndExpand(phrases, nextOneOrder)
		nextOneOrder = "sequential"
	}

	// 开始练习
	index := 0
	reader := bufio.NewReader(os.Stdin)

	for {
		// 获取当前短语
		phraseLine := phrases[index]
		entry := ParseEntryLine(phraseLine)
		phrase := entry.Text
		prompt := PromptFor(Phrases, phraseLine)

		// 显示短语
		fmt.Printf("请输入: %s\n", prompt)

		// 读取用户输入
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// 检查是否退出
		if input == "q" {
			fmt.Println("练习结束。")
			break
		}

		// 检查输入是否正确
		if input == phrase {
			fmt.Println("正确！")
			if showTranslation {
				// 译写模式下题面就是译文，不必再打一遍；确认原文更有用
				if prompt != phrase {
					fmt.Printf("原文: %s\n", phrase)
				}
				textLabel, noteLabel := ExampleLabels(Phrases)
				for _, field := range entry.DisplayFields(textLabel, noteLabel) {
					if field.Label == "翻译" && prompt != phrase {
						continue // 译写模式下题面就是译文，不必再打一遍
					}
					fmt.Printf("%s: %s\n", field.Label, field.Value)
				}
			}
			// 获取下一个短语的索引
			index = GetNextIndex(index, len(phrases), nextOneOrder)
		} else {
			fmt.Println("错误，请重新输入。")
			fmt.Printf("正确答案: %s\n", phrase)
		}

		fmt.Println() // 空行分隔
	}

	return nil
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
