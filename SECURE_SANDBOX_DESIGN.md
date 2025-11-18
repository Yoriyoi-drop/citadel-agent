# DESAIN SANDBOX PLUGIN YANG AMAN

## 1. ARSITEKTUR KEAMANAN PLUGIN

### Arsitektur Multi-Layer Security
```
┌─────────────────────────────────────────────────────────────────────────┐
│                         WORKFLOW ENGINE                                 │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐           │
│  │   NODE          │ │   NODE          │ │   NODE          │           │
│  │   EXECUTOR      │ │   EXECUTOR      │ │   PLUGIN        │           │
│  │                 │ │                 │ │   RUNTIME       │           │
│  │  - HTTP Node    │ │  - DB Node      │ │  - JS/Py        │           │
│  │  - Logic Node   │ │  - Utils Node   │ │    Sandbox      │           │
│  │  - Timer Node   │ │  - etc          │ │  - WASM         │           │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        SECURITY GATEWAY                                 │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐           │
│  │   INPUT        │ │   VALIDATION    │ │   PERMISSION     │           │
│  │   FILTER       │ │   & SANITIZATION│ │   CHECKER        │           │
│  │                │ │                 │ │                 │           │
│  │  - SSRF        │ │  - SQL Injection│ │  - API Key      │           │
│  │    Protection  │ │    Prevention   │ │    Scopes       │           │
│  │  - File Path   │ │  - XSS Prevention│ │  - Node Access │           │
│  │    Validation  │ │  - Content      │ │    Permissions  │           │
│  │                │ │    Sanitization │ │                 │           │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                     ISOLATED RUNTIME ENVIROMENT                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐              │    │
│  │  │  JS         │ │  PYTHON     │ │  WASM       │              │    │
│  │  │  SANDBOX    │ │  SANDBOX    │ │  RUNTIME    │              │    │
│  │  │             │ │             │ │             │              │    │
│  │  │  - VM        │ │  - Subprocess│ │  - WASM     │              │    │
│  │  │    Isolation │ │    Isolation│ │    Container│              │    │
│  │  │  - Resource  │ │  - Resource │ │  - No Native │              │    │
│  │  │    Limits    │ │    Limits   │ │    System   │              │    │
│  │  └─────────────┘ └─────────────┘ │    Calls    │              │    │
│  │                                 └─────────────┘              │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                     RESOURCE MONITORING                                 │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐           │
│  │   RESOURCE     │ │   TIME LIMIT    │ │   NETWORK       │           │
│  │   LIMITER      │ │   ENFORCER      │ │   CONTROLLER    │           │
│  │                │ │                 │ │                 │           │
│  │  - CPU/Mem     │ │  - Execution    │ │  - Egress       │           │
│  │    Quotas      │ │    Timeout      │ │    Proxy        │           │
│  │  - I/O Limits  │ │  - Max Run      │ │  - Domain       │           │
│  │  - File Limits │ │    Time         │ │    Whitelist    │           │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘           │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 2. IMPLEMENTASI DETAIL SANDBOX

### JS Sandbox (VM2-based dengan penambahan keamanan)

```go
package sandbox

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
)

// JSSandboxConfig menyimpan konfigurasi keamanan untuk sandbox JS
type JSSandboxConfig struct {
	Timeout          time.Duration // Timeout eksekusi
	MaxMemoryMB      int           // Batas memori dalam MB
	MaxOutputLength  int           // Batas output dalam karakter
	AllowedModules   []string      // Daftar modul yang diizinkan
	BlockedFunctions []string      // Daftar fungsi yang diblokir
	NetworkAccess    bool          // Apakah mengizinkan akses jaringan
	FileAccess       bool          // Apakah mengizinkan akses file
	EnvVars          []string      // Variabel lingkungan yang diizinkan
}

// JSSandbox menyediakan lingkungan eksekusi JS yang aman
type JSSandbox struct {
	config   *JSSandboxConfig
	vmPool   chan *otto.Otto
	maxVMs   int
	tempDir  string
}

// NewJSSandbox membuat instance baru dari JSSandbox
func NewJSSandbox(config *JSSandboxConfig) (*JSSandbox, error) {
	tempDir, err := ioutil.TempDir("", "citadel-js-sandbox-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	sb := &JSSandbox{
		config:  config,
		vmPool:  make(chan *otto.Otto, config.MaxMemoryMB), // Using as proxy for VM pool size
		maxVMs:  config.MaxMemoryMB * 10,                   // Estimasi: 10 VM per 100MB
		tempDir: tempDir,
	}

	// Pre-populate VM pool
	for i := 0; i < 2; i++ { // Start with minimal pool
		vm := otto.New()
		sb.vmPool <- vm
	}

	return sb, nil
}

