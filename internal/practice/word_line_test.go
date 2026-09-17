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
