package trex_modules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Loader manages loading Python modules
type Loader struct {
	paths []string
}

// NewLoader creates a new module loader
func NewLoader(paths string) *Loader {
	var pathList []string
	if paths != "" {
		pathList = strings.Split(paths, ":")
	}
	return &Loader{paths: pathList}
}

// FindModule searches for a module in configured paths
func (l *Loader) FindModule(moduleName string) (string, error) {
	// Check in configured paths
	for _, path := range l.paths {
		modulePath := filepath.Join(path, moduleName+".py")
		if _, err := os.Stat(modulePath); err == nil {
			return modulePath, nil
		}
	}

	// Check current directory
	if _, err := os.Stat(moduleName + ".py"); err == nil {
		return moduleName + ".py", nil
	}

	return "", os.ErrNotExist
}

// ValidateModuleOutput ensures output is valid JSON
func ValidateModuleOutput(output string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetModuleInfo returns information about a module
func (l *Loader) GetModuleInfo(moduleName string) map[string]interface{} {
	path, err := l.FindModule(moduleName)
	if err != nil {
		return nil
	}

	return map[string]interface{}{
		"name": moduleName,
		"path": path,
	}
}
