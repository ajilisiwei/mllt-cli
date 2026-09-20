package srs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ajilisiwei/mllt-cli/internal/config"
	"github.com/ajilisiwei/mllt-cli/internal/practice"
)

var intervals = []time.Duration{
	0,
	5 * time.Minute,
	30 * time.Minute,
	12 * time.Hour,
	24 * time.Hour,
	48 * time.Hour,
	96 * time.Hour,
	7 * 24 * time.Hour,
	15 * 24 * time.Hour,
	30 * 24 * time.Hour,
}

// ItemState 表示单个练习项的记忆状态。
type ItemState struct {
	Stage int       `json:"stage"`
	DueAt time.Time `json:"due_at"`
}

// Schedule 表示某个资源文件的记忆计划。
type Schedule struct {
	Items        map[string]ItemState `json:"items"`
	filePath     string               `json:"-"`
	resourceType string               `json:"-"`
}

// Load 根据资源类型和文件名加载记忆计划，并确保所有条目存在。
func Load(resourceType, fileName string, items []string) (*Schedule, error) {
	baseDir := practice.GetUserDataDir()
	currentLanguage := config.AppConfig.CurrentLanguage

	safeFileName := sanitizeFileName(fileName)
	path := filepath.Join(baseDir, "srs", currentLanguage, resourceType)
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("创建SRS目录失败: %w", err)
	}

	path = filepath.Join(path, safeFileName+".json")

	schedule := &Schedule{
		Items:        make(map[string]ItemState),
		filePath:     path,
		resourceType: resourceType,
	}

	if data, err := os.ReadFile(path); err == nil {
		if len(data) > 0 {
			if err := json.Unmarshal(data, schedule); err != nil {
				return nil, fmt.Errorf("解析SRS文件失败: %w", err)
			}
		}
		schedule.filePath = path
		schedule.resourceType = resourceType
	}

	schedule.ensureItems(items)
	if err := schedule.Save(); err != nil {
		return nil, err
	}

	return schedule, nil
}

