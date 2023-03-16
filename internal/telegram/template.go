package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

var (
	t *templates = new(templates)
)

type templates struct {
	Menu  string `json:"menu,omitempty"`
	Start string `json:"start,omitempty"`
}

func LoadTemplates(path string) error {
	var templates templates

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't read file %s: %w", path, err)
	}

	if err := json.NewDecoder(bytes.NewReader(content)).Decode(&templates); err != nil {
		return fmt.Errorf("can't decode file content: %w", err)
	}

	if templates.Menu == "" {
		return fmt.Errorf("missing MENU template")
	}
	if templates.Start == "" {
		return fmt.Errorf("missing START template")
	}

	*t = templates
	return nil
}

func GetTemplate() *templates {
	return t
}
