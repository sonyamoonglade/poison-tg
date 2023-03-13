package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type templates struct {
	Menu  string `json: "menu,omitempty"`
	Start string `json: "start,omitempty"`
}

type TemplateManager struct {
	t templates
}

func Load(path string) (*TemplateManager, error) {

	var templates templates

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read file %s: %w", path, err)
	}

	if err := json.NewDecoder(bytes.NewReader(content)).Decode(&templates); err != nil {
		return nil, fmt.Errorf("can't decode file content: %w", err)

	}

	if templates.Menu == "" {
		return nil, fmt.Errorf("missing MENU template")
	}

	if templates.Start == "" {
		return nil, fmt.Errorf("missing START template")
	}

	return &TemplateManager{
		t: templates,
	}, nil
}

func (tm *TemplateManager) Menu() string {
	return tm.t.Menu
}

func (tm *TemplateManager) Start() string {
	return tm.t.Start
}