// Save 将记忆计划写回磁盘。
func (s *Schedule) Save() error {
	if s == nil || s.filePath == "" {
		return nil
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// Order 根据记忆计划返回条目的练习顺序（索引数组）。
func (s *Schedule) Order(items []string) []int {
	type entry struct {
		index int
		due   time.Time
		stage int
	}

	entries := make([]entry, 0, len(items))

	for idx, item := range items {
		state := s.getState(item)
		entries = append(entries, entry{
			index: idx,
			due:   state.DueAt,
			stage: state.Stage,
		})
	}

	now := time.Now()
	sort.Slice(entries, func(i, j int) bool {
		ai := entries[i]
		aj := entries[j]

		di := ai.due
		dj := aj.due

		if di.IsZero() && dj.IsZero() {
			if ai.stage == aj.stage {
				return ai.index < aj.index
			}
			return ai.stage < aj.stage
		}

		if di.IsZero() {
			return true
		}
		if dj.IsZero() {
			return false
		}

		dueI := di
		dueJ := dj
		if dueI.Before(now) && dueJ.Before(now) {
			if dueI.Equal(dueJ) {
				if ai.stage == aj.stage {
					return ai.index < aj.index
				}
				return ai.stage < aj.stage
			}
			return dueI.Before(dueJ)
		}

		if dueI.Before(now) {
			return true
		}
		if dueJ.Before(now) {
			return false
		}

		if dueI.Equal(dueJ) {
			if ai.stage == aj.stage {
				return ai.index < aj.index
			}
			return ai.stage < aj.stage
		}
		return dueI.Before(dueJ)
	})

	ordered := make([]int, 0, len(entries))
	for _, e := range entries {
		ordered = append(ordered, e.index)
	}
	return ordered
}

// DueItems 返回今天真正该练的条目下标：先是已经到期的复习项（按到期时间从早到晚），
// 再补上没练过的新项。
//
// 艾宾浩斯的价值在于每天只练该练的那部分——一个五千词的文件如果每次都全量铺开，
// 间隔重复就退化成了随机刷题。上限传 0 表示不限制。
func (s *Schedule) DueItems(items []string, newLimit, reviewLimit int) []int {
	type candidate struct {
		index int
		due   time.Time
	}

	var reviews []candidate
	var fresh []int

	now := time.Now()
	for idx, item := range items {
		state := s.getState(item)
		switch {
		case state.DueAt.IsZero():
			fresh = append(fresh, idx)
		case state.DueAt.After(now):
			// 还没到复习时间，今天不练
		default:
			reviews = append(reviews, candidate{index: idx, due: state.DueAt})
		}
	}

	sort.SliceStable(reviews, func(i, j int) bool {
		return reviews[i].due.Before(reviews[j].due)
	})

	if reviewLimit > 0 && len(reviews) > reviewLimit {
		reviews = reviews[:reviewLimit]
	}
	if newLimit > 0 && len(fresh) > newLimit {
		fresh = fresh[:newLimit]
	}

	ordered := make([]int, 0, len(reviews)+len(fresh))
	for _, r := range reviews {
		ordered = append(ordered, r.index)
	}
	return append(ordered, fresh...)
}

// NextDue 返回最近一个尚未到期条目的复习时间；没有这样的条目时返回零值。
// 用来在今天练完之后告诉用户下次该什么时候回来。
func (s *Schedule) NextDue(items []string) time.Time {
	var next time.Time

	now := time.Now()
	for _, item := range items {
		due := s.getState(item).DueAt
		if due.IsZero() || !due.After(now) {
			continue
		}
		if next.IsZero() || due.Before(next) {
			next = due
		}
	}

	return next
}

// RecordResult 根据练习结果更新记忆计划。
func (s *Schedule) RecordResult(item string, correct bool) error {
	if s == nil {
		return nil
	}

	state := s.getState(item)
	if correct {
		if state.Stage < len(intervals)-1 {
			state.Stage++
		}
	} else {
		state.Stage = 0
	}

	state.DueAt = time.Now().Add(intervals[state.Stage])
	s.setState(item, state)
	return s.Save()
}

// RemoveItem 从记忆计划中移除指定条目。
func (s *Schedule) RemoveItem(item string) error {
	if s == nil {
		return nil
	}

	key := s.keyFor(item)
	delete(s.Items, key)
	return s.Save()
}

func (s *Schedule) ensureItems(items []string) {
	if s.Items == nil {
		s.Items = make(map[string]ItemState)
	}
	for _, item := range items {
		key := s.keyFor(item)
		if _, exists := s.Items[key]; !exists {
			s.Items[key] = ItemState{Stage: 0}
		}
	}
}

func (s *Schedule) keyFor(item string) string {
	// 单词条目以剥离音标后的单词为键，避免同一个词因脏数据产生两条记忆记录
	if s.resourceType == practice.Words {
		if key := strings.TrimSpace(practice.WordPrimaryText(item)); key != "" {
			return key
		}
		return strings.TrimSpace(item)
	}

	primary, _ := practice.ParseLine(item)
	key := strings.TrimSpace(primary)
	if key == "" {
		key = strings.TrimSpace(item)
	}
	return key
}

func (s *Schedule) getState(item string) ItemState {
	key := s.keyFor(item)
	if state, ok := s.Items[key]; ok {
		return state
	}
	return ItemState{Stage: 0}
}

func (s *Schedule) setState(item string, state ItemState) {
	key := s.keyFor(item)
	if s.Items == nil {
		s.Items = make(map[string]ItemState)
	}
	s.Items[key] = state
}

func sanitizeFileName(name string) string {
	sanitized := strings.TrimSuffix(name, ".txt")
	sanitized = strings.TrimSpace(sanitized)
	sanitized = strings.ReplaceAll(sanitized, string(os.PathSeparator), "_")
	if sanitized == "" {
		sanitized = "default"
	}
	return sanitized
}
