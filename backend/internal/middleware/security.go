// backend/internal/middleware/security.go
package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/citadel-agent/backend/internal/security"
)

// SecurityMiddleware provides security middleware for the application
type SecurityMiddleware struct {
	policy *security.SecurityPolicy
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(policy *security.SecurityPolicy) *SecurityMiddleware {
	return &SecurityMiddleware{
		policy: policy,
	}
}

// RequestValidation validates incoming requests
func (sm *SecurityMiddleware) RequestValidation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get request details
		reqURL := c.OriginalURL()
		method := c.Method()
		
		headers := make(map[string]string)
		c.Request().Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = string(value)
		})

		// Validate request against security policy
		if err := sm.policy.ValidateRequest(reqURL, method, headers); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": fmt.Sprintf("Request validation failed: %v", err),
			})
		}

		// Check for suspicious patterns in body
		body := string(c.Request().Body())
		if sm.policy.ContainsSuspiciousPatterns(body) {
			return c.Status(400).JSON(fiber.Map{
				"error": "Request body contains suspicious patterns",
			})
		}

		return c.Next()
	}
}

// RateLimiting provides rate limiting middleware
func (sm *SecurityMiddleware) RateLimiting() fiber.Handler {
	// Using an in-memory store for rate limiting (in production, use Redis or similar)
	rateLimitStore := make(map[string]*RateLimitRecord)

	return func(c *fiber.Ctx) error {
		// Get client IP
		clientIP := c.IP()
		
		// Create a key for rate limiting (IP + endpoint)
		key := fmt.Sprintf("%s:%s", clientIP, c.Path())
		
		// Get or create rate limit record
		record, exists := rateLimitStore[key]
		if !exists {
			record = &RateLimitRecord{
				Count:     0,
				LastReset: time.Now(),
			}
			rateLimitStore[key] = record
		}

		// Check if we need to reset the counter (based on window)
		rateLimit, endpoint := sm.getRateLimitForEndpoint(c.Path())
		if rateLimit != nil {
			if time.Since(record.LastReset) >= rateLimit.Window {
				record.Count = 0
				record.LastReset = time.Now()
			}

			// Increment count
			record.Count++

			// Check if limit exceeded
			if record.Count > rateLimit.Requests {
				return c.Status(429).JSON(fiber.Map{
					"error": rateLimit.Message,
					"retry_after": rateLimit.Window.Seconds(),
				})
			}
		}

		// Set rate limit headers
		if rateLimit != nil {
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimit.Requests))
			c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rateLimit.Requests-record.Count))
			c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", record.LastReset.Add(rateLimit.Window).Unix()))
		}

		return c.Next()
	}
}

// getRateLimitForEndpoint returns the appropriate rate limit for an endpoint
func (sm *SecurityMiddleware) getRateLimitForEndpoint(endpoint string) (*security.RateLimit, string) {
	// Check specific endpoints first
	if strings.HasPrefix(endpoint, "/api/v1/auth") {
		return sm.policy.RateLimits["auth"], "auth"
	}

	// Default API rate limit
	return sm.policy.RateLimits["api"], "api"
}

// ContentSecurityPolicy sets Content Security Policy headers
func (sm *SecurityMiddleware) ContentSecurityPolicy() fiber.Handler {
	cspHeader := sm.policy.GetCSPHeader()
	
	return func(c *fiber.Ctx) error {
		if cspHeader != "" {
			c.Set("Content-Security-Policy", cspHeader)
		}
		return c.Next()
	}
}

// CORS provides Cross-Origin Resource Sharing middleware
func (sm *SecurityMiddleware) CORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set CORS headers
		c.Set("Access-Control-Allow-Origin", "*") // In production, specify specific origins
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset")

		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(200)
		}

		return c.Next()
	}
}

// HSTS provides HTTP Strict Transport Security
func (sm *SecurityMiddleware) HSTS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		return c.Next()
	}
}

// XSSProtection sets headers to help prevent XSS attacks
func (sm *SecurityMiddleware) XSSProtection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY") // or "SAMEORIGIN"
		c.Set("X-XSS-Protection", "1; mode=block")
		return c.Next()
	}
}

// RequestSizeLimit limits the size of request bodies
func (sm *SecurityMiddleware) RequestSizeLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		contentLength := c.Request().Header.ContentLength()
		if contentLength > sm.policy.MaxRequestSize {
			return c.Status(413).JSON(fiber.Map{
				"error": "Request size exceeds limit",
			})
		}

		return c.Next()
	}
}

// Authentication middleware validates JWT tokens
func (sm *SecurityMiddleware) Authentication() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authorization header missing",
			})
		}

		// Check if it's a Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Extract token
		token := authHeader[7:]

		// In a real implementation, you would validate the JWT token here
		// For now, we'll just pass the request through
		// This is where you would decode the JWT and extract user information
		// For example:
		// claims, err := validateJWT(token)
		// if err != nil {
		//     return c.Status(401).JSON(fiber.Map{
		//         "error": "Invalid token",
		//     })
		// }
		// c.Locals("user_id", claims.UserID)
		// c.Locals("user_role", claims.Role)

		// For now, set a placeholder
		c.Locals("user_id", "placeholder_user_id")
		c.Locals("user_role", "admin")

		return c.Next()
	}
}

// Authorization middleware checks if user has required permissions
func (sm *SecurityMiddleware) Authorization(requiredPermissions []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract user info from context (set by authentication middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "User not authenticated",
			})
		}

		// In a real implementation, you would check permissions here
		// For now, we'll allow all permissions for the placeholder user
		// This is where you would check the user's permissions against the required ones

		return c.Next()
	}
}

// RateLimitRecord holds rate limiting information
type RateLimitRecord struct {
	Count     int
	LastReset time.Time
}

// SecurityHeaders combines multiple security headers
func (sm *SecurityMiddleware) SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set various security headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Set Content Security Policy if available
		if cspHeader := sm.policy.GetCSPHeader(); cspHeader != "" {
			c.Set("Content-Security-Policy", cspHeader)
		}

		// Set HSTS if using HTTPS
		if c.Scheme() == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		return c.Next()
	}
}

// NetworkAccessControl validates network requests
func (sm *SecurityMiddleware) NetworkAccessControl() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if this is an external request (for nodes that make HTTP requests)
		// This would be used in the context of workflow execution

		// For now, we'll just pass through
		// In a real implementation, this would validate outbound requests from workflow nodes
		return c.Next()
	}
}

// AuditLogging logs security-relevant events
func (sm *SecurityMiddleware) AuditLogging() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Log request details if audit logging is enabled
		if sm.policy.AuditLoggingEnabled {
			// In a real implementation, you would log this to a secure audit trail
			// For now, just continue
			fmt.Printf("AUDIT: %s %s from %s at %s\n", 
				c.Method(), 
				c.OriginalURL(), 
				c.IP(), 
				time.Now().Format(time.RFC3339))
		}

		return c.Next()
	}
}