package ui

import (
	"strings"
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

// 单词与音标黏连时，期望输入必须只包含单词本身（回归 "fool/fuːl/" 问题）
func TestGetExpectedInputStripsPhonetic(t *testing.T) {
	setupPracticeSessionTest(t)

	words := &PracticeSession{resourceType: practice.Words}
	sentences := &PracticeSession{resourceType: practice.Sentences}

	tests := []struct {
		name     string
		session  *PracticeSession
		item     string
		expected string
	}{
		{"缺少制表符", words, "fool/fuːl/\tn. 傻瓜，愚人", "fool"},
		{"空格代替制表符", words, "gay /ɡeɪ/\tadj. 快乐的", "gay"},
		{"规范三列格式", words, "abandon\t/əˈbændən/\tv. 遗弃", "abandon"},
		{"词条本身含斜杠", words, "CI/CD\t/siː aɪ siː diː/\tn. 持续集成", "CI/CD"},
		{"句子不受音标规则影响", sentences, "Are you kidding me?", "Are you kidding me?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.session.getExpectedInput(tt.item); got != tt.expected {
				t.Errorf("getExpectedInput(%q) = %q, want %q", tt.item, got, tt.expected)
			}
		})
	}
}

// 单词条目应把音标单独成行展示，而不是混在正文或翻译里
func TestGetCurrentItemSplitsPhonetic(t *testing.T) {
	setupPracticeSessionTest(t)

	session := &PracticeSession{
		resourceType:   practice.Words,
		items:          []string{"fool/fuːl/\tn. 傻瓜，愚人"},
		practiceOrder:  []int{0},
		completedCount: 0,
	}

	config.AppConfig.ShowTranslation = true
	want := "fool\n音标: /fuːl/\n翻译: n. 傻瓜，愚人"
	if got := session.getCurrentItem(); got != want {
		t.Errorf("显示翻译时 getCurrentItem() = %q, want %q", got, want)
	}

	config.AppConfig.ShowTranslation = false
	if got := session.getCurrentItem(); got != "fool" {
		t.Errorf("隐藏翻译时 getCurrentItem() = %q, want %q", got, "fool")
	}
}

// 短语条目应把释义、例句、例句翻译分行展示
func TestGetCurrentItemSplitsExample(t *testing.T) {
	setupPracticeSessionTest(t)

	session := &PracticeSession{
		resourceType:   practice.Phrases,
		items:          []string{"look into ->> 调查；了解一下 ->> I'll look into it and get back to you. ->> 我查一下再回复你。"},
		practiceOrder:  []int{0},
		completedCount: 0,
	}

	config.AppConfig.ShowTranslation = true
	want := "look into\n翻译: 调查；了解一下\n例句: I'll look into it and get back to you.\n译文: 我查一下再回复你。"
	if got := session.getCurrentItem(); got != want {
		t.Errorf("getCurrentItem() = %q\nwant %q", got, want)
	}

	// 打字答案只取正文，不受例句影响
	if got := session.getExpectedInput(session.items[0]); got != "look into" {
		t.Errorf("getExpectedInput() = %q, want %q", got, "look into")
	}

	config.AppConfig.ShowTranslation = false
	if got := session.getCurrentItem(); got != "look into" {
		t.Errorf("隐藏翻译时 getCurrentItem() = %q, want %q", got, "look into")
	}
}

// 旧的两段格式必须原样兼容
func TestGetCurrentItemKeepsLegacyTwoFieldEntries(t *testing.T) {
	setupPracticeSessionTest(t)

	session := &PracticeSession{
		resourceType:   practice.Sentences,
		items:          []string{"I'll check and let you know. ->> 我查一下再告知你。"},
		practiceOrder:  []int{0},
		completedCount: 0,
	}

	config.AppConfig.ShowTranslation = true
	want := "I'll check and let you know.\n翻译: 我查一下再告知你。"
	if got := session.getCurrentItem(); got != want {
		t.Errorf("getCurrentItem() = %q\nwant %q", got, want)
	}
}

// 展开出来的例句项要能像普通句子一样练：显示正文+翻译，答案就是例句本身
func TestExpandedExamplePracticesAsSentence(t *testing.T) {
	setupPracticeSessionTest(t)

	items := practice.ExpandEntries([]string{
		"sing along ->> 跟着唱 ->> Everyone was singing along. ->> 全场都跟着唱。",
	})
	if len(items) != 2 {
		t.Fatalf("展开后 %d 项，期望 2 项", len(items))
	}

	session := &PracticeSession{
		resourceType:   practice.Phrases,
		items:          items,
		practiceOrder:  []int{0, 1},
		completedCount: 1, // 第二项就是例句
	}

	config.AppConfig.ShowTranslation = true
	want := "Everyone was singing along.\n翻译: 全场都跟着唱。"
	if got := session.getCurrentItem(); got != want {
		t.Errorf("例句项显示 = %q\nwant %q", got, want)
	}
	if got := session.getExpectedInput(items[1]); got != "Everyone was singing along." {
		t.Errorf("例句项答案 = %q", got)
	}

	// 短语项本身不变
	if got := session.getExpectedInput(items[0]); got != "sing along" {
		t.Errorf("短语项答案 = %q", got)
	}
}

