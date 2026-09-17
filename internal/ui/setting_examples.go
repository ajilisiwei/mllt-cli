package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ajilisiwei/mllt-cli/internal/config"
)

// PracticeExamplesMenu 例句练习设置菜单
type PracticeExamplesMenu struct {
	list     list.Model
	quitting bool
}

// PracticeExamplesMenuItem 例句练习设置菜单项
type PracticeExamplesMenuItem struct {
	enabled     bool
	title       string
	description string
	isCurrent   bool
}

// 实现 list.Item 接口
func (i PracticeExamplesMenuItem) Title() string {
	if i.isCurrent {
		return "✔ " + i.title
	}
	return i.title
}
func (i PracticeExamplesMenuItem) Description() string { return i.description }
func (i PracticeExamplesMenuItem) FilterValue() string { return i.title }

// NewPracticeExamplesMenu 创建例句练习设置菜单
func NewPracticeExamplesMenu() *PracticeExamplesMenu {
	currentEnabled := config.AppConfig.PracticeExamples

	items := []list.Item{
		PracticeExamplesMenuItem{
			enabled:     true,
			title:       "例句也要练",
			description: "先打短语，再打用到它的整句",
			isCurrent:   currentEnabled,
		},
		PracticeExamplesMenuItem{
			enabled:     false,
			title:       "只练短语",
			description: "例句仅作展示，不进入练习",
			isCurrent:   !currentEnabled,
		},
		MenuItem{
			title:       "返回设置菜单",
			description: "返回到设置菜单",
			action:      func() (tea.Model, error) { return NewSettingMenu(), nil },
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "例句练习设置"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &PracticeExamplesMenu{list: l}
}

// Init 初始化模型
func (m PracticeExamplesMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m PracticeExamplesMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			return m.transitionTo(NewSettingMenu())

		case "enter":
			switch i := m.list.SelectedItem().(type) {
			case PracticeExamplesMenuItem:
				config.AppConfig.PracticeExamples = i.enabled
				if err := config.SaveConfig(); err != nil {
					return m, nil
				}
				return m.transitionTo(NewPracticeExamplesMenu())
			case MenuItem:
				if i.action == nil {
					return m, nil
				}
				next, err := i.action()
				if err != nil {
					return m, nil
				}
				return m.transitionTo(next)
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// transitionTo 切换到另一个模型，并把当前窗口大小传递过去
func (m PracticeExamplesMenu) transitionTo(next tea.Model) (tea.Model, tea.Cmd) {
	width, height := m.list.Width(), m.list.Height()+4
	if width > 0 && height > 4 {
		updated, _ := next.Update(tea.WindowSizeMsg{Width: width, Height: height})
		return updated, nil
	}
	return next, nil
}

// View 渲染视图
func (m PracticeExamplesMenu) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}
