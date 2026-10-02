package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ajilisiwei/mllt-cli/internal/manage"
	"github.com/ajilisiwei/mllt-cli/internal/practice"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ManageMenu 资源管理菜单模型
type ManageMenu struct {
	list     list.Model
	quitting bool
}

// 创建新的资源管理菜单
func NewManageMenu() *ManageMenu {
	items := []list.Item{
		MenuItem{
			title:       "删除资源",
			description: "删除练习资源",
			action:      func() (tea.Model, error) { return NewResourceTypeMenu("delete"), nil },
		},
		MenuItem{
			title:       "导入资源",
			description: "导入练习资源",
			action:      func() (tea.Model, error) { return NewResourceTypeMenu("import"), nil },
		},
		MenuItem{
			title:       "返回主菜单",
			description: "返回到主菜单",
			action:      func() (tea.Model, error) { return NewMainMenu(), nil },
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "资源管理"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &ManageMenu{
		list: l,
	}
}

// Init 初始化模型
func (m ManageMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m ManageMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// 返回到主菜单
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
func (m ManageMenu) View() string {
	if m.quitting {
		return "再见！"
	}

	return m.list.View()
}

// ResourceTypeMenu 资源类型菜单模型
type ResourceTypeMenu struct {
	list     list.Model
	action   string // "delete" 或 "import"
	quitting bool
}

// 创建新的资源类型菜单
func NewResourceTypeMenu(action string) *ResourceTypeMenu {
	items := []list.Item{
		MenuItem{
			title:       "单词",
			description: getActionTitle(action) + "单词资源",
			action: func() (tea.Model, error) {
				if action == "delete" {
					return NewManageResourceMenu(practice.Words, action), nil
				} else {
					return NewImportView(practice.Words), nil
				}
			},
		},
		MenuItem{
			title:       "短语",
			description: getActionTitle(action) + "短语资源",
			action: func() (tea.Model, error) {
				if action == "delete" {
					return NewManageResourceMenu(practice.Phrases, action), nil
				} else {
					return NewImportView(practice.Phrases), nil
				}
			},
		},
		MenuItem{
			title:       "句子",
			description: getActionTitle(action) + "句子资源",
			action: func() (tea.Model, error) {
				if action == "delete" {
					return NewManageResourceMenu(practice.Sentences, action), nil
				} else {
					return NewImportView(practice.Sentences), nil
				}
			},
		},
		MenuItem{
			title:       "对话",
			description: getActionTitle(action) + "对话资源",
			action: func() (tea.Model, error) {
				if action == "delete" {
					return NewManageResourceMenu(practice.Dialogues, action), nil
				} else {
					return NewImportView(practice.Dialogues), nil
				}
			},
		},
		MenuItem{
			title:       "文章",
			description: getActionTitle(action) + "文章资源",
			action: func() (tea.Model, error) {
				if action == "delete" {
					return NewManageResourceMenu(practice.Articles, action), nil
				} else {
					return NewImportView(practice.Articles), nil
				}
			},
		},
		MenuItem{
			title:       "返回资源管理菜单",
			description: "返回到资源管理菜单",
			action:      func() (tea.Model, error) { return NewManageMenu(), nil },
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = getActionTitle(action) + "资源"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &ResourceTypeMenu{
		list:   l,
		action: action,
	}
}

// 获取操作标题
func getActionTitle(action string) string {
	switch action {
	case "delete":
		return "删除"
	case "import":
		return "导入"
	default:
		return "管理"
	}
}

// Init 初始化模型
func (m ResourceTypeMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m ResourceTypeMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// 返回到管理菜单
			manageMenu := NewManageMenu()
			width, height := m.list.Width(), m.list.Height()+4
			if width > 0 && height > 4 {
				updatedModel, _ := manageMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
				return updatedModel, nil
			}
			return manageMenu, nil

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
func (m ResourceTypeMenu) View() string {
	if m.quitting {
		return "再见！"
	}

	return m.list.View()
}

// ManageResourceMenu 资源管理菜单模型
type ManageResourceMenu struct {
	list         list.Model
	resourceType string
	action       string
	folders      []practice.ResourceFolder
	quitting     bool
}

// ManageFolderItem 表示资源管理中的文件夹条目
type ManageFolderItem struct {
	folder       practice.ResourceFolder
	action       string
	resourceType string
}

func (i ManageFolderItem) Title() string {
	return i.folder.DisplayName
}

func (i ManageFolderItem) Description() string {
	return fmt.Sprintf("包含 %d 个资源", len(i.folder.Files))
}

func (i ManageFolderItem) FilterValue() string {
	return i.folder.DisplayName
}

// 创建新的资源管理菜单
func NewManageResourceMenu(resourceType, action string) *ManageResourceMenu {
	folders, err := manage.ListResourceFolders(resourceType)
	if err != nil {
		folders = []practice.ResourceFolder{}
	}

	items := []list.Item{}
	for _, folder := range folders {
		folderCopy := folder
		items = append(items, ManageFolderItem{
			folder:       folderCopy,
			action:       action,
			resourceType: resourceType,
		})
	}

	items = append(items, MenuItem{
		title:       "返回资源类型菜单",
		description: "返回到资源类型菜单",
		action:      func() (tea.Model, error) { return NewResourceTypeMenu(action), nil },
	})

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = getActionTitle(action) + getResourceTypeTitle(resourceType) + "文件夹"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &ManageResourceMenu{
		list:         l,
		resourceType: resourceType,
		action:       action,
		folders:      folders,
	}
}

// Init 初始化模型
func (m ManageResourceMenu) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m ManageResourceMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			resourceTypeMenu := NewResourceTypeMenu(m.action)
			width, height := m.list.Width(), m.list.Height()+4
			if width > 0 && height > 4 {
				updatedModel, _ := resourceTypeMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
				return updatedModel, nil
			}
			return resourceTypeMenu, nil

		case "enter":
			switch selected := m.list.SelectedItem().(type) {
			case ManageFolderItem:
				folderMenu := NewManageFolderDetailMenu(m.resourceType, m.action, selected.folder)
				width, height := m.list.Width(), m.list.Height()+4
				if width > 0 && height > 4 {
					updatedModel, _ := folderMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
					return updatedModel, nil
				}
				return folderMenu, nil
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
func (m ManageResourceMenu) View() string {
	if m.quitting {
		return "再见！"
	}

	return m.list.View()
}

// ManageFolderDetailMenu 文件夹详情菜单
type ManageFolderDetailMenu struct {
	list         list.Model
	resourceType string
	action       string
	folder       practice.ResourceFolder
	folderNotice string
	quitting     bool
}

// NewManageFolderDetailMenu 创建文件夹详情菜单
func NewManageFolderDetailMenu(resourceType, action string, folder practice.ResourceFolder) *ManageFolderDetailMenu {
	items := make([]list.Item, 0, len(folder.Files)+3)
	folderDir := folder.DirName
	folderDisplay := folder.DisplayName
	folderNotice := ""

	if len(folder.Files) == 0 {
		items = append(items, MenuItem{
			title:       "（该文件夹暂无资源）",
			description: "按 Esc 返回上一层",
			action:      nil,
		})
	} else {
		for _, file := range folder.Files {
			fileName := file
			identifier := practice.BuildResourceIdentifier(folderDir, fileName)
			display := practice.FormatResourceDisplayName(identifier)
			itemIdentifier := identifier
			itemDisplay := display
			items = append(items, MenuItem{
				title:       itemDisplay,
				description: getActionTitle(action) + itemDisplay,
				action: func() (tea.Model, error) {
					if action == "delete" {
						return NewDeleteConfirmView(resourceType, folderDir, itemIdentifier, itemDisplay), nil
					}
					return nil, nil
				},
			})
		}
	}

	if action == "delete" && folderDir != practice.DefaultFolderDir {
		if len(folder.Files) == 0 {
			items = append(items, MenuItem{
				title:       "删除该文件夹",
				description: "删除空文件夹",
				action: func() (tea.Model, error) {
					return NewDeleteFolderConfirmView(resourceType, folderDir, folderDisplay), nil
				},
			})
		} else {
			if folderDisplay != "" {
				folderNotice = fmt.Sprintf("文件夹“%s”非空，无法直接删除\n删除文件夹所有资源后，文件夹会自动删除", folderDisplay)
			} else {
				folderNotice = "文件夹非空，无法直接删除\n删除文件夹所有资源后，文件夹会自动删除"
			}
		}
	}

	items = append(items, MenuItem{
		title:       "返回文件夹列表",
		description: "返回上一层",
		action:      func() (tea.Model, error) { return NewManageResourceMenu(resourceType, action), nil },
	})

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("%s - %s", getResourceTypeTitle(resourceType), folderDisplay)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	return &ManageFolderDetailMenu{
		list:         l,
		resourceType: resourceType,
		action:       action,
		folder:       folder,
		folderNotice: folderNotice,
	}
}

// NewManageFolderDetailMenuByDir 根据目录重新构建文件夹详情菜单
func NewManageFolderDetailMenuByDir(resourceType, action, folderDir string) tea.Model {
	folder, err := manage.GetResourceFolder(resourceType, folderDir)
	if err != nil {
		return NewManageResourceMenu(resourceType, action)
	}
	return NewManageFolderDetailMenu(resourceType, action, *folder)
}

func (m ManageFolderDetailMenu) Init() tea.Cmd {
	return nil
}

func (m ManageFolderDetailMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			folderMenu := NewManageResourceMenu(m.resourceType, m.action)
			width, height := m.list.Width(), m.list.Height()+4
			if width > 0 && height > 4 {
				updatedModel, _ := folderMenu.Update(tea.WindowSizeMsg{Width: width, Height: height})
				return updatedModel, nil
			}
			return folderMenu, nil
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

func (m ManageFolderDetailMenu) View() string {
	if m.quitting {
		return "再见！"
	}

	view := m.list.View()
	if m.folderNotice != "" {
		view += "\n\n" + RenderText(m.folderNotice)
	}
	return view
}

// ImportView 导入视图模型
type ImportView struct {
	resourceType          string
	folderInput           textinput.Model
	fileInput             textinput.Model
	message               string
	width                 int
	height                int
	quitting              bool
	focus                 string
	folderOptions         []folderOption
	filteredFolderOptions []folderOption
	selectedFolderIndex   int
	folderDropdownVisible bool
	fileOptions           []fileOption
	filteredFileOptions   []fileOption
	selectedFileIndex     int
	fileDropdownVisible   bool
	fileVisibleOffset     int
}

type folderOption struct {
	dir     string
	display string
}

type fileOption struct {
	path  string
	label string
	isDir bool
}

const maxFileSuggestionDisplay = 10

func buildFolderOptions(resourceType string) []folderOption {
	folders, err := manage.ListResourceFolders(resourceType)
	if err != nil {
		return []folderOption{}
	}
	options := make([]folderOption, 0, len(folders))
	for _, folder := range folders {
		options = append(options, folderOption{
			dir:     folder.DirName,
			display: folder.DisplayName,
		})
	}
	return options
}

// 创建新的导入视图
func NewImportView(resourceType string) *ImportView {
	folderInput := textinput.New()
	folderInput.Placeholder = "默认"
	folderInput.Focus()
	folderInput.Width = 40

	fileInput := textinput.New()
	fileInput.Placeholder = "输入文件路径..."
	fileInput.Width = 40

	options := buildFolderOptions(resourceType)

	return &ImportView{
		resourceType:          resourceType,
		folderInput:           folderInput,
		fileInput:             fileInput,
		focus:                 "folder",
		folderOptions:         options,
		filteredFolderOptions: options,
		selectedFolderIndex:   0,
		folderDropdownVisible: len(options) > 0,
		fileOptions:           []fileOption{},
		filteredFileOptions:   []fileOption{},
		selectedFileIndex:     0,
		fileDropdownVisible:   false,
		fileVisibleOffset:     0,
	}
}

// Init 初始化模型
func (m ImportView) Init() tea.Cmd {
	return textinput.Blink
}

// Update 更新模型
func (m ImportView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.width > 0 {
			m.folderInput.Width = msg.Width - 20
			m.fileInput.Width = msg.Width - 20
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			// 返回到资源类型菜单
			newModel := NewResourceTypeMenu("import")
			// 传递当前窗口大小给新模型
			if m.width > 0 && m.height > 0 {
				updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				return updatedModel, nil
			}
			return newModel, nil

		case "tab":
			if m.focus == "folder" {
				m.applySelectedFolderSuggestion()
				m.setFocus("file")
			} else {
				if m.fileDropdownVisible && len(m.filteredFileOptions) > 0 {
					m.applySelectedFileSuggestion()
				} else {
					m.setFocus("folder")
				}
			}
			return m, nil

		case "enter":
			if m.focus == "folder" {
				m.applySelectedFolderSuggestion()
				m.setFocus("file")
				return m, nil
			}
			return m.handleImport()

		case "up":
			if m.focus == "folder" && m.folderDropdownVisible && len(m.filteredFolderOptions) > 0 {
				m.moveFolderSelection(-1)
				m.applySelectedFolderSuggestion()
				return m, nil
			}
			if m.focus == "file" && m.fileDropdownVisible && len(m.filteredFileOptions) > 0 {
				m.moveFileSelection(-1)
				m.applySelectedFileSuggestion()
				return m, nil
			}

		case "down":
			if m.focus == "folder" && m.folderDropdownVisible && len(m.filteredFolderOptions) > 0 {
				m.moveFolderSelection(1)
				m.applySelectedFolderSuggestion()
				return m, nil
			}
			if m.focus == "file" && m.fileDropdownVisible && len(m.filteredFileOptions) > 0 {
				m.moveFileSelection(1)
				m.applySelectedFileSuggestion()
				return m, nil
			}
		}
	}

	// 处理文本输入
	var cmd tea.Cmd
	if m.focus == "folder" {
		m.folderInput, cmd = m.folderInput.Update(msg)
		m.updateFolderSuggestions()
	} else {
		m.fileInput, cmd = m.fileInput.Update(msg)
		m.updateFileSuggestions()
	}
	return m, cmd
}

func (m *ImportView) setFocus(target string) {
	if m.focus == target {
		return
	}
	switch target {
	case "folder":
		m.folderInput.Focus()
		m.fileInput.Blur()
		m.updateFolderSuggestions()
	case "file":
		m.fileInput.Focus()
		m.folderInput.Blur()
		m.updateFileSuggestions()
	}
	m.focus = target
}

func (m *ImportView) updateFolderSuggestions() {
	query := strings.ToLower(strings.TrimSpace(m.folderInput.Value()))
	if query == "" {
		m.filteredFolderOptions = m.folderOptions
		m.selectedFolderIndex = 0
		m.folderDropdownVisible = len(m.filteredFolderOptions) > 0
		return
	}

	filtered := make([]folderOption, 0, len(m.folderOptions))
	for _, option := range m.folderOptions {
		if strings.HasPrefix(strings.ToLower(option.display), query) {
			filtered = append(filtered, option)
		}
	}

	m.filteredFolderOptions = filtered
	if len(filtered) == 0 {
		m.folderDropdownVisible = false
		m.selectedFolderIndex = 0
		return
	}
	if m.selectedFolderIndex >= len(filtered) {
		m.selectedFolderIndex = 0
	}
	m.folderDropdownVisible = true
}

func (m *ImportView) moveFolderSelection(delta int) {
	if len(m.filteredFolderOptions) == 0 {
		return
	}
	count := len(m.filteredFolderOptions)
	m.selectedFolderIndex = (m.selectedFolderIndex + delta + count) % count
}

func (m *ImportView) applySelectedFolderSuggestion() {
	if !m.folderDropdownVisible || len(m.filteredFolderOptions) == 0 {
		return
	}
	option := m.filteredFolderOptions[m.selectedFolderIndex]
	m.folderInput.SetValue(option.display)
	m.folderInput.CursorEnd()
}

func (m *ImportView) updateFileSuggestions() {
	value := m.fileInput.Value()
	trimmed := strings.TrimSpace(value)
	expanded := expandUserPath(trimmed)
	separator := string(os.PathSeparator)

	searchDir := expanded
	prefix := ""
	inputEndsWithSep := strings.HasSuffix(trimmed, separator)

	if trimmed == "" {
		searchDir = "."
	} else if inputEndsWithSep {
		searchDir = expanded
	} else {
		searchDir = filepath.Dir(expanded)
		if searchDir == "" {
			searchDir = "."
		}
		prefix = filepath.Base(expanded)
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		m.fileOptions = nil
		m.filteredFileOptions = nil
		m.fileDropdownVisible = false
		m.selectedFileIndex = 0
		return
	}

	lowerPrefix := strings.ToLower(prefix)
	options := make([]fileOption, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if lowerPrefix != "" && !strings.HasPrefix(strings.ToLower(name), lowerPrefix) {
			continue
		}
		fullPath := filepath.Join(searchDir, name)
		labelPath := fullPath
		if entry.IsDir() {
			fullPath += separator
			labelPath += separator
		}
		labelPath = collapseUserPath(labelPath)
		options = append(options, fileOption{
			path:  fullPath,
			label: labelPath,
			isDir: entry.IsDir(),
		})
	}

	if len(options) == 0 {
		m.fileOptions = nil
		m.filteredFileOptions = nil
		m.fileDropdownVisible = false
		m.selectedFileIndex = 0
		m.fileVisibleOffset = 0
		return
	}

	m.fileOptions = options
	m.filteredFileOptions = options
	m.selectedFileIndex = 0
	m.fileDropdownVisible = true
	m.fileVisibleOffset = 0
}

func (m *ImportView) moveFileSelection(delta int) {
	if len(m.filteredFileOptions) == 0 {
		return
	}
	count := len(m.filteredFileOptions)
	m.selectedFileIndex = (m.selectedFileIndex + delta + count) % count

	if m.selectedFileIndex < m.fileVisibleOffset {
		m.fileVisibleOffset = m.selectedFileIndex
	} else if m.selectedFileIndex >= m.fileVisibleOffset+maxFileSuggestionDisplay {
		m.fileVisibleOffset = m.selectedFileIndex - (maxFileSuggestionDisplay - 1)
	}

	maxOffset := count - maxFileSuggestionDisplay
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.fileVisibleOffset > maxOffset {
		m.fileVisibleOffset = maxOffset
	}
}

func (m *ImportView) applySelectedFileSuggestion() {
	if !m.fileDropdownVisible || len(m.filteredFileOptions) == 0 {
		return
	}
	option := m.filteredFileOptions[m.selectedFileIndex]
	m.fileInput.SetValue(option.label)
	m.fileInput.CursorEnd()
	// 当选择的是目录时，保持目录状态以便继续深入
	if option.isDir {
		m.updateFileSuggestions()
	} else {
		m.fileDropdownVisible = false
	}
}

func expandUserPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			if trimmed == "~" {
				return home
			}
			rest := strings.TrimPrefix(trimmed, "~")
			rest = strings.TrimPrefix(rest, string(os.PathSeparator))
			return filepath.Join(home, rest)
		}
	}
	return trimmed
}

func collapseUserPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	home = filepath.Clean(home)
	separator := string(os.PathSeparator)
	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, home+separator) || cleaned == home {
		rel, err := filepath.Rel(home, cleaned)
		if err == nil {
			if rel == "." || rel == "" {
				return "~"
			}
			collapsed := filepath.Join("~", rel)
			if strings.HasSuffix(path, separator) && !strings.HasSuffix(collapsed, separator) {
				collapsed += separator
			}
			return collapsed
		}
	}
	if strings.HasSuffix(path, separator) && !strings.HasSuffix(cleaned, separator) {
		return cleaned + separator
	}
	return path
}
func (m ImportView) handleImport() (tea.Model, tea.Cmd) {
	filePath := strings.TrimSpace(m.fileInput.Value())
	if filePath == "" {
		m.message = "请输入文件路径"
		return m, nil
	}

	expandedPath := expandUserPath(filePath)
	if expandedPath == "" {
		m.message = "请输入有效的文件路径"
		return m, nil
	}

	folderValue := strings.TrimSpace(m.folderInput.Value())
	normalizedFolder, changed := practice.NormalizeFolderName(folderValue)
	if changed && folderValue != "" && normalizedFolder != folderValue {
		m.folderInput.SetValue(normalizedFolder)
		m.folderInput.CursorEnd()
		m.message = fmt.Sprintf("文件夹名称已调整为 %s，请确认后再次导入", normalizedFolder)
		return m, nil
	}

	err := manage.ImportResourceForTest(m.resourceType, normalizedFolder, expandedPath)
	if err != nil {
		m.message = "导入失败: " + err.Error()
		return m, nil
	}

	newModel := NewResourceTypeMenu("import")
	if m.width > 0 && m.height > 0 {
		updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		return updatedModel, nil
	}
	return newModel, nil
}

// View 渲染视图
func (m ImportView) View() string {
	if m.quitting {
		return "导入已取消！"
	}

	var s strings.Builder

	// 标题
	title := "导入" + getResourceTypeTitle(m.resourceType)
	s.WriteString(RenderTitle(title) + "\n\n")

	// 文件夹输入
	s.WriteString(RenderHighlight("选择文件夹:") + "\n")
	s.WriteString(m.folderInput.View() + "\n")
	if m.folderDropdownVisible && len(m.filteredFolderOptions) > 0 {
		for idx, option := range m.filteredFolderOptions {
			marker := "  "
			if idx == m.selectedFolderIndex {
				marker = "> "
			}
			s.WriteString(marker + option.display + "\n")
		}
	}
	s.WriteString("\n")

	// 文件路径输入
	s.WriteString(RenderHighlight("请输入文件路径:") + "\n")
	s.WriteString(m.fileInput.View() + "\n\n")
	if m.fileDropdownVisible && len(m.filteredFileOptions) > 0 {
		start := m.fileVisibleOffset
		if start < 0 {
			start = 0
		}
		end := start + maxFileSuggestionDisplay
		if end > len(m.filteredFileOptions) {
			end = len(m.filteredFileOptions)
		}
		if start > 0 {
			s.WriteString("  ...\n")
		}
		for idx := start; idx < end; idx++ {
			option := m.filteredFileOptions[idx]
			marker := "  "
			if idx == m.selectedFileIndex {
				marker = "> "
			}
			s.WriteString(marker + option.label + "\n")
		}
		if end < len(m.filteredFileOptions) {
			s.WriteString("  ...\n")
		}
		s.WriteString("\n")
	}

	// 显示消息
	if m.message != "" {
		if strings.HasPrefix(m.message, "导入失败") {
			s.WriteString(RenderError(m.message) + "\n\n")
		} else {
			s.WriteString(RenderSuccess(m.message) + "\n\n")
		}
	}

	// 显示提示
	s.WriteString(RenderText("按 Tab 切换输入框，按 Enter 导入，按 Esc 返回") + "\n")

	return s.String()
}

// DeleteConfirmView 删除确认视图模型
type DeleteConfirmView struct {
	resourceType string
	folderDir    string
	resourceName string
	displayName  string
	message      string
	width        int
	height       int
	quitting     bool
}

// 创建新的删除确认视图
func NewDeleteConfirmView(resourceType, folderDir, resourceIdentifier, displayName string) *DeleteConfirmView {
	return &DeleteConfirmView{
		resourceType: resourceType,
		folderDir:    folderDir,
		resourceName: resourceIdentifier,
		displayName:  displayName,
	}
}

// Init 初始化模型
func (m DeleteConfirmView) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m DeleteConfirmView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "y", "Y":
			// 确认删除
			err := manage.DeleteResourceWithoutConfirm(m.resourceType, m.resourceName)
			if err != nil {
				m.message = "删除失败: " + err.Error()
				return m, nil
			}
			newModel := NewManageFolderDetailMenuByDir(m.resourceType, "delete", m.folderDir)
			// 传递当前窗口大小
			if m.width > 0 && m.height > 0 {
				updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				return updatedModel, nil
			}
			return newModel, nil

		case "n", "N", "esc":
			// 取消删除，返回资源列表
			newModel := NewManageFolderDetailMenuByDir(m.resourceType, "delete", m.folderDir)
			// 传递当前窗口大小
			if m.width > 0 && m.height > 0 {
				updatedModel, _ := newModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				return updatedModel, nil
			}
			return newModel, nil
		}
	}

	return m, nil
}

