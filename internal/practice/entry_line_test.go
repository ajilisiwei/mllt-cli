package practice

import "testing"

func TestParseEntryLine(t *testing.T) {
	cases := []struct {
		name string
		line string
		want Entry
	}{
		{
			name: "四段：正文、释义、例句、例句翻译",
			line: "look into ->> 调查；了解一下 ->> I'll look into it and get back to you. ->> 我查一下再回复你。",
			want: Entry{
				Text:        "look into",
				Meaning:     "调查；了解一下",
				Example:     "I'll look into it and get back to you.",
				ExampleNote: "我查一下再回复你。",
			},
		},
		{
			name: "三段：缺例句翻译",
			line: "hang up ->> 挂断电话 ->> Don't hang up on me.",
			want: Entry{
				Text:    "hang up",
				Meaning: "挂断电话",
				Example: "Don't hang up on me.",
			},
		},
		{
			name: "两段：兼容旧格式",
			line: "come up with ->> 想出（主意、答案等）",
			want: Entry{Text: "come up with", Meaning: "想出（主意、答案等）"},
		},
		{
			name: "整句条目",
			line: "I'll check and let you know. ->> 我查一下再告知你。",
			want: Entry{Text: "I'll check and let you know.", Meaning: "我查一下再告知你。"},
		},
		{
			name: "无分隔符",
			line: "Are you kidding me?",
			want: Entry{Text: "Are you kidding me?"},
		},
		{
			name: "制表符分隔也能识别正文",
			line: "take off\t起飞；脱下",
			want: Entry{Text: "take off", Meaning: "起飞；脱下"},
		},
		{
			name: "多余空白会被清理",
			line: "set up  ->>  安排；搭建  ->>  Let's set up a call.  ->>  我们约个电话吧。 ",
			want: Entry{
				Text:        "set up",
				Meaning:     "安排；搭建",
				Example:     "Let's set up a call.",
				ExampleNote: "我们约个电话吧。",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseEntryLine(tc.line)
			if got != tc.want {
				t.Errorf("ParseEntryLine(%q)\n got = %+v\nwant = %+v", tc.line, got, tc.want)
			}
		})
	}
}

// 新解析器不能改变打字答案，否则会影响已有的练习进度
func TestParseEntryLineKeepsExpectedInput(t *testing.T) {
	line := "look into ->> 调查 ->> I'll look into it. ->> 我查一下。"

	primary, _ := ParseLine(line)
	if primary != "look into" {
		t.Fatalf("ParseLine 正文 = %q，期望 %q", primary, "look into")
	}
	if entry := ParseEntryLine(line); entry.Text != primary {
		t.Fatalf("ParseEntryLine 正文 = %q，与 ParseLine 不一致", entry.Text)
	}
}