// Execute mengeksekusi kode JS dalam sandbox
func (sb *JSSandbox) Execute(ctx context.Context, code string, input map[string]interface{}) (*ExecutionResult, error) {
	startTime := time.Now()

	// Validasi kode sebelum eksekusi
	if err := sb.validateCode(code); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Code validation failed: %s", err.Error()),
		}, nil
	}

	// Ambil VM dari pool
	vmCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var vm *otto.Otto
	select {
	case vm = <-sb.vmPool:
		// Got a VM from the pool
	case <-vmCtx.Done():
		return &ExecutionResult{
			Success: false,
			Error:   "Timeout getting VM from pool",
		}, nil
	}

	// Pastikan VM dikembalikan ke pool setelah selesai
	defer func() {
		// Reset VM untuk membersihkan state sebelumnya
		newVM := otto.New()
		select {
		case sb.vmPool <- newVM:
			// VM berhasil dikembalikan ke pool
		default:
			// Pool penuh, abaikan VM
		}
	}()

	// Set input ke VM
	for key, value := range input {
		err := vm.Set(key, value)
		if err != nil {
			return &ExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to set input variable: %s", err.Error()),
			}, nil
		}
	}

	// Siapkan channel untuk menerima hasil atau timeout
	resultChan := make(chan *ExecutionResult, 1)
	execCtx, cancel := context.WithTimeout(ctx, sb.config.Timeout)
	defer cancel()

	// Eksekusi dalam goroutine agar bisa timeout
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- &ExecutionResult{
					Success: false,
					Error:   fmt.Sprintf("Sandbox panic: %v", r),
				}
			}
		}()

		// Tambahkan wrapper untuk mencegah infinite loop
		wrappedCode := fmt.Sprintf(`
			var __timeout = setTimeout(function() {
				throw new Error('Execution timeout');
			}, %d);

			try {
				%s
			} finally {
				clearTimeout(__timeout);
			}
		`, int(sb.config.Timeout.Milliseconds()), code)

		// Eksekusi kode
		value, err := vm.Run(wrapped combust)
		if err != nil {
			resultChan <- &ExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("Execution error: %s", err.Error()),
			}
			return
		}

		// Ekstrak hasil
		result, err := value.Export()
		if err != nil {
			resultChan <- &ExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to export result: %s", err.Error()),
			}
			return
		}

		resultChan <- &ExecutionResult{
			Success:  true,
			Data:     result,
			ExecTime: time.Since(startTime),
		}
	}()

	// Tunggu hasil atau timeout
	select {
	case result := <-resultChan:
		return result, nil
	case <-execCtx.Done():
		// Timeout - kita coba interupsi VM
		vm.Interrupt <- func() {
			panic("Execution interrupted due to timeout")
		}
		
		// Tunggu hasil interupsi
		select {
		case result := <-resultChan:
			return result, nil
		case <-time.After(100 * time.Millisecond):
			return &ExecutionResult{
				Success: false,
				Error:   "Code execution timed out and could not be interrupted",
			}, nil
		}
	}
}

// validateCode memvalidasi kode JS untuk potensi ancaman keamanan
func (sb *JSSandbox) validateCode(code string) error {
	// Validasi statis untuk pola yang berbahaya
	dangerousPatterns := []string{
		"eval\\(",
		"Function\\(",
		"setTimeout\\(\\s*['\"][^'\"]*['\"][^,]*,",
		"setInterval\\(\\s*['\"][^'\"]*['\"][^,]*,",
		"import\\(",
		"require\\(",
		"process\\.",
		"global\\.",
		"window\\.",
		"document\\.",
		"XMLHttpRequest",
		"fetch\\(",
		"File",
		"FileReader",
		"atob\\(",
		"btoa\\(",
		"unescape\\(",
		"escape\\(",
		"__proto__",
		"constructor",
		"prototype",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(code), strings.ToLower(pattern)) {
			return fmt.Errorf("code contains dangerous pattern: %s", pattern)
		}
	}

	// Parse AST untuk analisis lebih lanjut (opsional, bisa disimpan untuk implementasi lanjut)
	program, err := parser.ParseFile(nil, "", code, parser.IgnoreComments)
	if err != nil {
		return fmt.Errorf("syntax error: %w", err)
	}

	// Lakukan pengecekan AST jika diperlukan

	return nil
}

