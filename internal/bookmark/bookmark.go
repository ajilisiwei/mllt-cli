package bookmark

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ajilisiwei/mllt-cli/internal/practice"
)

// Special list names
const (
	MarkedList   = "标记"
	FavoriteList = "收藏"
	// WrongList 收集答错过的条目。与「标记」不同，它不会被排除出正常练习——
	// 错过的东西恰恰是最该继续练的。
	WrongList = "错题"
)

var supportedLists = map[string]struct{}{
	MarkedList:   {},
	FavoriteList: {},
	WrongList:    {},
}

// Add stores an item in the target special list. Returns true if the item was newly added.
func Add(resourceType, listName, item string) (bool, error) {
	if !isSupportedList(listName) {
		return false, fmt.Errorf("不支持的列表类型: %s", listName)
	}

	cleanedItem := normalizeItem(item)
	if cleanedItem == "" {
		return false, fmt.Errorf("无法标记空内容")
	}

	items, err := GetItems(resourceType, listName)
	if err != nil {
		return false, err
	}

	for _, existing := range items {
		if normalizeItem(existing) == cleanedItem {
			return false, nil
		}
	}

	items = append(items, cleanedItem)
	if err := practice.WriteResourceFile(resourceType, listName, items); err != nil {
		return false, err
	}

	return true, nil
}

// Remove deletes an item from the target special list. Returns true if the item existed.
func Remove(resourceType, listName, item string) (bool, error) {
	if !isSupportedList(listName) {
		return false, fmt.Errorf("不支持的列表类型: %s", listName)
	}

	cleanedItem := normalizeItem(item)
	if cleanedItem == "" {
		return false, fmt.Errorf("无法取消空内容")
	}

	items, err := GetItems(resourceType, listName)
	if err != nil {
		return false, err
	}

	updated := items[:0]
	removed := false
	for _, existing := range items {
		if normalizeItem(existing) == cleanedItem {
			removed = true
			continue
		}
		updated = append(updated, normalizeItem(existing))
	}

	if !removed {
		return false, nil
	}

	if err := practice.WriteResourceFile(resourceType, listName, updated); err != nil {
		return false, err
	}

	return true, nil
}

// Contains checks whether the item already exists in the target list.
func Contains(resourceType, listName, item string) (bool, error) {
	if !isSupportedList(listName) {
		return false, fmt.Errorf("不支持的列表类型: %s", listName)
	}

	cleanedItem := normalizeItem(item)
	if cleanedItem == "" {
		return false, nil
	}

	items, err := GetItems(resourceType, listName)
	if err != nil {
		return false, err
	}

	for _, existing := range items {
		if normalizeItem(existing) == cleanedItem {
			return true, nil
		}
	}
	return false, nil
}

// GetItems returns all stored items for the given list.
func GetItems(resourceType, listName string) ([]string, error) {
	if !isSupportedList(listName) {
		return nil, fmt.Errorf("不支持的列表类型: %s", listName)
	}

	items, err := practice.ReadResourceFile(resourceType, listName)
	if err != nil {
		return nil, err
	}

	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := normalizeItem(item); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned, nil
}

// IsSpecialList tells whether the given file name is one of the special lists.
func IsSpecialList(fileName string) bool {
	_, ok := supportedLists[fileName]
	return ok
}

// SupportsMark indicates whether the resource type allows mark/unmark commands.
func SupportsMark(resourceType string) bool {
	return resourceType != practice.Articles
}

func isSupportedList(listName string) bool {
	name := filepath.Base(listName)
	name = strings.TrimSuffix(name, ".txt")
	_, ok := supportedLists[name]
	return ok
}

func normalizeItem(item string) string {
	return strings.TrimSpace(item)
}
