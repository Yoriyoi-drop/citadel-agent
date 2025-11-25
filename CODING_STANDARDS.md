# Coding Standards & Naming Conventions

## File Naming

### Components (PascalCase)
- React components: `NodeEditor.tsx`, `WorkflowBuilder.tsx`
- UI components: `Button.tsx`, `Input.tsx`

### Utilities & Helpers (camelCase)
- Store files: `workflowStore.ts`, `nodeStore.ts`, `authStore.ts`
- Utility files: `utils.ts`, `constants.ts`
- Hook files: `use-toast.ts`, `use-workflow.ts`

### Types & Interfaces (camelCase)
- Type definition files: `workflow.ts`, `node.ts`

### Test Files (match source)
- Component tests: `NodeEditor.test.tsx`
- Utility tests: `utils.test.ts`

## Code Organization

### Directory Structure
```
src/
├── app/              # Next.js App Router pages
├── components/       # React components
│   ├── ui/          # UI primitives (shadcn)
│   └── workflow/    # Feature-specific components
├── stores/          # Zustand state management
├── types/           # TypeScript type definitions
├── lib/             # Utilities and helpers
│   ├── constants.ts
│   └── utils.ts
└── hooks/           # Custom React hooks
```

## TypeScript Best Practices

### Interfaces vs Types
- Use `interface` for object shapes that might be extended
- Use `type` for unions, intersections, and primitives

### Generic Naming
- Use descriptive names: `TNode`, `TWorkflow` (prefix with T)
- Or use full words: `NodeType`, `WorkflowType`

### Props Naming
- Component props: `[ComponentName]Props`
- Example: `NodeEditorProps`, `ConfigFieldProps`

## Component Structure

```typescript
// 1. Imports
import React from 'react';
import { useStore } from '@/stores/store';

// 2. Types/Interfaces
interface ComponentProps {
  // ...
}

// 3. Constants (if any)
const CONSTANT_VALUE = 'value';

// 4. Component
export function Component({ prop }: ComponentProps) {
  // a. Hooks
  const state = useStore();
  
  // b. State
  const [local, setLocal] = useState();
  
  // c. Effects
  useEffect(() => {}, []);
  
  // d. Handlers
  const handleClick = () => {};
  
  // e. Render
  return <div>Content</div>;
}

// 5. Sub-components or exports
```

## Variable Naming

### Boolean Variables
- Prefix with `is`, `has`, `should`, `can`
- Examples: `isLoading`, `hasError`, `shouldRender`, `canEdit`

### Event Handlers
- Prefix with `handle` or `on`
- Examples: `handleClick`, `handleSubmit`, `onNodeSelect`

### Constants
- SCREAMING_SNAKE_CASE for true constants
- camelCase for configuration objects

```typescript
const MAX_RETRIES = 3;
const API_ENDPOINT = 'https://api.example.com';

const config = {
  autoSave: true,
  retryCount: 3
};
```

## Import Order

1. External libraries
2. Internal absolute imports (@/)
3. Relative imports
4. Types
5. Styles

```typescript
import React from 'react';
import { useRouter } from 'next/navigation';

import { Button } from '@/components/ui/button';
import { useWorkflowStore } from '@/stores/workflowStore';

import { ConfigField } from './ConfigField';

import type { NodeType } from '@/types/workflow';
```

## Comments

### JSDoc for Functions
```typescript
/**
 * Format duration in milliseconds to human readable string
 * @param ms - Duration in milliseconds
 * @returns Formatted duration string (e.g., "1.5s", "2m")
 */
export function formatDuration(ms: number): string {
  // Implementation
}
```

### Inline Comments
- Use for complex logic explanation
- Keep brief and to the point
- Prefer self-documenting code

## Testing

### Test File Naming
- Match source file: `Component.test.tsx`
- Place in same directory or `__tests__` folder

### Test Structure
```typescript
describe('ComponentName', () => {
  test('should do something', () => {
    // Arrange
    // Act
    // Assert
  });
});
```

## Git Commit Messages

### Format
```
type(scope): subject

body (optional)

footer (optional)
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance

### Examples
```
feat(workflow): add node execution with error handling
fix(ui): resolve missing closing tag in NodeEditor
docs(readme): update setup instructions
refactor(node): extract ConfigField component
```
