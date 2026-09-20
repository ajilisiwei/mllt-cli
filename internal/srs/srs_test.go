package srs

import (
	"testing"
	"time"
)

func newSchedule(states map[string]ItemState) *Schedule {
	s := &Schedule{Items: map[string]ItemState{}}
	for k, v := range states {
		s.Items[k] = v
	}
	return s
}

func TestDueItemsPrioritizesReviewsThenNew(t *testing.T) {
	now := time.Now()
	items := []string{"alpha", "beta", "gamma", "delta", "epsilon"}

	schedule := newSchedule(map[string]ItemState{
		"alpha": {Stage: 3, DueAt: now.Add(-2 * time.Hour)}, // 已到期，最早
		"beta":  {Stage: 2, DueAt: now.Add(24 * time.Hour)}, // 未到期，今天不该练
		"gamma": {Stage: 1, DueAt: now.Add(-30 * time.Minute)},
		// delta、epsilon 没有记录，属于新项
	})

	got := schedule.DueItems(items, 1, 0)

	// alpha 和 gamma 到期（alpha 更早），再补 1 个新项
	want := []int{0, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("DueItems = %v，期望 %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 项 = %d，期望 %d", i, got[i], want[i])
		}
	}
}

func TestDueItemsSkipsItemsNotYetDue(t *testing.T) {
	now := time.Now()
	items := []string{"alpha", "beta"}
	schedule := newSchedule(map[string]ItemState{
		"alpha": {Stage: 4, DueAt: now.Add(48 * time.Hour)},
		"beta":  {Stage: 4, DueAt: now.Add(72 * time.Hour)},
	})

	if got := schedule.DueItems(items, 10, 0); len(got) != 0 {
		t.Fatalf("没有到期内容时应返回空，实际 %v", got)
	}
}

func TestDueItemsRespectsNewLimit(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	schedule := newSchedule(nil)

	if got := schedule.DueItems(items, 2, 0); len(got) != 2 {
		t.Fatalf("新项上限 2，实际取了 %d 项: %v", len(got), got)
	}
	if got := schedule.DueItems(items, 0, 0); len(got) != len(items) {
		t.Fatalf("上限为 0 表示不限，实际取了 %d 项", len(got))
	}
}

func TestDueItemsRespectsReviewLimit(t *testing.T) {
	now := time.Now()
	items := []string{"a", "b", "c"}
	schedule := newSchedule(map[string]ItemState{
		"a": {Stage: 1, DueAt: now.Add(-3 * time.Hour)},
		"b": {Stage: 1, DueAt: now.Add(-2 * time.Hour)},
		"c": {Stage: 1, DueAt: now.Add(-1 * time.Hour)},
	})

	got := schedule.DueItems(items, 0, 2)
	if len(got) != 2 {
		t.Fatalf("复习上限 2，实际 %d 项: %v", len(got), got)
	}
	// 上限生效时保留最早到期的
	if got[0] != 0 || got[1] != 1 {
		t.Errorf("应保留最早到期的两项，实际 %v", got)
	}
}

func TestRecordResultResetsStageOnFailure(t *testing.T) {
	schedule := newSchedule(map[string]ItemState{"a": {Stage: 5}})

	_ = schedule.RecordResult("a", false)
	if got := schedule.getState("a").Stage; got != 0 {
		t.Errorf("答错后阶段 = %d，期望回到 0", got)
	}

	_ = schedule.RecordResult("a", true)
	if got := schedule.getState("a").Stage; got != 1 {
		t.Errorf("答对后阶段 = %d，期望 1", got)
	}
}
