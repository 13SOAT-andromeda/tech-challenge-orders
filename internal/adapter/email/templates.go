package email

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type EmailTemplates string

const (
	ORDER_APPROVAL EmailTemplates = "order-approval"
)

func LoadTemplate(template EmailTemplates) (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("template not found")
	}

	baseDir := filepath.Dir(filename)
	templatePath := filepath.Join(baseDir, "templates", string(template)+".html")

	content, err := os.ReadFile(templatePath)

	if err != nil {
		return "", fmt.Errorf("error on loading template: %w", err)
	}

	return string(content), nil
}
