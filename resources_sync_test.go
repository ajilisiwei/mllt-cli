package mlltcli

import (
	"os"
	"path/filepath"
	"testing"
)

const probeResource = "english/words/default/六级单词.txt"

func TestSyncBundledResourcesInstallsVersion(t *testing.T) {
	baseDir := t.TempDir()

	if err := syncBundledResources(baseDir); err != nil {
		t.Fatalf("首次同步失败: %v", err)
	}

	want, err := embeddedResourceVersion()
	if err != nil {
		t.Fatalf("读取内置资源版本失败: %v", err)
	}
	if got := installedResourceVersion(filepath.Join(baseDir, resourcesDir)); got != want {
		t.Fatalf("版本标记 = %q，期望 %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(baseDir, resourcesDir, probeResource)); err != nil {
		t.Fatalf("内置资源未落盘: %v", err)
	}
}

func TestSyncBundledResourcesRefreshesOnVersionChange(t *testing.T) {
	baseDir := t.TempDir()
	if err := syncBundledResources(baseDir); err != nil {
		t.Fatalf("首次同步失败: %v", err)
	}

	resourcePath := filepath.Join(baseDir, resourcesDir, probeResource)
	stale := []byte("stale\t/steɪl/\tadj. 过期的\n")
	if err := os.WriteFile(resourcePath, stale, 0o644); err != nil {
		t.Fatal(err)
	}
	// 伪造一个旧版本，模拟老用户的安装目录
	if err := os.WriteFile(filepath.Join(baseDir, resourcesDir, resourceVersionFile), []byte("0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := syncBundledResources(baseDir); err != nil {
		t.Fatalf("升级同步失败: %v", err)
	}

	refreshed, err := os.ReadFile(resourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(refreshed) == string(stale) {
		t.Fatal("版本变化后资源文件未被刷新")
	}

	backup := findBackup(t, filepath.Join(baseDir, resourceBackupDirName))
	if string(backup) != string(stale) {
		t.Fatalf("备份内容 = %q，期望 %q", backup, stale)
	}
}

func TestSyncBundledResourcesKeepsUserEditsOnSameVersion(t *testing.T) {
	baseDir := t.TempDir()
	if err := syncBundledResources(baseDir); err != nil {
		t.Fatalf("首次同步失败: %v", err)
	}

	resourcePath := filepath.Join(baseDir, resourcesDir, probeResource)
	edited := []byte("mine\t/maɪn/\tpron. 我的\n")
	if err := os.WriteFile(resourcePath, edited, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := syncBundledResources(baseDir); err != nil {
		t.Fatalf("同版本同步失败: %v", err)
	}

	got, err := os.ReadFile(resourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(edited) {
		t.Fatal("版本未变化时不应覆盖用户修改")
	}
}

// findBackup 返回备份目录中 probeResource 的内容。
func findBackup(t *testing.T, backupRoot string) []byte {
	t.Helper()

	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		t.Fatalf("读取备份目录失败: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("备份目录数量 = %d，期望 1", len(entries))
	}

	data, err := os.ReadFile(filepath.Join(backupRoot, entries[0].Name(), filepath.FromSlash(probeResource)))
	if err != nil {
		t.Fatalf("读取备份文件失败: %v", err)
	}
	return data
}

// 版本标记文件本身不应产生备份，否则每次升级都会留下一个只含版本号的空目录
func TestSyncBundledResourcesDoesNotBackUpVersionMarker(t *testing.T) {
	baseDir := t.TempDir()
	if err := syncBundledResources(baseDir); err != nil {
		t.Fatalf("首次同步失败: %v", err)
	}

	// 只改版本标记，其余资源保持一致
	if err := os.WriteFile(filepath.Join(baseDir, resourcesDir, resourceVersionFile), []byte("0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := syncBundledResources(baseDir); err != nil {
		t.Fatalf("升级同步失败: %v", err)
	}

	if _, err := os.Stat(filepath.Join(baseDir, resourceBackupDirName)); !os.IsNotExist(err) {
		t.Fatalf("不应产生备份目录，err = %v", err)
	}
}

func TestShouldRefresh(t *testing.T) {
	cases := []struct {
		name      string
		installed string
		bundled   string
		want      bool
	}{
		{"版本更新则刷新", "3", "4", true},
		{"版本相同不刷新", "4", "4", false},
		{"旧版本的二进制不得回退资源", "4", "3", false},
		{"首次安装没有版本标记", "", "3", true},
		{"从日期式旧版本号平滑升级", "2026-09-17.2", "3", true},
		{"无法解析且内容相同则不刷新", "2026-09-17.2", "2026-09-17.2", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldRefresh(tc.installed, tc.bundled); got != tc.want {
				t.Errorf("shouldRefresh(%q, %q) = %v, want %v", tc.installed, tc.bundled, got, tc.want)
			}
		})
	}
}
