package practice

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ajilisiwei/mllt-cli/internal/config"
)

// SentencePractice 句子练习
func SentencePractice(fileName string) error {
	// 读取句子列表
	sentences, err := ReadResourceFile(Sentences, fileName)
	if err != nil {
		return err
	}

	// 检查句子列表是否为空
	if len(sentences) == 0 {
		fmt.Println("句子列表为空，请先添加句子。")
		return nil
	}

	fmt.Printf("开始练习句子列表: %s\n", fileName)
	fmt.Println("输入 'q' 退出练习。")

	// 获取配置
	nextOneOrder := config.AppConfig.NextOneOrder
	showTranslation := config.AppConfig.ShowTranslation

	// 例句／对比句也作为练习项：先在条目层定好顺序，再展开，保证配对的两句紧挨着出
	if config.AppConfig.PracticeExamples {
		sentences = orderAndExpand(sentences, nextOneOrder)
		nextOneOrder = "sequential"
	}

	// 开始练习
	index := 0
	reader := bufio.NewReader(os.Stdin)

	for {
		// 获取当前句子
		sentenceLine := sentences[index]
		entry := ParseEntryLine(sentenceLine)
		sentence := entry.Text
		prompt := PromptFor(Sentences, sentenceLine)

		// 显示句子
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
		if input == sentence {
			fmt.Println("正确！")
			if showTranslation {
				// 译写模式下题面就是译文，不必再打一遍；确认原文更有用
				if prompt != sentence {
					fmt.Printf("原文: %s\n", sentence)
				}
				textLabel, noteLabel := ExampleLabels(Sentences)
				for _, field := range entry.DisplayFields(textLabel, noteLabel) {
					if field.Label == "翻译" && prompt != sentence {
						continue // 译写模式下题面就是译文，不必再打一遍
					}
					fmt.Printf("%s: %s\n", field.Label, field.Value)
				}
			}
			// 获取下一个句子的索引
			index = GetNextIndex(index, len(sentences), nextOneOrder)
		} else {
			fmt.Println("错误，请重新输入。")
			fmt.Printf("正确答案: %s\n", sentence)
		}

		fmt.Println() // 空行分隔
	}

	return nil
}

// ListSentenceFiles 列出句子文件
func ListSentenceFiles() ([]string, error) {
	return GetResourceFiles(Sentences)
}