// View 渲染视图
func (m DeleteConfirmView) View() string {
	if m.quitting {
		return "再见！"
	}

	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("确认删除"))
	s.WriteString("\n\n")
	s.WriteString("确认删除以下资源吗？\n\n")
	s.WriteString("资源类型: " + getResourceTypeTitle(m.resourceType) + "\n")
	s.WriteString("所在文件夹: " + practice.FolderDisplayName(m.folderDir) + "\n")
	s.WriteString("文件名: " + m.displayName + "\n\n")
	s.WriteString("按 Y 确认删除，按 N 或 Esc 取消\n")

	if m.message != "" {
		s.WriteString("\n")
		s.WriteString(m.message)
	}

	return s.String()
}

// DeleteFolderConfirmView 删除文件夹确认视图
type DeleteFolderConfirmView struct {
	resourceType  string
	folderDir     string
	folderDisplay string
	message       string
	width         int
	height        int
	quitting      bool
}

func NewDeleteFolderConfirmView(resourceType, folderDir, folderDisplay string) *DeleteFolderConfirmView {
	return &DeleteFolderConfirmView{
		resourceType:  resourceType,
		folderDir:     folderDir,
		folderDisplay: folderDisplay,
	}
}

func (m DeleteFolderConfirmView) Init() tea.Cmd {
	return nil
}

