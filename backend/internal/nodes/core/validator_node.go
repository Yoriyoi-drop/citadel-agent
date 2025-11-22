package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/nodes/utils"
)

// ValidatorNodeConfig represents the configuration for a Validator node
type ValidatorNodeConfig struct {
	StructTags  map[string]string `json:"struct_tags"`  // Map field names to validation tags
	FieldName   string            `json:"field_name"`   // Field to validate when not using struct tags
	Validation  string            `json:"validation"`   // Validation rule (e.g. "required,email")
	CustomRules map[string]string `json:"custom_rules,omitempty"` // Custom validation functions
}

// ValidatorNode validates input data using go-playground/validator
type ValidatorNode struct {
	config   ValidatorNodeConfig
	validate *validator.Validate
}

// NewValidatorNode creates a new Validator node with the given configuration
func NewValidatorNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Extract config values
	structTags := make(map[string]string)
	if tags, exists := config["struct_tags"]; exists {
		if tagsMap, ok := tags.(map[string]interface{}); ok {
			for k, v := range tagsMap {
				if vStr, ok := v.(string); ok {
					structTags[k] = vStr
				}
			}
		}
	}

	fieldName := utils.GetStringValue(config["field_name"], "")
	validation := utils.GetStringValue(config["validation"], "")

	customRules := make(map[string]string)
	if rules, exists := config["custom_rules"]; exists {
		if rulesMap, ok := rules.(map[string]interface{}); ok {
			for k, v := range rulesMap {
				if vStr, ok := v.(string); ok {
					customRules[k] = vStr
				}
			}
		}
	}

	validatorConfig := ValidatorNodeConfig{
		StructTags:  structTags,
		FieldName:   fieldName,
		Validation:  validation,
		CustomRules: customRules,
	}

	return &ValidatorNode{
		config:   validatorConfig,
		validate: validator.New(),
	}, nil
}

// Execute implements the NodeInstance interface
func (v *ValidatorNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Prepare validation based on configuration
	var validationErrors []string

	if len(v.config.StructTags) > 0 {
		// Validate a struct using the provided tags
		validationErrors = v.validateStruct(input)
	} else if v.config.FieldName != "" && v.config.Validation != "" {
		// Validate a single field
		fieldValue := input[v.config.FieldName]
		err := v.validate.Var(fieldValue, v.config.Validation)
		if err != nil {
			validatorErrs, ok := err.(validator.ValidationErrors)
			if ok {
				for _, errDetail := range validatorErrs {
					validationErrors = append(validationErrors, fmt.Sprintf("%s: %s", errDetail.Field(), errDetail.Tag()))
				}
			} else {
				validationErrors = append(validationErrors, err.Error())
			}
		}
	} else {
		return nil, fmt.Errorf("either struct_tags or both field_name and validation must be provided")
	}

	// Check if there were validation errors
	if len(validationErrors) > 0 {
		return map[string]interface{}{
			"success": false,
			"valid":   false,
			"errors":  validationErrors,
			"input":   input,
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Validation passed
	return map[string]interface{}{
		"success": true,
		"valid":   true,
		"validated_data": input,
		"timestamp": time.Now().Unix(),
	}, nil
}

// validateStruct validates a map as a struct using the provided struct tags
func (v *ValidatorNode) validateStruct(input map[string]interface{}) []string {
	var errors []string

	// For each field specified in struct tags, validate it
	for fieldName, validationTag := range v.config.StructTags {
		fieldValue, exists := input[fieldName]
		if !exists {
			// Check if the field is required
			if strings.Contains(validationTag, "required") {
				errors = append(errors, fmt.Sprintf("%s: required field is missing", fieldName))
			}
			continue
		}

		// Validate the field value by validating against a temporary variable
		err := v.validate.Var(fieldValue, validationTag)
		if err != nil {
			validatorErrs, ok := err.(validator.ValidationErrors)
			if ok {
				for _, errDetail := range validatorErrs {
					errors = append(errors, fmt.Sprintf("%s: %s", fieldName, errDetail.Tag()))
				}
			} else {
				errors = append(errors, fmt.Sprintf("%s: %s", fieldName, err.Error()))
			}
		}
	}

	return errors
}