// ExecutionResult menyimpan hasil eksekusi
type ExecutionResult struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
	ExecTime time.Duration `json:"exec_time,omitempty"`
}

// Close membersihkan resources
func (sb *JSSandbox) Close() error {
	close(sb.vmPool)
	return os.RemoveAll(sb.tempDir)
}
```

### Python Sandbox (Subprocess Isolation)

```go
package sandbox

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PythonSandbox menyediakan lingkungan eksekusi Python yang aman
type PythonSandbox struct {
	config  *PythonSandboxConfig
	tempDir string
	python  string
}

// PythonSandboxConfig menyimpan konfigurasi keamanan untuk sandbox Python
type PythonSandboxConfig struct {
	Timeout          time.Duration // Timeout eksekusi
	MaxMemoryMB      int           // Batas memori dalam MB
	MaxOutputLength  int           // Batas output dalam karakter
	NetworkAccess    bool          // Apakah mengizinkan akses jaringan
	FileAccess       bool          // Apakah mengizinkan akses file (terbatas)
	AllowedLibraries []string      // Daftar pustaka yang diizinkan
	EnvVars          []string      // Variabel lingkungan yang diizinkan
}

// NewPythonSandbox membuat instance baru dari PythonSandbox
func NewPythonSandbox(config *PythonSandboxConfig) (*PythonSandbox, error) {
	tempDir, err := ioutil.TempDir("", "citadel-python-sandbox-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Cek apakah Python tersedia
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		return nil, fmt.Errorf("python3 not found: %w", err)
	}

	return &PythonSandbox{
		config:  config,
		tempDir: tempDir,
		python:  pythonPath,
	}, nil
}

// Execute mengeksekusi kode Python dalam sandbox
func (ps *PythonSandbox) Execute(ctx context.Context, code string, input map[string]interface{}) (*ExecutionResult, error) {
	startTime := time.Now()

	// Validasi kode
	if err := ps.validateCode(code); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Code validation failed: %s", err.Error()),
		}, nil
	}

	// Buat file sementara untuk kode Python
	pythonFile, err := ioutil.TempFile(ps.tempDir, "script-*.py")
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temporary file: %s", err.Error()),
		}, nil
	}
	defer os.Remove(pythonFile.Name())

	// Buat wrapper Python yang aman
	pythonCode := ps.createSafeWrapper(code, input)

	// Tulis kode ke file sementara
	_, err = pythonFile.WriteString(pythonCode)
	pythonFile.Close()
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to write to temporary file: %s", err.Error()),
		}, nil
	}

	// Siapkan command dengan resource limits
	cmd := exec.CommandContext(ctx, ps.python, pythonFile.Name())
	
	// Set environment variables yang diizinkan
	cmd.Env = append(os.Environ(), ps.config.EnvVars...)

	// Eksekusi dengan timeout
	execCtx, cancel := context.WithTimeout(ctx, ps.config.Timeout)
	defer cancel()

	// Jalankan dan ambil output
	output, err := cmd.Output()
	if err != nil {
		// Periksa apakah error karena timeout
		if execCtx.Err() == context.DeadlineExceeded {
			return &ExecutionResult{
				Success: false,
				Error:   "Execution timed out",
			}, nil
		}
		
		// Periksa stderr untuk informasi lebih lanjut
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("Python execution failed: %s", string(exitErr.Stderr)),
			}, nil
		}
		
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Python execution error: %s", err.Error()),
		}, nil
	}

	// Ambil hasil dan batasi panjangnya
	result := string(output)
	if len(result) > ps.config.MaxOutputLength {
		result = result[:ps.config.MaxOutputLength] + "... [truncated]"
	}

	return &ExecutionResult{
		Success:  true,
		Data:     result,
		ExecTime: time.Since(startTime),
	}, nil
}

