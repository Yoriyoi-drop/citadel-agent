package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// AWSS3ManagerConfig mewakili konfigurasi untuk node AWS S3 Manager
type AWSS3ManagerConfig struct {
	AccessKeyID     string                 `json:"access_key_id"`      // ID kunci akses AWS
	SecretAccessKey string                 `json:"secret_access_key"`  // Kunci akses rahasia AWS
	Region          string                 `json:"region"`             // Wilayah AWS
	BucketName      string                 `json:"bucket_name"`        // Nama bucket S3
	Operation       string                 `json:"operation"`          // Operasi (upload, download, list, delete)
	EnableEncryption bool                 `json:"enable_encryption"`   // Apakah mengaktifkan enkripsi
	EncryptionKey   string                 `json:"encryption_key"`     // Kunci enkripsi (jika diaktifkan)
	EnableMultipart bool                  `json:"enable_multipart"`    // Apakah mengaktifkan upload multipart
	PartSize        int64                  `json:"part_size"`          // Ukuran bagian untuk multipart (dalam MB)
	MaxRetries      int                    `json:"max_retries"`        // Jumlah maksimum percobaan ulang
	Timeout         int                    `json:"timeout"`            // Waktu timeout dalam detik
	EnableCaching   bool                   `json:"enable_caching"`     // Apakah mengaktifkan caching
	CacheTTL        int                    `json:"cache_ttl"`          // Waktu cache dalam detik
	EnableProfiling bool                   `json:"enable_profiling"`   // Apakah mengaktifkan profiling
	ReturnRawResults bool                 `json:"return_raw_results"`  // Apakah mengembalikan hasil mentah
	CustomParams    map[string]interface{} `json:"custom_params"`      // Parameter khusus untuk operasi S3
}

// AWSS3ManagerNode mewakili node yang mengelola operasi AWS S3
type AWSS3ManagerNode struct {
	config *AWSS3ManagerConfig
}

