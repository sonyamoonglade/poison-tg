package telegram

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestExtractOrderID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "extracts short order ID from valid requisites",
			input:    "Payment received [ABC123] for order #12345",
			expected: "ABC123",
		},
		{
			name:     "extracts short order ID from empty brackets",
			input:    "[ABC123]",
			expected: "ABC123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := extractShortOrderIDFromRequisites(tt.input)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestInjectStringData(t *testing.T) {
	tests := []struct {
		name        string
		callback    int
		str         string
		expected    string
		expectedErr error
	}{
		{"basic", 123, "foo", "foo:123", nil},
		{"empty string", 456, "", ":456", nil},
		{"negative callback", -789, "bar", "bar:-789", nil},
		{"callback with leading zeroes", 001, "baz", "baz:1", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := injectStringData(test.callback, test.str)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestParseStringCallbackData(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expected    string
		expectedCb  int
		expectedErr error
	}{
		{"basic", "foo:123", "foo", 123, nil},
		{"empty string", ":456", "", 456, nil},
		{"negative callback", "bar:-789", "bar", -789, nil},
		{"callback with leading zeroes", "baz:001", "baz", 1, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, cb, err := parseStringCallbackData(test.data)
			require.Equal(t, test.expected, result)
			require.Equal(t, test.expectedCb, cb)
			require.Equal(t, test.expectedErr, err)
		})
	}
}