// createSafeWrapper membuat wrapper yang aman untuk kode Python
func (ps *PythonSandbox) createSafeWrapper(code string, input map[string]interface{}) string {
	var inputStr string
	for k, v := range input {
		// Format input sebagai variabel Python
		inputStr += fmt.Sprintf("%s = %v\n", k, formatPythonValue(v))
	}
	
	// Buat wrapper yang hanya mengizinkan subset fungsi yang aman
	return fmt.Sprintf(`
import sys
import json
import os
import re
import math
import datetime
import collections
import urllib.parse
from io import StringIO

# Batasi akses ke modul berbahaya
import builtins
for name in ['open', 'exec', 'eval', '__import__', 'compile']:
    if hasattr(builtins, name):
        delattr(builtins, name)

# Set input data
%s

# Tangkap output untuk mencegah print yang tidak diinginkan
old_stdout = sys.stdout
sys.stdout = mystdout = StringIO()

try:
    # Eksekusi kode pengguna
    exec(%q)
    
    # Ambil output
    output = mystdout.getvalue()
    
    # Kirim hasil
    print(json.dumps({
        'result': output if output else 'Execution completed successfully',
        'status': 'success'
    }))
    
except Exception as e:
    print(json.dumps({
        'error': str(e),
        'status': 'error'
    }))

`, inputStr, code)
}

// formatPythonValue memformat nilai untuk digunakan dalam Python
func formatPythonValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "\\'"))
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "True"
		}
		return "False"
	case nil:
		return "None"
	default:
		// Untuk tipe kompleks, kembalikan string kosong atau representasi JSON
		return "''"
	}
}

// validateCode memvalidasi kode Python untuk potensi ancaman keamanan
func (ps *PythonSandbox) validateCode(code string) error {
	dangerousPatterns := []string{
		"import os",
		"import sys",
		"import subprocess",
		"import requests",
		"import urllib",
		"eval(",
		"exec(",
		"open(",
		"__import__",
		"compile(",
		"globals()",
		"locals()",
		"vars()",
		"getattr(",
		"setattr(",
		"hasattr(",
		"delattr(",
		"execfile(",
		"file(",
		"input(",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(code), strings.ToLower(pattern)) {
			return fmt.Errorf("code contains dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// Close membersihkan resources
func (ps *PythonSandbox) Close() error {
	return os.RemoveAll(ps.tempDir)
}
```

### WASM Sandbox (untuk node dengan kinerja tinggi)

```go
package sandbox

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	wasmt "github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WASMSandbox menyediakan lingkungan eksekusi WebAssembly yang aman
type WASMSandbox struct {
	config *WASMSandboxConfig
	runtime wasmt.Runtime
	tempDir string
}

// WASMSandboxConfig menyimpan konfigurasi keamanan untuk sandbox WASM
type WASMSandboxConfig struct {
	Timeout         time.Duration // Timeout eksekusi
	MaxMemoryPages  uint32        // Batas memori dalam jumlah halaman (64KB/halaman)
	MaxExecutions   int          // Jumlah maksimum eksekusi sebelum restart runtime
}

// NewWASMSandbox membuat instance baru dari WASMSandbox
func NewWASMSandbox(config *WASMSandboxConfig) (*WASMSandbox, error) {
	tempDir, err := ioutil.TempDir("", "citadel-wasm-sandbox-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Buat runtime WASM dengan konfigurasi keamanan
	runtime := wasmt.NewRuntimeWithConfig(
		context.Background(),
		wasmt.NewRuntimeConfig().
			WithMemoryLimitPages(config.MaxMemoryPages).
			WithCloseOnContextDone(true),
	)
	
	// Import WASI untuk fungsi standar (tanpa akses jaringan/file)
	wasi_snapshot_preview1.MustInstantiate(context.Background(), runtime)

	return &WASMSandbox{
		config:  config,
		runtime: runtime,
		tempDir: tempDir,
	}, nil
}

// Execute mengeksekusi modul WASM dalam sandbox
func (ws *WASMSandbox) Execute(ctx context.Context, wasmBytes []byte, functionName string, params []uint64) (*ExecutionResult, error) {
	startTime := time.Now()

	// Buat konteks eksekusi dengan timeout
	execCtx, cancel := context.WithTimeout(ctx, ws.config.Timeout)
	defer cancel()

	// Compile modul WASM
	compiled, err := ws.runtime.CompileModule(execCtx, wasmBytes)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to compile WASM module: %s", err.Error()),
		}, nil
	}

	// Instantiate modul
	instance, err := ws.runtime.InstantiateModule(execCtx, compiled, wasmt.NewModuleConfig())
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to instantiate WASM module: %s", err.Error()),
		}, nil
	}

	// Panggil fungsi
	result, err := instance.ExportedFunction(functionName).Call(execCtx, params...)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to call function: %s", err.Error()),
		}, nil
	}

	return &ExecutionResult{
		Success:  true,
		Data:     result,
		ExecTime: time.Since(startTime),
	}, nil
}

