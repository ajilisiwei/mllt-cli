package practice

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/ajilisiwei/mllt-cli/internal/config"
)

// 资源类型
const (
	Words     = "words"
	Phrases   = "phrases"
	Sentences = "sentences"
	Articles  = "articles"
)

// 资源文件夹常量
const (
	DefaultFolderDir     = "default"
	DefaultFolderDisplay = "默认"
)

// ResourceFolder 表示一个资源文件夹及其文件
type ResourceFolder struct {
	DirName     string
	DisplayName string
	Files       []string
}

// 分隔符
const (
	Separator = " ->> "
)

// 初始化随机数生成器
func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetResourcePath 获取资源文件路径
// 优先返回用户资源路径，若不存在则回退到基础资源目录
func normalizeFolderDir(folder string) string {
	sanitized, _ := sanitizeFolderName(folder)
	return sanitized
}

// NormalizeFolderName 对用户输入的文件夹名进行规范化处理，返回规范化后的名称和是否发生了修改
func NormalizeFolderName(folder string) (string, bool) {
	return sanitizeFolderName(folder)
}

var windowsReservedNames = map[string]struct{}{
	"con": {}, "prn": {}, "aux": {}, "nul": {},
	"com1": {}, "com2": {}, "com3": {}, "com4": {}, "com5": {}, "com6": {}, "com7": {}, "com8": {}, "com9": {},
	"lpt1": {}, "lpt2": {}, "lpt3": {}, "lpt4": {}, "lpt5": {}, "lpt6": {}, "lpt7": {}, "lpt8": {}, "lpt9": {},
}

const invalidFolderChars = "<>:\"/\\|?*"

func sanitizeFolderName(folder string) (string, bool) {
	original := folder
	folder = strings.TrimSpace(folder)
	folder = strings.ReplaceAll(folder, "..", "")
	folder = strings.ReplaceAll(folder, "\\", "/")
	folder = strings.Trim(folder, "/")

	if folder == "" || folder == "." || folder == DefaultFolderDisplay {
		return DefaultFolderDir, !strings.EqualFold(strings.TrimSpace(original), DefaultFolderDir)
	}

	replaced := false
	var builder strings.Builder
	for _, r := range folder {
		if isInvalidFolderRune(r) {
			replaced = true
			builder.WriteRune('_')
			continue
		}
		builder.WriteRune(r)
	}

	sanitized := strings.TrimSpace(builder.String())

	if runtime.GOOS == "windows" {
		sanitized = strings.Trim(sanitized, ". ")
		lower := strings.ToLower(sanitized)
		if _, reserved := windowsReservedNames[lower]; reserved {
			replaced = true
			sanitized = "_" + sanitized
		}
	}

	if sanitized == "" {
		return DefaultFolderDir, true
	}

	return sanitized, replaced || sanitized != folder
}

func isInvalidFolderRune(r rune) bool {
	if r == 0 || r == '/' {
		return true
	}
	if unicode.IsControl(r) {
		return true
	}
	if strings.ContainsRune(invalidFolderChars, r) {
		return true
	}
	return false
}

func FolderDisplayName(dir string) string {
	if dir == "" || dir == DefaultFolderDir {
		return DefaultFolderDisplay
	}
	return dir
}

func splitResourceIdentifier(name string) (string, string) {
	trimmed := strings.TrimSpace(name)
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = strings.TrimSuffix(trimmed, ".txt")
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return DefaultFolderDir, ""
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) == 1 {
		return DefaultFolderDir, parts[0]
	}
	folder := normalizeFolderDir(parts[0])
	fileName := strings.Join(parts[1:], "/")
	return folder, fileName
}

func BuildResourceIdentifier(folderDir, fileName string) string {
	folderDir = normalizeFolderDir(folderDir)
	fileName = strings.TrimSuffix(fileName, ".txt")
	fileName = strings.TrimSpace(fileName)
	fileName = strings.Trim(fileName, "/")
	if fileName == "" {
		return folderDir
	}
	if folderDir == "" || folderDir == DefaultFolderDir {
		return fileName
	}
	return folderDir + "/" + fileName
}

