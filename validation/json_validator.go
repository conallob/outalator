package validation

import (
	"encoding/json"
	"fmt"
)

const (
	// MaxMetadataKeys is the maximum number of keys allowed in metadata maps
	MaxMetadataKeys = 100

	// MaxMetadataKeyLength is the maximum length of a metadata key
	MaxMetadataKeyLength = 255

	// MaxMetadataValueLength is the maximum length of a metadata value
	MaxMetadataValueLength = 4096

	// MaxCustomFieldsSize is the maximum size in bytes for custom_fields JSON
	MaxCustomFieldsSize = 65536 // 64KB

	// MaxCustomFieldsDepth is the maximum nesting depth for custom_fields
	MaxCustomFieldsDepth = 10
)

// ValidateMetadata validates a metadata map
func ValidateMetadata(metadata map[string]string) error {
	if metadata == nil {
		return nil
	}

	if len(metadata) > MaxMetadataKeys {
		return fmt.Errorf("metadata exceeds maximum of %d keys (has %d)",
			MaxMetadataKeys, len(metadata))
	}

	for key, value := range metadata {
		if len(key) == 0 {
			return fmt.Errorf("metadata key cannot be empty")
		}

		if len(key) > MaxMetadataKeyLength {
			return fmt.Errorf("metadata key '%s' exceeds maximum length of %d bytes (has %d)",
				key, MaxMetadataKeyLength, len(key))
		}

		if len(value) > MaxMetadataValueLength {
			return fmt.Errorf("metadata value for key '%s' exceeds maximum length of %d bytes (has %d)",
				key, MaxMetadataValueLength, len(value))
		}
	}

	return nil
}

// ValidateCustomFields validates custom_fields for size and depth
func ValidateCustomFields(customFields map[string]any) error {
	if customFields == nil {
		return nil
	}

	// Check size by marshaling to JSON
	jsonBytes, err := json.Marshal(customFields)
	if err != nil {
		return fmt.Errorf("invalid custom_fields: %w", err)
	}

	if len(jsonBytes) > MaxCustomFieldsSize {
		return fmt.Errorf("custom_fields exceeds maximum size of %d bytes (has %d)",
			MaxCustomFieldsSize, len(jsonBytes))
	}

	// Check nesting depth
	depth := calculateDepth(customFields)
	if depth > MaxCustomFieldsDepth {
		return fmt.Errorf("custom_fields exceeds maximum nesting depth of %d (has %d)",
			MaxCustomFieldsDepth, depth)
	}

	return nil
}

// calculateDepth calculates the maximum nesting depth of a value
func calculateDepth(v any) int {
	switch val := v.(type) {
	case map[string]any:
		maxDepth := 0
		for _, item := range val {
			depth := calculateDepth(item)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return maxDepth + 1

	case []any:
		maxDepth := 0
		for _, item := range val {
			depth := calculateDepth(item)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return maxDepth + 1

	default:
		return 1
	}
}

// ValidateSourceMetadata validates source-specific metadata
func ValidateSourceMetadata(sourceMetadata map[string]any) error {
	// Use the same validation as custom fields
	return ValidateCustomFields(sourceMetadata)
}
