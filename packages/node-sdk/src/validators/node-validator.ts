// validators/node-validator.ts
import { NodeDefinition, NodeInputConfig, NodeOutputConfig } from '../interfaces/node';

export class NodeValidator {
  static validateDefinition(definition: NodeDefinition): { isValid: boolean; errors: string[] } {
    const errors: string[] = [];

    // Validate required fields
    if (!definition.id) {
      errors.push('Node ID is required');
    }

    if (!definition.type) {
      errors.push('Node type is required');
    }

    if (!definition.metadata) {
      errors.push('Node metadata is required');
    } else {
      if (!definition.metadata.name) {
        errors.push('Node metadata.name is required');
      }
      if (!definition.metadata.description) {
        errors.push('Node metadata.description is required');
      }
      if (!definition.metadata.version) {
        errors.push('Node metadata.version is required');
      }
      if (!definition.metadata.category) {
        errors.push('Node metadata.category is required');
      }
    }

    // Validate inputs
    if (definition.inputs) {
      definition.inputs.forEach((input: NodeInputConfig, index: number) => {
        if (!input.id) {
          errors.push(`Input at index ${index} must have an ID`);
        }
        if (!input.name) {
          errors.push(`Input at index ${index} must have a name`);
        }
        if (!input.type) {
          errors.push(`Input at index ${index} must have a type`);
        }
      });
    }

    // Validate outputs
    if (definition.outputs) {
      definition.outputs.forEach((output: NodeOutputConfig, index: number) => {
        if (!output.id) {
          errors.push(`Output at index ${index} must have an ID`);
        }
        if (!output.name) {
          errors.push(`Output at index ${index} must have a name`);
        }
        if (!output.type) {
          errors.push(`Output at index ${index} must have a type`);
        }
      });
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  }

  static validateInput(
    input: any,
    expectedInputs: NodeInputConfig[]
  ): { isValid: boolean; errors: string[] } {
    const errors: string[] = [];

    if (!expectedInputs || expectedInputs.length === 0) {
      return { isValid: true, errors };
    }

    // Check required inputs
    expectedInputs.forEach((expectedInput) => {
      if (expectedInput.required && (input[expectedInput.id] === undefined || input[expectedInput.id] === null)) {
        errors.push(`Required input '${expectedInput.id}' is missing`);
      }
    });

    // Validate input types
    Object.keys(input).forEach((key) => {
      const expectedInput = expectedInputs.find((inp) => inp.id === key);
      if (expectedInput) {
        const inputValue = input[key];
        let isValid = true;

        switch (expectedInput.type) {
          case 'string':
            isValid = typeof inputValue === 'string';
            break;
          case 'number':
            isValid = typeof inputValue === 'number';
            break;
          case 'boolean':
            isValid = typeof inputValue === 'boolean';
            break;
          case 'object':
            isValid = typeof inputValue === 'object' && inputValue !== null && !Array.isArray(inputValue);
            break;
          case 'array':
            isValid = Array.isArray(inputValue);
            break;
          case 'json':
            try {
              if (typeof inputValue === 'string') {
                JSON.parse(inputValue);
              }
            } catch {
              isValid = false;
            }
            break;
          default:
            // For custom types, we just check if it exists
            isValid = inputValue !== undefined;
        }

        if (!isValid) {
          errors.push(`Input '${key}' has invalid type. Expected '${expectedInput.type}', got '${typeof inputValue}'`);
        }
      }
    });

    return {
      isValid: errors.length === 0,
      errors,
    };
  }
}