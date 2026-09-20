package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ajilisiwei/mllt-cli/internal/config"
)

// PracticeDirectionMenu 练习方向设置菜单
type PracticeDirectionMenu struct {
	list     list.Model
	quitting bool
}

// PracticeDirectionMenuItem 练习方向设置菜单项
type PracticeDirectionMenuItem struct {
	direction   string
	title       string
	description string
	isCurrent   bool
}

// 实现 list.Item 接口
func (i PracticeDirectionMenuItem) Title() string {
	if i.isCurrent {
		return "✔ " + i.title
	}
	return i.title
}
func (i PracticeDirectionMenuItem) Description() string { return i.description }
func (i PracticeDirectionMenuItem) FilterValue() string { return i.title }

// NewPracticeDirectionMenu 创建练习方向设置菜单
func NewPracticeDirectionMenu() *PracticeDirectionMenu {
	current := config.AppConfig.PracticeDirection
	if current == "" {
		current = practiceDirectionCopy
	}

	items := []list.Item{
		PracticeDirectionMenuItem{
			direction:   practiceDirectionCopy,
			title:       "抄写：看英文打英文",
			description: "屏幕给出英文原文，练手感与拼写",
			isCurrent:   current == practiceDirectionCopy,
		},
		PracticeDirectionMenuItem{
			direction:   practiceDirectionTranslate,
			title:       "译写：看中文写英文",
			description: "只给中文（单词给音标加释义），练真正的产出能力",
			isCurrent:   current == practiceDirectionTranslate,
		},
		MenuItem{
			title:       "返回设置菜单",
			description: "返回到设置菜单",
			action:      func() (tea.Model, error) { return NewSettingMenu(), nil },
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "练习方向设置"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &PracticeDirectionMenu{list: l}
}

// Init 初始化模型
func (m PracticeDirectionMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m PracticeDirectionMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			case PracticeDirectionMenuItem:
				config.AppConfig.PracticeDirection = i.direction
				if err := config.SaveConfig(); err != nil {
					return m, nil
				}
				return m.transitionTo(NewPracticeDirectionMenu())
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
func (m PracticeDirectionMenu) transitionTo(next tea.Model) (tea.Model, tea.Cmd) {
	width, height := m.list.Width(), m.list.Height()+4
	if width > 0 && height > 4 {
		updated, _ := next.Update(tea.WindowSizeMsg{Width: width, Height: height})
		return updated, nil
	}
	return next, nil
}

// View 渲染视图
func (m PracticeDirectionMenu) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}
