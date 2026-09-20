package main

import (
	"fmt"
	"os"
	"strconv"

	mlltcli "github.com/ajilisiwei/mllt-cli"
	"github.com/ajilisiwei/mllt-cli/internal/config"
	"github.com/ajilisiwei/mllt-cli/internal/lang"
	"github.com/ajilisiwei/mllt-cli/internal/manage"
	"github.com/ajilisiwei/mllt-cli/internal/practice"
	"github.com/ajilisiwei/mllt-cli/internal/ui"
	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mllt-cli",
	Short: "一个多语言打字学习终端工具",
	Long: `这是一个支持多种语言的打字学习终端工具，
用户可以在终端中进行不同语言的单词、短语、句子、文章等的打字练习。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 显示主菜单
		fmt.Println("欢迎使用语言学习终端！")
		// 启动UI界面
		p := tea.NewProgram(ui.NewMainMenu(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("启动UI界面失败: %v\n", err)
			os.Exit(1)
		}
	},
}

// langCmd 表示lang子命令
var langCmd = &cobra.Command{
	Use:   "lang",
	Short: "语言管理模块",
	Long:  `语言管理模块，用于列出支持的语言和切换当前练习的语言。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("语言管理模块")
		// 这里将来会调用lang模块的功能
	},
}

// langLsCmd 表示lang ls子命令
var langLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "列出支持的语言",
	Long:  `列出支持的语言列表，当前选中的语言前面会有"·"标记。`,
	Run: func(cmd *cobra.Command, args []string) {
		lang.PrintLanguages()
	},
}

// langStCmd 表示lang st子命令
var langStCmd = &cobra.Command{
	Use:   "st [language]",
	Short: "切换练习语言",
	Long:  `切换当前练习的语言，例如：mllt-cli lang st japanese。`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		language := args[0]
		if err := lang.SwitchLanguage(language); err != nil {
			fmt.Println(err)
		}
	},
	ValidArgs: lang.ListLanguages(),
}

// practiceCmd 表示practice子命令
var practiceCmd = &cobra.Command{
	Use:   "practice",
	Short: "练习模块",
	Long:  `练习模块，用于进行单词、短语、句子、文章等的打字练习。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("练习模块")
		// 这里将来会调用practice模块的功能
	},
}

// practiceWordsCmd 表示practice words子命令
var practiceWordsCmd = &cobra.Command{
	Use:   "words [file]",
	Short: "单词练习",
	Long:  `单词练习功能，从指定的单词列表文件中读取单词进行练习。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有指定文件，则列出可用的单词列表文件
		if len(args) == 0 {
			files, err := practice.ListWordFiles()
			if err != nil {
				fmt.Println("获取单词列表文件失败:", err)
				return
			}

			if len(files) == 0 {
				fmt.Println("没有可用的单词列表文件，请先创建或导入单词列表。")
				return
			}

			fmt.Println("可用的单词列表文件:")
			for i, file := range files {
				fmt.Printf("%d. %s\n", i+1, file)
			}
			return
		}

		// 指定了文件，进行单词练习
		fileName := args[0]
		if err := practice.WordPractice(fileName); err != nil {
			fmt.Println("单词练习失败:", err)
		}
	},
}

// practicePhrasesCmd 表示practice phrases子命令
var practicePhrasesCmd = &cobra.Command{
	Use:   "phrases [file]",
	Short: "短语练习",
	Long:  `短语练习功能，从指定的短语列表文件中读取短语进行练习。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有指定文件，则列出可用的短语列表文件
		if len(args) == 0 {
			files, err := practice.ListPhraseFiles()
			if err != nil {
				fmt.Println("获取短语列表文件失败:", err)
				return
			}

			if len(files) == 0 {
				fmt.Println("没有可用的短语列表文件，请先创建或导入短语列表。")
				return
			}

			fmt.Println("可用的短语列表文件:")
			for i, file := range files {
				fmt.Printf("%d. %s\n", i+1, file)
			}
			return
		}

		// 指定了文件，进行短语练习
		fileName := args[0]
		if err := practice.PhrasePractice(fileName); err != nil {
			fmt.Println("短语练习失败:", err)
		}
	},
}

// practiceSentencesCmd 表示practice sentences子命令
var practiceSentencesCmd = &cobra.Command{
	Use:   "sentences [file]",
	Short: "句子练习",
	Long:  `句子练习功能，从指定的句子列表文件中读取句子进行练习。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有指定文件，则列出可用的句子列表文件
		if len(args) == 0 {
			files, err := practice.ListSentenceFiles()
			if err != nil {
				fmt.Println("获取句子列表文件失败:", err)
				return
			}

			if len(files) == 0 {
				fmt.Println("没有可用的句子列表文件，请先创建或导入句子列表。")
				return
			}

			fmt.Println("可用的句子列表文件:")
			for i, file := range files {
				fmt.Printf("%d. %s\n", i+1, file)
			}
			return
		}

		// 指定了文件，进行句子练习
		fileName := args[0]
		if err := practice.SentencePractice(fileName); err != nil {
			fmt.Println("句子练习失败:", err)
		}
	},
}

