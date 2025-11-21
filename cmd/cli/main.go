package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// DeviceCodeResponse represents response for device code
type DeviceCodeResponse struct {
	UserCode        string `json:"user_code"`
	DeviceCode      string `json:"device_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// TokenResponse represents JWT token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Credentials represents stored credentials
type Credentials struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

// CLIAuth handles CLI authentication
type CLIAuth struct {
	apiURL string
}

// NewCLIAuth creates a new CLI auth instance
func NewCLIAuth(apiURL string) *CLIAuth {
	if apiURL == "" {
		apiURL = "http://localhost:5001"
	}
	return &CLIAuth{apiURL: apiURL}
}

// Login initiates the login process
func (c *CLIAuth) Login(provider string) error {
	fmt.Printf("Initiating login with %s...\n", strings.Title(provider))

	// Start device flow
	deviceCode, err := c.initiateDeviceFlow(provider)
	if err != nil {
		return fmt.Errorf("failed to initiate device flow: %w", err)
	}

	fmt.Printf("\nTo sign in, use a web browser to open: %s\n", deviceCode.VerificationURI)
	fmt.Printf("Enter code: %s\n", deviceCode.UserCode)
	fmt.Println("Waiting for approval...")

	// Poll for verification
	credentials, err := c.pollForVerification(deviceCode.DeviceCode, deviceCode.Interval)
	if err != nil {
		return fmt.Errorf("failed to verify device: %w", err)
	}

	// Save credentials
	if err := c.saveCredentials(credentials); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("\n✅ Login successful!")
	return nil
}

// initiateDeviceFlow initiates the OAuth device flow
func (c *CLIAuth) initiateDeviceFlow(provider string) (*DeviceCodeResponse, error) {
	url := fmt.Sprintf("%s/auth/device", c.apiURL)
	
	payload := map[string]string{
		"provider": provider,
	}
	
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Citadel-Agent-CLI/1.0")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var deviceCodeResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceCodeResp); err != nil {
		return nil, err
	}
	
	return &deviceCodeResp, nil
}

// pollForVerification polls the server for device verification
func (c *CLIAuth) pollForVerification(deviceCode string, interval int) (*Credentials, error) {
	url := fmt.Sprintf("%s/auth/device/verify", c.apiURL)
	
	payload := map[string]string{
		"provider":   "github", // This would be dynamic in a full implementation
		"device_code": deviceCode,
	}
	
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	
	// Set timeout for the entire polling process
	timeout := time.After(10 * time.Minute) // Same as device code expiry
	
	for {
		select {
		case <-ticker.C:
			req, err := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
			if err != nil {
				return nil, err
			}
			
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "Citadel-Agent-CLI/1.0")
			
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				// Continue polling on network errors
				fmt.Printf("Network error, retrying...: %v\n", err)
				continue
			}
			
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			
			if resp.StatusCode == http.StatusOK {
				// Success! We got the tokens
				var tokenResp TokenResponse
				if err := json.Unmarshal(body, &tokenResp); err != nil {
					return nil, err
				}
				
				credentials := &Credentials{
					AccessToken:  tokenResp.AccessToken,
					RefreshToken: tokenResp.RefreshToken,
					Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
				}
				
				return credentials, nil
			} else if resp.StatusCode == http.StatusAccepted {
				// Still pending, continue polling
				continue
			} else {
				// Error occurred
				return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
			}
			
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for device verification")
		}
	}
}

// saveCredentials saves credentials to a local file
func (c *CLIAuth) saveCredentials(credentials *Credentials) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	
	configDir := filepath.Join(usr.HomeDir, ".config", "citadel-agent")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}
	
	credsPath := filepath.Join(configDir, "creds")
	
	file, err := os.OpenFile(credsPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return json.NewEncoder(file).Encode(credentials)
}

// loadCredentials loads credentials from a local file
func (c *CLIAuth) loadCredentials() (*Credentials, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	
	credsPath := filepath.Join(usr.HomeDir, ".config", "citadel-agent", "creds")
	
	file, err := os.Open(credsPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var credentials Credentials
	if err := json.NewDecoder(file).Decode(&credentials); err != nil {
		return nil, err
	}
	
	return &credentials, nil
}

// GetAccessToken returns the current access token, refreshing if necessary
func (c *CLIAuth) GetAccessToken() (string, error) {
	credentials, err := c.loadCredentials()
	if err != nil {
		return "", fmt.Errorf("not logged in, please run 'citadel-agent login'")
	}
	
	// Check if token is expired
	if time.Now().After(credentials.Expiry) {
		// In a real implementation, we would refresh the token here
		// For now, we'll just return an error
		return "", fmt.Errorf("access token expired, please re-login")
	}
	
	return credentials.AccessToken, nil
}

// Logout removes stored credentials
func (c *CLIAuth) Logout() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	
	credsPath := filepath.Join(usr.HomeDir, ".config", "citadel-agent", "creds")
	
	// Remove the credentials file
	if err := os.Remove(credsPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not currently logged in")
		}
		return err
	}
	
	fmt.Println("✅ Logged out successfully!")
	return nil
}

func main() {
	cliAuth := NewCLIAuth("")
	
	if len(os.Args) < 2 {
		fmt.Println("Usage: citadel-agent-cli login [provider] | logout")
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	switch command {
	case "login":
		provider := "github" // default provider
		if len(os.Args) > 2 {
			provider = os.Args[2]
		}
		
		if err := cliAuth.Login(provider); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "logout":
		if err := cliAuth.Logout(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "whoami":
		token, err := cliAuth.GetAccessToken()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Current access token: %s...\n", token[:20]) // Just show first 20 chars
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: citadel-agent-cli login [provider] | logout | whoami")
		os.Exit(1)
	}
}