func FormatResourceDisplayName(identifier string) string {
	folderDir, fileName := splitResourceIdentifier(identifier)
	displayFolder := FolderDisplayName(folderDir)
	if fileName == "" {
		return displayFolder
	}
	if folderDir == DefaultFolderDir {
		return fileName
	}
	return displayFolder + "/" + fileName
}

func GetResourcePath(resourceType string, fileName string) string {
	currentLanguage := config.AppConfig.CurrentLanguage
	folderDir, baseName := splitResourceIdentifier(fileName)
	if baseName == "" {
		baseName = folderDir
		folderDir = DefaultFolderDir
	}
	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		baseName = "resource"
	}
	if !strings.HasSuffix(baseName, ".txt") {
		baseName += ".txt"
	}

	userRoot := filepath.Join(getUserDataBaseDir(), currentLanguage, resourceType)
	baseRoot := filepath.Join(getResourceBaseDir(), currentLanguage, resourceType)

	folderDir = normalizeFolderDir(folderDir)

	userPath := filepath.Join(userRoot, folderDir, baseName)
	if _, err := os.Stat(userPath); err == nil {
		return userPath
	}

	basePath := filepath.Join(baseRoot, folderDir, baseName)
	if _, err := os.Stat(basePath); err == nil {
		return basePath
	}

	if folderDir == DefaultFolderDir {
		legacyUserPath := filepath.Join(userRoot, baseName)
		if _, err := os.Stat(legacyUserPath); err == nil {
			return legacyUserPath
		}
		legacyBasePath := filepath.Join(baseRoot, baseName)
		if _, err := os.Stat(legacyBasePath); err == nil {
			return legacyBasePath
		}
	}

	return userPath
}

// getResourceBaseDir 获取资源基础目录
func getResourceBaseDir() string {
	// 检查是否在测试环境中
	if isTestEnvironment() {
		return "resources"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 如果无法获取用户主目录，使用当前目录下的resources
		return "resources"
	}
	return filepath.Join(homeDir, ".mllt-cli", "resources")
}

func getUserDataBaseDir() string {
	if isTestEnvironment() {
		return filepath.Join("resources", "user-data")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("resources", "user-data")
	}
	return filepath.Join(homeDir, ".mllt-cli", "user-data")
}

// GetUserDataDir 返回用于存储用户练习数据的目录
func GetUserDataDir() string {
	return getUserDataBaseDir()
}

// isTestEnvironment 检查是否在测试环境中
func isTestEnvironment() bool {
	if os.Getenv("MLLTCLI_TEST") == "1" {
		return true
	}
	base := filepath.Base(os.Args[0])
	return strings.HasSuffix(base, ".test")
}

// GetResourceFiles 获取指定类型的资源文件列表
func GetResourceFolders(resourceType string) ([]ResourceFolder, error) {
	currentLanguage := config.AppConfig.CurrentLanguage
	baseRoot := filepath.Join(getResourceBaseDir(), currentLanguage, resourceType)
	userRoot := filepath.Join(getUserDataBaseDir(), currentLanguage, resourceType)

	folderMap := make(map[string]map[string]struct{})

	collect := func(root string) error {
		entries, err := os.ReadDir(root)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				folderDir := normalizeFolderDir(name)
				files, err := os.ReadDir(filepath.Join(root, name))
				if err != nil {
					return err
				}
				for _, f := range files {
					if f.IsDir() {
						continue
					}
					fname := strings.TrimSuffix(f.Name(), ".txt")
					if fname == "" {
						continue
					}
					if _, ok := folderMap[folderDir]; !ok {
						folderMap[folderDir] = make(map[string]struct{})
					}
					folderMap[folderDir][fname] = struct{}{}
				}
			} else {
				if !strings.HasSuffix(name, ".txt") {
					continue
				}
				fname := strings.TrimSuffix(name, ".txt")
				if fname == "" {
					continue
				}
				if _, ok := folderMap[DefaultFolderDir]; !ok {
					folderMap[DefaultFolderDir] = make(map[string]struct{})
				}
				folderMap[DefaultFolderDir][fname] = struct{}{}
			}
		}
		return nil
	}

	if err := collect(baseRoot); err != nil {
		return nil, fmt.Errorf("读取基础资源目录失败: %w", err)
	}
	if err := collect(userRoot); err != nil {
		return nil, fmt.Errorf("读取用户资源目录失败: %w", err)
	}

	if len(folderMap) == 0 {
		folderMap[DefaultFolderDir] = make(map[string]struct{})
	}

	var folders []ResourceFolder
	for dir, filesMap := range folderMap {
		var files []string
		for name := range filesMap {
			files = append(files, name)
		}
		sort.Strings(files)
		folders = append(folders, ResourceFolder{
			DirName:     dir,
			DisplayName: FolderDisplayName(dir),
			Files:       files,
		})
	}

	sort.Slice(folders, func(i, j int) bool {
		if folders[i].DirName == DefaultFolderDir {
			return true
		}
		if folders[j].DirName == DefaultFolderDir {
			return false
		}
		return folders[i].DirName < folders[j].DirName
	})

	return folders, nil
}

