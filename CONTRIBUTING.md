# Contributing to Citadel Agent

Thank you for your interest in contributing to Citadel Agent! We appreciate your help in building the world's most powerful self-hosted workflow automation platform.

## Table of Contents
1. [How Can I Contribute?](#how-can-i-contribute)
2. [Development Setup](#development-setup)
3. [Coding Standards](#coding-standards)
4. [Pull Requests](#pull-requests)
5. [Reporting Issues](#reporting-issues)

## How Can I Contribute?

### 🐛 Bug Reports
We're not perfect (yet!), and bug reports are incredibly valuable to us. When reporting a bug:
- Use a clear and descriptive title
- Describe the exact steps to reproduce
- Expected vs actual behavior
- Include relevant logs or error messages

### 💡 Feature Requests
Have ideas for improvement? Great!
- Search existing issues first
- Explain why this feature would be valuable
- Provide examples of how you'd use it
- Consider implementation complexity

### 🧪 Testing
- Write unit tests for new features
- Improve existing tests
- Report failing tests
- Performance testing

### 📝 Documentation
- Fix typos
- Improve clarity
- Add examples
- Translate to other languages

## Development Setup

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose
- Git

### Local Development

```bash
# 1. Fork and clone the repository
git clone https://github.com/your-username/citadel-agent.git
cd citadel-agent

# 2. Create a new branch
git checkout -b feature-amazing-feature

# 3. Install backend dependencies
cd backend && go mod download

# 4. Install frontend dependencies
cd ../frontend && npm install

# 5. Set up environment
cp .env.example .env
# Edit .env with your local configuration

# 6. Start the development stack
docker-compose up -d postgres redis temporal

# 7. Run backend separately
cd ../backend && go run cmd/api/main.go

# 8. Run frontend separately
cd ../frontend && npm run dev
```

### Frontend Development
```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Run tests
npm test
```

### Backend Development
```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Run specific package tests
go test ./internal/nodes/http

# Build
go build -o bin/server cmd/api/main.go

# Run linter
golangci-lint run
```

## Coding Standards

### Go Standards
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Write clear, idiomatic Go code
- Always handle errors appropriately
- Write tests for all new functionality

### Naming Conventions
```go
// Use descriptive names
func ValidateWorkflow(w *Workflow) error { /* */ }

// Use context for cancellation
func ProcessNode(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) { /* */ }

// Use struct tags for JSON
type NodeConfig struct {
    Name     string `json:"name"`
    Enabled  bool   `json:"enabled"`
}
```

### React/TypeScript Standards
- Follow [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript)
- Use TypeScript for all new code
- Write reusable, composable components
- Use functional components with hooks
- Follow React best practices

```tsx
// Use functional components
const MyComponent: React.FC<MyProps> = ({ prop }) => {
  return <div>{prop}</div>;
};

// Use TypeScript interfaces
interface MyProps {
  title: string;
  count?: number;
}
```

### Documentation Standards
- Document all exported functions/types
- Write clear, concise comments
- Include usage examples
- Keep documentation up-to-date

```go
// ProcessNode executes a workflow node with the given inputs.
//
// Example:
//   result, err := ProcessNode(ctx, map[string]interface{}{"data": "input"})
func ProcessNode(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
  // implementation
}
```

## Pull Requests

### Before Submitting
- Follow the coding standards
- Write clear, detailed descriptions
- Include tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### PR Checklist
- [ ] Code follows project standards
- [ ] New functionality is properly tested
- [ ] Documentation is updated
- [ ] Breaking changes are documented
- [ ] Related issues are linked (if applicable)

### PR Description Template
```
## Summary
Brief description of the changes

## Changes
- Added/Fixed/Improved feature A
- Resolved issue B
- Updated documentation for C

## Testing
- Tested manually
- Added unit tests
- Verified performance impact

## Breaking Changes
- List of breaking changes (if any)

## Related Issues
- Fixes #123
- Related to #456
```

## Reporting Issues

### Good Issue Reports
- Clear, descriptive title
- Steps to reproduce
- Environment details (OS, version, etc.)
- Expected vs actual behavior
- Relevant logs or screenshots

### Before Creating an Issue
- Search for existing issues
- Check the documentation
- Try the latest version
- Consider if it's a usage question (use discussions instead)

### Security Issues
For security-related issues, please contact us directly at security@citadel-agent.com instead of creating a public issue.

## Development Philosophy

### Keep It Simple
- Prefer simple solutions over complex ones
- Avoid premature optimization
- Focus on user needs

### Make It Robust
- Handle edge cases
- Fail gracefully
- Log appropriately
- Consider security implications

### Maintain Quality
- Write tests for new features
- Refactor existing code when improving
- Keep functions small and focused
- Document complex logic

## Need Help?

- Check the documentation
- Look at existing code examples
- Ask in our community Discord
- Create a draft PR to discuss your approach

Thank you for contributing to Citadel Agent! Together, we're building the most powerful self-hosted automation platform the world has ever seen.

---

*This guide is a living document. If you see anything that could be improved, please submit a PR!* ❤️