// NewAWSS3ManagerNode membuat node AWS S3 Manager baru
func NewAWSS3ManagerNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var s3Config AWSS3ManagerConfig
	err = json.Unmarshal(jsonData, &s3Config)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if s3Config.Region == "" {
		s3Config.Region = "us-east-1"
	}

	if s3Config.Operation == "" {
		s3Config.Operation = "upload"
	}

	if s3Config.MaxRetries == 0 {
		s3Config.MaxRetries = 3
	}

	if s3Config.Timeout == 0 {
		s3Config.Timeout = 300 // default timeout 300 detik
	}

	if s3Config.PartSize == 0 {
		s3Config.PartSize = 5 // default 5MB untuk multipart
	}

	return &AWSS3ManagerNode{
		config: &s3Config,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (a *AWSS3ManagerNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	accessKeyID := a.config.AccessKeyID
	if inputAccessKeyID, ok := input["access_key_id"].(string); ok && inputAccessKeyID != "" {
		accessKeyID = inputAccessKeyID
	}

	secretAccessKey := a.config.SecretAccessKey
	if inputSecretKey, ok := input["secret_access_key"].(string); ok && inputSecretKey != "" {
		secretAccessKey = inputSecretKey
	}

	region := a.config.Region
	if inputRegion, ok := input["region"].(string); ok && inputRegion != "" {
		region = inputRegion
	}

	bucketName := a.config.BucketName
	if inputBucketName, ok := input["bucket_name"].(string); ok && inputBucketName != "" {
		bucketName = inputBucketName
	}

	operation := a.config.Operation
	if inputOperation, ok := input["operation"].(string); ok && inputOperation != "" {
		operation = inputOperation
	}

	enableEncryption := a.config.EnableEncryption
	if inputEnableEncryption, ok := input["enable_encryption"].(bool); ok {
		enableEncryption = inputEnableEncryption
	}

	encryptionKey := a.config.EncryptionKey
	if inputEncryptionKey, ok := input["encryption_key"].(string); ok && inputEncryptionKey != "" {
		encryptionKey = inputEncryptionKey
	}

	enableMultipart := a.config.EnableMultipart
	if inputEnableMultipart, ok := input["enable_multipart"].(bool); ok {
		enableMultipart = inputEnableMultipart
	}

	partSize := a.config.PartSize
	if inputPartSize, ok := input["part_size"].(float64); ok {
		partSize = int64(inputPartSize)
	}

	maxRetries := a.config.MaxRetries
	if inputMaxRetries, ok := input["max_retries"].(float64); ok {
		maxRetries = int(inputMaxRetries)
	}

	timeout := a.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	enableCaching := a.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := a.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := a.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	returnRawResults := a.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := a.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	// Validasi input yang diperlukan
	if accessKeyID == "" || secretAccessKey == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "access_key_id dan secret_access_key diperlukan untuk operasi S3",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if bucketName == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "bucket_name diperlukan untuk operasi S3",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Validasi operasi
	validOperations := map[string]bool{
		"upload":   true,
		"download": true,
		"list":     true,
		"delete":   true,
		"copy":     true,
		"move":     true,
	}
	
	if !validOperations[operation] {
		return map[string]interface{}{
			"success":   false,
			"error":     fmt.Sprintf("operasi '%s' tidak didukung. Operasi yang didukung: upload, download, list, delete, copy, move", operation),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks operasi dengan timeout
	s3Ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Lakukan operasi S3
	s3Result, err := a.performS3Operation(s3Ctx, operation, input)
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
		"region":               region,
		"bucket_name":          bucketName,
		"enable_encryption":    enableEncryption,
		"enable_multipart":     enableMultipart,
		"part_size":            partSize,
		"max_retries":          maxRetries,
		"s3_result":            s3Result,
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"config":               a.config,
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

// performS3Operation melakukan operasi S3 yang spesifik
func (a *AWSS3ManagerNode) performS3Operation(ctx context.Context, operation string, input map[string]interface{}) (map[string]interface{}, error) {
	// Simulasikan waktu pemrosesan
	time.Sleep(100 * time.Millisecond)

	result := make(map[string]interface{})

	switch operation {
	case "upload":
		result = a.performS3Upload(input)
	case "download":
		result = a.performS3Download(input)
	case "list":
		result = a.performS3List(input)
	case "delete":
		result = a.performS3Delete(input)
	case "copy":
		result = a.performS3Copy(input)
	case "move":
		result = a.performS3Move(input)
	default:
		result = map[string]interface{}{
			"operation": operation,
			"status":    "not_implemented",
			"message":   fmt.Sprintf("Operasi %s belum diimplementasikan dalam simulasi ini", operation),
		}
	}

	result["operation_completed"] = true
	result["timestamp"] = time.Now().Unix()
	result["processing_time"] = time.Since(time.Now().Add(-100 * time.Millisecond)).Seconds()

	return result, nil
}

// performS3Upload melakukan operasi upload ke S3
func (a *AWSS3ManagerNode) performS3Upload(input map[string]interface{}) map[string]interface{} {
	filePath := ""
	if path, exists := input["file_path"]; exists {
		if pathStr, ok := path.(string); ok {
			filePath = pathStr
		}
	}

	objectKey := ""
	if key, exists := input["object_key"]; exists {
		if keyStr, ok := key.(string); ok {
			objectKey = keyStr
		}
	}

	if objectKey == "" {
		// Gunakan nama file dari path atau buat otomatis
		objectKey = fmt.Sprintf("uploaded_file_%d", time.Now().Unix())
	}

	// Simulasikan upload
	result := map[string]interface{}{
		"operation":    "upload",
		"object_key":   objectKey,
		"file_path":    filePath,
		"uploaded":     true,
		"bucket":       a.config.BucketName,
		"region":       a.config.Region,
		"size":         "2.5 MB", // Simulasikan ukuran file
		"encryption":   a.config.EnableEncryption,
		"multipart":    a.config.EnableMultipart,
		"etag":         "1a2b3c4d5e6f7g8h9i0j", // Simulasikan ETag
		"version_id":   fmt.Sprintf("version_%d", time.Now().Unix()),
		"upload_speed": "5.2 MB/s", // Simulasikan kecepatan upload
	}

	return result
}

// performS3Download melakukan operasi download dari S3
func (a *AWSS3ManagerNode) performS3Download(input map[string]interface{}) map[string]interface{} {
	objectKey := ""
	if key, exists := input["object_key"]; exists {
		if keyStr, ok := key.(string); ok {
			objectKey = keyStr
		}
	}

	if objectKey == "" {
		return map[string]interface{}{
			"operation": "download",
			"error":     "object_key diperlukan untuk operasi download",
		}
	}

	// Simulasikan download
	result := map[string]interface{}{
		"operation":    "download",
		"object_key":   objectKey,
		"downloaded":   true,
		"bucket":       a.config.BucketName,
		"region":       a.config.Region,
		"size":         "2.5 MB", // Simulasikan ukuran file
		"download_path": "/tmp/downloaded_file", // Simulasikan path download
		"checksum":     "a1b2c3d4e5f6g7h8i9j0",
		"download_speed": "8.1 MB/s", // Simulasikan kecepatan download
	}

	return result
}

// performS3List melakukan operasi list objek di S3
func (a *AWSS3ManagerNode) performS3List(input map[string]interface{}) map[string]interface{} {
	prefix := ""
	if pref, exists := input["prefix"]; exists {
		if prefStr, ok := pref.(string); ok {
			prefix = prefStr
		}
	}

	maxKeys := 10
	if max, exists := input["max_keys"]; exists {
		if maxFloat, ok := max.(float64); ok {
			maxKeys = int(maxFloat)
		}
	}

	// Simulasikan list objek
	objects := []map[string]interface{}{
		{
			"key":          "folder1/document1.pdf",
			"size":         1024000, // 1MB dalam bytes
			"last_modified": time.Now().Unix() - 86400, // 1 hari yang lalu
			"etag":         "1a2b3c4d5e6f",
		},
		{
			"key":          "folder2/image1.jpg",
			"size":         2048000, // 2MB dalam bytes
			"last_modified": time.Now().Unix() - 172800, // 2 hari yang lalu
			"etag":         "2b3c4d5e6f7g",
		},
		{
			"key":          "folder3/data.json",
			"size":         512000, // 0.5MB dalam bytes
			"last_modified": time.Now().Unix() - 259200, // 3 hari yang lalu
			"etag":         "3c4d5e6f7g8h",
		},
	}

	// Batasi jumlah objek sesuai maxKeys
	if len(objects) > maxKeys {
		objects = objects[:maxKeys]
	}

	result := map[string]interface{}{
		"operation": "list",
		"bucket":    a.config.BucketName,
		"prefix":    prefix,
		"objects":   objects,
		"count":     len(objects),
		"truncated": len(objects) == maxKeys, // Menunjukkan apakah hasil dipotong
		"region":    a.config.Region,
	}

	return result
}

// performS3Delete melakukan operasi delete dari S3
func (a *AWSS3ManagerNode) performS3Delete(input map[string]interface{}) map[string]interface{} {
	objectKey := ""
	if key, exists := input["object_key"]; exists {
		if keyStr, ok := key.(string); ok {
			objectKey = keyStr
		}
	}

	if objectKey == "" {
		return map[string]interface{}{
			"operation": "delete",
			"error":     "object_key diperlukan untuk operasi delete",
		}
	}

	// Simulasikan delete
	result := map[string]interface{}{
		"operation":  "delete",
		"object_key": objectKey,
		"deleted":    true,
		"bucket":     a.config.BucketName,
		"region":     a.config.Region,
		"timestamp":  time.Now().Unix(),
	}

	return result
}

// performS3Copy melakukan operasi copy di S3
func (a *AWSS3ManagerNode) performS3Copy(input map[string]interface{}) map[string]interface{} {
	sourceKey := ""
	if key, exists := input["source_key"]; exists {
		if keyStr, ok := key.(string); ok {
			sourceKey = keyStr
		}
	}

	destinationKey := ""
	if key, exists := input["destination_key"]; exists {
		if keyStr, ok := key.(string); ok {
			destinationKey = keyStr
		}
	}

	if sourceKey == "" || destinationKey == "" {
		return map[string]interface{}{
			"operation": "copy",
			"error":     "source_key dan destination_key diperlukan untuk operasi copy",
		}
	}

	// Simulasikan copy
	result := map[string]interface{}{
		"operation":      "copy",
		"source_key":     sourceKey,
		"destination_key": destinationKey,
		"copied":         true,
		"source_bucket":  a.config.BucketName,
		"dest_bucket":    a.config.BucketName, // Dalam contoh ini sama
		"region":         a.config.Region,
		"size":           "2.5 MB", // Simulasikan ukuran file
	}

	return result
}

// performS3Move melakukan operasi move di S3
func (a *AWSS3ManagerNode) performS3Move(input map[string]interface{}) map[string]interface{} {
	sourceKey := ""
	if key, exists := input["source_key"]; exists {
		if keyStr, ok := key.(string); ok {
			sourceKey = keyStr
		}
	}

	destinationKey := ""
	if key, exists := input["destination_key"]; exists {
		if keyStr, ok := key.(string); ok {
			destinationKey = keyStr
		}
	}

	if sourceKey == "" || destinationKey == "" {
		return map[string]interface{}{
			"operation": "move",
			"error":     "source_key dan destination_key diperlukan untuk operasi move",
		}
	}

	// Simulasikan move (copy + delete)
	result := map[string]interface{}{
		"operation":       "move",
		"source_key":      sourceKey,
		"destination_key": destinationKey,
		"moved":           true,
		"source_bucket":   a.config.BucketName,
		"dest_bucket":     a.config.BucketName, // Dalam contoh ini sama
		"region":          a.config.Region,
		"size":            "2.5 MB", // Simulasikan ukuran file
		"delete_original": true,
	}

	return result
}

// GetType mengembalikan jenis node
func (a *AWSS3ManagerNode) GetType() string {
	return "aws_s3_manager"
}

// GetID mengembalikan ID unik untuk instance node
func (a *AWSS3ManagerNode) GetID() string {
	return "aws_s3_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterAWSS3ManagerNode mendaftarkan node AWS S3 Manager dengan engine
func RegisterAWSS3ManagerNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("aws_s3_manager", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewAWSS3ManagerNode(config)
	})
}