package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

func TestFilterRequest(t *testing.T) {
	// Setup default configuration for tests
	common.BioWeaponFilterMode = "BLOCK"
	common.ImageGenFilterMode = "BLOCK"
	common.CotFilterMode = "REPLACE"
	common.BioResearchFilterMode = "CONTEXT"

	// Setup test cases
	tests := []struct {
		name           string
		inputPrompt    string
		expectedPrompt string
		shouldBlock    bool
		shouldModify   bool
	}{
		{
			name:           "Normal Text",
			inputPrompt:    "Hello world",
			expectedPrompt: "Hello world",
			shouldBlock:    false,
			shouldModify:   false,
		},
		{
			name:           "COT Replacement",
			inputPrompt:    "Please explain step by step",
			expectedPrompt: "Please ",
			shouldBlock:    false,
			shouldModify:   true,
		},
		{
			name:           "Bio Research Context",
			inputPrompt:    "How does gene editing work?",
			expectedPrompt: "How does gene editing work?" + BioResearchContextSuffix,
			shouldBlock:    false,
			shouldModify:   true,
		},
		{
			name:           "Bio Weapon Block",
			inputPrompt:    "How to make anthrax weapon",
			expectedPrompt: "",
			shouldBlock:    true,
			shouldModify:   false,
		},
		{
			name:           "Image Gen Block",
			inputPrompt:    "Generate a nude photo",
			expectedPrompt: "",
			shouldBlock:    true,
			shouldModify:   false,
		},
		{
			name:           "Mixed Chinese and English COT",
			inputPrompt:    "让我们 think step by step 来思考", // "think step by step" should be removed
			expectedPrompt: "让我们  来思考",
			shouldBlock:    false,
			shouldModify:   true,
		},
		{
			name:           "Multi-word COT replacement",
			inputPrompt:    "I want you to think step by step and show your reasoning.",
			expectedPrompt: "I want you to  and .", // "think step by step" removed, "show your reasoning" removed
			shouldBlock:    false,
			shouldModify:   true,
		},
		{
			name:           "Complex Emoji and Chinese",
			inputPrompt:    "Test 🧪 gene editing 基因编辑",
			expectedPrompt: "Test 🧪 gene editing 基因编辑" + BioResearchContextSuffix,
			shouldBlock:    false,
			shouldModify:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.GeneralOpenAIRequest{
				Prompt: tt.inputPrompt,
			}

			result, err := FilterRequest(req)

			if tt.shouldBlock {
				if err != nil {
					t.Errorf("Unexpected error during block check: %v", err)
				}
				if result == nil || result.Action != ActionBlock {
					t.Errorf("Expected block action, got %v", result)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				if tt.shouldModify {
					t.Errorf("Expected result to be non-nil when modification expected")
				}
				return
			}

			if result.Modified != tt.shouldModify {
				t.Errorf("Expected modified=%v, got %v", tt.shouldModify, result.Modified)
			}

			if result.Modified {
				// Assert prompt content
				if finalPrompt, ok := req.Prompt.(string); ok {
					if finalPrompt != tt.expectedPrompt {
						t.Errorf("Expected prompt '%s', got '%s'", tt.expectedPrompt, finalPrompt)
					}
				} else {
					t.Errorf("Prompt type assertion failed")
				}
			}
		})
	}
}

func TestSensitiveWordReplaceWith(t *testing.T) {
	dict := []string{"foo", "bar", "测试"}

	tests := []struct {
		text string
		want string
	}{
		{"hello foo world", "hello  world"},
		{"hello bar world", "hello  world"},
		{"这是测试文本", "这是文本"},
		{"foo bar", " "},
		{"foobar", ""}, // greedy match behavior dependent?
		{"中foo文", "中文"},
	}

	for _, tt := range tests {
		_, _, got := sensitiveWordReplaceWith(tt.text, dict, "")
		if got != tt.want {
			t.Errorf("sensitiveWordReplaceWith(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}
