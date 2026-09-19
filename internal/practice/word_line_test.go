package practice

import "testing"

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
			name: "完整七列",
			line: "abandon\t/əˈbændən/\tv. 放弃，抛弃\tThey abandoned the plan after the first test.\t第一次测试后他们就放弃了那个方案。\tabandon a plan｜abandon ship\tabandoned adj. 被遗弃的 · abandonment n. 放弃",
			want: WordEntry{
				Word:        "abandon",
				Phonetic:    "/əˈbændən/",
				Meaning:     "v. 放弃，抛弃",
				Example:     "They abandoned the plan after the first test.",
				ExampleNote: "第一次测试后他们就放弃了那个方案。",
				Collocation: "abandon a plan｜abandon ship",
				Family:      "abandoned adj. 被遗弃的 · abandonment n. 放弃",
			},
		},
		{
			name: "六列：有例句有搭配，无词族",
			line: "apply\t/əˈplaɪ/\tv. 申请；应用\tShe applied for the job last week.\t她上周申请了那份工作。\tapply for｜apply to",
			want: WordEntry{
				Word:        "apply",
				Phonetic:    "/əˈplaɪ/",
				Meaning:     "v. 申请；应用",
				Example:     "She applied for the job last week.",
				ExampleNote: "她上周申请了那份工作。",
				Collocation: "apply for｜apply to",
			},
		},
		{
			name: "五列：只到例句翻译",
			line: "able\t/ˈeɪbl/\tadj. 有能力的\tShe was able to fix it herself.\t她自己就能修好。",
			want: WordEntry{
				Word:        "able",
				Phonetic:    "/ˈeɪbl/",
				Meaning:     "adj. 有能力的",
				Example:     "She was able to fix it herself.",
				ExampleNote: "她自己就能修好。",
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
			line: "issue\tˈɪʃuː\tn. 问题；议题\tThat's a separate issue.\t那是另一个问题。",
			want: WordEntry{
				Word:        "issue",
				Phonetic:    "/ˈɪʃuː/",
				Meaning:     "n. 问题；议题",
				Example:     "That's a separate issue.",
				ExampleNote: "那是另一个问题。",
			},
		},
		{
			name: "多余的空白要清掉",
			line: "term \t /tɜːrm/ \t n. 术语 \t Let's agree on terms. \t 我们先把说法统一。 ",
			want: WordEntry{
				Word:        "term",
				Phonetic:    "/tɜːrm/",
				Meaning:     "n. 术语",
				Example:     "Let's agree on terms.",
				ExampleNote: "我们先把说法统一。",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseWordEntry(tc.line); got != tc.want {
				t.Errorf("ParseWordEntry(%q)\n got = %+v\nwant = %+v", tc.line, got, tc.want)
			}
		})
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