func GetResourceFiles(resourceType string) ([]string, error) {
	currentLanguage := config.AppConfig.CurrentLanguage
	baseDir := filepath.Join(getResourceBaseDir(), currentLanguage, resourceType)
	userDir := filepath.Join(getUserDataBaseDir(), currentLanguage, resourceType)

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("创建基础资源目录失败: %w", err)
	}
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("创建用户资源目录失败: %w", err)
	}

	folders, err := GetResourceFolders(resourceType)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, folder := range folders {
		for _, file := range folder.Files {
			if folder.DirName == DefaultFolderDir {
				fileNames = append(fileNames, file)
			} else {
				fileNames = append(fileNames, folder.DirName+"/"+file)
			}
		}
	}

	sort.Strings(fileNames)

	return fileNames, nil
}

// ReadResourceFile 读取资源文件内容
func ReadResourceFile(resourceType string, fileName string) ([]string, error) {
	filePath, err := resolveReadableResourcePath(resourceType, fileName)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 对于非单词类型的资源，尝试解析换行符分隔格式
	if resourceType != Words {
		if lines, err := parseNewlineFormat(file); err == nil && len(lines) > 0 {
			return lines, nil
		}
		// 如果解析换行符格式失败，回退到原来的逐行解析
		file.Seek(0, 0) // 重置文件指针
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return lines, nil
}

// WriteResourceFile 将内容写入资源文件（覆盖写入）
func WriteResourceFile(resourceType string, fileName string, lines []string) error {
	filePath := getWritableResourcePath(resourceType, fileName)

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("创建资源目录失败: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("打开资源文件失败: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("写入资源文件失败: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("刷新资源文件失败: %w", err)
	}

	return nil
}

func resolveReadableResourcePath(resourceType, fileName string) (string, error) {
	currentLanguage := config.AppConfig.CurrentLanguage
	folderDir, baseName := splitResourceIdentifier(fileName)
	if baseName == "" {
		baseName = folderDir
		folderDir = DefaultFolderDir
	}
	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		return "", fmt.Errorf("无效的资源名称")
	}
	if !strings.HasSuffix(baseName, ".txt") {
		baseName += ".txt"
	}

	folderDir = normalizeFolderDir(folderDir)
	userRoot := filepath.Join(getUserDataBaseDir(), currentLanguage, resourceType)
	baseRoot := filepath.Join(getResourceBaseDir(), currentLanguage, resourceType)

	searchPaths := []string{
		filepath.Join(userRoot, folderDir, baseName),
		filepath.Join(baseRoot, folderDir, baseName),
	}
	if folderDir == DefaultFolderDir {
		searchPaths = append(searchPaths,
			filepath.Join(userRoot, baseName),
			filepath.Join(baseRoot, baseName),
		)
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	targetDir := filepath.Join(userRoot, folderDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("创建用户资源目录失败: %w", err)
	}

	path := filepath.Join(targetDir, baseName)
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("创建资源文件失败: %w", err)
	}
	file.Close()

	return path, nil
}

func getWritableResourcePath(resourceType, fileName string) string {
	currentLanguage := config.AppConfig.CurrentLanguage
	folderDir, baseName := splitResourceIdentifier(fileName)
	if baseName == "" {
		baseName = folderDir
		folderDir = DefaultFolderDir
	}
	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		baseName = "resource"
	}
	if !strings.HasSuffix(baseName, ".txt") {
		baseName += ".txt"
	}
	folderDir = normalizeFolderDir(folderDir)

	dir := filepath.Join(getUserDataBaseDir(), currentLanguage, resourceType, folderDir)
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, baseName)
}

