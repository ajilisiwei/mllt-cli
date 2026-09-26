package practice

import (
	"strings"
	"testing"
)

func TestParseWordLine(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		word     string
		phonetic string
		meaning  string
	}{
		{
			name:     "missing tab between word and phonetic",
			line:     "fool/fuːl/\tn. 傻瓜，愚人，笨蛋",
			word:     "fool",
			phonetic: "/fuːl/",
			meaning:  "n. 傻瓜，愚人，笨蛋",
		},
		{
			name:     "space instead of tab between word and phonetic",
			line:     "gay /ɡeɪ/\tadj. 同性恋的；快乐的",
			word:     "gay",
			phonetic: "/ɡeɪ/",
			meaning:  "adj. 同性恋的；快乐的",
		},
		{
			name:     "multiple spaces before phonetic",
			line:     "hamburger   /ˈhæmbɜːrɡər/\tn. 汉堡包",
			word:     "hamburger",
			phonetic: "/ˈhæmbɜːrɡər/",
			meaning:  "n. 汉堡包",
		},
		{
			name:     "well formed three column entry",
			line:     "abandon\t/əˈbændən/\tv. 遗弃；离开；放弃",
			word:     "abandon",
			phonetic: "/əˈbændən/",
			meaning:  "v. 遗弃；离开；放弃",
		},
		{
			name:     "phonetic containing a comma",
			line:     "as/əz,æz/\tprep. 作为",
			word:     "as",
			phonetic: "/əz,æz/",
			meaning:  "prep. 作为",
		},
		{
			name:     "space separated legacy format",
			line:     "abandon / əˈbændən / vt.丢弃；放弃，抛弃  ",
			word:     "abandon",
			phonetic: "/əˈbændən/",
			meaning:  "vt.丢弃；放弃，抛弃",
		},
		{
			name:     "arrow separated format",
			line:     "computer ->> /kəmˈpjuːtər/ n. 计算机，电脑",
			word:     "computer",
			phonetic: "/kəmˈpjuːtər/",
			meaning:  "n. 计算机，电脑",
		},
		{
			name:     "entry without phonetic",
			line:     "apple ->> 苹果",
			word:     "apple",
			phonetic: "",
			meaning:  "苹果",
		},
		{
			name:     "tab separated entry without phonetic",
			line:     "hello\tn. 你好",
			word:     "hello",
			phonetic: "",
			meaning:  "n. 你好",
		},
		{
			name:     "slash inside the word is not a phonetic",
			line:     "and/or ->> 和／或",
			word:     "and/or",
			phonetic: "",
			meaning:  "和／或",
		},
		{
			name:     "unit with slash is not a phonetic",
			line:     "km/h ->> 千米每小时",
			word:     "km/h",
			phonetic: "",
			meaning:  "千米每小时",
		},
		{
			name:     "bare word",
			line:     "standalone",
			word:     "standalone",
			phonetic: "",
			meaning:  "",
		},
		{
			name:     "phonetic without a word keeps the raw text",
			line:     "/fuːl/\tn. 傻瓜",
			word:     "/fuːl/",
			phonetic: "",
			meaning:  "n. 傻瓜",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			word, phonetic, meaning := ParseWordLine(tc.line)
			if word != tc.word {
				t.Errorf("word = %q, expected %q", word, tc.word)
			}
			if phonetic != tc.phonetic {
				t.Errorf("phonetic = %q, expected %q", phonetic, tc.phonetic)
			}
			if meaning != tc.meaning {
				t.Errorf("meaning = %q, expected %q", meaning, tc.meaning)
			}
		})
	}
}

func TestWordPrimaryText(t *testing.T) {
	cases := []struct {
		line     string
		expected string
	}{
		{"fool/fuːl/\tn. 傻瓜", "fool"},
		{"gay /ɡeɪ/\tadj. 快乐的", "gay"},
		{"abandon\t/əˈbændən/\tv. 遗弃", "abandon"},
		{"and/or ->> 和／或", "and/or"},
	}

	for _, tc := range cases {
		if got := WordPrimaryText(tc.line); got != tc.expected {
			t.Errorf("WordPrimaryText(%q) = %q, expected %q", tc.line, got, tc.expected)
		}
	}
}

// 非单词资源不应被音标规则影响。
func TestParseLineKeepsSentencesIntact(t *testing.T) {
	line := "Are you kidding me?"
	original, translation := ParseLine(line)
	if original != line || translation != "" {
		t.Fatalf("ParseLine(%q) = (%q, %q)", line, original, translation)
	}
}

