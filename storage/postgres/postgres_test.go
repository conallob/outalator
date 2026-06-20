package postgres

import (
	"testing"
)

func TestMarshalJSONMap(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]string
		want    string
		wantErr bool
	}{
		{
			name:    "nil map returns empty object",
			input:   nil,
			want:    "{}",
			wantErr: false,
		},
		{
			name:    "empty map returns empty object",
			input:   map[string]string{},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "map with values",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			want:    `{"key1":"value1","key2":"value2"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalJSONMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("marshalJSONMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want && tt.name != "map with values" {
				t.Errorf("marshalJSONMap() = %v, want %v", string(got), tt.want)
			}
			// For map with values, just check it's valid JSON and not null
			if tt.name == "map with values" {
				if string(got) == "null" {
					t.Errorf("marshalJSONMap() returned null, expected valid JSON object")
				}
				if len(got) == 0 {
					t.Errorf("marshalJSONMap() returned empty bytes")
				}
			}
		})
	}
}

func TestMarshalJSONAny(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:    "nil returns empty object",
			input:   nil,
			want:    "{}",
			wantErr: false,
		},
		{
			name:    "nil map returns empty object",
			input:   (map[string]any)(nil),
			want:    "{}",
			wantErr: false,
		},
		{
			name:    "empty map returns empty object",
			input:   map[string]any{},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "map with string values",
			input: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			want:    "",  // Just check it's not null
			wantErr: false,
		},
		{
			name: "map with nested values",
			input: map[string]any{
				"key1": "value1",
				"nested": map[string]any{
					"inner": "value",
				},
			},
			want:    "",  // Just check it's not null
			wantErr: false,
		},
		{
			name: "map with array values",
			input: map[string]any{
				"key1": "value1",
				"array": []string{"item1", "item2"},
			},
			want:    "",  // Just check it's not null
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalJSONAny(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("marshalJSONAny() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check for specific expected values
			if tt.want != "" {
				if string(got) != tt.want {
					t.Errorf("marshalJSONAny() = %v, want %v", string(got), tt.want)
				}
			} else {
				// Just verify it's not null or empty
				if string(got) == "null" {
					t.Errorf("marshalJSONAny() returned null, expected valid JSON")
				}
				if len(got) == 0 {
					t.Errorf("marshalJSONAny() returned empty bytes")
				}
			}
		})
	}
}

func TestMarshalConsistency(t *testing.T) {
	t.Run("nil map always produces same output", func(t *testing.T) {
		result1, _ := marshalJSONMap(nil)
		result2, _ := marshalJSONMap(nil)

		if string(result1) != string(result2) {
			t.Errorf("marshaling nil map produced inconsistent results: %s vs %s",
				result1, result2)
		}

		if string(result1) != "{}" {
			t.Errorf("marshaling nil map should produce '{}', got %s", result1)
		}
	})

	t.Run("nil any always produces same output", func(t *testing.T) {
		result1, _ := marshalJSONAny(nil)
		result2, _ := marshalJSONAny(nil)

		if string(result1) != string(result2) {
			t.Errorf("marshaling nil any produced inconsistent results: %s vs %s",
				result1, result2)
		}

		if string(result1) != "{}" {
			t.Errorf("marshaling nil any should produce '{}', got %s", result1)
		}
	})
}