// parseNewlineFormat 解析换行符分隔格式的资源文件
// 格式：原文\n翻译\n\n（空行分隔不同的条目）
func parseNewlineFormat(file *os.File) ([]string, error) {
	var lines []string
	var currentOriginal string
	var currentTranslation string
	var lineCount int

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		if line == "" {
			// 遇到空行，如果有当前条目，则保存
			if currentOriginal != "" {
				if currentTranslation != "" {
					// 有翻译的情况
					lines = append(lines, currentOriginal+" ->> "+currentTranslation)
				} else {
					// 没有翻译的情况
					lines = append(lines, currentOriginal)
				}
				currentOriginal = ""
				currentTranslation = ""
			}
		} else if currentOriginal == "" {
			if containsInlineSeparator(line) {
				return nil, fmt.Errorf("检测到行内分隔符，回退到逐行解析")
			}
			// 第一行是原文
			currentOriginal = line
		} else if currentTranslation == "" {
			// 第二行是翻译
			currentTranslation = line
		} else {
			// 如果已经有原文和翻译，但还有内容，说明格式不对
			// 回退到普通格式解析
			return nil, fmt.Errorf("格式不符合换行符分隔规则")
		}
	}

	// 处理文件末尾的条目
	if currentOriginal != "" {
		if currentTranslation != "" {
			lines = append(lines, currentOriginal+" ->> "+currentTranslation)
		} else {
			lines = append(lines, currentOriginal)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 如果没有解析到任何内容，或者格式不符合要求，返回错误
	if len(lines) == 0 {
		return nil, fmt.Errorf("没有找到符合换行符分隔格式的内容")
	}

	return lines, nil
}

func containsInlineSeparator(line string) bool {
	if strings.Contains(line, " ->> ") {
		return true
	}
	return false
}

// ParseLine 解析行内容，返回原文和翻译
// 支持多种分隔符，按优先级顺序：" ->> ", 制表符, 空格, "/", ":", "："
func ParseLine(line string) (string, string) {
	// 首先检查 " ->> " 分隔符
	if strings.Contains(line, " ->> ") {
		parts := strings.SplitN(line, " ->> ", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}

	// 然后检查制表符分隔符
	// 第一个制表符之前的文本作为原文，之后的作为翻译
	tabIndex := strings.Index(line, "\t")
	if tabIndex > 0 {
		return strings.TrimSpace(line[:tabIndex]), strings.TrimSpace(line[tabIndex+1:])
	}

	// 然后检查空格分隔符
	// 仅当空格后的文本包含非ASCII字符时，才认为它是翻译（避免纯英文句子被拆分）
	spaceIndex := strings.Index(line, " ")
	if spaceIndex > 0 {
		original := strings.TrimSpace(line[:spaceIndex])
		translation := strings.TrimSpace(line[spaceIndex+1:])
		if translation != "" && containsNonASCII(translation) {
			return original, translation
		}
	}

	// 定义其他分隔符优先级列表
	separators := []string{"/", ":", "："}

	// 按优先级尝试其他分隔符
	for _, sep := range separators {
		if strings.Contains(line, sep) {
			parts := strings.SplitN(line, sep, 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			}
		}
	}

	// 如果没有找到任何分隔符，整行作为原文，翻译为空
	return line, ""
}

// containsNonASCII 判断字符串中是否包含非ASCII字符
func containsNonASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return true
		}
	}
	return false
}

// GetNextIndex 获取下一个索引
func GetNextIndex(currentIndex, total int, order string) int {
	if order == "sequential" {
		return (currentIndex + 1) % total
	}
	// 随机模式
	return rand.Intn(total)
}
