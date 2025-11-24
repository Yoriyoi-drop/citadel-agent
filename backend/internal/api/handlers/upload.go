package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	// MaxMemory for multipart form parsing (10MB)
	MaxMemory = 10 << 20
	// ChunkSize for streaming uploads (1MB)
	ChunkSize = 1 << 20
)

// FileUploadHandler handles file uploads with streaming
type FileUploadHandler struct {
	uploadDir       string
	maxFileSize     int64
	allowedTypes    map[string]bool
	enableStreaming bool
}

// NewFileUploadHandler creates a new file upload handler
func NewFileUploadHandler(uploadDir string, maxFileSize int64, allowedTypes []string) *FileUploadHandler {
	allowed := make(map[string]bool)
	for _, t := range allowedTypes {
		allowed[t] = true
	}

	return &FileUploadHandler{
		uploadDir:       uploadDir,
		maxFileSize:     maxFileSize,
		allowedTypes:    allowed,
		enableStreaming: true,
	}
}

// UploadFile handles file upload with streaming for large files
func (h *FileUploadHandler) UploadFile(c *fiber.Ctx) error {
	// Parse multipart form with limited memory
	if _, err := c.MultipartForm(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse multipart form",
		})
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided",
		})
	}

	// Validate file size
	if file.Size > h.maxFileSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error": fmt.Sprintf("File too large. Maximum size: %d bytes", h.maxFileSize),
		})
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	ext = strings.TrimPrefix(ext, ".")
	if len(h.allowedTypes) > 0 && !h.allowedTypes[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("File type not allowed: %s", ext),
		})
	}

	// Use streaming for large files
	if h.enableStreaming && file.Size > ChunkSize {
		return h.streamUpload(c, file)
	}

	// Use standard upload for small files
	return h.standardUpload(c, file)
}

// streamUpload handles large file uploads with streaming
func (h *FileUploadHandler) streamUpload(c *fiber.Ctx, fileHeader *multipart.FileHeader) error {
	// Open source file
	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open uploaded file",
		})
	}
	defer src.Close()

	// Create destination file
	destPath := filepath.Join(h.uploadDir, fileHeader.Filename)
	dst, err := os.Create(destPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create destination file",
		})
	}
	defer dst.Close()

	// Stream file in chunks
	buffer := make([]byte, ChunkSize)
	totalBytes := int64(0)

	for {
		n, err := src.Read(buffer)
		if err != nil && err != io.EOF {
			os.Remove(destPath) // Clean up on error
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read file chunk",
			})
		}

		if n == 0 {
			break
		}

		written, err := dst.Write(buffer[:n])
		if err != nil {
			os.Remove(destPath) // Clean up on error
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to write file chunk",
			})
		}

		totalBytes += int64(written)

		// Check if we exceeded max size during streaming
		if totalBytes > h.maxFileSize {
			os.Remove(destPath) // Clean up
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": "File size exceeded during upload",
			})
		}
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"filename": fileHeader.Filename,
		"size":     totalBytes,
		"path":     destPath,
		"method":   "streaming",
	})
}

// standardUpload handles small file uploads
func (h *FileUploadHandler) standardUpload(c *fiber.Ctx, fileHeader *multipart.FileHeader) error {
	destPath := filepath.Join(h.uploadDir, fileHeader.Filename)

	// Save file
	if err := c.SaveFile(fileHeader, destPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"filename": fileHeader.Filename,
		"size":     fileHeader.Size,
		"path":     destPath,
		"method":   "standard",
	})
}

// UploadMultipleFiles handles multiple file uploads
func (h *FileUploadHandler) UploadMultipleFiles(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse multipart form",
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No files provided",
		})
	}

	results := make([]map[string]interface{}, 0, len(files))
	errors := make([]string, 0)

	for _, file := range files {
		// Validate each file
		if file.Size > h.maxFileSize {
			errors = append(errors, fmt.Sprintf("%s: file too large", file.Filename))
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		ext = strings.TrimPrefix(ext, ".")
		if len(h.allowedTypes) > 0 && !h.allowedTypes[ext] {
			errors = append(errors, fmt.Sprintf("%s: file type not allowed", file.Filename))
			continue
		}

		// Save file
		destPath := filepath.Join(h.uploadDir, file.Filename)
		if err := c.SaveFile(file, destPath); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to save", file.Filename))
			continue
		}

		results = append(results, map[string]interface{}{
			"filename": file.Filename,
			"size":     file.Size,
			"path":     destPath,
		})
	}

	return c.JSON(fiber.Map{
		"success": len(errors) == 0,
		"files":   results,
		"errors":  errors,
		"total":   len(files),
		"saved":   len(results),
	})
}
