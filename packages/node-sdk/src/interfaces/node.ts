// interfaces/node.ts
export interface NodeInput {
  [key: string]: any;
}

export interface NodeOutput {
  [key: string]: any;
}

export interface NodeMetadata {
  name: string;
  description: string;
  version: string;
  category: string;
  icon?: string;
  tags?: string[];
}

export interface NodeDefinition {
  id: string;
  type: string;
  metadata: NodeMetadata;
  inputs: NodeInputConfig[];
  outputs: NodeOutputConfig[];
  credentials?: CredentialConfig[];
  settings?: SettingConfig[];
}

export interface NodeInputConfig {
  id: string;
  name: string;
  type: string;
  required: boolean;
  description?: string;
  default?: any;
  validation?: ValidationRule[];
}

export interface NodeOutputConfig {
  id: string;
  name: string;
  type: string;
  description?: string;
}

export interface CredentialConfig {
  id: string;
  name: string;
  type: string;
  required: boolean;
  description?: string;
}

export interface SettingConfig {
  id: string;
  name: string;
  type: string;
  required: boolean;
  default?: any;
  description?: string;
}

export interface ValidationRule {
  type: string;
  value: any;
  message: string;
}

export interface NodeExecutionContext {
  input: NodeInput;
  credentials?: any;
  settings?: any;
  workflow?: any;
  execution?: any;
}

export interface NodeExecutionResult {
  success: boolean;
  data?: NodeOutput;
  error?: string;
  metadata?: {
    executionTime: number;
    timestamp: Date;
  };
}

export interface NodeInterface {
  execute(context: NodeExecutionContext): Promise<NodeExecutionResult>;
}