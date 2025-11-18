package ui

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ajilisiwei/mllt-cli/internal/bookmark"
	"github.com/ajilisiwei/mllt-cli/internal/config"
	"github.com/ajilisiwei/mllt-cli/internal/practice"
	"github.com/ajilisiwei/mllt-cli/internal/sound"
	"github.com/ajilisiwei/mllt-cli/internal/srs"
	"github.com/ajilisiwei/mllt-cli/internal/statistics"
)

type commandOption struct {
	name        string
	description string
}

var baseCommandOptions = []commandOption{
	{name: "exit", description: "退出当前练习会话"},
	{name: "help", description: "显示当前练习会话的帮助信息"},
	{name: "mark", description: "标记当前内容，下次练习跳过"},
	{name: "unmark", description: "取消标记当前内容"},
	{name: "favorite", description: "收藏当前内容，可在收藏列表中查看"},
	{name: "unfavorite", description: "取消收藏当前内容"},
}

func cloneCommandOptions(options []commandOption) []commandOption {
	cloned := make([]commandOption, len(options))
	copy(cloned, options)
	return cloned
}

func filterCommandOptions(options []commandOption, query string) []commandOption {
	trimmed := strings.TrimSpace(strings.ToLower(query))
	if trimmed == "" {
		return cloneCommandOptions(options)
	}

	filtered := make([]commandOption, 0, len(options))
	for _, option := range options {
		if strings.HasPrefix(option.name, trimmed) {
			filtered = append(filtered, option)
		}
	}
	return filtered
}

func sessionCommandOptions(resourceType string) []commandOption {
	options := make([]commandOption, 0, len(baseCommandOptions))
	for _, option := range baseCommandOptions {
		if !bookmark.SupportsMark(resourceType) && (option.name == "mark" || option.name == "unmark") {
			continue
		}
		options = append(options, option)
	}
	return options
}

// PracticeSession 练习会话模型
type PracticeSession struct {
	resourceType    string
	fileName        string
	items           []string        // 练习项目
	currentIndex    int             // 当前项目索引
	textInput       textinput.Model // 文本输入
	progress        progress.Model  // 进度条
	width           int             // 窗口宽度
	height          int             // 窗口高度
	startTime       time.Time       // 开始时间
	endTime         time.Time       // 结束时间
	correct         int             // 正确数量
	incorrect       int             // 错误数量
	displayFileName string          // 用于展示的文件名
	orderMode       string          // 练习顺序模式
	state           string          // 状态："practicing", "finished"
	result          string          // 结果信息
	quitting        bool
	// v0.2 新增：错误状态跟踪
	lastInputWrong bool   // 上次输入是否错误
	wrongInput     string // 错误的输入内容
	expectedText   string // 期望的正确文本
	// v0.2 新增：随机顺序支持
	practiceOrder    []int // 练习顺序索引列表
	completedCount   int   // 已完成的项目数量
	initialItemCount int   // 初始练习项目数量
	srsEnabled       bool
	srsSchedule      *srs.Schedule
	statsLogged      bool
	// v0.5 新增：命令模式支持
	commandOptions          []commandOption
	filteredCommands        []commandOption
	selectedCommandIndex    int
	commandDropdownVisible  bool
	inCommandMode           bool
	commandFeedback         string
	commandFeedbackIsError  bool
	inputDefaultTextStyle   lipgloss.Style
	inputDefaultCursorStyle lipgloss.Style
}

