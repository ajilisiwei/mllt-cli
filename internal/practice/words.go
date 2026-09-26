package practice

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ajilisiwei/mllt-cli/internal/config"
)

// WordPractice 单词练习
func WordPractice(fileName string) error {
	// 读取单词列表
	words, err := ReadResourceFile(Words, fileName)
	if err != nil {
		return err
	}

	// 检查单词列表是否为空
	if len(words) == 0 {
		fmt.Println("单词列表为空，请先添加单词。")
		return nil
	}

	fmt.Printf("开始练习单词列表: %s\n", fileName)
	fmt.Println("输入 'q' 退出练习。")

	// 获取配置
	nextOneOrder := config.AppConfig.NextOneOrder
	showTranslation := config.AppConfig.ShowTranslation

	// 例句也作为练习项：先打单词，再打用到它的整句
	if config.AppConfig.PracticeExamples {
		words = orderAndExpand(words, nextOneOrder)
		nextOneOrder = "sequential"
	}

	// 开始练习
	index := 0
	reader := bufio.NewReader(os.Stdin)

	for {
		// 获取当前单词
		wordLine := words[index]
		entry := ParseWordEntry(wordLine)
		word := entry.Word
		prompt := PromptFor(Words, wordLine)

		// 显示题面。译写模式下音标已经在题面里，答对后再一并展示
		fmt.Printf("请输入: %s\n", prompt)
		if showTranslation && entry.Phonetic != "" && prompt == word {
			fmt.Printf("音标: %s\n", entry.Phonetic)
		}
		answered := prompt == word

		// 读取用户输入
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// 检查是否退出
		if input == "q" {
			fmt.Println("练习结束。")
			break
		}

		// 检查输入是否正确
		if input == word {
			fmt.Println("正确！")
			if showTranslation {
				// 译写模式下题面就是音标加释义，不必再打一遍；确认拼写更有用
				if !answered {
					fmt.Printf("原文: %s\n", word)
				}
				for _, field := range entry.DisplayFields() {
					if answered && field.Label == "音标" {
						continue // 题面已经打过了
					}
					fmt.Printf("%s: %s\n", field.Label, field.Value)
				}
			}
			// 获取下一个单词的索引
			index = GetNextIndex(index, len(words), nextOneOrder)
		} else {
			fmt.Println("错误，请重新输入。")
			fmt.Printf("正确答案: %s\n", word)
		}

		fmt.Println() // 空行分隔
	}

	return nil
}

// ListWordFiles 列出单词文件
func ListWordFiles() ([]string, error) {
	return GetResourceFiles(Words)
}
