package validation

import (
	"strings"
	"testing"
)

func TestValidateMetadata(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil metadata is valid",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "empty metadata is valid",
			input:   map[string]string{},
			wantErr: false,
		},
		{
			name: "valid metadata",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name: "empty key is invalid",
			input: map[string]string{
				"": "value",
			},
			wantErr: true,
			errMsg:  "key cannot be empty",
		},
		{
			name: "key too long",
			input: map[string]string{
				strings.Repeat("a", MaxMetadataKeyLength+1): "value",
			},
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name: "value too long",
			input: map[string]string{
				"key": strings.Repeat("a", MaxMetadataValueLength+1),
			},
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name: "too many keys",
			input: func() map[string]string {
				m := make(map[string]string)
				for i := 0; i < MaxMetadataKeys+1; i++ {
					m[string(rune(i))] = "value"
				}
				return m
			}(),
			wantErr: true,
			errMsg:  "exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetadata(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateMetadata() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateCustomFields(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil custom fields is valid",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "empty custom fields is valid",
			input:   map[string]any{},
			wantErr: false,
		},
		{
			name: "valid custom fields",
			input: map[string]any{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
			wantErr: false,
		},
		{
			name: "nested custom fields",
			input: map[string]any{
				"key1": "value1",
				"nested": map[string]any{
					"inner": "value",
				},
			},
			wantErr: false,
		},
		{
			name: "deep nesting exceeds limit",
			input: func() map[string]any {
				result := make(map[string]any)
				current := result
				for i := 0; i < MaxCustomFieldsDepth+1; i++ {
					next := make(map[string]any)
					current["nested"] = next
					current = next
				}
				current["value"] = "deep"
				return result
			}(),
			wantErr: true,
			errMsg:  "exceeds maximum nesting depth",
		},
		{
			name: "size exceeds limit",
			input: map[string]any{
				"large": strings.Repeat("a", MaxCustomFieldsSize),
			},
			wantErr: true,
			errMsg:  "exceeds maximum size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCustomFields(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCustomFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateCustomFields() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestCalculateDepth(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  int
	}{
		{
			name:  "string has depth 1",
			input: "value",
			want:  1,
		},
		{
			name:  "number has depth 1",
			input: 123,
			want:  1,
		},
		{
			name: "flat map has depth 1",
			input: map[string]any{
				"key": "value",
			},
			want: 2,
		},
		{
			name: "nested map has depth 2",
			input: map[string]any{
				"outer": map[string]any{
					"inner": "value",
				},
			},
			want: 3,
		},
		{
			name: "deeply nested map",
			input: map[string]any{
				"l1": map[string]any{
					"l2": map[string]any{
						"l3": "value",
					},
				},
			},
			want: 4,
		},
		{
			name:  "array has depth 1",
			input: []any{"a", "b", "c"},
			want:  2,
		},
		{
			name: "nested array",
			input: []any{
				[]any{"a", "b"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateDepth(tt.input)
			if got != tt.want {
				t.Errorf("calculateDepth() = %v, want %v", got, tt.want)
			}
		})
	}
}
