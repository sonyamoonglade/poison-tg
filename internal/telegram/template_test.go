package telegram

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadTemplates(t *testing.T) {
	testcases := []struct {
		description string
		templates   templates
		expectedErr string
	}{
		{
			description: "missing MENU template",
			templates: templates{
				Start:               "start_template",
				CartPreviewStartFMT: "cart_preview_start_template",
				CartPreviewEndFMT:   "cart_preview_end_template",
				CartPositionFMT:     "cart_position_template",
				CalculatorOutput:    "output",
			},
			expectedErr: "missing Menu template",
		},
		{
			description: "missing START template",
			templates: templates{
				Menu:                "menu_template",
				CartPreviewStartFMT: "cart_preview_start_template",
				CartPreviewEndFMT:   "cart_preview_end_template",
				CartPositionFMT:     "cart_position_template",
				CalculatorOutput:    "output",
			},
			expectedErr: "missing Start template",
		},
		{
			description: "empty file",
			templates:   templates{},
			expectedErr: "can't decode file content. File is empty",
		},
	}

	for _, tc := range testcases {
		tempFile, err := ioutil.TempFile("", "test_templates.json")
		if err != nil {
			t.Fatal(err)
		}

		jsonData, err := json.Marshal(tc.templates)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tempFile.Write(jsonData); err != nil {
			t.Fatal(err)
		}
		tempFile.Close()
		defer os.Remove(tempFile.Name())

		t.Run(tc.description, func(t *testing.T) {
			err := LoadTemplates(tempFile.Name())

			if err == nil {
				t.Errorf("LoadTemplates() should have failed")
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("expected '%s', but got '%s'", tc.expectedErr, err.Error())
			}
		})
	}
}