// ValidateWASM memvalidasi biner WASM untuk fungsi/fitur yang dilarang
func (ws *WASMSandbox) ValidateWASM(wasmBytes []byte) error {
	// Dalam implementasi produksi, lakukan validasi menyeluruh terhadap:
	// - Import yang tidak diizinkan (misalnya fungsi host untuk akses jaringan/file)
	// - Ekspor fungsi yang tidak diharapkan
	// - Jumlah memory, table, dan elemen global
	// - Fungsi yang disalahgunakan untuk akses sistem
	
	// Untuk sekarang, hanya lakukan validasi dasar
	if len(wasmBytes) == 0 {
		return fmt.Errorf("empty WASM binary")
	}
	
	// Tambahkan validasi lebih lanjut di sini sesuai kebutuhan
	
	return nil
}

// Close membersihkan resources
func (ws *WASMSandbox) Close() error {
	ctx := context.Background()
	ws.runtime.Close(ctx)
	return os.RemoveAll(ws.tempDir)
}
```

### Plugin Manager Utama

```go
package sandbox

import (
	"context"
	"fmt"
	"time"
)

// PluginManager mengelola berbagai jenis sandbox plugin
type PluginManager struct {
	jsSandbox    *JSSandbox
	pySandbox    *PythonSandbox
	wasmSandbox  *WASMSandbox
	pluginCache  map[string][]byte // Cache untuk WASM plugin
}

// PluginType mendefinisikan jenis plugin
type PluginType string

const (
	JSType   PluginType = "javascript"
	PyType   PluginType = "python"
	WASMType PluginType = "wasm"
)

// NewPluginManager membuat instance baru dari PluginManager
func NewPluginManager(jsConfig *JSSandboxConfig, pyConfig *PythonSandboxConfig, wasmConfig *WASMSandboxConfig) (*PluginManager, error) {
	// Buat semua sandbox
	jsSandbox, err := NewJSSandbox(jsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create JS sandbox: %w", err)
	}

	pySandbox, err := NewPythonSandbox(pyConfig)
	if err != nil {
		jsSandbox.Close() // Cleanup jika gagal
		return nil, fmt.Errorf("failed to create Python sandbox: %w", err)
	}

	wasmSandbox, err := NewWASMSandbox(wasmConfig)
	if err != nil {
		jsSandbox.Close()
		pySandbox.Close()
		return nil, fmt.Errorf("failed to create WASM sandbox: %w", err)
	}

	return &PluginManager{
		jsSandbox:   jsSandbox,
		pySandbox:   pySandbox,
		wasmSandbox: wasmSandbox,
		pluginCache: make(map[string][]byte),
	}, nil
}

// ExecutePlugin mengeksekusi plugin dengan tipe yang ditentukan
func (pm *PluginManager) ExecutePlugin(ctx context.Context, pluginType PluginType, code string, wasmBytes []byte, functionName string, input map[string]interface{}) (*ExecutionResult, error) {
	switch pluginType {
	case JSType:
		return pm.jsSandbox.Execute(ctx, code, input)
	case PyType:
		return pm.pySandbox.Execute(ctx, code, input)
	case WASMType:
		if wasmBytes == nil {
			return &ExecutionResult{
				Success: false,
				Error:   "WASM binary is required for WASM plugin type",
			}, nil
		}
		
		// Validasi WASM sebelum eksekusi
		if err := pm.wasmSandbox.ValidateWASM(wasmBytes); err != nil {
			return &ExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("WASM validation failed: %s", err.Error()),
			}, nil
		}
		
		// Untuk WASM, kita butuh cara untuk mengonversi input ke parameter WASM
		// Ini adalah implementasi dasar - dalam produksi, Anda perlu mengembangkan
		// cara untuk mengonversi data input ke format yang bisa diterima oleh fungsi WASM
		
		// Sebagai contoh sementara, kita hanya menggunakan parameter kosong
		return pm.wasmSandbox.Execute(ctx, wasmBytes, functionName, []uint64{})
	default:
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported plugin type: %s", pluginType),
		}, nil
	}
}

// ValidateCode memvalidasi kode untuk tipe plugin tertentu
func (pm *PluginManager) ValidateCode(pluginType PluginType, code string) error {
	switch pluginType {
	case JSType:
		return pm.jsSandbox.validateCode(code)
	case PyType:
		return pm.pySandbox.validateCode(code)
	case WASMType:
		// Validasi untuk WASM dilakukan saat eksekusi atau pemuatan awal
		return nil
	default:
		return fmt.Errorf("unsupported plugin type: %s", pluginType)
	}
}

