package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CSRFConfig holds CSRF protection configuration
type CSRFConfig struct {
	TokenLength    int
	TokenLookup    string // "header:X-CSRF-Token" or "form:csrf_token"
	CookieName     string
	CookieDomain   string
	CookiePath     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite string
	Expiration     time.Duration
	KeyGenerator   func() (string, error)
}

// CSRFMiddleware provides CSRF protection
type CSRFMiddleware struct {
	config CSRFConfig
	tokens sync.Map // map[string]time.Time
}

// NewCSRFMiddleware creates a new CSRF middleware
func NewCSRFMiddleware(config CSRFConfig) *CSRFMiddleware {
	// Set defaults
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.TokenLookup == "" {
		config.TokenLookup = "header:X-CSRF-Token"
	}
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.CookiePath == "" {
		config.CookiePath = "/"
	}
	if config.CookieSameSite == "" {
		config.CookieSameSite = "Lax"
	}
	if config.Expiration == 0 {
		config.Expiration = 24 * time.Hour
	}
	if config.KeyGenerator == nil {
		config.KeyGenerator = defaultKeyGenerator
	}

	middleware := &CSRFMiddleware{
		config: config,
	}

	// Start cleanup goroutine
	go middleware.cleanupExpiredTokens()

	return middleware
}

// Protect returns a Fiber handler that validates CSRF tokens
func (m *CSRFMiddleware) Protect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip CSRF for safe methods
		method := c.Method()
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return c.Next()
		}

		// Get token from request
		token := m.extractToken(c)
		if token == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CSRF token missing",
			})
		}

		// Validate token
		if !m.validateToken(token) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Invalid CSRF token",
			})
		}

		return c.Next()
	}
}

// GenerateToken generates a new CSRF token and sets it in cookie
func (m *CSRFMiddleware) GenerateToken(c *fiber.Ctx) error {
	token, err := m.config.KeyGenerator()
	if err != nil {
		return err
	}

	// Store token with expiration
	m.tokens.Store(token, time.Now().Add(m.config.Expiration))

	// Set cookie
	cookie := &fiber.Cookie{
		Name:     m.config.CookieName,
		Value:    token,
		Path:     m.config.CookiePath,
		Domain:   m.config.CookieDomain,
		Expires:  time.Now().Add(m.config.Expiration),
		Secure:   m.config.CookieSecure,
		HTTPOnly: m.config.CookieHTTPOnly,
		SameSite: m.config.CookieSameSite,
	}
	c.Cookie(cookie)

	// Also set in response header for SPA
	c.Set("X-CSRF-Token", token)

	return nil
}

// extractToken extracts CSRF token from request
func (m *CSRFMiddleware) extractToken(c *fiber.Ctx) string {
	// Try header first
	token := c.Get("X-CSRF-Token")
	if token != "" {
		return token
	}

	// Try form field
	token = c.FormValue("csrf_token")
	if token != "" {
		return token
	}

	// Try cookie
	token = c.Cookies(m.config.CookieName)
	return token
}

// validateToken validates a CSRF token
func (m *CSRFMiddleware) validateToken(token string) bool {
	if token == "" {
		return false
	}

	// Check if token exists and not expired
	expiration, exists := m.tokens.Load(token)
	if !exists {
		return false
	}

	expirationTime, ok := expiration.(time.Time)
	if !ok {
		return false
	}

	// Check if expired
	if time.Now().After(expirationTime) {
		m.tokens.Delete(token)
		return false
	}

	return true
}

// cleanupExpiredTokens periodically removes expired tokens
func (m *CSRFMiddleware) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.tokens.Range(func(key, value interface{}) bool {
			expirationTime, ok := value.(time.Time)
			if ok && now.After(expirationTime) {
				m.tokens.Delete(key)
			}
			return true
		})
	}
}

// defaultKeyGenerator generates a random token
func defaultKeyGenerator() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetToken returns the current CSRF token for the request
func (m *CSRFMiddleware) GetToken(c *fiber.Ctx) string {
	return c.Cookies(m.config.CookieName)
}
