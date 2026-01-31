package ui

import (
	"testing"

	"github.com/ajilisiwei/mllt-cli/internal/config"
	"github.com/ajilisiwei/mllt-cli/internal/practice"
)

// 测试前的准备工作
func setupPracticeSessionTest(t *testing.T) {
	// 确保配置已加载
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
}

// 测试getShowTranslationConfig方法
func TestGetShowTranslationConfig(t *testing.T) {
	setupPracticeSessionTest(t)

	// 创建测试会话
	session := &PracticeSession{}

	// 测试全局翻译设置为true
	config.AppConfig.ShowTranslation = true
	if !session.getShowTranslationConfig() {
		t.Error("全局设置为true时应该显示翻译")
	}

	// 测试全局翻译设置为false
	config.AppConfig.ShowTranslation = false
	if session.getShowTranslationConfig() {
		t.Error("全局设置为false时不应该显示翻译")
	}
}

// 测试getExpectedInput方法
func TestGetExpectedInput(t *testing.T) {
	setupPracticeSessionTest(t)

	// 创建测试会话
	session := &PracticeSession{
		resourceType: practice.Words,
	}

	tests := []struct {
		name     string
		item     string
		expected string
	}{
		{
			name:     "单词带翻译",
			item:     "apple ->> 苹果",
			expected: "apple",
		},
		{
			name:     "短语带翻译",
			item:     "good morning ->> 早上好",
			expected: "good morning",
		},
		{
			name:     "只有正文无翻译",
			item:     "hello",
			expected: "hello",
		},
		{
			name:     "空字符串",
			item:     "",
			expected: "",
		},
		{
			name:     "带前后空格的项目",
			item:     "  world ->> 世界  ",
			expected: "world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := session.getExpectedInput(tt.item)
			if result != tt.expected {
				t.Errorf("getExpectedInput() = %v, want %v", result, tt.expected)
			}
		})
	}

	// 测试短语类型
	session.resourceType = practice.Phrases
	result := session.getExpectedInput("good afternoon ->> 下午好")
	expected := "good afternoon"
	if result != expected {
		t.Errorf("短语类型 getExpectedInput() = %v, want %v", result, expected)
	}

	// 测试句子类型（应该分离翻译）
	session.resourceType = practice.Sentences
	result = session.getExpectedInput("How are you? ->> 你好吗？")
	expected = "How are you?"
	if result != expected {
		t.Errorf("句子类型 getExpectedInput() = %v, want %v", result, expected)
	}
}