// 打乱顺序后，例句仍必须紧跟在它的短语后面，否则打整句时已经没有上下文
func TestExpandInPracticeOrderKeepsPairsTogether(t *testing.T) {
	items := []string{
		"be tone-deaf ->> 五音不全 ->> I'm completely tone-deaf. ->> 我完全五音不全。",
		"apple ->> 苹果",
		"go viral ->> 爆火 ->> That clip went viral overnight. ->> 那段视频一夜爆火。",
	}

	expanded, isExample, order := expandInPracticeOrder(items, []int{2, 0, 1})

	want := []string{
		"go viral ->> 爆火 ->> That clip went viral overnight. ->> 那段视频一夜爆火。",
		"That clip went viral overnight. ->> 那段视频一夜爆火。",
		"be tone-deaf ->> 五音不全 ->> I'm completely tone-deaf. ->> 我完全五音不全。",
		"I'm completely tone-deaf. ->> 我完全五音不全。",
		"apple ->> 苹果",
	}
	if len(expanded) != len(want) {
		t.Fatalf("展开后 %d 项，期望 %d 项: %q", len(expanded), len(want), expanded)
	}
	for i := range want {
		if expanded[i] != want[i] {
			t.Errorf("第 %d 项 = %q\n期望 %q", i, expanded[i], want[i])
		}
	}

	wantExample := []bool{false, true, false, true, false}
	for i := range wantExample {
		if isExample[i] != wantExample[i] {
			t.Errorf("第 %d 项 isExample = %v, 期望 %v", i, isExample[i], wantExample[i])
		}
	}

	for i := range order {
		if order[i] != i {
			t.Fatalf("展开后练习顺序应为恒等，第 %d 项 = %d", i, order[i])
		}
	}
}

// 扩展列齐全的单词条目要分行展示，并且例句能变成练习项
func TestWordEntryWithAllColumns(t *testing.T) {
	setupPracticeSessionTest(t)

	line := strings.Join([]string{
		"abandon",
		"/əˈbændən/",
		"v. 放弃，抛弃",
		"They abandoned the plan after the first test.",
		"第一次测试后他们就放弃了那个方案。",
		"abandon a plan｜abandon ship",
		"abandoned adj. 被遗弃的 · abandonment n. 放弃",
	}, "\t")

	items := practice.ExpandEntries([]string{line})
	if len(items) != 2 {
		t.Fatalf("展开后 %d 项，期望 2 项: %q", len(items), items)
	}

	session := &PracticeSession{
		resourceType:   practice.Words,
		items:          items,
		practiceOrder:  []int{0, 1},
		completedCount: 0,
	}
	config.AppConfig.ShowTranslation = true

	want := strings.Join([]string{
		"abandon",
		"音标: /əˈbændən/",
		"翻译: v. 放弃，抛弃",
		"例句: They abandoned the plan after the first test.",
		"译文: 第一次测试后他们就放弃了那个方案。",
		"搭配: abandon a plan｜abandon ship",
		"词族: abandoned adj. 被遗弃的 · abandonment n. 放弃",
	}, "\n")
	if got := session.getCurrentItem(); got != want {
		t.Errorf("单词项显示 =\n%s\n期望 =\n%s", got, want)
	}
	if got := session.getExpectedInput(items[0]); got != "abandon" {
		t.Errorf("单词项答案 = %q", got)
	}

	// 例句项：打整句，显示翻译
	session.completedCount = 1
	wantExample := "They abandoned the plan after the first test.\n翻译: 第一次测试后他们就放弃了那个方案。"
	if got := session.getCurrentItem(); got != wantExample {
		t.Errorf("例句项显示 =\n%s\n期望 =\n%s", got, wantExample)
	}
	if got := session.getExpectedInput(items[1]); got != "They abandoned the plan after the first test." {
		t.Errorf("例句项答案 = %q", got)
	}
}

// 三列的老词库不受扩展列影响
func TestLegacyWordEntryUnchanged(t *testing.T) {
	setupPracticeSessionTest(t)

	session := &PracticeSession{
		resourceType:   practice.Words,
		items:          practice.ExpandEntries([]string{"ability\t/əˈbɪləti/\tn. 能力，能耐；才能"}),
		practiceOrder:  []int{0},
		completedCount: 0,
	}
	if len(session.items) != 1 {
		t.Fatalf("没有例句的词条不该被展开，实际 %d 项", len(session.items))
	}

	config.AppConfig.ShowTranslation = true
	want := "ability\n音标: /əˈbɪləti/\n翻译: n. 能力，能耐；才能"
	if got := session.getCurrentItem(); got != want {
		t.Errorf("getCurrentItem() = %q, want %q", got, want)
	}
}
