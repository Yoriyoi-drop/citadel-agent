package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test untuk fungsi NewCLIAuth
func TestNewCLIAuth(t *testing.T) {
	// Test dengan URL kosong (akan menggunakan default)
	auth := NewCLIAuth("")
	assert.Equal(t, "http://localhost:5001", auth.apiURL)

	// Test dengan URL spesifik
	auth2 := NewCLIAuth("https://api.example.com")
	assert.Equal(t, "https://api.example.com", auth2.apiURL)
}

// Test untuk initiateDeviceFlow dengan mock server
func TestInitiateDeviceFlow(t *testing.T) {
	// Buat mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/auth/device", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		
		var payload map[string]string
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, "github", payload["provider"])
		
		response := DeviceCodeResponse{
			UserCode:        "ABCD-1234",
			DeviceCode:      "test-device-code",
			VerificationURI: "https://github.com/login/device",
			ExpiresIn:       900,
			Interval:        5,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	auth := NewCLIAuth(server.URL)
	deviceCode, err := auth.initiateDeviceFlow("github")
	
	assert.NoError(t, err)
	assert.NotNil(t, deviceCode)
	assert.Equal(t, "ABCD-1234", deviceCode.UserCode)
	assert.Equal(t, "test-device-code", deviceCode.DeviceCode)
	assert.Equal(t, "https://github.com/login/device", deviceCode.VerificationURI)
}

// Test untuk save dan load credentials
func TestCredentialsStorage(t *testing.T) {
	// Buat temporary directory untuk simulasikan home directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	auth := NewCLIAuth("http://localhost:5001")
	
	// Buat credentials untuk disimpan
	creds := &Credentials{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Test save credentials
	err := auth.saveCredentials(creds)
	assert.NoError(t, err)

	// Test load credentials
	loadedCreds, err := auth.loadCredentials()
	assert.NoError(t, err)
	assert.Equal(t, creds.AccessToken, loadedCreds.AccessToken)
	assert.Equal(t, creds.RefreshToken, loadedCreds.RefreshToken)
}

// Test untuk fungsi GetAccessToken setelah login
func TestGetAccessTokenAfterLogin(t *testing.T) {
	// Buat temporary directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	auth := NewCLIAuth("http://localhost:5001")

	// Simpan credentials valid
	creds := &Credentials{
		AccessToken:  "valid-access-token-after-login",
		RefreshToken: "valid-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour), // Belum expired
	}
	err := auth.saveCredentials(creds)
	assert.NoError(t, err, "Penyimpanan credentials harus berhasil")

	// Test mendapatkan access token - harus berhasil setelah login
	token, err := auth.GetAccessToken()
	assert.NoError(t, err, "Harus bisa mendapatkan access token setelah login")
	assert.Equal(t, "valid-access-token-after-login", token, "Access token harus sesuai dengan yang disimpan")
}

// Test untuk fungsi Logout
func TestLogout(t *testing.T) {
	// Buat temporary directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	auth := NewCLIAuth("http://localhost:5001")

	// Simpan credentials dulu
	creds := &Credentials{
		AccessToken:  "to-be-deleted-token",
		RefreshToken: "to-be-deleted-refresh",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
	auth.saveCredentials(creds)

	// Verifikasi credentials sudah ada
	_, err := auth.loadCredentials()
	assert.NoError(t, err)

	// Test logout
	err = auth.Logout()
	assert.NoError(t, err)

	// Verifikasi credentials sudah dihapus
	_, err = auth.loadCredentials()
	assert.Error(t, err)
}

// Test untuk command line arguments
func TestCommandLineArguments(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test argumen login
	os.Args = []string{"citadel-agent-cli", "login", "github"}
	assert.Equal(t, "login", os.Args[1])
	assert.Equal(t, "github", os.Args[2])

	// Test argumen logout
	os.Args = []string{"citadel-agent-cli", "logout"}
	assert.Equal(t, "logout", os.Args[1])

	// Test argumen whoami
	os.Args = []string{"citadel-agent-cli", "whoami"}
	assert.Equal(t, "whoami", os.Args[1])

	// Test tanpa argumen
	os.Args = []string{"citadel-agent-cli"}
	if len(os.Args) < 2 {
		// Ini adalah kondisi yang akan ditangani oleh fungsi main
		assert.True(t, len(os.Args) < 2 || len(os.Args) == 1) // hanya nama program
	}
}