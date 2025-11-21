// packages/plugin-system/src/types/plugin.ts
export enum PluginType {
  NODE = 'node',
  TRIGGER = 'trigger',
  ACTION = 'action',
  UTILITY = 'utility',
  INTEGRATION = 'integration',
  AI_AGENT = 'ai_agent',
  DATA_PROCESSOR = 'data_processor'
}

export enum PluginStatus {
  DRAFT = 'draft',
  PUBLISHED = 'published',
  DEPRECATED = 'deprecated',
  UNINSTALLED = 'uninstalled',
  INSTALLED = 'installed',
  UPDATING = 'updating',
  ERROR = 'error'
}

export enum PluginGrade {
  BASIC = 'basic',
  INTERMEDIATE = 'intermediate',
  ADVANCED = 'advanced',
  ELITE = 'elite'
}

export interface PluginMetadata {
  id: string;
  name: string;
  description: string;
  version: string;
  author: string;
  license: string;
  category: string;
  grade: PluginGrade;
  tags?: string[];
  icon?: string;
  documentationUrl?: string;
  repositoryUrl?: string;
  minCitadelVersion?: string;
  homepage?: string;
  bugs?: string;
  keywords?: string[];
  dependencies?: PluginDependency[];
  supportedLanguages?: string[];
}

export interface PluginDependency {
  id: string;
  version: string;
  optional?: boolean;
}

export interface PluginManifest {
  schemaVersion: string;
  metadata: PluginMetadata;
  files: PluginFile[];
  permissions: PluginPermission[];
  configuration?: PluginConfiguration[];
  hooks?: PluginHooks;
}

export interface PluginFile {
  path: string;
  hash: string;
  size: number;
  executable: boolean;
}

export interface PluginPermission {
  type: 'file_system' | 'network' | 'database' | 'credentials' | 'system';
  resource: string;
  access: 'read' | 'write' | 'execute';
  required: boolean;
}

export interface PluginConfiguration {
  id: string;
  name: string;
  type: 'string' | 'number' | 'boolean' | 'object' | 'array' | 'json';
  required: boolean;
  default?: any;
  description?: string;
  validation?: ValidationRule[];
}

export interface ValidationRule {
  type: 'required' | 'pattern' | 'min' | 'max' | 'enum' | 'custom';
  value: any;
  message: string;
}

export interface PluginHooks {
  install?: string;
  uninstall?: string;
  update?: string;
  validate?: string;
}

export interface SecurityPolicy {
  allowedDomains?: string[];
  allowedProtocols?: string[];
  maxFileSize?: number;
  maxExecutionTime?: number;
  memoryLimit?: number;
  cpuLimit?: number;
  networkAccess?: boolean;
  fileSystemAccess?: boolean;
  environmentVariables?: string[];
}

export interface PluginInstallationReport {
  success: boolean;
  errors: string[];
  warnings: string[];
  installedFiles: string[];
  permissionsGranted: PluginPermission[];
  securityPolicy: SecurityPolicy;
}