// 创建新的练习会话
func NewPracticeSession(resourceType, fileName string) *PracticeSession {
	items, err := practice.ReadResourceFile(resourceType, fileName)
	if err != nil {
		items = []string{}
	}

	normalizedItems := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			normalizedItems = append(normalizedItems, trimmed)
		}
	}

	// 过滤已标记或收藏的内容（特殊列表除外）
	if !bookmark.IsSpecialList(fileName) && len(normalizedItems) > 0 {
		normalizedItems = filterExcludedItems(resourceType, normalizedItems)
	}

	// 创建文本输入
	ti := textinput.New()
	ti.Placeholder = "输入这里..."
	ti.Prompt = ""
	ti.Focus()
	ti.Width = 40

	// 创建进度条
	p := progress.New(progress.WithDefaultGradient())

	// 初始化练习顺序
	practiceOrder := make([]int, len(normalizedItems))
	for i := range practiceOrder {
		practiceOrder[i] = i
	}

	orderMode := strings.ToLower(config.AppConfig.NextOneOrder)
	if orderMode == "" {
		orderMode = "random"
	}

	var schedule *srs.Schedule
	srsEnabled := false

	if len(practiceOrder) > 0 && resourceType != practice.Articles && orderMode == "ebbinghaus" {
		if sch, err := srs.Load(resourceType, fileName, normalizedItems); err == nil {
			ordered := sch.Order(normalizedItems)
			if len(ordered) == len(practiceOrder) {
				practiceOrder = ordered
				schedule = sch
				srsEnabled = true
			}
		} else {
			orderMode = "sequential"
		}
	}

	if !srsEnabled && len(practiceOrder) > 1 && resourceType != practice.Articles {
		switch orderMode {
		case "random":
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(practiceOrder), func(i, j int) {
				practiceOrder[i], practiceOrder[j] = practiceOrder[j], practiceOrder[i]
			})
		case "sequential":
			// already sequential
		default:
			orderMode = "sequential"
		}
	}

	sessionOptions := sessionCommandOptions(resourceType)

	session := &PracticeSession{
		resourceType:            resourceType,
		fileName:                fileName,
		items:                   normalizedItems,
		currentIndex:            0,
		textInput:               ti,
		progress:                p,
		startTime:               time.Now(),
		orderMode:               orderMode,
		state:                   "practicing",
		practiceOrder:           practiceOrder,
		completedCount:          0,
		initialItemCount:        len(practiceOrder),
		srsEnabled:              srsEnabled,
		srsSchedule:             schedule,
		displayFileName:         practice.FormatResourceDisplayName(fileName),
		commandOptions:          sessionOptions,
		filteredCommands:        cloneCommandOptions(sessionOptions),
		selectedCommandIndex:    0,
		commandDropdownVisible:  false,
		inCommandMode:           false,
		inputDefaultTextStyle:   ti.TextStyle,
		inputDefaultCursorStyle: ti.CursorStyle,
	}

	if len(normalizedItems) == 0 {
		session.state = "finished"
		session.endTime = time.Now()
		session.result = emptyListMessage(fileName)
	}

	return session
}

func filterExcludedItems(resourceType string, items []string) []string {
	filtered := make([]string, 0, len(items))

	markedSet := make(map[string]struct{})
	if marked, err := bookmark.GetItems(resourceType, bookmark.MarkedList); err == nil {
		for _, item := range marked {
			markedSet[item] = struct{}{}
		}
	}

	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := markedSet[trimmed]; ok {
			continue
		}
		filtered = append(filtered, trimmed)
	}

	return filtered
}

func emptyListMessage(fileName string) string {
	if bookmark.IsSpecialList(fileName) {
		return fmt.Sprintf("当前\"%s\"列表为空。按 Enter 或 Esc 返回练习菜单。", fileName)
	}
	return "当前资源没有可练习内容，可能已经全部标记。按 Enter 或 Esc 返回练习菜单。"
}

// Init 初始化模型
func (m PracticeSession) Init() tea.Cmd {
	return textinput.Blink
}

// Update 更新模型
func (m *PracticeSession) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typedMsg.Width
		m.height = typedMsg.Height
		m.progress.Width = typedMsg.Width - 20
		m.textInput.Width = typedMsg.Width - 20
		return m, nil

	case tea.KeyMsg:
		key := typedMsg.String()
		m.playKeyboardSound(typedMsg)
		switch key {
		case "ctrl+c":
			m.quitting = true
			sound.StopAllSounds()
			m.logStatistics(false)
			return m, tea.Quit
		case "esc":
			sound.StopAllSounds()
			m.logStatistics(false)
			practiceMenu := NewPracticeMenu()
			if m.width > 0 && m.height > 4 {
				updatedModel, _ := practiceMenu.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				return updatedModel, nil
			}
			return practiceMenu, nil
		case "enter":
			if m.state == "finished" {
				sound.StopAllSounds()
				practiceMenu := NewPracticeMenu()
				if m.width > 0 && m.height > 4 {
					updatedModel, _ := practiceMenu.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
					return updatedModel, nil
				}
				return practiceMenu, nil
			}

			trimmed := strings.TrimSpace(m.textInput.Value())
			if strings.HasPrefix(trimmed, ">") {
				return m.handleCommand(trimmed)
			}
			return m.handleAnswerSubmission()
		case "tab":
			if m.commandDropdownVisible && len(m.filteredCommands) > 0 {
				m.applySelectedCommandSuggestion()
				return m, nil
			}
		}

		if m.commandDropdownVisible && len(m.filteredCommands) > 0 {
			switch key {
			case "up":
				m.moveCommandSelection(-1)
				m.applySelectedCommandSuggestion()
				return m, nil
			case "down":
				m.moveCommandSelection(1)
				m.applySelectedCommandSuggestion()
				return m, nil
			}
		}
	}

	if m.state == "finished" {
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.updateCommandDropdown()

	return m, cmd
}

