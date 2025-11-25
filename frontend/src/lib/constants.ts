/**
 * Node Categories
 */
export const NODE_CATEGORIES = {
    TRIGGER: 'trigger',
    ACTION: 'action',
    TRANSFORM: 'transform',
    UTILITY: 'utility',
    AI: 'ai',
    DATABASE: 'database',
    COMMUNICATION: 'communication',
} as const;

/**
 * Node Status
 */
export const NODE_STATUS = {
    IDLE: 'idle',
    RUNNING: 'running',
    SUCCESS: 'success',
    ERROR: 'error',
} as const;

/**
 * Workflow Settings
 */
export const WORKFLOW_ERROR_HANDLING = {
    STOP: 'stop',
    CONTINUE: 'continue',
    RETRY: 'retry',
} as const;

/**
 * Execution Status
 */
export const EXECUTION_STATUS = {
    PENDING: 'pending',
    RUNNING: 'running',
    COMPLETED: 'completed',
    FAILED: 'failed',
    CANCELLED: 'cancelled',
} as const;

/**
 * Default Workflow Settings
 */
export const DEFAULT_WORKFLOW_SETTINGS = {
    autoSave: true,
    errorHandling: WORKFLOW_ERROR_HANDLING.STOP,
    retryCount: 3,
};

/**
 * Port Types
 */
export const PORT_TYPES = {
    STRING: 'string',
    NUMBER: 'number',
    BOOLEAN: 'boolean',
    OBJECT: 'object',
    ARRAY: 'array',
    FILE: 'file',
} as const;

/**
 * Config Field Types
 */
export const CONFIG_FIELD_TYPES = {
    STRING: 'string',
    NUMBER: 'number',
    BOOLEAN: 'boolean',
    SELECT: 'select',
    MULTISELECT: 'multiselect',
    TEXTAREA: 'textarea',
    PASSWORD: 'password',
    FILE: 'file',
    JSON: 'json',
} as const;

/**
 * UI Constants
 */
export const UI_CONSTANTS = {
    ITEMS_PER_PAGE: 20,
    DEBOUNCE_DELAY: 300,
    TOAST_DURATION: 3000,
    MAX_RETRIES: 3,
};

/**
 * API Endpoints
 */
export const API_ENDPOINTS = {
    WORKFLOWS: '/api/workflows',
    NODES: '/api/nodes',
    EXECUTIONS: '/api/executions',
    TEMPLATES: '/api/templates',
    AI_CHAT: '/api/ai/chat',
} as const;
