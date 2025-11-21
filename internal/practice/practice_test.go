package practice

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ajilisiwei/mllt-cli/internal/config"
)

// 测试前的准备工作
func setupTest(t *testing.T) {
	// 确保配置已加载
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 确保资源目录存在
	resourceDirs := []string{
		filepath.Join("resources", config.AppConfig.CurrentLanguage, Words),
		filepath.Join("resources", config.AppConfig.CurrentLanguage, Phrases),
		filepath.Join("resources", config.AppConfig.CurrentLanguage, Sentences),
		filepath.Join("resources", config.AppConfig.CurrentLanguage, Articles),
	}

	for _, dir := range resourceDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("创建目录失败: %v", err)
			}
		}
	}

	// 创建测试文件
	testFiles := map[string][]string{
		Words:     {"test_words.txt", "apple ->> 苹果\nbanana ->> 香蕉\norange ->> 橙子"},
		Phrases:   {"test_phrases.txt", "good morning ->> 早上好\ngood afternoon ->> 下午好"},
		Sentences: {"test_sentences.txt", "How are you? ->> 你好吗？\nI am fine. ->> 我很好。"},
		Articles:  {"test_article.txt", "This is a test article.\nIt has multiple lines.\nFor testing purposes."},
	}

	for resourceType, fileInfo := range testFiles {
		filePath := filepath.Join("resources", config.AppConfig.CurrentLanguage, resourceType, fileInfo[0])
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
		defer file.Close()

		if _, err := file.WriteString(fileInfo[1]); err != nil {
			t.Fatalf("写入测试文件失败: %v", err)
		}
	}
}

// 测试后的清理工作
func cleanupTest(t *testing.T) {
	// 删除测试文件
	testFiles := []string{
		filepath.Join("resources", config.AppConfig.CurrentLanguage, Words, "test_words.txt"),
		filepath.Join("resources", config.AppConfig.CurrentLanguage, Phrases, "test_phrases.txt"),
		filepath.Join("resources", config.AppConfig.CurrentLanguage, Sentences, "test_sentences.txt"),
		filepath.Join("resources", config.AppConfig.CurrentLanguage, Articles, "test_article.txt"),
	}

	for _, file := range testFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			t.Logf("删除测试文件失败: %v", err)
		}
	}
}

func TestGetResourcePath(t *testing.T) {
	// 确保配置已加载
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 测试获取资源路径
	path := GetResourcePath(Words, "test.txt")
	expectedPath := filepath.Join("resources", "user-data", config.AppConfig.CurrentLanguage, Words, DefaultFolderDir, "test.txt")

	if path != expectedPath {
		t.Errorf("资源路径不匹配，期望 %s，实际 %s", expectedPath, path)
	}
}

func TestGetResourceFiles(t *testing.T) {
	// 设置测试环境
	setupTest(t)
	defer cleanupTest(t)

	// 测试获取资源文件列表
	files, err := GetResourceFiles(Words)
	if err != nil {
		t.Errorf("获取资源文件列表失败: %v", err)
	}

	// 验证文件列表包含测试文件
	found := false
	for _, file := range files {
		if file == "test_words" {
			found = true
			break
		}
	}

	if !found {
		t.Error("资源文件列表中没有找到测试文件")
	}
}

func TestReadResourceFile(t *testing.T) {
	// 设置测试环境
	setupTest(t)
	defer cleanupTest(t)

	// 测试读取资源文件
	lines, err := ReadResourceFile(Words, "test_words.txt")
	if err != nil {
		t.Errorf("读取资源文件失败: %v", err)
	}

	// 验证文件内容
	if len(lines) != 3 {
		t.Errorf("文件行数不匹配，期望 3，实际 %d", len(lines))
	}

	expectedLine := "apple ->> 苹果"
	if lines[0] != expectedLine {
		t.Errorf("文件内容不匹配，期望 %s，实际 %s", expectedLine, lines[0])
	}
}

func TestParseLineSentenceWithoutTranslation(t *testing.T) {
	line := "People often talk about the beggar who sits on the corner of Main Street."
	original, translation := ParseLine(line)

	if original != line {
		t.Fatalf("纯英文句子应保持原样，期望 %q，实际 %q", line, original)
	}
	if translation != "" {
		t.Fatalf("纯英文句子不应解析出翻译，实际 %q", translation)
	}
}

func TestParseLineWithSpaceSeparatedTranslation(t *testing.T) {
	line := "beggar /ˈbɛgər/ n. 乞丐"
	original, translation := ParseLine(line)

	if original != "beggar" {
		t.Fatalf("应正确解析原文，期望 %q，实际 %q", "beggar", original)
	}
	expectedTranslation := "/ˈbɛgər/ n. 乞丐"
	if translation != expectedTranslation {
		t.Fatalf("应解析出翻译，期望 %q，实际 %q", expectedTranslation, translation)
	}
}