// 测试isInputCorrect方法
func TestIsInputCorrect(t *testing.T) {
	setupPracticeSessionTest(t)

	// 创建测试会话
	session := &PracticeSession{}

	// 测试完全匹配模式
	config.AppConfig.CorrectnessMatchMode = "exact_match"

	tests := []struct {
		name          string
		userInput     string
		expectedInput string
		want          bool
	}{
		{
			name:          "完全匹配 - 正确",
			userInput:     "apple",
			expectedInput: "apple",
			want:          true,
		},
		{
			name:          "完全匹配 - 错误",
			userInput:     "aple",
			expectedInput: "apple",
			want:          false,
		},
		{
			name:          "完全匹配 - 大小写不同",
			userInput:     "Apple",
			expectedInput: "apple",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := session.isInputCorrect(tt.userInput, tt.expectedInput)
			if result != tt.want {
				t.Errorf("isInputCorrect() = %v, want %v", result, tt.want)
			}
		})
	}

	// 测试单词匹配模式
	config.AppConfig.CorrectnessMatchMode = "word_match"

	wordMatchTests := []struct {
		name          string
		userInput     string
		expectedInput string
		want          bool
	}{
		{
			name:          "单词匹配 - 大小写不同但正确",
			userInput:     "Apple",
			expectedInput: "apple",
			want:          true,
		},
		{
			name:          "单词匹配 - 标点符号不同但正确",
			userInput:     "Hello, world!",
			expectedInput: "Hello world",
			want:          true,
		},
		{
			name:          "单词匹配 - 多余空格但正确",
			userInput:     "good  morning",
			expectedInput: "good morning",
			want:          true,
		},
		{
			name:          "单词匹配 - 单词错误",
			userInput:     "god morning",
			expectedInput: "good morning",
			want:          false,
		},
	}

	for _, tt := range wordMatchTests {
		t.Run(tt.name, func(t *testing.T) {
			result := session.isInputCorrect(tt.userInput, tt.expectedInput)
			if result != tt.want {
				t.Errorf("isInputCorrect() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestIsInputCorrectForItem(t *testing.T) {
	setupPracticeSessionTest(t)

	session := &PracticeSession{}
	item := "People often use the phrase '2' in conversation. ->> In everyday conversations, people frequently mention 'a great number of' when they want to emphasize quantity."

	config.AppConfig.CorrectnessMatchMode = "exact_match"
	exactMatchCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "只输入原文",
			input: "People often use the phrase '2' in conversation.",
			want:  true,
		},
		{
			name:  "输入原文和翻译",
			input: "People often use the phrase '2' in conversation. In everyday conversations, people frequently mention 'a great number of' when they want to emphasize quantity.",
			want:  true,
		},
		{
			name:  "输入带箭头的完整原始行",
			input: "People often use the phrase '2' in conversation. ->> In everyday conversations, people frequently mention 'a great number of' when they want to emphasize quantity.",
			want:  true,
		},
		{
			name:  "仅输入翻译",
			input: "In everyday conversations, people frequently mention 'a great number of' when they want to emphasize quantity.",
			want:  false,
		},
	}

	for _, tt := range exactMatchCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := session.isInputCorrectForItem(tt.input, item); got != tt.want {
				t.Errorf("isInputCorrectForItem() = %v, want %v", got, tt.want)
			}
		})
	}

	config.AppConfig.CorrectnessMatchMode = "word_match"
	wordMatchCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "忽略标点仍判定正确",
			input: "People often use the phrase 2 in conversation In everyday conversations people frequently mention a great number of when they want to emphasize quantity",
			want:  true,
		},
		{
			name:  "单词错误仍判定错误",
			input: "People often use the phrase two in chat. In common talks, folks often mention a great number when they want to emphasize quantity.",
			want:  false,
		},
	}

	for _, tt := range wordMatchCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := session.isInputCorrectForItem(tt.input, item); got != tt.want {
				t.Errorf("isInputCorrectForItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试normalizeForWordMatch方法
func TestNormalizeForWordMatch(t *testing.T) {
	// 创建测试会话
	session := &PracticeSession{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "转换大小写",
			input:    "Hello World",
			expected: "hello world",
		},
		{
			name:     "移除标点符号",
			input:    "Hello, world!",
			expected: "hello world",
		},
		{
			name:     "处理多余空格",
			input:    "  hello   world  ",
			expected: "hello world",
		},
		{
			name:     "复杂标点符号",
			input:    "It's a beautiful day, isn't it?",
			expected: "its a beautiful day isnt it",
		},
		{
			name:     "数字保留",
			input:    "I have 5 apples.",
			expected: "i have 5 apples",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := session.normalizeForWordMatch(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeForWordMatch() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// 测试getCurrentItem方法
func TestGetCurrentItem(t *testing.T) {
	setupPracticeSessionTest(t)

	// 创建测试会话
	session := &PracticeSession{
		resourceType:   practice.Words,
		items:          []string{"apple ->> 苹果", "banana ->> 香蕉", "orange ->> 橙子"},
		practiceOrder:  []int{0, 1, 2},
		completedCount: 0,
	}

	// 测试不显示翻译的情况
	config.AppConfig.ShowTranslation = false
	tests := []struct {
		name           string
		completedCount int
		expected       string
	}{
		{
			name:           "第一个项目 - 不显示翻译",
			completedCount: 0,
			expected:       "apple",
		},
		{
			name:           "第二个项目 - 不显示翻译",
			completedCount: 1,
			expected:       "banana",
		},
		{
			name:           "第三个项目 - 不显示翻译",
			completedCount: 2,
			expected:       "orange",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session.completedCount = tt.completedCount
			result := session.getCurrentItem()
			if result != tt.expected {
				t.Errorf("getCurrentItem() = %v, want %v", result, tt.expected)
			}
		})
	}

	// 测试显示翻译的情况
	config.AppConfig.ShowTranslation = true
	session.completedCount = 0
	result := session.getCurrentItem()
	expected := "apple\n翻译: 苹果"
	if result != expected {
		t.Errorf("显示翻译 getCurrentItem() = %v, want %v", result, expected)
	}

	// 测试短语类型 - 不显示翻译
	session.resourceType = practice.Phrases
	session.items = []string{"good morning ->> 早上好", "good afternoon ->> 下午好"}
	session.completedCount = 0
	config.AppConfig.ShowTranslation = false
	result = session.getCurrentItem()
	expected = "good morning"
	if result != expected {
		t.Errorf("短语类型不显示翻译 getCurrentItem() = %v, want %v", result, expected)
	}

	// 测试短语类型 - 显示翻译
	config.AppConfig.ShowTranslation = true
	result = session.getCurrentItem()
	expected = "good morning\n翻译: 早上好"
	if result != expected {
		t.Errorf("短语类型显示翻译 getCurrentItem() = %v, want %v", result, expected)
	}

	// 测试句子类型 - 不显示翻译
	session.resourceType = practice.Sentences
	session.items = []string{"How are you? ->> 你好吗？"}
	session.completedCount = 0
	config.AppConfig.ShowTranslation = false
	result = session.getCurrentItem()
	expected = "How are you?"
	if result != expected {
		t.Errorf("句子类型不显示翻译 getCurrentItem() = %v, want %v", result, expected)
	}

	// 测试句子类型 - 显示翻译
	config.AppConfig.ShowTranslation = true
	result = session.getCurrentItem()
	expected = "How are you?\n翻译: 你好吗？"
	if result != expected {
		t.Errorf("句子类型显示翻译 getCurrentItem() = %v, want %v", result, expected)
	}

	// 测试文章类型 - 不显示翻译
	session.resourceType = practice.Articles
	session.items = []string{"This is a test. ->> 这是一个测试。"}
	session.completedCount = 0
	config.AppConfig.ShowTranslation = false
	result = session.getCurrentItem()
	expected = "This is a test."
	if result != expected {
		t.Errorf("文章类型不显示翻译 getCurrentItem() = %v, want %v", result, expected)
	}

	// 测试文章类型 - 显示翻译
	config.AppConfig.ShowTranslation = true
	result = session.getCurrentItem()
	expected = "This is a test.\n翻译: 这是一个测试。"
	if result != expected {
		t.Errorf("文章类型显示翻译 getCurrentItem() = %v, want %v", result, expected)
	}

	// 测试超出范围的情况
	session.completedCount = 10
	result = session.getCurrentItem()
	if result != "" {
		t.Errorf("超出范围 getCurrentItem() = %v, want empty string", result)
	}
}
