package practice

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ajilisiwei/mllt-cli/internal/config"
)

// describeFunc 把一行原始记录拆成命令行练习需要的三样东西：
// 要输入的正文、屏幕上的题面（译写模式下是中文）、答对后展示的字段。
type describeFunc func(line string) (text string, prompt string, fields []Field)

// runPractice 是四种资源共用的命令行练习循环。
//
// 各类型的差别只在 describe 一个函数里，循环本身——取题顺序、展开、判定、
// 答错给答案——没有任何不同，所以不该为每种资源各抄一遍。
func runPractice(resourceType, fileName, label string, describe describeFunc) error {
	items, err := ReadResourceFile(resourceType, fileName)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Printf("%s列表为空，请先添加内容。\n", label)
		return nil
	}

	fmt.Printf("开始练习%s列表: %s\n", label, fileName)
	fmt.Println("输入 'q' 退出练习。")

	nextOneOrder := config.AppConfig.NextOneOrder
	showTranslation := config.AppConfig.ShowTranslation

	// 先在条目层定好顺序再展开，例句和对话轮次才会紧跟着它的条目出。
	// 对话无视「例句练习」开关：后续轮次不是例句，是内容本身。
	if config.AppConfig.PracticeExamples || resourceType == Dialogues {
		items = orderAndExpand(items, nextOneOrder)
		nextOneOrder = "sequential"
	}

	index := 0
	reader := bufio.NewReader(os.Stdin)

	for {
		line := items[index]
		text, prompt, fields := describe(line)

		fmt.Printf("请输入: %s\n", prompt)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "q" {
			fmt.Println("练习结束。")
			break
		}

		if input != text {
			fmt.Println("错误，请重新输入。")
			fmt.Printf("正确答案: %s\n", text)
			fmt.Println()
			continue
		}

		fmt.Println("正确！")
		if showTranslation {
			translated := prompt != text
			// 译写模式下题面已经是译文了，再打一遍没意义；确认原文更有用
			if translated {
				fmt.Printf("原文: %s\n", text)
			}
			for _, field := range fields {
				if translated && isPromptField(field.Label) {
					continue
				}
				fmt.Printf("%s: %s\n", field.Label, field.Value)
			}
		}

		index = GetNextIndex(index, len(items), nextOneOrder)
		fmt.Println()
	}

	return nil
}

// isPromptField 判断某个字段是否就是译写模式下的题面内容。
func isPromptField(label string) bool {
	switch label {
	case "翻译", "译文", "译文1", "音标":
		return true
	}
	return false
}