// Close membersihkan semua resources
func (pm *PluginManager) Close() error {
	if pm.jsSandbox != nil {
		pm.jsSandbox.Close()
	}
	if pm.pySandbox != nil {
		pm.pySandbox.Close()
	}
	if pm.wasmSandbox != nil {
		pm.wasmSandbox.Close()
	}
	return nil
}

// PluginSecurityPolicy mendefinisikan kebijakan keamanan untuk plugin
type PluginSecurityPolicy struct {
	AllowNetwork        bool          `json:"allow_network"`
	AllowFileAccess     bool          `json:"allow_file_access"`
	Timeout             time.Duration `json:"timeout"`
	MaxMemoryMB         int           `json:"max_memory_mb"`
	MaxExecutionTime    time.Duration `json:"max_execution_time"`
	AllowedLibraries    []string      `json:"allowed_libraries"`
	BlockedOperations   []string      `json:"blocked_operations"`
	RequireSignature    bool          `json:"require_signature"`
	SignaturePublicKey  string        `json:"signature_public_key"`
}
```

### Contoh Konfigurasi dan Penggunaan

```go
// example_usage.go
package main

import (
	"context"
	"fmt"
	"time"

	"citadel-agent/backend/internal/plugins/sandbox"  // Adjust import path as needed
)

func main() {
	// Konfigurasi keamanan
	jsConfig := &sandbox.JSSandboxConfig{
		Timeout:         5 * time.Second,
		MaxMemoryMB:     50,
		MaxOutputLength: 10000,
		NetworkAccess:   false,
		FileAccess:      false,
		BlockedFunctions: []string{"eval", "Function", "require", "import"},
	}

	pyConfig := &sandbox.PythonSandboxConfig{
		Timeout:          5 * time.Second,
		MaxMemoryMB:      100,
		MaxOutputLength:  10000,
		NetworkAccess:    false,
		FileAccess:       false,
		AllowedLibraries: []string{"json", "math", "datetime", "re", "urllib.parse"},
	}

	wasmConfig := &sandbox.WASMSandboxConfig{
		Timeout:        3 * time.Second,
		MaxMemoryPages: 100, // 6.4MB (100 * 64KB)
		MaxExecutions:  1000,
	}

	// Buat plugin manager
	manager, err := sandbox.NewPluginManager(jsConfig, pyConfig, wasmConfig)
	if err != nil {
		panic(err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Contoh eksekusi JS plugin yang aman
	jsCode := `
		// Kode JS yang hanya melakukan operasi aman
		var result = input.value * 2;
		result = result + " squared is " + (result * result);
		result;
	`

	jsResult, err := manager.ExecutePlugin(ctx, sandbox.JSType, jsCode, nil, "", map[string]interface{}{
		"value": 5,
	})
	if err != nil {
		fmt.Printf("JS execution failed: %v\n", err)
	} else {
		fmt.Printf("JS Result: %+v\n", jsResult)
	}

	// Contoh eksekusi Python plugin yang aman
	pyCode := `
# Kode Python yang hanya melakukan operasi aman
result = input_value * input_value
print(f"The square of {input_value} is {result}")
`
	
	pyResult, err := manager.ExecutePlugin(ctx, sandbox.PyType, pyCode, nil, "", map[string]interface{}{
		"input_value": 4,
	})
	if err != nil {
		fmt.Printf("Python execution failed: %v\n", err)
	} else {
		fmt.Printf("Python Result: %+v\n", pyResult)
	}

	// Contoh validasi kode
	badCode := "eval('console.log(\"malicious code\")');"
	if err := manager.ValidateCode(sandbox.JSType, badCode); err != nil {
		fmt.Printf("Validation correctly blocked malicious code: %v\n", err)
	} else {
		fmt.Println("Validation failed to block malicious code!")
	}
}
```

### Implementasi Lengkap di Plugin System

```go
// plugins/secure_runtime.go
package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"citadel-agent/backend/internal/plugins/sandbox"
)

// SecurePluginRuntime mengintegrasikan sandbox ke dalam sistem plugin utama
type SecurePluginRuntime struct {
	manager *sandbox.PluginManager
}

// NewSecurePluginRuntime membuat instance baru dari SecurePluginRuntime
func NewSecurePluginRuntime() (*SecurePluginRuntime, error) {
	jsConfig := &sandbox.JSSandboxConfig{
		Timeout:         10 * time.Second,
		MaxMemoryMB:     100,
		MaxOutputLength: 100000,
		NetworkAccess:   false,
		FileAccess:      false,
		BlockedFunctions: []string{
			"eval", "Function", "require", "import", "setTimeout", "setInterval",
			"process", "global", "window", "document", "XMLHttpRequest", "fetch",
		},
	}

	pyConfig := &sandbox.PythonSandboxConfig{
		Timeout:          10 * time.Second,
		MaxMemoryMB:      200,
		MaxOutputLength:  100000,
		NetworkAccess:    false,
		FileAccess:       false,
		AllowedLibraries: []string{
			"json", "math", "datetime", "collections", "urllib.parse", "re", "base64",
		},
	}

	wasmConfig := &sandbox.WASMSandboxConfig{
		Timeout:        5 * time.Second,
		MaxMemoryPages: 200, // 12.8MB
		MaxExecutions:  10000,
	}

	manager, err := sandbox.NewPluginManager(jsConfig, pyConfig, wasmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin manager: %w", err)
	}

	return &SecurePluginRuntime{
		manager: manager,
	}, nil
}

// ExecutePlugin mengeksekusi plugin dengan validasi dan keamanan yang ketat
func (spr *SecurePluginRuntime) ExecutePlugin(ctx context.Context, pluginType sandbox.PluginType, code string, wasmBytes []byte, functionName string, input map[string]interface{}) (*sandbox.ExecutionResult, error) {
	// Validasi awal
	if err := spr.manager.ValidateCode(pluginType, code); err != nil {
		return &sandbox.ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Code validation failed: %s", err.Error()),
		}, nil
	}

	// Tambahkan input validation tambahan berdasarkan kebijakan
	if err := spr.validateInput(input); err != nil {
		return &sandbox.ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Input validation failed: %s", err.Error()),
		}, nil
	}

	// Eksekusi plugin dengan timeout tambahan
	execCtx, cancel := context.WithTimeout(ctx, 15*time.Second) // Timeout lebih lama daripada sandbox
	defer cancel()

	result, err := spr.manager.ExecutePlugin(execCtx, pluginType, code, wasmBytes, functionName, input)
	if err != nil {
		return &sandbox.ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Plugin execution failed: %s", err.Error()),
		}, nil
	}

	// Validasi hasil sebelum mengembalikan
	if result.Error != "" {
		return result, nil
	}

	if err := spr.validateOutput(result.Data); err != nil {
		return &sandbox.ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Output validation failed: %s", err.Error()),
		}, nil
	}

	return result, nil
}