func TestParseWordEntry(t *testing.T) {
	cases := []struct {
		name string
		line string
		want WordEntry
	}{
		{
			name: "定长列加三个例句",
			line: strings.Join([]string{
				"abandon", "/əˈbændən/", "v. 放弃，抛弃",
				"abandon a plan｜abandon ship", "abandoned adj. 被遗弃的",
				"They abandoned the plan after the first test.", "第一次测试后他们就放弃了那个方案。",
				"We had to abandon ship.", "我们不得不弃船。",
				"Don't abandon hope just yet.", "先别放弃希望。",
			}, "\t"),
			want: WordEntry{
				Word: "abandon", Phonetic: "/əˈbændən/", Meaning: "v. 放弃，抛弃",
				Collocation: "abandon a plan｜abandon ship",
				Family:      "abandoned adj. 被遗弃的",
				Examples: []Contrast{
					{Text: "They abandoned the plan after the first test.", Note: "第一次测试后他们就放弃了那个方案。"},
					{Text: "We had to abandon ship.", Note: "我们不得不弃船。"},
					{Text: "Don't abandon hope just yet.", Note: "先别放弃希望。"},
				},
				Example:     "They abandoned the plan after the first test.",
				ExampleNote: "第一次测试后他们就放弃了那个方案。",
			},
		},
		{
			name: "只有一个例句",
			line: "apply\t/əˈplaɪ/\tv. 申请\tapply for\t\tShe applied for the job.\t她申请了那份工作。",
			want: WordEntry{
				Word: "apply", Phonetic: "/əˈplaɪ/", Meaning: "v. 申请",
				Collocation: "apply for",
				Examples:    []Contrast{{Text: "She applied for the job.", Note: "她申请了那份工作。"}},
				Example:     "She applied for the job.", ExampleNote: "她申请了那份工作。",
			},
		},
		{
			name: "三列：现有词库的标准格式",
			line: "ability\t/əˈbɪləti/\tn. 能力，能耐；才能",
			want: WordEntry{Word: "ability", Phonetic: "/əˈbɪləti/", Meaning: "n. 能力，能耐；才能"},
		},
		{
			name: "两列：缺制表符的历史脏数据仍要兜住",
			line: "fool/fuːl/\tn. 傻瓜",
			want: WordEntry{Word: "fool", Phonetic: "/fuːl/", Meaning: "n. 傻瓜"},
		},
		{
			name: "例句练习项：正文加翻译，中间列留空",
			line: "They abandoned the plan.\t\t他们放弃了那个方案。",
			want: WordEntry{Word: "They abandoned the plan.", Meaning: "他们放弃了那个方案。"},
		},
		{
			name: "音标列没写斜杠也能识别",
			line: "issue\tˈɪʃuː\tn. 问题；议题",
			want: WordEntry{Word: "issue", Phonetic: "/ˈɪʃuː/", Meaning: "n. 问题；议题"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseWordEntry(tc.line)

			if got.Word != tc.want.Word || got.Phonetic != tc.want.Phonetic ||
				got.Meaning != tc.want.Meaning || got.Collocation != tc.want.Collocation ||
				got.Family != tc.want.Family {
				t.Errorf("定长列\n got = %+v\nwant = %+v", got, tc.want)
			}
			if len(got.Examples) != len(tc.want.Examples) {
				t.Fatalf("例句数量 = %d，期望 %d: %+v", len(got.Examples), len(tc.want.Examples), got.Examples)
			}
			for i := range tc.want.Examples {
				if got.Examples[i] != tc.want.Examples[i] {
					t.Errorf("第 %d 个例句 = %+v，期望 %+v", i+1, got.Examples[i], tc.want.Examples[i])
				}
			}
			if got.Example != tc.want.Example || got.ExampleNote != tc.want.ExampleNote {
				t.Errorf("兼容字段 = (%q, %q)", got.Example, got.ExampleNote)
			}
		})
	}
}

// 三个例句要展开成三个独立的练习项，并且紧跟在单词后面
func TestWordExamplesAllExpand(t *testing.T) {
	line := strings.Join([]string{
		"abandon", "/əˈbændən/", "v. 放弃", "abandon a plan", "",
		"They abandoned the plan.", "他们放弃了那个方案。",
		"We had to abandon ship.", "我们不得不弃船。",
		"Don't abandon hope.", "别放弃希望。",
	}, "\t")

	blocks := ExpandEntryBlocks([]string{line})
	if len(blocks) != 1 {
		t.Fatalf("应该只有一个块，实际 %d 个", len(blocks))
	}
	if len(blocks[0]) != 4 {
		t.Fatalf("块内 %d 项，期望 4 项（单词 + 三个例句）: %q", len(blocks[0]), blocks[0])
	}

	want := []string{"abandon", "They abandoned the plan.", "We had to abandon ship.", "Don't abandon hope."}
	for i, w := range want {
		if got := WordPrimaryText(blocks[0][i]); got != w {
			t.Errorf("块内第 %d 项 = %q，期望 %q", i, got, w)
		}
	}
}

// 扩展列不能改变打字答案，否则已有的练习进度会失效
func TestParseWordEntryKeepsExpectedInput(t *testing.T) {
	lines := []string{
		"abandon\t/əˈbændən/\tv. 放弃\tThey abandoned it.\t他们放弃了。\tabandon a plan\tabandonment n. 放弃",
		"ability\t/əˈbɪləti/\tn. 能力",
		"fool/fuːl/\tn. 傻瓜",
	}
	want := []string{"abandon", "ability", "fool"}

	for i, line := range lines {
		if got := WordPrimaryText(line); got != want[i] {
			t.Errorf("WordPrimaryText(%q) = %q, want %q", line, got, want[i])
		}
	}
}
