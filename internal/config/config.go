package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 表示应用程序的配置
type Config struct {
	// 支持的语言列表
	Languages []string `mapstructure:"languages"`
	// 当前选中的语言
	CurrentLanguage string `mapstructure:"current_language"`
	// 单词练习配置
	Words WordsConfig `mapstructure:"words"`
	// 短语练习配置
	Phrases PhrasesConfig `mapstructure:"phrases"`
	// 句子练习配置
	Sentences SentencesConfig `mapstructure:"sentences"`
	// 文章练习配置
	Articles ArticlesConfig `mapstructure:"articles"`
	// v0.2 新增：正确性匹配模式，可选值：exact_match（完全匹配）、word_match（单词匹配）
	CorrectnessMatchMode string `mapstructure:"correctness_match_mode"`
	// v0.2 新增：全局下一个资源的出现顺序，可选值：random（随机）、sequential（顺序）
	NextOneOrder string `mapstructure:"next_one_order"`
	// v0.3 新增：是否启用键盘按键声音
	InputKeyboardSound bool `mapstructure:"input_keyboard_sound"`
	// v0.3 新增：全局是否显示翻译
	ShowTranslation bool `mapstructure:"show_translation"`
	// v0.4 新增：短语的例句是否也作为练习内容，而不只是展示
	PracticeExamples bool `mapstructure:"practice_examples"`
	// v0.4 新增：艾宾浩斯模式下每天引入的新条目上限，0 表示不限
	DailyNewLimit int `mapstructure:"daily_new_limit"`
	// v0.4 新增：艾宾浩斯模式下每天的复习上限，0 表示不限
	DailyReviewLimit int `mapstructure:"daily_review_limit"`
	// v0.4 新增：练习方向，copy（看英文抄写）或 translate（看中文译写）
	PracticeDirection string `mapstructure:"practice_direction"`
}

// WordsConfig 表示单词练习的配置
type WordsConfig struct {
	// 单词练习配置（目前为空，预留扩展）
}

// PhrasesConfig 表示短语练习的配置
type PhrasesConfig struct {
	// 短语练习配置（目前为空，预留扩展）
}

// SentencesConfig 表示句子练习的配置
type SentencesConfig struct {
	// 句子练习配置（目前为空，预留扩展）
}

// ArticlesConfig 表示文章练习的配置
type ArticlesConfig struct {
	// 文章练习配置（目前为空，预留扩展）
}

// 全局配置实例
var AppConfig Config

// LoadConfig 加载配置文件
// 开发阶段读取项目内配置文件，生产环境读取用户目录配置文件
func LoadConfig() error {
	// 设置配置文件的名称和路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 老版本的配置文件里没有这个键，缺省按「例句也要练」处理
	viper.SetDefault("practice_examples", true)
	viper.SetDefault("daily_new_limit", 50)
	viper.SetDefault("daily_review_limit", 0)
	viper.SetDefault("practice_direction", "copy")

	// 判断是否为开发环境
	isDevMode := isInDevelopmentMode()

	if isDevMode {
		// 开发环境：优先读取项目内的配置文件
		viper.AddConfigPath("./config")
		viper.AddConfigPath(".")
		// 添加测试环境下的配置文件路径
		viper.AddConfigPath("../../config")
	} else {
		// 生产环境：读取用户目录下的配置文件
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户主目录失败: %w", err)
		}
		configDir := filepath.Join(homeDir, ".mllt-cli")
		viper.AddConfigPath(configDir)

		// 获取当前执行文件的目录作为备用路径
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			viper.AddConfigPath(filepath.Join(execDir, "config"))
		}
	}

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 将配置文件的内容解析到结构体中
	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// isInDevelopmentMode 判断是否为开发模式
// 通过检查项目根目录是否存在go.mod文件来判断
func isInDevelopmentMode() bool {
	// 检查当前目录及上级目录是否存在go.mod文件
	paths := []string{"./go.mod", "../go.mod", "../../go.mod"}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// SaveConfig 保存配置到文件
func SaveConfig() error {
	// 将结构体的内容写入到配置文件中
	for k, v := range map[string]interface{}{
		"languages":              AppConfig.Languages,
		"current_language":       AppConfig.CurrentLanguage,
		"words":                  AppConfig.Words,
		"phrases":                AppConfig.Phrases,
		"sentences":              AppConfig.Sentences,
		"articles":               AppConfig.Articles,
		"correctness_match_mode": AppConfig.CorrectnessMatchMode,
		"next_one_order":         AppConfig.NextOneOrder,
		"input_keyboard_sound":   AppConfig.InputKeyboardSound,
		"show_translation":       AppConfig.ShowTranslation,
		"practice_examples":      AppConfig.PracticeExamples,
		"daily_new_limit":        AppConfig.DailyNewLimit,
		"daily_review_limit":     AppConfig.DailyReviewLimit,
		"practice_direction":     AppConfig.PracticeDirection,
	} {
		viper.Set(k, v)
	}

	// 保存配置文件
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}