func (m DeleteFolderConfirmView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "y", "Y":
			if err := manage.DeleteResourceFolder(m.resourceType, m.folderDir); err != nil {
				m.message = "删除失败: " + err.Error()
				return m, nil
			}
			menu := NewManageResourceMenu(m.resourceType, "delete")
			if m.width > 0 && m.height > 0 {
				updatedModel, _ := menu.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				return updatedModel, nil
			}
			return menu, nil
		case "n", "N", "esc":
			menu := NewManageFolderDetailMenuByDir(m.resourceType, "delete", m.folderDir)
			if m.width > 0 && m.height > 0 {
				updatedModel, _ := menu.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				return updatedModel, nil
			}
			return menu, nil
		}
	}

	return m, nil
}

func (m DeleteFolderConfirmView) View() string {
	if m.quitting {
		return "再见！"
	}

	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(TitleStyle.Render("确认删除文件夹"))
	s.WriteString("\n\n")
	s.WriteString("确认删除以下文件夹吗？\n\n")
	s.WriteString("资源类型: " + getResourceTypeTitle(m.resourceType) + "\n")
	s.WriteString("文件夹: " + m.folderDisplay + "\n\n")
	s.WriteString("按 Y 确认删除，按 N 或 Esc 取消\n")

	if m.message != "" {
		s.WriteString("\n")
		s.WriteString(m.message)
	}

	return s.String()
}