func (m *PracticeSession) handleAnswerSubmission() (tea.Model, tea.Cmd) {
	m.setCommandFeedback("", false)
	value := m.textInput.Value()
	userInput := strings.TrimSpace(value)
	if strings.Contains(userInput, " ->> ") {
		parts := strings.Split(userInput, " ->> ")
		if len(parts) > 0 {
			userInput = strings.TrimSpace(parts[0])
		}
	}

	originalItem := m.getCurrentRawItem()
	expectedInput := m.getExpectedInput(originalItem)

	isCorrect := m.isInputCorrectForItem(userInput, originalItem)
	m.recordSpacedRepetition(originalItem, isCorrect)

	if isCorrect {
		m.correct++
		m.clearErrorState()
		m.textInput.SetValue("")
		m.updateCommandDropdown()
		m.advanceToNextItem()
	} else {
		m.incorrect++
		m.lastInputWrong = true
		m.wrongInput = userInput
		m.expectedText = expectedInput
		m.textInput.SetValue("")
		m.updateCommandDropdown()
	}

	return m, nil
}

func (m *PracticeSession) handleCommand(raw string) (tea.Model, tea.Cmd) {
	defer m.resetCommandInput()

	commandText := strings.TrimSpace(strings.TrimPrefix(raw, ">"))
	if commandText == "" {
		m.setCommandFeedback("请输入命令。", true)
		return m, nil
	}

	parts := strings.Fields(commandText)
	if len(parts) == 0 {
		m.setCommandFeedback("请输入命令。", true)
		return m, nil
	}

	commandName := strings.ToLower(parts[0])

	switch commandName {
	case "exit":
		sound.StopAllSounds()
		practiceMenu := NewPracticeMenu()
		if m.width > 0 && m.height > 4 {
			updatedModel, _ := practiceMenu.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			return updatedModel, nil
		}
		return practiceMenu, nil
	case "help":
		m.setCommandFeedback(m.renderCommandHelpDetail(), false)
		return m, nil
	case "mark":
		return m.handleMarkCommand()
	case "unmark":
		return m.handleUnmarkCommand()
	case "favorite":
		return m.handleFavoriteCommand()
	case "unfavorite":
		return m.handleUnfavoriteCommand()
	default:
		m.setCommandFeedback(fmt.Sprintf("未知命令: %s", commandText), true)
		return m, nil
	}
}

func (m *PracticeSession) handleMarkCommand() (tea.Model, tea.Cmd) {
	if !bookmark.SupportsMark(m.resourceType) {
		m.setCommandFeedback("当前资源类型不支持标记功能。", true)
		return m, nil
	}

	item := m.getCurrentRawItem()
	if item == "" {
		m.setCommandFeedback("没有可标记的内容。", true)
		return m, nil
	}

	added, err := bookmark.Add(m.resourceType, bookmark.MarkedList, item)
	if err != nil {
		m.setCommandFeedback(fmt.Sprintf("标记失败: %v", err), true)
		return m, nil
	}
	if !added {
		m.setCommandFeedback("该内容已标记。", false)
		return m, nil
	}

	m.setCommandFeedback("已标记当前内容，下次练习将跳过。", false)
	m.clearErrorState()
	m.removeItemFromSRS(item)
	m.advanceToNextItem()
	return m, nil
}

