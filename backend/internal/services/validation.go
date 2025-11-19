package services

import (
	"errors"
	"fmt"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

// Error returns a string representation of all validation errors
func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no validation errors"
	}
	
	msg := "validation errors: "
	for i, err := range e.Errors {
		if i > 0 {
			msg += "; "
		}
		msg += err.Error()
	}
	return msg
}

// Add adds a validation error to the collection
func (e *ValidationErrors) Add(field, message string, value interface{}) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// BusinessError represents a business logic error
type BusinessError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

// Error returns the error message
func (e *BusinessError) Error() string {
	return fmt.Sprintf("business error (%s): %s", e.Code, e.Message)
}

// CreateBusinessError creates a new business error
func CreateBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// AddDetail adds a detail to the business error
func (e *BusinessError) AddDetail(key string, value interface{}) {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
}

// Validation constants
const (
	ErrCodeRequiredField      = "REQUIRED_FIELD"
	ErrCodeInvalidFormat      = "INVALID_FORMAT"
	ErrCodeInvalidValue       = "INVALID_VALUE"
	ErrCodeDuplicateValue     = "DUPLICATE_VALUE"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeInvalidState       = "INVALID_STATE"
	ErrCodeBusinessRule       = "BUSINESS_RULE_VIOLATION"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeValidationError    = "VALIDATION_ERROR"
	ErrCodeInternalError      = "INTERNAL_ERROR"
)

// Validation functions
func ValidateRequired(value interface{}, fieldName string) error {
	if value == nil {
		return CreateBusinessError(ErrCodeRequiredField, fmt.Sprintf("%s is required", fieldName))
	}
	
	switch v := value.(type) {
	case string:
		if v == "" {
			return CreateBusinessError(ErrCodeRequiredField, fmt.Sprintf("%s is required", fieldName))
		}
	case int, int32, int64:
		if v == 0 {
			return CreateBusinessError(ErrCodeRequiredField, fmt.Sprintf("%s is required", fieldName))
		}
	}
	
	return nil
}

func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	
	// Simple email validation
	// In a real application, use a more comprehensive validation
	if len(email) > 254 {
		return CreateBusinessError(ErrCodeInvalidFormat, "email is too long")
	}
	
	if len(email) < 5 {
		return CreateBusinessError(ErrCodeInvalidFormat, "email is too short")
	}
	
	if !contains(email, "@") || !contains(email, ".") {
		return CreateBusinessError(ErrCodeInvalidFormat, "invalid email format")
	}
	
	return nil
}

func ValidateStringLength(value string, min, max int, fieldName string) error {
	if len(value) < min {
		return CreateBusinessError(ErrCodeInvalidValue, fmt.Sprintf("%s must be at least %d characters", fieldName, min))
	}
	if len(value) > max {
		return CreateBusinessError(ErrCodeInvalidValue, fmt.Sprintf("%s must be at most %d characters", fieldName, max))
	}
	return nil
}

func ValidateURL(url string) error {
	if url == "" {
		return nil // Allow empty URL
	}
	
	// Simple URL validation
	// In a real application, use net/url package for proper validation
	if !startsWith(url, "http://") && !startsWith(url, "https://") {
		return CreateBusinessError(ErrCodeInvalidFormat, "URL must start with http:// or https://")
	}
	
	return nil
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		contains(s[1:], substr)))
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// Common validation functions for services
func ValidateWorkflowInput(name, description string) *ValidationErrors {
	errs := &ValidationErrors{}
	
	if name == "" {
		errs.Add("name", "Workflow name is required", name)
	} else if len(name) > 255 {
		errs.Add("name", "Workflow name must be at most 255 characters", name)
	}
	
	if description != "" && len(description) > 1000 {
		errs.Add("description", "Workflow description must be at most 1000 characters", description)
	}
	
	return errs
}

func ValidateNodeInput(workflowID, nodeType, name string) *ValidationErrors {
	errs := &ValidationErrors{}
	
	if workflowID == "" {
		errs.Add("workflow_id", "Workflow ID is required", workflowID)
	}
	
	if nodeType == "" {
		errs.Add("type", "Node type is required", nodeType)
	}
	
	if name == "" {
		errs.Add("name", "Node name is required", name)
	} else if len(name) > 255 {
		errs.Add("name", "Node name must be at most 255 characters", name)
	}
	
	return errs
}

func ValidateUserInput(email, username, firstName, lastName string) *ValidationErrors {
	errs := &ValidationErrors{}
	
	if email == "" {
		errs.Add("email", "Email is required", email)
	} else if err := ValidateEmail(email); err != nil {
		errs.Add("email", err.Error(), email)
	}
	
	if username == "" {
		errs.Add("username", "Username is required", username)
	} else if len(username) < 3 {
		errs.Add("username", "Username must be at least 3 characters", username)
	} else if len(username) > 50 {
		errs.Add("username", "Username must be at most 50 characters", username)
	}
	
	if firstName != "" && len(firstName) > 100 {
		errs.Add("first_name", "First name must be at most 100 characters", firstName)
	}
	
	if lastName != "" && len(lastName) > 100 {
		errs.Add("last_name", "Last name must be at most 100 characters", lastName)
	}
	
	return errs
}