// backend/internal/runtimes/types.go
package runtimes

// RuntimeType represents the type of runtime
type RuntimeType string

const (
	RuntimeGo      RuntimeType = "go"
	RuntimeJS      RuntimeType = "javascript"
	RuntimePython  RuntimeType = "python"
	RuntimeJava    RuntimeType = "java"
	RuntimeRuby    RuntimeType = "ruby"
	RuntimePHP     RuntimeType = "php"
	RuntimeRust    RuntimeType = "rust"
	RuntimeCSharp  RuntimeType = "csharp"
	RuntimeShell   RuntimeType = "shell"
)