package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// RESTAPIClientConfig mewakili konfigurasi untuk node REST API Client
type RESTAPIClientConfig struct {
	BaseURL         string                 `json:"base_url"`          // URL dasar API
	APIKey          string                 `json:"api_key"`           // Kunci API
	AuthType        string                 `json:"auth_type"`         // Jenis autentikasi (bearer, api_key, basic, oauth)
	Timeout         int                    `json:"timeout"`           // Waktu timeout dalam detik
	MaxRetries      int                    `json:"max_retries"`       // Jumlah maksimum percobaan ulang
	EnableRetry     bool                   `json:"enable_retry"`      // Apakah mengaktifkan retry
	RetryDelay      int                    `json:"retry_delay"`       // Delay antar retry dalam detik
	EnableCaching   bool                   `json:"enable_caching"`    // Apakah mengaktifkan caching
	CacheTTL        int                    `json:"cache_ttl"`         // Waktu cache dalam detik
	RequestHeaders  map[string]string      `json:"request_headers"`   // Header permintaan tambahan
	ResponseHeaders []string               `json:"response_headers"`  // Header respons yang ingin diambil
	EnableProfiling bool                   `json:"enable_profiling"`  // Apakah mengaktifkan profiling
	ReturnRawResults bool                 `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams    map[string]interface{} `json:"custom_params"`     // Parameter khusus untuk permintaan
	Preprocessing   PreprocessingConfig    `json:"preprocessing"`     // Konfigurasi pra-pemrosesan
	Postprocessing  PostprocessingConfig   `json:"postprocessing"`    // Konfigurasi pasca-pemrosesan
	SSLVerification bool                   `json:"ssl_verification"`  // Apakah memverifikasi SSL
	EnableRateLimit bool                   `json:"enable_rate_limit"` // Apakah mengaktifkan pembatasan laju
	RateLimit       int                    `json:"rate_limit"`        // Batas permintaan per detik
}

// PreprocessingConfig mewakili konfigurasi pra-pemrosesan
type PreprocessingConfig struct {
	NormalizeParams bool                   `json:"normalize_params"`  // Apakah menormalkan parameter
	ValidateParams  bool                   `json:"validate_params"`   // Apakah memvalidasi parameter
	TransformParams bool                   `json:"transform_params"`  // Apakah mentransformasi parameter
	TransformRules  map[string]interface{} `json:"transform_rules"`   // Aturan transformasi
}

// PostprocessingConfig mewakili konfigurasi pasca-pemrosesan
type PostprocessingConfig struct {
	FilterResponse  bool                   `json:"filter_response"`   // Apakah memfilter respons
	ResponseMapping map[string]string      `json:"response_mapping"`  // Pemetaan field respons
	ExtractFields   []string               `json:"extract_fields"`    // Field untuk diekstraksi
	DataTransform   map[string]interface{} `json:"data_transform"`    // Transformasi data
}

// RESTAPIClientNode mewakili node yang membuat permintaan ke REST API
type RESTAPIClientNode struct {
	config *RESTAPIClientConfig
}

// NewRESTAPIClientNode membuat node REST API Client baru
func NewRESTAPIClientNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var apiConfig RESTAPIClientConfig
	err = json.Unmarshal(jsonData, &apiConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if apiConfig.AuthType == "" {
		apiConfig.AuthType = "bearer"
	}

	if apiConfig.MaxRetries == 0 {
		apiConfig.MaxRetries = 3
	}

	if apiConfig.Timeout == 0 {
		apiConfig.Timeout = 60 // default timeout 60 detik
	}

	if apiConfig.RetryDelay == 0 {
		apiConfig.RetryDelay = 1 // default delay 1 detik
	}

	if apiConfig.RequestHeaders == nil {
		apiConfig.RequestHeaders = make(map[string]string)
	}

	if apiConfig.Preprocessing.TransformRules == nil {
		apiConfig.Preprocessing.TransformRules = make(map[string]interface{})
	}

	if apiConfig.Postprocessing.ResponseMapping == nil {
		apiConfig.Postprocessing.ResponseMapping = make(map[string]string)
	}

	return &RESTAPIClientNode{
		config: &apiConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (r *RESTAPIClientNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	baseURL := r.config.BaseURL
	if inputBaseURL, ok := input["base_url"].(string); ok && inputBaseURL != "" {
		baseURL = inputBaseURL
	}

	apiKey := r.config.APIKey
	if inputAPIKey, ok := input["api_key"].(string); ok && inputAPIKey != "" {
		apiKey = inputAPIKey
	}

	authType := r.config.AuthType
	if inputAuthType, ok := input["auth_type"].(string); ok && inputAuthType != "" {
		authType = inputAuthType
	}

	timeout := r.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	maxRetries := r.config.MaxRetries
	if inputMaxRetries, ok := input["max_retries"].(float64); ok {
		maxRetries = int(inputMaxRetries)
	}

	enableRetry := r.config.EnableRetry
	if inputEnableRetry, ok := input["enable_retry"].(bool); ok {
		enableRetry = inputEnableRetry
	}

	retryDelay := r.config.RetryDelay
	if inputRetryDelay, ok := input["retry_delay"].(float64); ok {
		retryDelay = int(inputRetryDelay)
	}

	enableCaching := r.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := r.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	requestHeaders := r.config.RequestHeaders
	if inputHeaders, ok := input["request_headers"].(map[string]interface{}); ok {
		requestHeaders = make(map[string]string)
		for k, v := range inputHeaders {
			if vStr, ok := v.(string); ok {
				requestHeaders[k] = vStr
			}
		}
	}

	responseHeaders := r.config.ResponseHeaders
	if inputRespHeaders, ok := input["response_headers"].([]interface{}); ok {
		responseHeaders = make([]string, len(inputRespHeaders))
		for i, val := range inputRespHeaders {
			responseHeaders[i] = fmt.Sprintf("%v", val)
		}
	}

	enableProfiling := r.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	returnRawResults := r.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := r.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	sslVerification := r.config.SSLVerification
	if inputSSLVerification, ok := input["ssl_verification"].(bool); ok {
		sslVerification = inputSSLVerification
	}

	enableRateLimit := r.config.EnableRateLimit
	if inputEnableRateLimit, ok := input["enable_rate_limit"].(bool); ok {
		enableRateLimit = inputEnableRateLimit
	}

	rateLimit := r.config.RateLimit
	if inputRateLimit, ok := input["rate_limit"].(float64); ok {
		rateLimit = int(inputRateLimit)
	}

	// Validasi input yang diperlukan
	if baseURL == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "base_url diperlukan untuk permintaan API",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Validasi metode permintaan
	method := "GET"
	if inputMethod, exists := input["method"]; exists {
		if methodStr, ok := inputMethod.(string); ok {
			method = methodStr
		}
	}

	// Validasi URL
	url := baseURL
	if inputURL, exists := input["url"]; exists {
		if urlStr, ok := inputURL.(string); ok {
			url = urlStr
		}
	}

	if url == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "url diperlukan untuk permintaan API",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks operasi dengan timeout
	apiCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Lakukan permintaan API
	apiResult, err := r.makeAPIRequest(apiCtx, method, url, input)
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
		"base_url":             baseURL,
		"request_url":          url,
		"method":               method,
		"auth_type":            authType,
		"api_key_used":         apiKey != "",
		"timeout":              timeout,
		"max_retries":          maxRetries,
		"enable_retry":         enableRetry,
		"retry_delay":          retryDelay,
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"api_result":           apiResult,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"request_headers":      requestHeaders,
		"response_headers":     responseHeaders,
		"ssl_verification":     sslVerification,
		"enable_rate_limit":    enableRateLimit,
		"rate_limit":           rateLimit,
		"config":               r.config,
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

// makeAPIRequest membuat permintaan ke API
func (r *RESTAPIClientNode) makeAPIRequest(ctx context.Context, method, url string, input map[string]interface{}) (map[string]interface{}, error) {
	// Simulasikan waktu pemrosesan
	time.Sleep(100 * time.Millisecond)

	// Simulasikan berbagai respons tergantung metode dan URL
	var responseBody map[string]interface{}
	var responseHeaders map[string]interface{}
	var statusCode int

	switch method {
	case "GET":
		// Simulasikan respons GET
		responseBody = map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":    1,
					"name":  "Item 1",
					"value": "value1",
				},
				{
					"id":    2,
					"name":  "Item 2",
					"value": "value2",
				},
			},
			"total":    2,
			"page":     1,
			"per_page": 10,
		}
		responseHeaders = map[string]interface{}{
			"content-type":  "application/json",
			"server":        "API-Server/1.0",
			"x-rate-limit":  "100",
			"x-rate-remaining": "99",
		}
		statusCode = 200
	case "POST":
		// Simulasikan respons POST
		responseBody = map[string]interface{}{
			"success": true,
			"message": "Data created successfully",
			"id":      123,
			"data":    input,
		}
		responseHeaders = map[string]interface{}{
			"content-type":   "application/json",
			"location":       fmt.Sprintf("%s/%d", url, 123),
		}
		statusCode = 201
	case "PUT":
		// Simulasikan respons PUT
		responseBody = map[string]interface{}{
			"success": true,
			"message": "Data updated successfully",
			"id":      input["id"],
			"data":    input,
		}
		responseHeaders = map[string]interface{}{
			"content-type": "application/json",
		}
		statusCode = 200
	case "DELETE":
		// Simulasikan respons DELETE
		responseBody = map[string]interface{}{
			"success": true,
			"message": "Data deleted successfully",
			"id":      input["id"],
		}
		responseHeaders = map[string]interface{}{
			"content-type": "application/json",
		}
		statusCode = 200
	default:
		// Metode tidak didukung
		responseBody = map[string]interface{}{
			"error":   "Method not allowed",
			"message": fmt.Sprintf("Method %s not supported", method),
		}
		statusCode = 405
	}

	// Simulasikan kemungkinan kesalahan
	if shouldFail, exists := input["simulate_error"]; exists {
		if shouldFailBool, ok := shouldFail.(bool); ok && shouldFailBool {
			responseBody = map[string]interface{}{
				"error":   "Simulated API error",
				"message": "This is a simulated error for testing purposes",
			}
			statusCode = 500
		}
	}

	// Jika URL mengandung kata "error", kembalikan error
	if contains(url, "error") {
		responseBody = map[string]interface{}{
			"error":   "API error",
			"message": "Error returned from API due to error in URL path",
		}
		statusCode = 400
	}

	result := map[string]interface{}{
		"request_made":   true,
		"method":         method,
		"url":            url,
		"status_code":    statusCode,
		"response_body":  responseBody,
		"response_headers": responseHeaders,
		"request_time":   time.Now().Unix(),
		"processing_time": time.Since(time.Now().Add(-100 * time.Millisecond)).Seconds(),
		"retry_count":    0, // Dalam simulasi, tidak ada retry
		"from_cache":     false, // Dalam simulasi, tidak dari cache
		"rate_limited":   false, // Dalam simulasi, tidak dibatasi lajunya
	}

	return result, nil
}

// GetType mengembalikan jenis node
func (r *RESTAPIClientNode) GetType() string {
	return "rest_api_client"
}

// GetID mengembalikan ID unik untuk instance node
func (r *RESTAPIClientNode) GetID() string {
	return "rest_api_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterRESTAPIClientNode mendaftarkan node REST API Client dengan engine
func RegisterRESTAPIClientNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("rest_api_client", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewRESTAPIClientNode(config)
	})
}

// contains adalah fungsi helper untuk memeriksa apakah string mengandung substring
func contains(text, substr string) bool {
	textLen := len(text)
	substrLen := len(substr)
	
	if substrLen > textLen {
		return false
	}
	
	for i := 0; i <= textLen-substrLen; i++ {
		match := true
		for j := 0; j < substrLen; j++ {
			if text[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	
	return false
}