// practiceArticlesCmd 表示practice articles子命令
var practiceArticlesCmd = &cobra.Command{
	Use:   "articles [file]",
	Short: "文章练习",
	Long:  `文章练习功能，从指定的文章文件中读取文章进行练习。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有指定文件，则列出可用的文章文件
		if len(args) == 0 {
			files, err := practice.ListArticleFiles()
			if err != nil {
				fmt.Println("获取文章文件失败:", err)
				return
			}

			if len(files) == 0 {
				fmt.Println("没有可用的文章文件，请先创建或导入文章。")
				return
			}

			fmt.Println("可用的文章文件:")
			for i, file := range files {
				fmt.Printf("%d. %s\n", i+1, file)
			}
			return
		}

		// 指定了文件，进行文章练习
		fileName := args[0]
		if err := practice.ArticlePractice(fileName); err != nil {
			fmt.Println("文章练习失败:", err)
		}
	},
}

// manageCmd 表示manage子命令
var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "资源管理模块",
	Long:  `资源管理模块，用于管理练习资源，包括删除和导入资源。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("资源管理模块")
		// 这里将来会调用manage模块的功能
	},
}

// manageDeleteCmd 表示manage delete子命令
var manageDeleteCmd = &cobra.Command{
	Use:   "delete [resourceType] [file]",
	Short: "删除资源",
	Long:  `删除资源，例如：mllt-cli manage delete words daily.txt。`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		// 获取资源类型
		resourceType := args[0]

		// 验证资源类型
		if !manage.ValidateResourceType(resourceType) {
			fmt.Printf("无效的资源类型: %s\n", resourceType)
			fmt.Println("有效的资源类型: words, phrases, sentences, articles")
			return
		}

		// 如果没有指定文件，则列出可用的资源文件
		if len(args) == 1 {
			files, err := manage.ListResourceFiles(resourceType)
			if err != nil {
				fmt.Printf("获取%s文件失败: %s\n", resourceType, err)
				return
			}

			if len(files) == 0 {
				fmt.Printf("没有可用的%s文件。\n", resourceType)
				return
			}

			fmt.Printf("可用的%s文件:\n", resourceType)
			for i, file := range files {
				fmt.Printf("%d. %s\n", i+1, file)
			}
			return
		}

		// 指定了文件，删除资源
		fileName := args[1]
		if err := manage.DeleteResource(resourceType, fileName); err != nil {
			fmt.Printf("删除%s文件失败: %s\n", resourceType, err)
		}
	},
}

// manageImportCmd 表示manage import子命令
var manageImportCmd = &cobra.Command{
	Use:   "import [resourceType] [file]",
	Short: "导入资源",
	Long:  `导入资源，例如：mllt-cli manage import words /path/to/words.txt。`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// 获取资源类型和文件路径
		resourceType := args[0]
		filePath := args[1]

		// 验证资源类型
		if !manage.ValidateResourceType(resourceType) {
			fmt.Printf("无效的资源类型: %s\n", resourceType)
			fmt.Println("有效的资源类型: words, phrases, sentences, articles")
			return
		}

		// 导入资源，默认导入到“默认”文件夹
		if err := manage.ImportResource(resourceType, practice.DefaultFolderDir, filePath); err != nil {
			fmt.Printf("导入%s文件失败: %s\n", resourceType, err)
		}
	},
}

// settingCmd 表示setting子命令
var settingCmd = &cobra.Command{
	Use:   "setting",
	Short: "设置模块",
	Long:  `设置模块，用于配置匹配模式、练习顺序、键盘声音和翻译显示等设置。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("设置模块")
		fmt.Println("使用 'mllt-cli setting --help' 查看可用的设置选项")
	},
}

// settingMatchModeCmd 表示setting match-mode子命令
var settingMatchModeCmd = &cobra.Command{
	Use:   "match-mode [mode]",
	Short: "设置匹配模式",
	Long:  `设置正确性匹配模式，可选值：exact_match（完全匹配）、word_match（单词匹配）。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// 显示当前匹配模式
			fmt.Printf("当前匹配模式: %s\n", config.AppConfig.CorrectnessMatchMode)
			fmt.Println("可用的匹配模式:")
			fmt.Println("  exact_match - 完全匹配")
			fmt.Println("  word_match  - 单词匹配")
			return
		}

		mode := args[0]
		if mode != "exact_match" && mode != "word_match" {
			fmt.Printf("无效的匹配模式: %s\n", mode)
			fmt.Println("可用的匹配模式: exact_match, word_match")
			return
		}

		config.AppConfig.CorrectnessMatchMode = mode
		if err := config.SaveConfig(); err != nil {
			fmt.Printf("保存配置失败: %s\n", err)
			return
		}
		fmt.Printf("匹配模式已设置为: %s\n", mode)
	},
	ValidArgs: []string{"exact_match", "word_match"},
}

