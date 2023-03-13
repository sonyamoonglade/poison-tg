package telegram

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("should load all values", func(t *testing.T) {
		mt, st := "menu_template", "start_template"
		content := fmt.Sprintf(`{"menu":"%s", "start":"%s"}`, mt, st)
		tmpFile, err := os.Create("test_templates.tmp.json")
		require.NoError(t, err)
		tmpFile.Write([]byte(content))

		templateManager, err := Load(tmpFile.Name())
		require.NoError(t, err)

		require.Equal(t, templateManager.Menu(), mt)
		require.Equal(t, templateManager.Start(), st)

		defer os.Remove(tmpFile.Name())
	})

	t.Run("shoud load but not all values are present in file", func(t *testing.T) {
		mt := "menu_template"
		content := fmt.Sprintf(`{"menu":"%s"}`, mt)
		tmpFile, err := os.Create("test_templates.tmp.json")
		require.NoError(t, err)
		tmpFile.Write([]byte(content))

		templateManager, err := Load(tmpFile.Name())
		require.Error(t, err)
		require.Equal(t, "missing START template", err.Error())
		require.Nil(t, templateManager)
		defer os.Remove(tmpFile.Name())
	})
}