func (m *PracticeSession) handleUnmarkCommand() (tea.Model, tea.Cmd) {
	if !bookmark.SupportsMark(m.resourceType) {
		m.setCommandFeedback("当前资源类型不支持取消标记。", true)
		return m, nil
	}

	item := m.getCurrentRawItem()
	if item == "" {
		m.setCommandFeedback("没有可取消标记的内容。", true)
		return m, nil
	}

	removed, err := bookmark.Remove(m.resourceType, bookmark.MarkedList, item)
	if err != nil {
		m.setCommandFeedback(fmt.Sprintf("取消标记失败: %v", err), true)
		return m, nil
	}
	if !removed {
		m.setCommandFeedback("当前内容未被标记。", false)
		return m, nil
	}

	m.setCommandFeedback("已取消标记当前内容。", false)
	m.clearErrorState()
	m.removeItemFromSRS(item)
	if bookmark.IsSpecialList(m.fileName) {
		m.removeCurrentItemFromSession(item)
	} else {
		m.advanceToNextItem()
	}
	return m, nil
}

func (m *PracticeSession) handleFavoriteCommand() (tea.Model, tea.Cmd) {
	item := m.getCurrentRawItem()
	if item == "" {
		m.setCommandFeedback("没有可收藏的内容。", true)
		return m, nil
	}

	added, err := bookmark.Add(m.resourceType, bookmark.FavoriteList, item)
	if err != nil {
		m.setCommandFeedback(fmt.Sprintf("收藏失败: %v", err), true)
		return m, nil
	}
	if !added {
		m.setCommandFeedback("该内容已在收藏列表中。", false)
		return m, nil
	}

	m.setCommandFeedback("已收藏当前内容，可在收藏列表中查看。", false)
	m.clearErrorState()
	m.removeItemFromSRS(item)
	m.advanceToNextItem()
	return m, nil
}

func (m *PracticeSession) handleUnfavoriteCommand() (tea.Model, tea.Cmd) {
	item := m.getCurrentRawItem()
	if item == "" {
		m.setCommandFeedback("没有可取消收藏的内容。", true)
		return m, nil
	}

	removed, err := bookmark.Remove(m.resourceType, bookmark.FavoriteList, item)
	if err != nil {
		m.setCommandFeedback(fmt.Sprintf("取消收藏失败: %v", err), true)
		return m, nil
	}
	if !removed {
		m.setCommandFeedback("当前内容未被收藏。", false)
		return m, nil
	}

	m.setCommandFeedback("已取消收藏当前内容。", false)
	m.clearErrorState()
	m.removeItemFromSRS(item)
	if bookmark.IsSpecialList(m.fileName) {
		m.removeCurrentItemFromSession(item)
	} else {
		m.advanceToNextItem()
	}
	return m, nil
}

func (m *PracticeSession) moveCommandSelection(delta int) {
	if len(m.filteredCommands) == 0 {
		m.selectedCommandIndex = 0
		return
	}

	count := len(m.filteredCommands)
	m.selectedCommandIndex = (m.selectedCommandIndex + delta) % count
	if m.selectedCommandIndex < 0 {
		m.selectedCommandIndex += count
	}
}

func (m *PracticeSession) applySelectedCommandSuggestion() {
	if len(m.filteredCommands) == 0 {
		return
	}

	selected := m.filteredCommands[m.selectedCommandIndex]
	m.textInput.SetValue("> " + selected.name)
	m.textInput.CursorEnd()
	m.updateCommandDropdown()
}

func (m *PracticeSession) updateCommandDropdown() {
	input := m.textInput.Value()
	trimmedLeading := strings.TrimLeft(input, " ")

	m.inCommandMode = strings.HasPrefix(strings.TrimSpace(input), ">")

	if !strings.HasPrefix(trimmedLeading, ">") {
		m.commandDropdownVisible = false
		m.filteredCommands = cloneCommandOptions(m.commandOptions)
		m.selectedCommandIndex = 0
		return
	}

	body := strings.TrimSpace(strings.TrimPrefix(trimmedLeading, ">"))
	shouldShowDropdown := strings.TrimSpace(trimmedLeading) == ">" || strings.HasPrefix(trimmedLeading, "> ")
	m.filteredCommands = filterCommandOptions(m.commandOptions, body)
	if len(m.filteredCommands) == 0 {
		m.commandDropdownVisible = false
		m.selectedCommandIndex = 0
		return
	}

	if m.selectedCommandIndex >= len(m.filteredCommands) {
		m.selectedCommandIndex = 0
	}

	m.commandDropdownVisible = shouldShowDropdown
}