// settingOrderCmd 表示setting order子命令
var settingOrderCmd = &cobra.Command{
	Use:   "order [order]",
	Short: "设置练习顺序",
	Long:  `设置练习资源的出现顺序，可选值：random（随机）、sequential（顺序）、ebbinghaus（艾宾浩斯间隔复习）。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// 显示当前顺序设置
			fmt.Printf("当前练习顺序: %s\n", config.AppConfig.NextOneOrder)
			fmt.Println("可用的顺序设置:")
			fmt.Println("  random     - 随机")
			fmt.Println("  sequential - 顺序")
			fmt.Println("  ebbinghaus - 艾宾浩斯间隔复习（答错的会更早重现）")
			return
		}

		order := args[0]
		if order != "random" && order != "sequential" && order != "ebbinghaus" {
			fmt.Printf("无效的顺序设置: %s\n", order)
			fmt.Println("可用的顺序设置: random, sequential, ebbinghaus")
			return
		}

		config.AppConfig.NextOneOrder = order
		if err := config.SaveConfig(); err != nil {
			fmt.Printf("保存配置失败: %s\n", err)
			return
		}
		fmt.Printf("练习顺序已设置为: %s\n", order)
	},
	ValidArgs: []string{"random", "sequential", "ebbinghaus"},
}

// settingKeyboardSoundCmd 表示setting keyboard-sound子命令
var settingKeyboardSoundCmd = &cobra.Command{
	Use:   "keyboard-sound [enable|disable]",
	Short: "设置键盘声音",
	Long:  `设置是否启用键盘按键声音，可选值：enable（启用）、disable（禁用）。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// 显示当前键盘声音设置
			status := "禁用"
			if config.AppConfig.InputKeyboardSound {
				status = "启用"
			}
			fmt.Printf("当前键盘声音设置: %s\n", status)
			fmt.Println("可用的设置:")
			fmt.Println("  enable  - 启用")
			fmt.Println("  disable - 禁用")
			return
		}

		setting := args[0]
		var enable bool
		switch setting {
		case "enable":
			enable = true
		case "disable":
			enable = false
		default:
			fmt.Printf("无效的设置: %s\n", setting)
			fmt.Println("可用的设置: enable, disable")
			return
		}

		config.AppConfig.InputKeyboardSound = enable
		if err := config.SaveConfig(); err != nil {
			fmt.Printf("保存配置失败: %s\n", err)
			return
		}
		status := "禁用"
		if enable {
			status = "启用"
		}
		fmt.Printf("键盘声音已设置为: %s\n", status)
	},
	ValidArgs: []string{"enable", "disable"},
}

// settingTranslationCmd 表示setting translation子命令
var settingTranslationCmd = &cobra.Command{
	Use:   "translation [show|hide]",
	Short: "设置翻译显示",
	Long:  `设置是否显示翻译内容，可选值：show（显示）、hide（隐藏）。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// 显示当前翻译显示设置
			status := "隐藏"
			if config.AppConfig.ShowTranslation {
				status = "显示"
			}
			fmt.Printf("当前翻译显示设置: %s\n", status)
			fmt.Println("可用的设置:")
			fmt.Println("  show - 显示")
			fmt.Println("  hide - 隐藏")
			return
		}

		setting := args[0]
		var show bool
		switch setting {
		case "show":
			show = true
		case "hide":
			show = false
		default:
			fmt.Printf("无效的设置: %s\n", setting)
			fmt.Println("可用的设置: show, hide")
			return
		}

		config.AppConfig.ShowTranslation = show
		if err := config.SaveConfig(); err != nil {
			fmt.Printf("保存配置失败: %s\n", err)
			return
		}
		status := "隐藏"
		if show {
			status = "显示"
		}
		fmt.Printf("翻译显示已设置为: %s\n", status)
	},
	ValidArgs: []string{"show", "hide"},
}

// settingExamplesCmd 表示setting examples子命令
var settingExamplesCmd = &cobra.Command{
	Use:   "examples [on|off]",
	Short: "设置例句是否参与练习",
	Long:  `设置短语的例句是否也作为练习内容，可选值：on（也练例句）、off（只练短语）。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			status := "只练短语"
			if config.AppConfig.PracticeExamples {
				status = "也练例句"
			}
			fmt.Printf("当前例句练习设置: %s\n", status)
			fmt.Println("可用的设置:")
			fmt.Println("  on  - 先打短语，再打用到它的整句")
			fmt.Println("  off - 例句仅作展示，不进入练习")
			return
		}

		var enable bool
		switch args[0] {
		case "on":
			enable = true
		case "off":
			enable = false
		default:
			fmt.Printf("无效的设置: %s\n", args[0])
			fmt.Println("可用的设置: on, off")
			return
		}

		config.AppConfig.PracticeExamples = enable
		if err := config.SaveConfig(); err != nil {
			fmt.Printf("保存配置失败: %s\n", err)
			return
		}
		status := "只练短语"
		if enable {
			status = "也练例句"
		}
		fmt.Printf("例句练习已设置为: %s\n", status)
	},
	ValidArgs: []string{"on", "off"},
}

