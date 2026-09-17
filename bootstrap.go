package mlltcli

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	appDirName      = ".mllt-cli"
	configFileName  = "config.yaml"
	resourcesDir    = "resources"
	assetsDir       = "assets"
	userDataDirName = "user-data"

	// resourceVersionFile 标记内置资源的版本，内容变化时触发一次覆盖式同步。
	resourceVersionFile = "VERSION"
	// resourceBackupDirName 存放覆盖前的旧资源，避免用户导入的内容被冲掉。
	resourceBackupDirName = "resources-backup"
)

//go:embed config/config.yaml
var defaultConfig []byte

//go:embed resources assets
var bundledFS embed.FS

// EnsureAssets ensures that default configuration, assets, and bundled
// resources are available in the user's application directory. Existing files
// are kept intact to avoid overwriting user-imported content.
func EnsureAssets() error {
	if isTestEnvironment() {
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %w", err)
	}

	baseDir := filepath.Join(homeDir, appDirName)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("创建基础目录失败: %w", err)
	}

	if err := ensureUserData(baseDir); err != nil {
		return err
	}

	if err := ensureConfig(baseDir); err != nil {
		return err
	}

	if err := syncBundledResources(baseDir); err != nil {
		return err
	}

	if err := copyEmbeddedTree(assetsDir, filepath.Join(baseDir, assetsDir)); err != nil {
		return err
	}

	return nil
}

func ensureConfig(baseDir string) error {
	targetPath := filepath.Join(baseDir, configFileName)
	if _, err := os.Stat(targetPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查配置文件失败: %w", err)
	}

	if err := os.WriteFile(targetPath, defaultConfig, 0o644); err != nil {
		return fmt.Errorf("写入默认配置失败: %w", err)
	}
	return nil
}

func ensureUserData(baseDir string) error {
	userDataDir := filepath.Join(baseDir, userDataDirName)
	if err := os.MkdirAll(userDataDir, 0o755); err != nil {
		return fmt.Errorf("创建用户数据目录失败: %w", err)
	}
	return nil
}

func copyEmbeddedTree(root, dest string) error {
	return fs.WalkDir(bundledFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == root {
			return os.MkdirAll(dest, 0o755)
		}

		rel := strings.TrimPrefix(path, root+"/")
		rel = filepath.FromSlash(rel)
		targetPath := filepath.Join(dest, rel)

		if d.IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
			return nil
		}

		if _, err := os.Stat(targetPath); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}

		srcFile, err := bundledFS.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("创建文件夹失败: %w", err)
		}

		destFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, srcFile); err != nil {
			return err
		}

		return nil
	})
}

func isTestEnvironment() bool {
	if os.Getenv("MLLTCLI_TEST") == "1" {
		return true
	}
	exeName := filepath.Base(os.Args[0])
	return strings.HasSuffix(exeName, ".test")
}