func (m *PracticeSession) resetCommandInput() {
	m.textInput.SetValue("")
	m.textInput.CursorEnd()
	m.commandDropdownVisible = false
	m.filteredCommands = cloneCommandOptions(m.commandOptions)
	m.selectedCommandIndex = 0
	m.inCommandMode = false
	m.updateCommandDropdown()
}

func (m *PracticeSession) setCommandFeedback(message string, isError bool) {
	clean := strings.TrimSpace(message)
	m.commandFeedback = clean
	m.commandFeedbackIsError = isError && clean != ""
}

func (m *PracticeSession) playKeyboardSound(msg tea.KeyMsg) {
	if m.state != "practicing" {
		return
	}
	if !config.AppConfig.InputKeyboardSound {
		return
	}

	char := rune(0)
	if len(msg.Runes) > 0 {
		char = msg.Runes[0]
	}
	sound.PlayTypingSound(char)
}

func (m *PracticeSession) clearErrorState() {
	m.lastInputWrong = false
	m.wrongInput = ""
	m.expectedText = ""
}

func (m *PracticeSession) recordSpacedRepetition(item string, correct bool) {
	if !m.srsEnabled || m.srsSchedule == nil || item == "" {
		return
	}
	_ = m.srsSchedule.RecordResult(item, correct)
}

func (m *PracticeSession) removeItemFromSRS(item string) {
	if !m.srsEnabled || m.srsSchedule == nil || item == "" {
		return
	}
	_ = m.srsSchedule.RemoveItem(item)
}

func (m *PracticeSession) removeCurrentItemFromSession(item string) {
	if len(m.practiceOrder) == 0 {
		m.finishSession()
		return
	}

	if m.completedCount >= len(m.practiceOrder) {
		m.finishSession()
		return
	}

	actualIndex := m.practiceOrder[m.completedCount]
	if m.srsEnabled && item != "" {
		_ = m.srsSchedule.RemoveItem(item)
	}
	if actualIndex >= 0 && actualIndex < len(m.items) {
		m.items = append(m.items[:actualIndex], m.items[actualIndex+1:]...)
	}

	m.practiceOrder = append(m.practiceOrder[:m.completedCount], m.practiceOrder[m.completedCount+1:]...)
	for i := range m.practiceOrder {
		if m.practiceOrder[i] > actualIndex {
			m.practiceOrder[i]--
		}
	}

	if len(m.practiceOrder) == 0 || m.completedCount >= len(m.practiceOrder) {
		m.finishSession()
	}
}

func (m *PracticeSession) advanceToNextItem() {
	if len(m.practiceOrder) == 0 {
		m.finishSession()
		return
	}

	m.completedCount++
	if m.completedCount >= len(m.practiceOrder) {
		m.finishSession()
	}
}

func (m *PracticeSession) finishSession() {
	if m.state == "finished" {
		return
	}

	m.state = "finished"
	m.endTime = time.Now()
	m.result = m.calculateResult()
	sound.StopAllSounds()
	m.logStatistics(true)
}

func (m *PracticeSession) logStatistics(completed bool) {
	if m.statsLogged {
		return
	}

	total := m.correct + m.incorrect
	if total == 0 && !completed {
		return
	}

	duration := time.Since(m.startTime)
	if !m.endTime.IsZero() {
		duration = m.endTime.Sub(m.startTime)
	}
	if duration < 0 {
		duration = 0
	}

	accuracy := 0.0
	if total > 0 {
		accuracy = float64(m.correct) / float64(total) * 100
	}

	record := statistics.SessionRecord{
		Timestamp:       time.Now(),
		ResourceType:    m.resourceType,
		FileName:        m.fileName,
		Total:           total,
		Correct:         m.correct,
		Incorrect:       m.incorrect,
		Accuracy:        accuracy,
		DurationSeconds: int64(duration.Seconds()),
		OrderMode:       m.orderMode,
		Completed:       completed,
	}

	if err := statistics.LogSession(record); err != nil {
		// 统计记录失败时不阻断用户流程，仅在命令反馈中提示
		m.commandFeedback = fmt.Sprintf("记录统计数据失败: %v", err)
		m.commandFeedbackIsError = true
	}

	m.statsLogged = true
}