// settingDailyCmd 表示setting daily子命令
var settingDailyCmd = &cobra.Command{
	Use:   "daily [new-limit] [review-limit]",
	Short: "设置艾宾浩斯模式的每日练习量",
	Long: `设置艾宾浩斯模式下每天引入的新条目数量和复习上限，0 表示不限。
只影响"下一个资源出现顺序"为 ebbinghaus 时的取题数量。`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Printf("每日新条目上限: %s\n", formatLimit(config.AppConfig.DailyNewLimit))
			fmt.Printf("每日复习上限:   %s\n", formatLimit(config.AppConfig.DailyReviewLimit))
			fmt.Println("\n用法: mllt-cli setting daily <新条目上限> [复习上限]，0 表示不限")
			return
		}

		newLimit, err := strconv.Atoi(args[0])
		if err != nil || newLimit < 0 {
			fmt.Printf("无效的新条目上限: %s\n", args[0])
			return
		}
		config.AppConfig.DailyNewLimit = newLimit

		if len(args) == 2 {
			reviewLimit, err := strconv.Atoi(args[1])
			if err != nil || reviewLimit < 0 {
				fmt.Printf("无效的复习上限: %s\n", args[1])
				return
			}
			config.AppConfig.DailyReviewLimit = reviewLimit
		}

		if err := config.SaveConfig(); err != nil {
			fmt.Printf("保存配置失败: %s\n", err)
			return
		}
		fmt.Printf("每日新条目上限已设为 %s，复习上限 %s\n",
			formatLimit(config.AppConfig.DailyNewLimit),
			formatLimit(config.AppConfig.DailyReviewLimit))
	},
}

// formatLimit 把 0 显示成"不限"
func formatLimit(limit int) string {
	if limit <= 0 {
		return "不限"
	}
	return fmt.Sprintf("%d", limit)
}

func init() {
	// 确保默认资源与配置已初始化
	if err := mlltcli.EnsureAssets(); err != nil {
		fmt.Println("警告: 初始化默认资源失败:", err)
	}

	// 加载配置文件
	if err := config.LoadConfig(); err != nil {
		fmt.Println("警告: 加载配置文件失败:", err)
		fmt.Println("将使用默认配置。")
	}

	// 添加子命令到根命令
	rootCmd.AddCommand(langCmd)
	rootCmd.AddCommand(practiceCmd)
	rootCmd.AddCommand(manageCmd)
	rootCmd.AddCommand(settingCmd)

	// 添加lang子命令
	langCmd.AddCommand(langLsCmd)
	langCmd.AddCommand(langStCmd)

	// 添加practice子命令
	practiceCmd.AddCommand(practiceWordsCmd)
	practiceCmd.AddCommand(practicePhrasesCmd)
	practiceCmd.AddCommand(practiceSentencesCmd)
	practiceCmd.AddCommand(practiceArticlesCmd)

	// 添加manage子命令
	manageCmd.AddCommand(manageDeleteCmd)
	manageCmd.AddCommand(manageImportCmd)

	// 添加setting子命令
	settingCmd.AddCommand(settingMatchModeCmd)
	settingCmd.AddCommand(settingOrderCmd)
	settingCmd.AddCommand(settingKeyboardSoundCmd)
	settingCmd.AddCommand(settingTranslationCmd)
	settingCmd.AddCommand(settingExamplesCmd)
	settingCmd.AddCommand(settingDailyCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
