package mlltcli

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// syncBundledResources 把内置资源同步到用户目录。
//
// 内置资源带有版本标记（resources/VERSION）：版本未变时只补齐缺失的文件，
// 保持既有行为；版本变化时会覆盖内容不同的同名文件，让词库修复能到达老用户。
// 由于导入功能可能直接写入内置资源文件，覆盖前一律先备份，避免用户内容丢失。
func syncBundledResources(baseDir string) error {
	dest := filepath.Join(baseDir, resourcesDir)

	want, err := embeddedResourceVersion()
	if err != nil {
		// 没有版本标记时退回到「只补缺失文件」的旧行为
		return copyEmbeddedTree(resourcesDir, dest)
	}

	if installedResourceVersion(dest) == want {
		return copyEmbeddedTree(resourcesDir, dest)
	}

	backupDir := filepath.Join(baseDir, resourceBackupDirName, time.Now().Format("20060102-150405"))
	if err := refreshEmbeddedTree(resourcesDir, dest, backupDir); err != nil {
		return err
	}

	versionPath := filepath.Join(dest, resourceVersionFile)
	if err := os.WriteFile(versionPath, []byte(want+"\n"), 0o644); err != nil {
		return fmt.Errorf("写入资源版本标记失败: %w", err)
	}

	return nil
}

func embeddedResourceVersion() (string, error) {
	data, err := fs.ReadFile(bundledFS, resourcesDir+"/"+resourceVersionFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func installedResourceVersion(dest string) string {
	data, err := os.ReadFile(filepath.Join(dest, resourceVersionFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// refreshEmbeddedTree 覆盖内容已过期的内置资源，覆盖前把原文件备份到 backupDir。
func refreshEmbeddedTree(root, dest, backupDir string) error {
	return fs.WalkDir(bundledFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return os.MkdirAll(dest, 0o755)
		}

		rel := filepath.FromSlash(strings.TrimPrefix(path, root+"/"))
		targetPath := filepath.Join(dest, rel)

		if d.IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
			return nil
		}

		bundled, err := fs.ReadFile(bundledFS, path)
		if err != nil {
			return err
		}

		current, err := os.ReadFile(targetPath)
		switch {
		case err == nil:
			if bytes.Equal(current, bundled) {
				return nil
			}
			if err := backupFile(filepath.Join(backupDir, rel), current); err != nil {
				return err
			}
		case !os.IsNotExist(err):
			return err
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("创建文件夹失败: %w", err)
		}
		if err := os.WriteFile(targetPath, bundled, 0o644); err != nil {
			return fmt.Errorf("更新资源文件失败: %w", err)
		}
		return nil
	})
}

func backupFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建资源备份目录失败: %w", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("备份资源文件失败: %w", err)
	}
	return nil
}