// View 渲染视图
func (m PracticeSession) View() string {
	if m.quitting {
		return "练习已中断！"
	}

	var s strings.Builder

	// 标题
	title := fmt.Sprintf("%s练习 - %s", getResourceTypeTitle(m.resourceType), m.displayFileName)
	s.WriteString(RenderTitle(title) + "\n\n")

	if m.state == "practicing" {
		total := len(m.practiceOrder)
		currentPosition := 0
		if total > 0 {
			currentPosition = m.completedCount + 1
			if currentPosition > total {
				currentPosition = total
			}
		}

		progressValue := 1.0
		if total > 0 {
			progressValue = float64(m.completedCount) / float64(total)
		}

		progressText := fmt.Sprintf("进度: %d/%d", currentPosition, total)
		s.WriteString(RenderText(progressText) + "\n")
		s.WriteString(m.progress.ViewAs(progressValue) + "\n\n")

		currentItem := m.getCurrentItem()
		if currentItem != "" {
			s.WriteString(RenderHighlight("当前项目:") + "\n")
			wrappedText := m.wrapText(currentItem, m.width-4)
			s.WriteString(RenderText(wrappedText) + "\n\n")
		} else {
			s.WriteString(RenderText("暂无可练习内容") + "\n\n")
		}

		s.WriteString(RenderHighlight("请输入:") + "\n")
		m.applyInputHighlight()
		s.WriteString(m.textInput.View() + "\n")
		dropdown := m.renderCommandDropdown()
		if dropdown != "" {
			s.WriteString(dropdown + "\n\n")
		} else {
			s.WriteString("\n")
		}

		if m.lastInputWrong {
			s.WriteString(RenderError("❌ 输入错误！") + "\n")
			s.WriteString(m.renderWordLevelError() + "\n")
		}

		if m.commandFeedback != "" {
			if m.commandFeedbackIsError {
				s.WriteString(RenderError(m.commandFeedback) + "\n\n")
			} else {
				s.WriteString(RenderSuccess(m.commandFeedback) + "\n\n")
			}
		}

		s.WriteString(RenderText(`按 Enter 提交，按 Esc 退出练习，输入"> help"获取命令说明`) + "\n\n")
		if m.inCommandMode {
			helpSummary := m.renderCommandHelpSummary()
			if helpSummary != "" {
				s.WriteString(helpSummary + "\n")
			}
		}
	} else if m.state == "finished" {
		if len(m.practiceOrder) == 0 && len(m.items) == 0 {
			s.WriteString(RenderHighlight("暂无练习内容") + "\n\n")
		} else {
			s.WriteString(RenderSuccess("练习完成！") + "\n\n")
		}
		s.WriteString(RenderText(m.result) + "\n\n")
		s.WriteString(RenderText("按 Enter 或 Esc 返回练习菜单") + "\n")
	}

	return s.String()
}

func (m PracticeSession) renderCommandDropdown() string {
	if !m.commandDropdownVisible || len(m.filteredCommands) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(RenderHighlight("命令建议:") + "\n")
	for idx, option := range m.filteredCommands {
		line := fmt.Sprintf("  %s - %s", option.name, option.description)
		if idx == m.selectedCommandIndex {
			builder.WriteString(RenderHighlight("➤ " + option.name + " - " + option.description))
		} else {
			builder.WriteString(RenderText(line))
		}
		builder.WriteString("\n")
	}

	return strings.TrimRight(builder.String(), "\n")
}

func (m PracticeSession) renderCommandHelpSummary() string {
	if len(m.commandOptions) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(RenderHighlight("命令说明:") + "\n")
	for _, option := range m.commandOptions {
		line := fmt.Sprintf("> %s - %s", option.name, option.description)
		builder.WriteString(RenderText(line) + "\n")
	}

	return strings.TrimRight(builder.String(), "\n")
}

func (m PracticeSession) renderCommandHelpDetail() string {
	if len(m.commandOptions) == 0 {
		return "暂无可用命令"
	}

	var builder strings.Builder
	builder.WriteString("可用命令:\n")
	for _, option := range m.commandOptions {
		builder.WriteString(fmt.Sprintf("  > %s - %s\n", option.name, option.description))
	}
	builder.WriteString(`提示：输入"> "后可使用↑↓选择命令，Tab 自动填充。`)
	return builder.String()
}

