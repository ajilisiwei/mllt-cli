package main

import "testing"

func TestCleanMeaning(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "补上被吞掉的词性分隔",
			in:   "v. 遗弃；离开；放弃；终止；陷入n. 放任，狂热",
			want: "v. 遗弃；离开；放弃 n. 放任，狂热",
		},
		{
			name: "整组人名义项直接丢弃",
			in:   "adj. 能；有能力的；能干的n. (Able)人名；(伊朗)阿布勒；(英)埃布尔",
			want: "adj. 能；有能力的；能干的",
		},
		{
			name: "全角括号的人名写法同样丢弃",
			in:   "n. 行动；活动；功能；战斗n. （Action）（英）埃克申（人名）",
			want: "n. 行动；活动；功能",
		},
		{
			name: "连续两个词性标记要都切开",
			in:   "adv.adj.向后地(的),相反地(的),追溯",
			want: "adv. adj. 向后地(的),相反地(的),追溯",
		},
		{
			name: "每个词性最多保留三个义项",
			in:   "n. 口音；重音；强调；特点；重音符号",
			want: "n. 口音；重音；强调",
		},
		{
			name: "最多保留三个词性",
			in:   "adj. 纯理论的n. 摘要v. 提取vi. 抽象化",
			want: "adj. 纯理论的 n. 摘要 v. 提取",
		},
		{
			name: "分号后跟词性时不产生多余空格",
			in:   "n. 程序； v. 编程",
			want: "n. 程序 v. 编程",
		},
		{
			name: "没有词性标记的释义原样保留",
			in:   "容器即服务（CaaS）",
			want: "容器即服务（CaaS）",
		},
		{
			name: "已经规范的释义不做改动",
			in:   "n. 能力，能耐；才能",
			want: "n. 能力，能耐；才能",
		},
		{
			name: "只剩人名时保留原文，宁可留噪音也不丢内容",
			in:   "n. (Ant)人名；(土、芬)安特",
			want: "n. (Ant)人名；(土、芬)安特",
		},
		{
			name: "释义里的英文单词不能被当成词性标记",
			in:   "abbr. 优势 (advantage)n. 广告；促销活动",
			want: "abbr. 优势 (advantage) n. 广告；促销活动",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cleanMeaning(tc.in); got != tc.want {
				t.Errorf("cleanMeaning(%q)\n got = %q\nwant = %q", tc.in, got, tc.want)
			}
		})
	}
}

// 清洗必须幂等，否则反复运行会越洗越少
func TestCleanMeaningIsIdempotent(t *testing.T) {
	inputs := []string{
		"v. 遗弃；离开；放弃；终止；陷入n. 放任，狂热",
		"adj. 能；有能力的n. (Able)人名；(伊朗)阿布勒",
		"adv.adj.向后地(的),相反地(的)",
		"n. 程序； v. 编程",
	}

	for _, in := range inputs {
		once := cleanMeaning(in)
		if twice := cleanMeaning(once); twice != once {
			t.Errorf("不幂等:\n输入 %q\n一次 %q\n两次 %q", in, once, twice)
		}
	}
}
