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
			if got.Text != tc.want.Text || got.Meaning != tc.want.Meaning ||
				got.Example != tc.want.Example || got.ExampleNote != tc.want.ExampleNote {
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

func TestExpandEntries(t *testing.T) {
	cases := []struct {
		name  string
		items []string
		want  []string
	}{
		{
			name:  "带例句的条目展开成短语和例句两项",
			items: []string{"sing along ->> 跟着唱 ->> Everyone was singing along. ->> 全场都跟着唱。"},
			want: []string{
				"sing along ->> 跟着唱 ->> Everyone was singing along. ->> 全场都跟着唱。",
				"Everyone was singing along. ->> 全场都跟着唱。",
			},
		},
		{
			name:  "例句没有翻译时只带正文",
			items: []string{"hang up ->> 挂断电话 ->> Don't hang up on me."},
			want: []string{
				"hang up ->> 挂断电话 ->> Don't hang up on me.",
				"Don't hang up on me.",
			},
		},
		{
			name:  "没有例句的条目保持不变",
			items: []string{"apple ->> 苹果", "I'll check and let you know. ->> 我查一下再告知你。"},
			want:  []string{"apple ->> 苹果", "I'll check and let you know. ->> 我查一下再告知你。"},
		},
		{
			name:  "空列表",
			items: nil,
			want:  []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExpandEntries(tc.items)
			if len(got) != len(tc.want) {
				t.Fatalf("展开后 %d 项，期望 %d 项: %q", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("第 %d 项 = %q, 期望 %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// 展开出来的例句项必须是合法条目：打字答案就是例句本身
func TestExpandedExampleIsPracticeable(t *testing.T) {
	items := ExpandEntries([]string{"sing along ->> 跟着唱 ->> Everyone was singing along. ->> 全场都跟着唱。"})

	example := items[1]
	entry := ParseEntryLine(example)
	if entry.Text != "Everyone was singing along." {
		t.Errorf("例句项正文 = %q", entry.Text)
	}
	if entry.Meaning != "全场都跟着唱。" {
		t.Errorf("例句项翻译 = %q", entry.Meaning)
	}
	if primary, _ := ParseLine(example); primary != entry.Text {
		t.Errorf("ParseLine 与 ParseEntryLine 不一致: %q vs %q", primary, entry.Text)
	}
}

func TestExpandEntryBlocksKeepsExampleWithItsPhrase(t *testing.T) {
	items := []string{
		"be tone-deaf ->> 五音不全 ->> I'm completely tone-deaf. ->> 我完全五音不全。",
		"apple ->> 苹果",
		"go viral ->> 爆火 ->> That clip went viral overnight. ->> 那段视频一夜爆火。",
	}

	blocks := ExpandEntryBlocks(items)
	if len(blocks) != len(items) {
		t.Fatalf("%d 个块，期望 %d 个", len(blocks), len(items))
	}

	wantSizes := []int{2, 1, 2}
	for i, want := range wantSizes {
		if len(blocks[i]) != want {
			t.Errorf("第 %d 个块有 %d 项，期望 %d 项", i, len(blocks[i]), want)
		}
		if blocks[i][0] != items[i] {
			t.Errorf("第 %d 个块的首项应是原条目，实际 %q", i, blocks[i][0])
		}
	}

	if blocks[0][1] != "I'm completely tone-deaf. ->> 我完全五音不全。" {
		t.Errorf("例句项 = %q", blocks[0][1])
	}

	// 无论条目顺序怎么打乱，例句都必须紧跟它的短语
	for _, order := range [][]int{{2, 0, 1}, {1, 2, 0}} {
		var flat []string
		for _, idx := range order {
			flat = append(flat, blocks[idx]...)
		}
		for i, line := range flat {
			entry := ParseEntryLine(line)
			if entry.Example == "" {
				continue
			}
			if i+1 >= len(flat) {
				t.Fatalf("顺序 %v：%q 的例句缺失", order, entry.Text)
			}
			if next, _ := ParseLine(flat[i+1]); next != entry.Example {
				t.Errorf("顺序 %v：%q 的下一项是 %q，期望它的例句 %q", order, entry.Text, next, entry.Example)
			}
		}
	}
}

func TestParseEntryLineCollectsAllContrasts(t *testing.T) {
	line := "He lives here. ->> 他住在这儿（一般现在时：长期状态）" +
		" ->> He is living here. ->> 他目前住在这儿（现在进行时：临时安排）" +
		" ->> He has lived here for ten years. ->> 他在这儿住了十年了（现在完成时：持续到现在）" +
		" ->> He lived here for ten years. ->> 他在这儿住过十年（一般过去时：现在已搬走）"

	entry := ParseEntryLine(line)

	if entry.Text != "He lives here." {
		t.Errorf("正文 = %q", entry.Text)
	}
	if entry.Meaning != "他住在这儿（一般现在时：长期状态）" {
		t.Errorf("注释 = %q", entry.Meaning)
	}
	if len(entry.Contrasts) != 3 {
		t.Fatalf("对比句数量 = %d，期望 3", len(entry.Contrasts))
	}

	want := []Contrast{
		{Text: "He is living here.", Note: "他目前住在这儿（现在进行时：临时安排）"},
		{Text: "He has lived here for ten years.", Note: "他在这儿住了十年了（现在完成时：持续到现在）"},
		{Text: "He lived here for ten years.", Note: "他在这儿住过十年（一般过去时：现在已搬走）"},
	}
	for i, w := range want {
		if entry.Contrasts[i] != w {
			t.Errorf("第 %d 个对比 = %+v，期望 %+v", i+1, entry.Contrasts[i], w)
		}
	}

	// 兼容字段仍指向第一个对比句
	if entry.Example != want[0].Text || entry.ExampleNote != want[0].Note {
		t.Errorf("兼容字段 = (%q, %q)", entry.Example, entry.ExampleNote)
	}
}

// 整套时态必须在同一个练习块里，否则对比就散了
func TestExpandEntryBlocksKeepsWholeParadigmTogether(t *testing.T) {
	line := "He lives here. ->> 一般现在时" +
		" ->> He is living here. ->> 现在进行时" +
		" ->> He has lived here. ->> 现在完成时"

	blocks := ExpandEntryBlocks([]string{line})
	if len(blocks) != 1 {
		t.Fatalf("应该只有一个块，实际 %d 个", len(blocks))
	}

	block := blocks[0]
	if len(block) != 3 {
		t.Fatalf("块内 %d 项，期望 3 项: %q", len(block), block)
	}

	wantTexts := []string{"He lives here.", "He is living here.", "He has lived here."}
	for i, want := range wantTexts {
		if got, _ := ParseLine(block[i]); got != want {
			t.Errorf("块内第 %d 项 = %q，期望 %q", i, got, want)
		}
	}
}

// 奇数段（最后一句没有注释）不能丢内容
func TestParseEntryLineHandlesMissingLastNote(t *testing.T) {
	entry := ParseEntryLine("hang up ->> 挂断电话 ->> Don't hang up on me.")

	if len(entry.Contrasts) != 1 {
		t.Fatalf("对比句数量 = %d，期望 1", len(entry.Contrasts))
	}
	if entry.Contrasts[0].Text != "Don't hang up on me." || entry.Contrasts[0].Note != "" {
		t.Errorf("对比 = %+v", entry.Contrasts[0])
	}
}