func (m *PracticeSession) applyInputHighlight() {
	if m.state != "practicing" {
		m.resetInputHighlight()
		return
	}

	value := m.textInput.Value()
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		m.resetInputHighlight()
		return
	}

	if strings.HasPrefix(trimmed, ">") {
		m.resetInputHighlight()
		return
	}

	item := m.getCurrentRawItem()
	if item == "" {
		m.resetInputHighlight()
		return
	}

	expected := strings.TrimSpace(m.getExpectedInput(item))
	if expected == "" {
		m.resetInputHighlight()
		return
	}

	if m.shouldHighlightInputMismatch(value, expected) {
		m.textInput.TextStyle = ErrorStyle
		m.textInput.CursorStyle = ErrorStyle
	} else {
		m.resetInputHighlight()
	}
}

func (m *PracticeSession) resetInputHighlight() {
	m.textInput.TextStyle = m.inputDefaultTextStyle
	m.textInput.CursorStyle = m.inputDefaultCursorStyle
}

func (m *PracticeSession) shouldHighlightInputMismatch(value, expected string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}

	matchMode := strings.ToLower(config.AppConfig.CorrectnessMatchMode)
	switch matchMode {
	case "word_match":
		expectedPrefix := m.normalizeForWordMatch(expected)
		inputPrefix := m.normalizeForWordMatch(value)
		if inputPrefix == "" {
			return false
		}
		return !strings.HasPrefix(expectedPrefix, inputPrefix)
	default:
		return !strings.HasPrefix(expected, value)
	}
}

// 获取当前项目
func (m PracticeSession) getCurrentItem() string {
	item := m.getCurrentRawItem()
	if item == "" {
		return ""
	}

	primary, translation := practice.ParseLine(item)
	primary = strings.TrimSpace(primary)
	translation = strings.TrimSpace(translation)

	if !m.getShowTranslationConfig() || translation == "" {
		if primary != "" {
			return primary
		}
		return item
	}

	if primary == "" {
		primary = item
	}

	return primary + "\n" + translation
}

func (m PracticeSession) getCurrentRawItem() string {
	if m.completedCount < 0 || m.completedCount >= len(m.practiceOrder) {
		return ""
	}

	actualIndex := m.practiceOrder[m.completedCount]
	if actualIndex < 0 || actualIndex >= len(m.items) {
		return ""
	}

	return m.items[actualIndex]
}

// 获取翻译显示配置
func (m PracticeSession) getShowTranslationConfig() bool {
	// v0.3 优化：使用全局翻译设置
	return config.AppConfig.ShowTranslation
}

// 获取期望输入
func (m PracticeSession) getExpectedInput(item string) string {
	// 对于所有资源类型，使用ParseLine函数正确解析多种分隔符，只返回正文部分
	content, _ := practice.ParseLine(item)
	if content != "" {
		return content
	}
	return item
}

func (m PracticeSession) isInputCorrectForItem(userInput, item string) bool {
	acceptedInputs := m.getAcceptedInputs(item)
	for _, expected := range acceptedInputs {
		if expected == "" {
			continue
		}
		if m.isInputCorrect(userInput, expected) {
			return true
		}
	}
	return false
}

func (m PracticeSession) getAcceptedInputs(item string) []string {
	expected := strings.TrimSpace(m.getExpectedInput(item))
	trimmedItem := strings.TrimSpace(item)
	_, translation := practice.ParseLine(item)
	translation = strings.TrimSpace(translation)

	acceptedSet := make(map[string]struct{})
	if expected != "" {
		acceptedSet[expected] = struct{}{}
	}
	if trimmedItem != "" {
		acceptedSet[trimmedItem] = struct{}{}
	}

	if expected != "" && translation != "" {
		withSpace := strings.TrimSpace(expected + " " + translation)
		if withSpace != "" {
			acceptedSet[withSpace] = struct{}{}
		}
		withArrow := strings.TrimSpace(expected + " ->> " + translation)
		if withArrow != "" {
			acceptedSet[withArrow] = struct{}{}
		}
	}

	if len(acceptedSet) == 0 && trimmedItem != "" {
		return []string{trimmedItem}
	}

	result := make([]string, 0, len(acceptedSet))
	for candidate := range acceptedSet {
		result = append(result, candidate)
	}
	return result
}

