package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/ajilisiwei/mllt-cli/internal/config"
)

// SettingMenuItem 设置菜单项
type SettingMenuItem struct {
	title       string
	description string
	action      func() (tea.Model, error)
}

// 实现list.Item接口
func (i SettingMenuItem) Title() string       { return i.title }
func (i SettingMenuItem) Description() string { return i.description }
func (i SettingMenuItem) FilterValue() string { return i.title }

// SettingMenu 设置菜单模型
type SettingMenu struct {
	list     list.Model
	quitting bool
}

// 创建新的设置菜单
func NewSettingMenu() *SettingMenu {
	// 创建菜单项
	items := []list.Item{
		SettingMenuItem{
			title:       "匹配模式设置",
			description: "设置正确性匹配模式（完全匹配/单词匹配）",
			action: func() (tea.Model, error) {
				return NewMatchModeMenu(), nil
			},
		},
		SettingMenuItem{
			title:       "顺序设置",
			description: "设置单词、短语、句子练习的出现顺序（不影响文章练习）",
			action: func() (tea.Model, error) {
				return NewOrderMenu(), nil
			},
		},
		SettingMenuItem{
			title:       "键盘声音设置",
			description: "设置是否启用键盘按键声音",
			action: func() (tea.Model, error) {
				return NewKeyboardSoundMenu(), nil
			},
		},
		SettingMenuItem{
			title:       "翻译显示设置",
			description: "设置是否显示翻译内容",
			action: func() (tea.Model, error) {
				return NewShowTranslationMenu(), nil
			},
		},
		SettingMenuItem{
			title:       "例句练习设置",
			description: "设置短语的例句是否也作为练习内容",
			action: func() (tea.Model, error) {
				return NewPracticeExamplesMenu(), nil
			},
		},
		MenuItem{
			title:       "返回主菜单",
			description: "返回到主菜单",
			action:      func() (tea.Model, error) { return NewMainMenu(), nil },
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "设置"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &SettingMenu{
		list: l,
	}
}

// Init 初始化模型
func (m SettingMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m SettingMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // 减去标题和状态栏的高度
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			mainMenu := NewMainMenu()
			width, height := m.list.Width(), m.list.Height()+4
			if width > 0 && height > 4 {
				updatedModel, _ := mainMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
				return updatedModel, nil
			}
			return mainMenu, nil

		case "enter":
			switch i := m.list.SelectedItem().(type) {
			case SettingMenuItem:
				if i.action != nil {
					newModel, err := i.action()
					if err != nil {
						// 处理错误
						return m, nil
					}
					// 传递当前窗口大小
					width, height := m.list.Width(), m.list.Height()+4
					if width > 0 && height > 4 {
						updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
						return updatedModel, nil
					}
					return newModel, nil
				}
			case MenuItem:
				if i.action != nil {
					newModel, err := i.action()
					if err != nil {
						// 处理错误
						return m, nil
					}
					// 如果返回的是主菜单，需要传递当前窗口大小
					if mainMenu, ok := newModel.(*MainMenu); ok {
						// 获取当前窗口大小并设置
						width, height := m.list.Width(), m.list.Height()+4
						if width > 0 && height > 4 {
							updatedModel, _ := mainMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
							return updatedModel, nil
						}
						return mainMenu, nil
					}
					return newModel, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View 渲染视图
func (m SettingMenu) View() string {
	if m.quitting {
		return "再见！"
	}

	return m.list.View()
}

// MatchModeMenu 匹配模式菜单模型
type MatchModeMenu struct {
	list     list.Model
	quitting bool
}

// MatchModeMenuItem 匹配模式菜单项
type MatchModeMenuItem struct {
	mode        string
	title       string
	description string
	isCurrent   bool
}

// 实现list.Item接口
func (i MatchModeMenuItem) Title() string {
	title := i.title
	if i.isCurrent {
		title = "✔ " + title
	}
	return title
}
func (i MatchModeMenuItem) Description() string { return i.description }
func (i MatchModeMenuItem) FilterValue() string { return i.title }

// 创建新的匹配模式菜单
func NewMatchModeMenu() *MatchModeMenu {
	currentMode := config.AppConfig.CorrectnessMatchMode
	if currentMode == "" {
		currentMode = "exact_match" // 默认值
	}

	// 创建菜单项
	items := []list.Item{
		MatchModeMenuItem{
			mode:        "exact_match",
			title:       "完全匹配",
			description: "用户输入必须完全匹配才能被认为是正确的",
			isCurrent:   currentMode == "exact_match",
		},
		MatchModeMenuItem{
			mode:        "word_match",
			title:       "单词匹配",
			description: "只要包含单词就被认为是正确的，忽略大小写和标点符号",
			isCurrent:   currentMode == "word_match",
		},
		MenuItem{
			title:       "返回设置菜单",
			description: "返回到设置菜单",
			action:      func() (tea.Model, error) { return NewSettingMenu(), nil },
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "匹配模式设置"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &MatchModeMenu{
		list: l,
	}
}

// Init 初始化模型
func (m MatchModeMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m MatchModeMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			settingMenu := NewSettingMenu()
			width, height := m.list.Width(), m.list.Height()+4
			if width > 0 && height > 4 {
				updatedModel, _ := settingMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
				return updatedModel, nil
			}
			return settingMenu, nil

		case "enter":
			switch i := m.list.SelectedItem().(type) {
			case MatchModeMenuItem:
				// 设置匹配模式
				config.AppConfig.CorrectnessMatchMode = i.mode
				config.SaveConfig()
				// 刷新菜单
				newModel := NewMatchModeMenu()
				// 传递当前窗口大小
				width, height := m.list.Width(), m.list.Height()+4
				if width > 0 && height > 4 {
					updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
					return updatedModel, nil
				}
				return newModel, nil
			case MenuItem:
				if i.action != nil {
					newModel, err := i.action()
					if err != nil {
						return m, nil
					}
					// 传递当前窗口大小
					width, height := m.list.Width(), m.list.Height()+4
					if width > 0 && height > 4 {
						updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
						return updatedModel, nil
					}
					return newModel, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View 渲染视图
func (m MatchModeMenu) View() string {
	if m.quitting {
		return "再见！"
	}

	return m.list.View()
}

// OrderMenu 顺序菜单模型
type OrderMenu struct {
	list     list.Model
	quitting bool
}

// OrderMenuItem 顺序菜单项
type OrderMenuItem struct {
	order       string
	title       string
	description string
	isCurrent   bool
}

// 实现list.Item接口
func (i OrderMenuItem) Title() string {
	title := i.title
	if i.isCurrent {
		title = "✔ " + title
	}
	return title
}
func (i OrderMenuItem) Description() string { return i.description }
func (i OrderMenuItem) FilterValue() string { return i.title }

// 创建新的顺序菜单
func NewOrderMenu() *OrderMenu {
	currentOrder := config.AppConfig.NextOneOrder
	if currentOrder == "" {
		currentOrder = "random" // 默认值
	}

	// 创建菜单项
	items := []list.Item{
		OrderMenuItem{
			order:       "random",
			title:       "随机顺序",
			description: "单词、短语、句子练习时资源的出现顺序是随机的（不影响文章练习）",
			isCurrent:   currentOrder == "random",
		},
		OrderMenuItem{
			order:       "sequential",
			title:       "顺序出现",
			description: "单词、短语、句子练习时资源的出现顺序是顺序的（不影响文章练习）",
			isCurrent:   currentOrder == "sequential",
		},
		OrderMenuItem{
			order:       "ebbinghaus",
			title:       "艾宾浩斯遗忘曲线",
			description: "按照艾宾浩斯遗忘曲线安排练习顺序（不影响文章练习）",
			isCurrent:   currentOrder == "ebbinghaus",
		},
		MenuItem{
			title:       "返回设置菜单",
			description: "返回到设置菜单",
			action:      func() (tea.Model, error) { return NewSettingMenu(), nil },
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "顺序设置"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &OrderMenu{
		list: l,
	}
}

// Init 初始化模型
func (m OrderMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m OrderMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			settingMenu := NewSettingMenu()
			width, height := m.list.Width(), m.list.Height()+4
			if width > 0 && height > 4 {
				updatedModel, _ := settingMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
				return updatedModel, nil
			}
			return settingMenu, nil

		case "enter":
			switch i := m.list.SelectedItem().(type) {
			case OrderMenuItem:
				// 设置顺序
				config.AppConfig.NextOneOrder = i.order
				config.SaveConfig()
				// 刷新菜单
				newModel := NewOrderMenu()
				// 传递当前窗口大小
				width, height := m.list.Width(), m.list.Height()+4
				if width > 0 && height > 4 {
					updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
					return updatedModel, nil
				}
				return newModel, nil
			case MenuItem:
				if i.action != nil {
					newModel, err := i.action()
					if err != nil {
						return m, nil
					}
					// 传递当前窗口大小
					width, height := m.list.Width(), m.list.Height()+4
					if width > 0 && height > 4 {
						updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
						return updatedModel, nil
					}
					return newModel, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View 渲染视图
func (m OrderMenu) View() string {
	if m.quitting {
		return ""
	}

	return m.list.View()
}

// KeyboardSoundMenu 键盘声音设置菜单
type KeyboardSoundMenu struct {
	list     list.Model
	quitting bool
}

// KeyboardSoundMenuItem 键盘声音设置菜单项
type KeyboardSoundMenuItem struct {
	enabled     bool
	title       string
	description string
	isCurrent   bool
}

// 实现list.Item接口
func (i KeyboardSoundMenuItem) Title() string {
	title := i.title
	if i.isCurrent {
		title = "✔ " + title
	}
	return title
}
func (i KeyboardSoundMenuItem) Description() string { return i.description }
func (i KeyboardSoundMenuItem) FilterValue() string { return i.title }

// 创建新的键盘声音设置菜单
func NewKeyboardSoundMenu() *KeyboardSoundMenu {
	currentEnabled := config.AppConfig.InputKeyboardSound

	// 创建菜单项
	items := []list.Item{
		KeyboardSoundMenuItem{
			enabled:     true,
			title:       "启用键盘声音",
			description: "输入时播放键盘按键声音",
			isCurrent:   currentEnabled,
		},
		KeyboardSoundMenuItem{
			enabled:     false,
			title:       "禁用键盘声音",
			description: "输入时不播放键盘按键声音",
			isCurrent:   !currentEnabled,
		},
		MenuItem{
			title:       "返回设置菜单",
			description: "返回到设置菜单",
			action:      func() (tea.Model, error) { return NewSettingMenu(), nil },
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "键盘声音设置"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &KeyboardSoundMenu{
		list: l,
	}
}

// Init 初始化模型
func (m KeyboardSoundMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m KeyboardSoundMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			settingMenu := NewSettingMenu()
			width, height := m.list.Width(), m.list.Height()+4
			if width > 0 && height > 4 {
				updatedModel, _ := settingMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
				return updatedModel, nil
			}
			return settingMenu, nil

		case "enter":
			switch i := m.list.SelectedItem().(type) {
			case KeyboardSoundMenuItem:
				// 更新配置
				config.AppConfig.InputKeyboardSound = i.enabled
				config.SaveConfig()
				// 刷新菜单
				newModel := NewKeyboardSoundMenu()
				// 传递当前窗口大小
				width, height := m.list.Width(), m.list.Height()+4
				if width > 0 && height > 4 {
					updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
					return updatedModel, nil
				}
				return newModel, nil
			case MenuItem:
				if i.action != nil {
					newModel, err := i.action()
					if err != nil {
						return m, nil
					}
					// 传递当前窗口大小
					width, height := m.list.Width(), m.list.Height()+4
					if width > 0 && height > 4 {
						updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
						return updatedModel, nil
					}
					return newModel, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View 渲染视图
func (m KeyboardSoundMenu) View() string {
	if m.quitting {
		return ""
	}

	return m.list.View()
}

// ShowTranslationMenu 显示翻译设置菜单
type ShowTranslationMenu struct {
	list     list.Model
	quitting bool
}

// ShowTranslationMenuItem 显示翻译设置菜单项
type ShowTranslationMenuItem struct {
	enabled     bool
	title       string
	description string
	isCurrent   bool
}

// 实现list.Item接口
func (i ShowTranslationMenuItem) Title() string {
	title := i.title
	if i.isCurrent {
		title = "✔ " + title
	}
	return title
}
func (i ShowTranslationMenuItem) Description() string { return i.description }
func (i ShowTranslationMenuItem) FilterValue() string { return i.title }

// 创建新的显示翻译设置菜单
func NewShowTranslationMenu() *ShowTranslationMenu {
	currentEnabled := config.AppConfig.ShowTranslation

	// 创建菜单项
	items := []list.Item{
		ShowTranslationMenuItem{
			enabled:     true,
			title:       "显示翻译",
			description: "练习时显示翻译内容",
			isCurrent:   currentEnabled,
		},
		ShowTranslationMenuItem{
			enabled:     false,
			title:       "隐藏翻译",
			description: "练习时不显示翻译内容",
			isCurrent:   !currentEnabled,
		},
		MenuItem{
			title:       "返回设置菜单",
			description: "返回到设置菜单",
			action:      func() (tea.Model, error) { return NewSettingMenu(), nil },
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "翻译显示设置"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &ShowTranslationMenu{
		list: l,
	}
}

// Init 初始化模型
func (m ShowTranslationMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m ShowTranslationMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			settingMenu := NewSettingMenu()
			width, height := m.list.Width(), m.list.Height()+4
			if width > 0 && height > 4 {
				updatedModel, _ := settingMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
				return updatedModel, nil
			}
			return settingMenu, nil

		case "enter":
			switch i := m.list.SelectedItem().(type) {
			case ShowTranslationMenuItem:
				// 更新配置
				config.AppConfig.ShowTranslation = i.enabled
				config.SaveConfig()
				// 刷新菜单
				newModel := NewShowTranslationMenu()
				// 传递当前窗口大小
				width, height := m.list.Width(), m.list.Height()+4
				if width > 0 && height > 4 {
					updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
					return updatedModel, nil
				}
				return newModel, nil
			case MenuItem:
				if i.action != nil {
					newModel, err := i.action()
					if err != nil {
						return m, nil
					}
					// 传递当前窗口大小
					width, height := m.list.Width(), m.list.Height()+4
					if width > 0 && height > 4 {
						updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
						return updatedModel, nil
					}
					return newModel, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View 渲染视图
func (m ShowTranslationMenu) View() string {
	if m.quitting {
		return ""
	}

	return m.list.View()
}
