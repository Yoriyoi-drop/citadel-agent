package nodes

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
)

// AIPromptGeneratorNode generates AI prompts based on task descriptions
type AIPromptGeneratorNode struct {
	defaultLanguage string
	defaultStyle    string
}

// Execute implements the NodeExecutor interface
func (a *AIPromptGeneratorNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	taskDescription, ok := input["task_description"].(string)
	if !ok {
		return &engine.ExecutionResult{
			Status: "error",
			Error:  "task_description is required and must be a string",
		}, nil
	}

	targetLanguage, _ := input["target_language"].(string)
	if targetLanguage == "" {
		targetLanguage = "en"
	}

	promptStyle, _ := input["prompt_style"].(string)
	if promptStyle == "" {
		promptStyle = "direct"
	}

	// Generate the prompt based on the task description and parameters
	generatedPrompt, err := a.generatePrompt(taskDescription, targetLanguage, promptStyle)
	if err != nil {
		return &engine.ExecutionResult{
			Status: "error",
			Error:  fmt.Sprintf("Failed to generate prompt: %v", err),
		}, nil
	}

	result := map[string]interface{}{
		"input_task_description": taskDescription,
		"target_language":        targetLanguage,
		"prompt_style":           promptStyle,
		"generated_prompt":       generatedPrompt,
		"timestamp":              time.Now().Unix(),
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data:   result,
	}, nil
}

// generatePrompt creates a well-structured prompt based on the task description
func (a *AIPromptGeneratorNode) generatePrompt(taskDescription, language, style string) (string, error) {
	// Add processing delay to simulate AI processing
	time.Sleep(200 * time.Millisecond)

	var prompt string
	switch style {
	case "chain_of_thought":
		prompt = fmt.Sprintf(
			"Please think step by step to solve the following task. First, identify the key components of the task. Next, determine the sequence of operations needed. Finally, execute the task.\n\nTask: %s\n\nTake a deep breath and work on this step by step.", 
			taskDescription)
	case "few_shot":
		prompt = fmt.Sprintf(
			"Here are some examples of similar tasks:\n\nExample 1: [Provide an example with input and output]\nExample 2: [Provide another example with input and output]\n\nNow, please perform the following task following the pattern of the examples:\n\nTask: %s", 
			taskDescription)
	case "instruction_following":
		prompt = fmt.Sprintf(
			"You are an expert assistant. Your goal is to follow the user's instructions precisely. Here is the instruction:\n\n%s\n\nPlease provide a complete and accurate response that fully addresses the instruction.", 
			taskDescription)
	default: // direct
		prompt = fmt.Sprintf("Task: %s\n\nProvide a detailed response.", taskDescription)
	}

	return prompt, nil
}