// 检查输入是否正确
func (m PracticeSession) isInputCorrect(userInput, expectedInput string) bool {
	// 获取匹配模式
	matchMode := config.AppConfig.CorrectnessMatchMode
	if matchMode == "" {
		matchMode = "exact_match" // 默认值
	}

	switch matchMode {
	case "exact_match":
		// 完全匹配
		return userInput == expectedInput
	case "word_match":
		// 单词匹配，忽略大小写和标点符号
		return m.normalizeForWordMatch(userInput) == m.normalizeForWordMatch(expectedInput)
	default:
		// 默认使用完全匹配
		return userInput == expectedInput
	}
}

// 标准化文本用于单词匹配
func (m PracticeSession) normalizeForWordMatch(text string) string {
	// 转换为小写
	text = strings.ToLower(text)
	// 移除标点符号，只保留字母、数字和空格
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	text = reg.ReplaceAllString(text, "")
	// 移除多余的空格
	text = strings.TrimSpace(text)
	reg = regexp.MustCompile(`\s+`)
	text = reg.ReplaceAllString(text, " ")
	return text
}

// 渲染单词级别的错误高亮
func (m PracticeSession) renderWordLevelError() string {
	userWords := strings.Fields(m.wrongInput)
	expectedWords := strings.Fields(m.expectedText)

	var result strings.Builder
	result.WriteString(RenderError("你的输入: "))

	// 构建用户输入的高亮文本
	var userInputParts []string
	maxLen := len(userWords)
	if len(expectedWords) > maxLen {
		maxLen = len(expectedWords)
	}

	for i := 0; i < maxLen; i++ {
		if i < len(userWords) {
			userWord := userWords[i]
			expectedWord := ""
			if i < len(expectedWords) {
				expectedWord = expectedWords[i]
			}

			// 如果单词不匹配，用红色高亮
			if userWord != expectedWord {
				userInputParts = append(userInputParts, RenderError(userWord))
			} else {
				userInputParts = append(userInputParts, userWord)
			}
		}
	}

	// 将用户输入拼接并换行
	userInputText := strings.Join(userInputParts, " ")
	wrappedUserInput := m.wrapText(userInputText, m.width-4) // 减去边距
	result.WriteString(wrappedUserInput)

	result.WriteString("\n")

	// 正确答案也需要换行，只显示正文部分
	correctAnswerText := "正确答案: " + m.expectedText
	wrappedCorrectAnswer := m.wrapText(correctAnswerText, m.width-4) // 减去边距
	result.WriteString(RenderSuccess(wrappedCorrectAnswer))

	return result.String()
}

// 计算练习结果
func (m PracticeSession) calculateResult() string {
	// 计算练习时间
	duration := m.endTime.Sub(m.startTime)
	durationStr := fmt.Sprintf("%d分%d秒", int(duration.Minutes()), int(duration.Seconds())%60)

	// 计算正确率
	total := m.correct + m.incorrect
	accuracy := 0.0
	if total > 0 {
		accuracy = float64(m.correct) / float64(total) * 100
	}

	// 计算速度（每分钟字符数）
	totalChars := 0
	for _, item := range m.items {
		expectedInput := m.getExpectedInput(item)
		totalChars += len(expectedInput)
	}

	cpm := 0.0
	if duration.Minutes() > 0 {
		cpm = float64(totalChars) / duration.Minutes()
	}

	return fmt.Sprintf(
		"练习时间: %s\n正确数量: %d\n错误数量: %d\n正确率: %.1f%%\n打字速度: %.1f CPM (每分钟字符数)",
		durationStr, m.correct, m.incorrect, accuracy, cpm,
	)
}

// wrapText 将文本按指定宽度换行
func (m PracticeSession) wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		// 如果当前行为空，直接添加单词
		if currentLine == "" {
			currentLine = word
		} else {
			// 检查添加新单词后是否超过宽度
			testLine := currentLine + " " + word
			if len(testLine) <= width {
				currentLine = testLine
			} else {
				// 超过宽度，将当前行添加到结果中，开始新行
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
	}

	// 添加最后一行
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}