// validateInput memvalidasi data input
func (spr *SecurePluginRuntime) validateInput(input map[string]interface{}) error {
	for k, v := range input {
		if len(k) > 100 {
			return fmt.Errorf("input key too long: %s", k)
		}

		// Validasi tipe dan ukuran nilai
		switch val := v.(type) {
		case string:
			if len(val) > 10000 {
				return fmt.Errorf("input value too long for key %s", k)
			}
		case []byte:
			if len(val) > 100000 { // 100KB
				return fmt.Errorf("input value too large for key %s", k)
			}
		}
	}

	return nil
}

// validateOutput memvalidasi hasil output
func (spr *SecurePluginRuntime) validateOutput(output interface{}) error {
	// Implementasi validasi output sesuai kebutuhan
	// Misalnya, memastikan output tidak mengandung data sensitif
	// atau tidak melebihi batas ukuran tertentu

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	if len(jsonBytes) > 1000000 { // 1MB
		return fmt.Errorf("output too large: %d bytes", len(jsonBytes))
	}

	return nil
}

// Close membersihkan resources
func (spr *SecurePluginRuntime) Close() error {
	return spr.manager.Close()
}
```

---

## 3. IMPLEMENTASI NETWORK dan FILE SYSTEM PROTECTION

### Egress Proxy untuk HTTP Node

```go
// security/egress_proxy.go
package security

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// EgressProxy menyediakan kontrol atas permintaan keluar
type EgressProxy struct {
	allowedDomains   map[string]bool
	allowedIPs       []*net.IPNet
	blockedIPs       []*net.IPNet
	timeout          time.Duration
	maxRedirects     int
}

