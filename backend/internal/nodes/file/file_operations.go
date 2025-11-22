package file

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// FileOperationConfig mewakili konfigurasi untuk node operasi file
type FileOperationConfig struct {
	Operation       string                   `json:"operation"`         // Operasi (upload, download, list, delete, move, copy)
	SourcePath      string                   `json:"source_path"`       // Path sumber
	DestinationPath string                   `json:"destination_path"`  // Path tujuan
	StorageProvider string                   `json:"storage_provider"`  // Penyedia penyimpanan (local, s3, gcs, azure)
	FilePath        string                   `json:"file_path"`         // Path file untuk operasi single file
	FileName        string                   `json:"file_name"`         // Nama file
	FilePattern     string                   `json:"file_pattern"`      // Pattern file untuk operasi batch
	MaxFileSize     int64                    `json:"max_file_size"`     // Ukuran maksimum file dalam bytes
	EnableEncryption bool                    `json:"enable_encryption"` // Apakah mengaktifkan enkripsi
	EncryptionKey   string                   `json:"encryption_key"`    // Kunci enkripsi
	EnableCompression bool                   `json:"enable_compression"` // Apakah mengaktifkan kompresi
	CompressionType string                   `json:"compression_type"`  // Jenis kompresi (gzip, zip, etc.)
	Timeout         int                      `json:"timeout"`           // Waktu timeout dalam detik
	MaxRetries      int                      `json:"max_retries"`       // Jumlah maksimum percobaan ulang
	EnableCaching   bool                     `json:"enable_caching"`    // Apakah mengaktifkan caching
	CacheTTL        int                      `json:"cache_ttl"`         // Waktu cache dalam detik
	EnableProfiling bool                     `json:"enable_profiling"`  // Apakah mengaktifkan profiling
	ReturnRawResults bool                    `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams    map[string]interface{}   `json:"custom_params"`     // Parameter khusus untuk operasi file
	Preprocessing   PreprocessingConfig      `json:"preprocessing"`     // Konfigurasi pra-pemrosesan
	Postprocessing  PostprocessingConfig     `json:"postprocessing"`    // Konfigurasi pasca-pemrosesan
	AccessControls  []AccessControl          `json:"access_controls"`   // Pengaturan akses
	Metadata        map[string]string        `json:"metadata"`          // Metadata file
	Recursive       bool                     `json:"recursive"`         // Apakah operasi bersifat rekursif
}

// AccessControl mewakili pengaturan kontrol akses
type AccessControl struct {
	Type        string `json:"type"`        // Jenis kontrol (user, group, role)
	Principal   string `json:"principal"`   // Entitas (username, groupname, etc.)
	Permissions string `json:"permissions"` // Izin (read, write, delete, etc.)
}

// PreprocessingConfig mewakili konfigurasi pra-pemrosesan
type PreprocessingConfig struct {
	NormalizeInput bool                   `json:"normalize_input"` // Apakah menormalkan input
	ValidateInput  bool                   `json:"validate_input"`  // Apakah memvalidasi input
	TransformInput bool                   `json:"transform_input"` // Apakah mentransformasi input
	TransformRules map[string]interface{} `json:"transform_rules"` // Aturan transformasi
}

// PostprocessingConfig mewakili konfigurasi pasca-pemrosesan
type PostprocessingConfig struct {
	FilterOutput  bool              `json:"filter_output"`   // Apakah memfilter output
	OutputMapping map[string]string `json:"output_mapping"`  // Pemetaan field output
	TransformOutput bool            `json:"transform_output"` // Apakah mentransformasi output
}

// FileOperationNode mewakili node yang melakukan operasi file
type FileOperationNode struct {
	config *FileOperationConfig
}

// NewFileNode membuat node operasi file baru
func NewFileNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var fileConfig FileOperationConfig
	err = json.Unmarshal(jsonData, &fileConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if fileConfig.Operation == "" {
		fileConfig.Operation = "upload"
	}

	if fileConfig.StorageProvider == "" {
		fileConfig.StorageProvider = "local"
	}

	if fileConfig.MaxRetries == 0 {
		fileConfig.MaxRetries = 3
	}

	if fileConfig.Timeout == 0 {
		fileConfig.Timeout = 300 // default timeout 300 detik
	}

	if fileConfig.MaxFileSize == 0 {
		fileConfig.MaxFileSize = 100 * 1024 * 1024 // default 100MB
	}

	if fileConfig.AccessControls == nil {
		fileConfig.AccessControls = []AccessControl{}
	}

	if fileConfig.Metadata == nil {
		fileConfig.Metadata = make(map[string]string)
	}

	return &FileOperationNode{
		config: &fileConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (f *FileOperationNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	operation := f.config.Operation
	if inputOperation, ok := input["operation"].(string); ok && inputOperation != "" {
		operation = inputOperation
	}

	sourcePath := f.config.SourcePath
	if inputSourcePath, ok := input["source_path"].(string); ok && inputSourcePath != "" {
		sourcePath = inputSourcePath
	}

	destinationPath := f.config.DestinationPath
	if inputDestinationPath, ok := input["destination_path"].(string); ok && inputDestinationPath != "" {
		destinationPath = inputDestinationPath
	}

	storageProvider := f.config.StorageProvider
	if inputStorageProvider, ok := input["storage_provider"].(string); ok && inputStorageProvider != "" {
		storageProvider = inputStorageProvider
	}

	filePath := f.config.FilePath
	if inputFilePath, ok := input["file_path"].(string); ok && inputFilePath != "" {
		filePath = inputFilePath
	}

	fileName := f.config.FileName
	if inputFileName, ok := input["file_name"].(string); ok && inputFileName != "" {
		fileName = inputFileName
	}

	filePattern := f.config.FilePattern
	if inputFilePattern, ok := input["file_pattern"].(string); ok && inputFilePattern != "" {
		filePattern = inputFilePattern
	}

	maxFileSize := f.config.MaxFileSize
	if inputMaxFileSize, ok := input["max_file_size"].(float64); ok {
		maxFileSize = int64(inputMaxFileSize)
	}

	enableEncryption := f.config.EnableEncryption
	if inputEnableEncryption, ok := input["enable_encryption"].(bool); ok {
		enableEncryption = inputEnableEncryption
	}

	encryptionKey := f.config.EncryptionKey
	if inputEncryptionKey, ok := input["encryption_key"].(string); ok && inputEncryptionKey != "" {
		encryptionKey = inputEncryptionKey
	}

	enableCompression := f.config.EnableCompression
	if inputEnableCompression, ok := input["enable_compression"].(bool); ok {
		enableCompression = inputEnableCompression
	}

	compressionType := f.config.CompressionType
	if inputCompressionType, ok := input["compression_type"].(string); ok && inputCompressionType != "" {
		compressionType = inputCompressionType
	}

	timeout := f.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	maxRetries := f.config.MaxRetries
	if inputMaxRetries, ok := input["max_retries"].(float64); ok {
		maxRetries = int(inputMaxRetries)
	}

	enableCaching := f.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := f.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := f.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	returnRawResults := f.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := f.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	accessControls := f.config.AccessControls
	if inputAccessControls, ok := input["access_controls"].([]interface{}); ok {
		accessControls = make([]AccessControl, len(inputAccessControls))
		for i, control := range inputAccessControls {
			if controlMap, ok := control.(map[string]interface{}); ok {
				var ac AccessControl
				if typ, exists := controlMap["type"]; exists {
					if typStr, ok := typ.(string); ok {
						ac.Type = typStr
					}
				}
				if principal, exists := controlMap["principal"]; exists {
					if principalStr, ok := principal.(string); ok {
						ac.Principal = principalStr
					}
				}
				if perms, exists := controlMap["permissions"]; exists {
					if permsStr, ok := perms.(string); ok {
						ac.Permissions = permsStr
					}
				}
				accessControls[i] = ac
			}
		}
	}

	metadata := f.config.Metadata
	if inputMetadata, ok := input["metadata"].(map[string]interface{}); ok {
		metadata = make(map[string]string)
		for k, v := range inputMetadata {
			if vStr, ok := v.(string); ok {
				metadata[k] = vStr
			}
		}
	}

	recursive := f.config.Recursive
	if inputRecursive, ok := input["recursive"].(bool); ok {
		recursive = inputRecursive
	}

	// Validasi input berdasarkan operasi
	switch operation {
	case "upload":
		if filePath == "" && sourcePath == "" {
			return map[string]interface{}{
				"success":   false,
				"error":     "file_path atau source_path diperlukan untuk operasi upload",
				"timestamp": time.Now().Unix(),
			}, nil
		}
		if destinationPath == "" {
			return map[string]interface{}{
				"success":   false,
				"error":     "destination_path diperlukan untuk operasi upload",
				"timestamp": time.Now().Unix(),
			}, nil
		}
	case "download":
		if filePath == "" && sourcePath == "" {
			return map[string]interface{}{
				"success":   false,
				"error":     "file_path atau source_path diperlukan untuk operasi download",
				"timestamp": time.Now().Unix(),
			}, nil
		}
		if destinationPath == "" {
			return map[string]interface{}{
				"success":   false,
				"error":     "destination_path diperlukan untuk operasi download",
				"timestamp": time.Now().Unix(),
			}, nil
		}
	case "delete":
		if filePath == "" && sourcePath == "" {
			return map[string]interface{}{
				"success":   false,
				"error":     "file_path atau source_path diperlukan untuk operasi delete",
				"timestamp": time.Now().Unix(),
			}, nil
		}
	}

	// Buat konteks operasi dengan timeout
	fileCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Lakukan operasi file
	fileResult, err := f.performFileOperation(fileCtx, operation, input)
	if err != nil {
		return map[string]interface{}{
			"success":   false,
			"error":     err.Error(),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Siapkan hasil akhir
	finalResult := map[string]interface{}{
		"success":              true,
		"operation":            operation,
		"storage_provider":     storageProvider,
		"source_path":          sourcePath,
		"destination_path":     destinationPath,
		"file_path":            filePath,
		"file_name":            fileName,
		"file_pattern":         filePattern,
		"max_file_size":        maxFileSize,
		"enable_encryption":    enableEncryption,
		"enable_compression":   enableCompression,
		"compression_type":     compressionType,
		"max_retries":          maxRetries,
		"recursive_operation":  recursive,
		"file_result":          fileResult,
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"access_controls":      accessControls,
		"metadata":             metadata,
		"config":               f.config,
	}

	// Tambahkan metrik kinerja jika profiling diaktifkan
	if enableProfiling {
		finalResult["performance_metrics"] = map[string]interface{}{
			"start_time": time.Now().Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
		}
	}

	return finalResult, nil
}

// performFileOperation melakukan operasi file yang ditentukan
func (f *FileOperationNode) performFileOperation(ctx context.Context, operation string, input map[string]interface{}) (map[string]interface{}, error) {
	// Simulasikan waktu pemrosesan
	time.Sleep(150 * time.Millisecond)

	var result map[string]interface{}

	switch operation {
	case "upload":
		result = f.performFileUpload(input)
	case "download":
		result = f.performFileDownload(input)
	case "list":
		result = f.performFileList(input)
	case "delete":
		result = f.performFileDelete(input)
	case "move":
		result = f.performFileMove(input)
	case "copy":
		result = f.performFileCopy(input)
	case "metadata":
		result = f.performFileMetadata(input)
	default:
		result = f.performDefaultOperation(input)
	}

	result["operation_completed"] = true
	result["operation_type"] = operation
	result["timestamp"] = time.Now().Unix()
	result["processing_time"] = time.Since(time.Now().Add(-150 * time.Millisecond)).Seconds()
	result["storage_provider"] = f.config.StorageProvider

	return result, nil
}

// performFileUpload melakukan operasi upload file
func (f *FileOperationNode) performFileUpload(input map[string]interface{}) map[string]interface{} {
	sourcePath := f.getFilePath(input)
	destinationPath := f.config.DestinationPath
	if dest, exists := input["destination_path"]; exists {
		if destStr, ok := dest.(string); ok {
			destinationPath = destStr
		}
	}

	result := map[string]interface{}{
		"operation":       "upload",
		"source_path":     sourcePath,
		"destination_path": destinationPath,
		"uploaded":        true,
		"file_size":       "2.5 MB", // Simulasikan ukuran file
		"upload_speed":    "5.2 MB/s", // Simulasikan kecepatan upload
		"encryption":      f.config.EnableEncryption,
		"compression":     f.config.EnableCompression,
		"checksum":        "a1b2c3d4e5f6g7h8i9j0", // Simulasikan checksum
		"version_id":      fmt.Sprintf("version_%d", time.Now().Unix()),
		"storage_class":   "standard",
		"metadata":        f.config.Metadata,
	}

	return result
}

// performFileDownload melakukan operasi download file
func (f *FileOperationNode) performFileDownload(input map[string]interface{}) map[string]interface{} {
	sourcePath := f.getFilePath(input)
	destinationPath := f.config.DestinationPath
	if dest, exists := input["destination_path"]; exists {
		if destStr, ok := dest.(string); ok {
			destinationPath = destStr
		}
	}

	result := map[string]interface{}{
		"operation":       "download",
		"source_path":     sourcePath,
		"destination_path": destinationPath,
		"downloaded":      true,
		"file_size":       "2.5 MB", // Simulasikan ukuran file
		"download_speed":  "8.1 MB/s", // Simulasikan kecepatan download
		"checksum":        "a1b2c3d4e5f6g7h8i9j0", // Simulasikan checksum
		"local_path":      "/tmp/downloaded_file",
		"permissions":     "rw-rw-rw-",
		"metadata":        f.config.Metadata,
	}

	return result
}

// performFileList melakukan operasi list file
func (f *FileOperationNode) performFileList(input map[string]interface{}) map[string]interface{} {
	directoryPath := f.config.SourcePath
	if source, exists := input["source_path"]; exists {
		if sourceStr, ok := source.(string); ok {
			directoryPath = sourceStr
		}
	}

	// Simulasikan daftar file
	files := []map[string]interface{}{
		{
			"name":         "document1.pdf",
			"size":         1024000, // 1MB dalam bytes
			"type":         "file",
			"modified":     time.Now().Unix() - 86400, // 1 hari yang lalu
			"permissions":  "rw-rw-rw-",
			"checksum":     "1a2b3c4d5e6f",
		},
		{
			"name":         "image1.jpg",
			"size":         2048000, // 2MB dalam bytes
			"type":         "file",
			"modified":     time.Now().Unix() - 172800, // 2 hari yang lalu
			"permissions":  "rw-rw-rw-",
			"checksum":     "2b3c4d5e6f7g",
		},
		{
			"name":         "data_folder",
			"size":         4096,
			"type":         "directory",
			"modified":     time.Now().Unix() - 259200, // 3 hari yang lalu
			"permissions":  "rwxrwxrwx",
			"checksum":     "",
		},
	}

	result := map[string]interface{}{
		"operation":    "list",
		"directory":    directoryPath,
		"files":        files,
		"file_count":   len(files),
		"total_size":   3076096, // Jumlah total ukuran file
		"recursive":    f.config.Recursive,
		"pattern":      f.config.FilePattern,
		"storage_type": f.config.StorageProvider,
	}

	return result
}

// performFileDelete melakukan operasi delete file
func (f *FileOperationNode) performFileDelete(input map[string]interface{}) map[string]interface{} {
	filePath := f.getFilePath(input)

	result := map[string]interface{}{
		"operation":   "delete",
		"file_path":   filePath,
		"deleted":     true,
		"timestamp":   time.Now().Unix(),
		"recursive":   f.config.Recursive,
		"files_deleted": 1, // Simulasikan jumlah file yang dihapus
		"storage_type": f.config.StorageProvider,
	}

	return result
}

// performFileMove melakukan operasi move file
func (f *FileOperationNode) performFileMove(input map[string]interface{}) map[string]interface{} {
	sourcePath := f.getFilePath(input)
	destinationPath := f.config.DestinationPath
	if dest, exists := input["destination_path"]; exists {
		if destStr, ok := dest.(string); ok {
			destinationPath = destStr
		}
	}

	result := map[string]interface{}{
		"operation":       "move",
		"source_path":     sourcePath,
		"destination_path": destinationPath,
		"moved":           true,
		"timestamp":       time.Now().Unix(),
		"file_size":       "2.5 MB", // Simulasikan ukuran file
		"storage_type":    f.config.StorageProvider,
		"original_path":   sourcePath,
		"new_path":        destinationPath,
	}

	return result
}

// performFileCopy melakukan operasi copy file
func (f *FileOperationNode) performFileCopy(input map[string]interface{}) map[string]interface{} {
	sourcePath := f.getFilePath(input)
	destinationPath := f.config.DestinationPath
	if dest, exists := input["destination_path"]; exists {
		if destStr, ok := dest.(string); ok {
			destinationPath = destStr
		}
	}

	result := map[string]interface{}{
		"operation":       "copy",
		"source_path":     sourcePath,
		"destination_path": destinationPath,
		"copied":          true,
		"timestamp":       time.Now().Unix(),
		"file_size":       "2.5 MB", // Simulasikan ukuran file
		"storage_type":    f.config.StorageProvider,
		"original_path":   sourcePath,
		"copy_path":       destinationPath,
	}

	return result
}

// performFileMetadata melakukan operasi metadata file
func (f *FileOperationNode) performFileMetadata(input map[string]interface{}) map[string]interface{} {
	filePath := f.getFilePath(input)

	result := map[string]interface{}{
		"operation":    "metadata",
		"file_path":    filePath,
		"exists":       true,
		"size":         2621440, // 2.5 MB dalam bytes
		"created":      time.Now().Unix() - 86400, // 1 hari yang lalu
		"modified":     time.Now().Unix() - 43200, // 12 jam yang lalu
		"accessed":     time.Now().Unix() - 1800, // 30 menit yang lalu
		"type":         "file",
		"extension":    "pdf",
		"permissions":  "rw-rw-rw-",
		"owner":        "user1",
		"group":        "users",
		"checksum":     "a1b2c3d4e5f6g7h8i9j0",
		"storage_type": f.config.StorageProvider,
		"metadata":     f.config.Metadata,
	}

	return result
}

// performDefaultOperation melakukan operasi default
func (f *FileOperationNode) performDefaultOperation(input map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"operation":     "default",
		"description":   "Operasi file default ketika tipe operasi tidak dikenali",
		"input_passed":  true,
		"storage_type":  f.config.StorageProvider,
		"provider_specific": map[string]interface{}{
			"provider": f.config.StorageProvider,
			"config":   f.config,
		},
	}

	return result
}

// getFilePath mendapatkan path file dari input atau konfigurasi
func (f *FileOperationNode) getFilePath(input map[string]interface{}) string {
	if filePath, exists := input["file_path"]; exists {
		if pathStr, ok := filePath.(string); ok {
			return pathStr
		}
	}
	
	if sourcePath, exists := input["source_path"]; exists {
		if pathStr, ok := sourcePath.(string); ok {
			return pathStr
		}
	}
	
	if f.config.FilePath != "" {
		return f.config.FilePath
	}
	
	if f.config.SourcePath != "" {
		return f.config.SourcePath
	}
	
	return "/default/file/path"
}

// GetType mengembalikan jenis node
func (f *FileOperationNode) GetType() string {
	return "file_operation"
}

// GetID mengembalikan ID unik untuk instance node
func (f *FileOperationNode) GetID() string {
	return "file_op_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterFileNode mendaftarkan node operasi file dengan engine
func RegisterFileNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("file_operation", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewFileNode(config)
	})
}