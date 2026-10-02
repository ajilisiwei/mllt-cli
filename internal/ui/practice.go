package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/ajilisiwei/mllt-cli/internal/practice"
)

// PracticeMenu 练习菜单模型
type PracticeMenu struct {
	list     list.Model
	quitting bool
}

// 创建新的练习菜单
func NewPracticeMenu() *PracticeMenu {
	items := []list.Item{
		MenuItem{
			title:       "单词练习",
			description: "进行单词打字练习",
			action:      func() (tea.Model, error) { return NewResourceSelectionMenu(practice.Words), nil },
		},
		MenuItem{
			title:       "短语练习",
			description: "进行短语打字练习",
			action:      func() (tea.Model, error) { return NewResourceSelectionMenu(practice.Phrases), nil },
		},
		MenuItem{
			title:       "句子练习",
			description: "进行句子打字练习",
			action:      func() (tea.Model, error) { return NewResourceSelectionMenu(practice.Sentences), nil },
		},
		MenuItem{
			title:       "对话练习",
			description: "进行成段对话的打字练习",
			action:      func() (tea.Model, error) { return NewResourceSelectionMenu(practice.Dialogues), nil },
		},
		MenuItem{
			title:       "文章练习",
			description: "进行文章打字练习",
			action:      func() (tea.Model, error) { return NewResourceSelectionMenu(practice.Articles), nil },
		},
		MenuItem{
			title:       "返回主菜单",
			description: "返回到主菜单",
			action:      func() (tea.Model, error) { return NewMainMenu(), nil },
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "练习菜单"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &PracticeMenu{
		list: l,
	}
}

// Init 初始化模型
func (m PracticeMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m PracticeMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			i, ok := m.list.SelectedItem().(MenuItem)
			if ok && i.action != nil {
				newModel, err := i.action()
				if err != nil {
					// 处理错误
					return m, nil
				}
				// 传递当前窗口大小给新模型
				width, height := m.list.Width(), m.list.Height()+4
				if width > 0 && height > 4 {
					updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
					return updatedModel, nil
				}
				return newModel, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View 渲染视图
func (m PracticeMenu) View() string {
	if m.quitting {
		return "再见！"
	}

	return m.list.View()
}

// ResourceSelectionMenu 资源选择菜单
type ResourceSelectionMenu struct {
	list         list.Model
	resourceType string
	folders      []practice.ResourceFolder
	quitting     bool
}

// 创建新的资源选择菜单
func NewResourceSelectionMenu(resourceType string) *ResourceSelectionMenu {
	folders, err := practice.GetResourceFolders(resourceType)
	if err != nil {
		folders = []practice.ResourceFolder{}
	}

	items := []list.Item{}
	for _, folder := range folders {
		folderCopy := folder
		items = append(items, ResourceFolderItem{folder: folderCopy})
	}

	items = append(items, MenuItem{
		title:       "返回练习菜单",
		description: "返回到练习菜单",
		action:      func() (tea.Model, error) { return NewPracticeMenu(), nil },
	})

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = getResourceTypeTitle(resourceType) + "文件夹"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &ResourceSelectionMenu{
		list:         l,
		resourceType: resourceType,
		folders:      folders,
	}
}

// ResourceFolderItem 表示一个文件夹条目
type ResourceFolderItem struct {
	folder practice.ResourceFolder
}

func (i ResourceFolderItem) Title() string {
	return i.folder.DisplayName
}

func (i ResourceFolderItem) Description() string {
	return fmt.Sprintf("包含 %d 个资源", len(i.folder.Files))
}

func (i ResourceFolderItem) FilterValue() string {
	return i.folder.DisplayName
}

// 获取资源类型标题
func getResourceTypeTitle(resourceType string) string {
	switch resourceType {
	case practice.Words:
		return "单词"
	case practice.Phrases:
		return "短语"
	case practice.Sentences:
		return "句子"
	case practice.Dialogues:
		return "对话"
	case practice.Articles:
		return "文章"
	default:
		return "资源"
	}
}

// Init 初始化模型
func (m ResourceSelectionMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m ResourceSelectionMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			practiceMenu := NewPracticeMenu()
			width, height := m.list.Width(), m.list.Height()+4
			if width > 0 && height > 4 {
				updatedModel, _ := practiceMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
				return updatedModel, nil
			}
			return practiceMenu, nil

		case "enter":
			switch selected := m.list.SelectedItem().(type) {
			case ResourceFolderItem:
				filesMenu := NewResourceFilesMenu(m.resourceType, selected.folder)
				width, height := m.list.Width(), m.list.Height()+4
				if width > 0 && height > 4 {
					updatedModel, _ := filesMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
					return updatedModel, nil
				}
				return filesMenu, nil
			case MenuItem:
				if selected.action != nil {
					newModel, err := selected.action()
					if err != nil {
						return m, nil
					}
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
func (m ResourceSelectionMenu) View() string {
	if m.quitting {
		return "再见！"
	}

	return m.list.View()
}

// ResourceFilesMenu 文件列表菜单
type ResourceFilesMenu struct {
	list         list.Model
	resourceType string
	folder       practice.ResourceFolder
	quitting     bool
}

// NewResourceFilesMenu 创建文件列表菜单
func NewResourceFilesMenu(resourceType string, folder practice.ResourceFolder) *ResourceFilesMenu {
	items := make([]list.Item, 0, len(folder.Files)+1)

	if len(folder.Files) == 0 {
		items = append(items, MenuItem{
			title:       "（该文件夹暂无资源）",
			description: "导入资源后即可在此练习",
			action:      nil,
		})
	} else {
		for _, file := range folder.Files {
			fileName := file
			identifier := practice.BuildResourceIdentifier(folder.DirName, fileName)
			display := practice.FormatResourceDisplayName(identifier)
			itemIdentifier := identifier
			itemDisplay := display
			items = append(items, MenuItem{
				title:       itemDisplay,
				description: "练习 " + itemDisplay,
				action: func() (tea.Model, error) {
					return NewPracticeSession(resourceType, itemIdentifier), nil
				},
			})
		}
	}

	items = append(items, MenuItem{
		title:       "返回文件夹列表",
		description: "返回上一层",
		action: func() (tea.Model, error) {
			return NewResourceSelectionMenu(resourceType), nil
		},
	})

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("%s - %s", getResourceTypeTitle(resourceType), folder.DisplayName)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &ResourceFilesMenu{
		list:         l,
		resourceType: resourceType,
		folder:       folder,
	}
}

func (m ResourceFilesMenu) Init() tea.Cmd {
	return nil
}

func (m ResourceFilesMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			selection := NewResourceSelectionMenu(m.resourceType)
			width, height := m.list.Width(), m.list.Height()+4
			if width > 0 && height > 4 {
				updatedModel, _ := selection.Update(tea.WindowSizeMsg{Width: width, Height: height})
				return updatedModel, nil
			}
			return selection, nil
		case "enter":
			item, ok := m.list.SelectedItem().(MenuItem)
			if ok && item.action != nil {
				newModel, err := item.action()
				if err != nil {
					return m, nil
				}
				width, height := m.list.Width(), m.list.Height()+4
				if width > 0 && height > 4 {
					updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: width, Height: height})
					return updatedModel, nil
				}
				return newModel, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ResourceFilesMenu) View() string {
	if m.quitting {
		return "再见！"
	}
	return m.list.View()
}
