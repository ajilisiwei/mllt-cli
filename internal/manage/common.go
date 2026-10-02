package manage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ajilisiwei/mllt-cli/internal/config"
	"github.com/ajilisiwei/mllt-cli/internal/practice"
)

// 资源类型
const (
	Words     = "words"
	Phrases   = "phrases"
	Sentences = "sentences"
	Dialogues = "dialogues"
	Articles  = "articles"
)

// GetResourcePath 获取资源文件完整路径（支持文件夹）
func GetResourcePath(resourceType, resourceIdentifier string) string {
	return practice.GetResourcePath(resourceType, resourceIdentifier)
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

// isTestEnvironment 检查是否在测试环境中
func isTestEnvironment() bool {
	if os.Getenv("MLLTCLI_TEST") == "1" {
		return true
	}
	base := filepath.Base(os.Args[0])
	return strings.HasSuffix(base, ".test")
}

// GetResourceFiles 获取指定类型的资源文件列表
func GetResourceFiles(resourceType string) ([]string, error) {
	folders, err := ListResourceFolders(resourceType)
	if err != nil {
		return nil, err
	}

	var flattened []string
	for _, folder := range folders {
		for _, file := range folder.Files {
			if folder.DirName == practice.DefaultFolderDir {
				flattened = append(flattened, file)
			} else {
				flattened = append(flattened, folder.DirName+"/"+file)
			}
		}
	}

	return flattened, nil
}

// ListResourceFolders 返回包含文件的资源文件夹列表
func ListResourceFolders(resourceType string) ([]practice.ResourceFolder, error) {
	if !ValidateResourceType(resourceType) {
		return nil, fmt.Errorf("无效的资源类型: %s", resourceType)
	}

	return practice.GetResourceFolders(resourceType)
}

// GetResourceFolder 根据目录名获取单个文件夹信息
func GetResourceFolder(resourceType, folderDir string) (*practice.ResourceFolder, error) {
	folders, err := ListResourceFolders(resourceType)
	if err != nil {
		return nil, err
	}

	normalized, _ := practice.NormalizeFolderName(folderDir)
	for _, folder := range folders {
		if normExisting, _ := practice.NormalizeFolderName(folder.DirName); normExisting == normalized {
			copy := folder
			return &copy, nil
		}
	}

	return nil, fmt.Errorf("未找到文件夹: %s", folderDir)
}

// DeleteResourceFolder 删除指定资源文件夹（仅限用户自建文件夹）
func DeleteResourceFolder(resourceType, folderDir string) error {
	normalized, _ := practice.NormalizeFolderName(folderDir)
	if normalized == practice.DefaultFolderDir {
		return fmt.Errorf("无法删除默认文件夹")
	}

	if !ValidateResourceType(resourceType) {
		return fmt.Errorf("无效的资源类型: %s", resourceType)
	}

	currentLanguage := config.AppConfig.CurrentLanguage
	userFolderPath := filepath.Join(getUserDataBaseDir(), currentLanguage, resourceType, normalized)
	baseFolderPath := filepath.Join(getResourceBaseDir(), currentLanguage, resourceType, normalized)

	// 检查基础资源中是否存在内容
	if entries, err := os.ReadDir(baseFolderPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				return fmt.Errorf("文件夹中仍包含系统资源，无法删除")
			}
		}
	}

	entries, err := os.ReadDir(userFolderPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("文件夹不存在或已删除")
		}
		return fmt.Errorf("读取文件夹失败: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("文件夹非空，无法删除")
		}
	}

	if err := os.Remove(userFolderPath); err != nil {
		return fmt.Errorf("删除文件夹失败: %w", err)
	}

	return nil
}

// ValidateResourceType 验证资源类型是否有效
func ValidateResourceType(resourceType string) bool {
	switch resourceType {
	case Words, Phrases, Sentences, Dialogues, Articles:
		return true
	default:
		return false
	}
}