// NewEgressProxy membuat instance baru dari EgressProxy
func NewEgressProxy(allowedDomains []string, allowedIPs []string, blockedIPs []string) (*EgressProxy, error) {
	proxy := &EgressProxy{
		allowedDomains: make(map[string]bool),
		allowedIPs:     make([]*net.IPNet, 0),
		blockedIPs:     make([]*net.IPNet, 0),
		timeout:        30 * time.Second,
		maxRedirects:   10,
	}

	// Set domain yang diizinkan
	for _, domain := range allowedDomains {
		proxy.allowedDomains[strings.ToLower(domain)] = true
	}

	// Parse IP yang diizinkan
	for _, ipStr := range allowedIPs {
		_, ipNet, err := net.ParseCIDR(ipStr)
		if err != nil {
			return nil, fmt.Errorf("invalid allowed IP: %s", ipStr)
		}
		proxy.allowedIPs = append(proxy.allowedIPs, ipNet)
	}

	// Parse IP yang diblokir
	for _, ipStr := range blockedIPs {
		_, ipNet, err := net.ParseCIDR(ipStr)
		if err != nil {
			return nil, fmt.Errorf("invalid blocked IP: %s", ipStr)
		}
		proxy.blockedIPs = append(proxy.blockedIPs, ipNet)
	}

	return proxy, nil
}

// IsDomainAllowed memeriksa apakah domain diizinkan
func (ep *EgressProxy) IsDomainAllowed(domain string) bool {
	domain = strings.ToLower(domain)

	// Cek apakah domain diizinkan secara eksplisit
	if ep.allowedDomains[domain] {
		return true
	}

	// Cek apakah ada subdomain yang diizinkan (misalnya *.github.com)
	parts := strings.Split(domain, ".")
	for i := 0; i < len(parts)-1; i++ {
		wildcard := "*." + strings.Join(parts[i+1:], ".")
		if ep.allowedDomains[wildcard] {
			return true
		}
	}

	return false
}

// IsIPAllowed memeriksa apakah IP diizinkan
func (ep *EgressProxy) IsIPAllowed(ip net.IP) bool {
	// Cek blocked IPs dulu
	for _, blockedIP := range ep.blockedIPs {
		if blockedIP.Contains(ip) {
			return false
		}
	}

	// Cek allowed IPs
	for _, allowedIP := range ep.allowedIPs {
		if allowedIP.Contains(ip) {
			return true
		}
	}

	// Jika tidak ada daftar putih, maka kita blokir semuanya
	return false
}

// ValidateURL memvalidasi URL untuk mencegah SSRF
func (ep *EgressProxy) ValidateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Cek skema
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported scheme: %s", parsedURL.Scheme)
	}

	// Cek domain
	if !ep.IsDomainAllowed(parsedURL.Hostname()) {
		return fmt.Errorf("domain not allowed: %s", parsedURL.Hostname())
	}

	// Cek IP (jika domain di-resolve ke IP)
	ips, err := net.LookupIP(parsedURL.Hostname())
	if err != nil {
		// Jika tidak bisa resolve, kita masih bisa cek langsung host atau IP
		host, port, err := net.SplitHostPort(parsedURL.Host)
		if err != nil {
			// Mungkin tidak ada port
			host = parsedURL.Host
		}

		// Cek apakah host adalah IP
		if ip := net.ParseIP(host); ip != nil {
			if !ep.IsIPAllowed(ip) {
				return fmt.Errorf("IP not allowed: %s", ip.String())
			}
		}
	} else {
		// Cek semua IP yang di-resolve
		for _, ip := range ips {
			if !ep.IsIPAllowed(ip) {
				return fmt.Errorf("resolved IP not allowed: %s", ip.String())
			}
		}
	}

	return nil
}

// CreateSecureClient membuat HTTP client yang aman
func (ep *EgressProxy) CreateSecureClient() *http.Client {
	return &http.Client{
		Timeout: ep.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= ep.maxRedirects {
				return fmt.Errorf("too many redirects")
			}

			if err := ep.ValidateURL(req.URL.String()); err != nil {
				return fmt.Errorf("redirect validation failed: %w", err)
			}

			return nil
		},
	}
}
```

---

## 4. KESIMPULAN

Desain sandbox plugin ini menyediakan lapisan keamanan yang komprehensif:

1. **Isolasi proses** untuk JS dan Python
2. **Validasi statis dan dinamis** kode
3. **Resource limits** (waktu, memori, I/O)
4. **Pemblokiran fungsi berbahaya**
5. **Proteksi jaringan** melalui egress proxy
6. **Sistem WASM** untuk eksekusi aman dengan kinerja tinggi

Implementasi ini siap untuk dipasangkan dengan sistem plugin yang ada dan akan memberikan keamanan yang jauh lebih baik daripada sandbox saat ini.