// backend/internal/nodes/file/file_operations.go
package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// FileOperationType represents the type of file operation
type FileOperationType string

const (
	FileRead     FileOperationType = "read"
	FileWrite    FileOperationType = "write"
	FileAppend   FileOperationType = "append"
	FileDelete   FileOperationType = "delete"
	FileExists   FileOperationType = "exists"
	FileList     FileOperationType = "list"
	FileCopy     FileOperationType = "copy"
	FileMove     FileOperationType = "move"
	FileMkdir    FileOperationType = "mkdir"
	FileChmod    FileOperationType = "chmod"
)

// FileNodeConfig represents the configuration for a file node
type FileNodeConfig struct {
	Operation   FileOperationType `json:"operation"`
	FilePath    string           `json:"file_path"`
	Destination string           `json:"destination,omitempty"`
	Content     string           `json:"content,omitempty"`
	Recursive   bool             `json:"recursive,omitempty"`
	Permissions os.FileMode      `json:"permissions,omitempty"`
}

// FileNode represents a file system operation node
type FileNode struct {
	config *FileNodeConfig
}

// NewFileNode creates a new file node
func NewFileNode(config *FileNodeConfig) *FileNode {
	return &FileNode{
		config: config,
	}
}

// Execute executes the file operation node
func (fn *FileNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	filePath := fn.config.FilePath
	if path, exists := inputs["file_path"]; exists {
		if pathStr, ok := path.(string); ok {
			filePath = pathStr
		}
	}

	content := fn.config.Content
	if cont, exists := inputs["content"]; exists {
		if contStr, ok := cont.(string); ok {
			content = contStr
		}
	}

	destination := fn.config.Destination
	if dest, exists := inputs["destination"]; exists {
		if destStr, ok := dest.(string); ok {
			destination = destStr
		}
	}

	// Perform the file operation based on type
	switch fn.config.Operation {
	case FileRead:
		return fn.executeRead(filePath)
	case FileWrite:
		return fn.executeWrite(filePath, content)
	case FileAppend:
		return fn.executeAppend(filePath, content)
	case FileDelete:
		return fn.executeDelete(filePath)
	case FileExists:
		return fn.executeExists(filePath)
	case FileList:
		return fn.executeList(filePath)
	case FileCopy:
		return fn.executeCopy(filePath, destination)
	case FileMove:
		return fn.executeMove(filePath, destination)
	case FileMkdir:
		return fn.executeMkdir(filePath, fn.config.Permissions)
	case FileChmod:
		return fn.executeChmod(filePath, fn.config.Permissions)
	default:
		return nil, fmt.Errorf("unsupported file operation: %s", fn.config.Operation)
	}
}

// executeRead reads a file
func (fn *FileNode) executeRead(filePath string) (map[string]interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"content": string(content),
		"size":    len(content),
		"file":    filePath,
	}, nil
}

// executeWrite writes content to a file
func (fn *FileNode) executeWrite(filePath, content string) (map[string]interface{}, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"file":    filePath,
		"size":    len(content),
	}, nil
}

// executeAppend appends content to a file
func (fn *FileNode) executeAppend(filePath, content string) (map[string]interface{}, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for appending: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return nil, fmt.Errorf("failed to append to file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"file":    filePath,
		"size":    len(content),
	}, nil
}

// executeDelete deletes a file
func (fn *FileNode) executeDelete(filePath string) (map[string]interface{}, error) {
	if err := os.Remove(filePath); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"file":    filePath,
	}, nil
}

// executeExists checks if a file exists
func (fn *FileNode) executeExists(filePath string) (map[string]interface{}, error) {
	_, err := os.Stat(filePath)
	exists := err == nil

	return map[string]interface{}{
		"success": true,
		"exists":  exists,
		"file":    filePath,
	}, nil
}

// executeList lists files in a directory
func (fn *FileNode) executeList(dirPath string) (map[string]interface{}, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	files := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		fileInfo, err := entry.Info()
		if err != nil {
			continue // Skip if we can't get info
		}

		fileInfoMap := map[string]interface{}{
			"name":    entry.Name(),
			"is_dir":  entry.IsDir(),
			"size":    fileInfo.Size(),
			"mod_time": fileInfo.ModTime().Unix(),
			"mode":    fileInfo.Mode().String(),
		}

		files = append(files, fileInfoMap)
	}

	return map[string]interface{}{
		"success": true,
		"directory": dirPath,
		"files":   files,
		"count":   len(files),
	}, nil
}

// executeCopy copies a file
func (fn *FileNode) executeCopy(srcPath, dstPath string) (map[string]interface{}, error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"source":  srcPath,
		"dest":    dstPath,
	}, nil
}

// executeMove moves/rename a file
func (fn *FileNode) executeMove(srcPath, dstPath string) (map[string]interface{}, error) {
	if err := os.Rename(srcPath, dstPath); err != nil {
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"source":  srcPath,
		"dest":    dstPath,
	}, nil
}

// executeMkdir creates a directory
func (fn *FileNode) executeMkdir(dirPath string, permissions os.FileMode) (map[string]interface{}, error) {
	if permissions == 0 {
		permissions = 0755
	}

	if fn.config.Recursive {
		if err := os.MkdirAll(dirPath, permissions); err != nil {
			return nil, fmt.Errorf("failed to create directory recursively: %w", err)
		}
	} else {
		if err := os.Mkdir(dirPath, permissions); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return map[string]interface{}{
		"success": true,
		"directory": dirPath,
	}, nil
}

// executeChmod changes file permissions
func (fn *FileNode) executeChmod(filePath string, permissions os.FileMode) (map[string]interface{}, error) {
	if permissions == 0 {
		permissions = 0644
	}

	if err := os.Chmod(filePath, permissions); err != nil {
		return nil, fmt.Errorf("failed to change file permissions: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"file":    filePath,
		"mode":    permissions.String(),
	}, nil
}

// RegisterFileNode registers the file node type with the engine
func RegisterFileNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("file_operation", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var op FileOperationType
		if opVal, exists := config["operation"]; exists {
			if opStr, ok := opVal.(string); ok {
				op = FileOperationType(opStr)
			}
		}

		var filePath string
		if pathVal, exists := config["file_path"]; exists {
			if pathStr, ok := pathVal.(string); ok {
				filePath = pathStr
			}
		}

		var content string
		if contentVal, exists := config["content"]; exists {
			if contentStr, ok := contentVal.(string); ok {
				content = contentStr
			}
		}

		var destination string
		if destVal, exists := config["destination"]; exists {
			if destStr, ok := destVal.(string); ok {
				destination = destStr
			}
		}

		var recursive bool
		if recVal, exists := config["recursive"]; exists {
			if recBool, ok := recVal.(bool); ok {
				recursive = recBool
			}
		}

		var permissions float64
		if permVal, exists := config["permissions"]; exists {
			if permFloat, ok := permVal.(float64); ok {
				permissions = permFloat
			}
		}

		nodeConfig := &FileNodeConfig{
			Operation:   op,
			FilePath:    filePath,
			Content:     content,
			Destination: destination,
			Recursive:   recursive,
			Permissions: os.FileMode(int(permissions)),
		}

		return NewFileNode(nodeConfig), nil
	})
}