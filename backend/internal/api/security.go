// backend/internal/api/security.go
package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityMiddleware provides various security headers and protections
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set security headers
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';")
		
		// Rate limiting could be implemented here if needed
		// For now, we'll add basic request size limiting
		if c.Request.ContentLength > 10*1024*1024 { // 10MB
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		
		// Check if origin is allowed
		if isOriginAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		// Handle exact matches and subdomain wildcards
		if allowedOrigin == origin {
			return true
		}
		if strings.HasPrefix(allowedOrigin, "*.") {
			allowedDomain := allowedOrigin[2:] // Remove "*."
			if strings.HasSuffix(origin, "."+allowedDomain) {
				return true
			}
		}
	}
	return false
}

// RequestLoggingMiddleware logs incoming requests (excluding sensitive paths)
func RequestLoggingMiddleware(excludePaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Skip logging for certain paths
		for _, excludePath := range excludePaths {
			if strings.HasPrefix(c.Request.URL.Path, excludePath) {
				c.Next()
				return
			}
		}
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(start)
		
		// Log request info (excluding sensitive data)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		ip := c.ClientIP()
		
		// Format log entry (in a real system you'd send to a logging service)
		logEntry := gin.H{
			"method":     method,
			"path":       path,
			"status":     statusCode,
			"ip":         ip,
			"duration":   duration.String(),
			"user_agent": c.GetHeader("User-Agent"),
			"time":       time.Now().Format(time.RFC3339),
		}
		
		// Log level based on status code
		if statusCode >= 400 {
			// In a real implementation, you'd use your logging framework
			_ = logEntry // Use the log entry in your logging framework
		}
	}
}

// InputValidationMiddleware validates common input patterns to prevent injection
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for common SQL injection patterns in query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsSQLInjectionPattern(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid input detected",
						"field": key,
					})
					c.Abort()
					return
				}
			}
		}
		
		// Check for common XSS patterns in JSON body if it's a POST/PUT request
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				// In a real system, we'd parse and validate the JSON body
				// For now, we'll just continue
			}
		}
		
		c.Next()
	}
}

// containsSQLInjectionPattern checks for common SQL injection patterns
func containsSQLInjectionPattern(input string) bool {
	// Convert to lowercase for comparison
	lowerInput := strings.ToLower(input)
	
	// Common SQL injection patterns
	patterns := []string{
		"' or 1=1",
		"'; drop table",
		"'; exec ",
		"union select",
		"insert into",
		"update ",
		"delete from",
		"xp_cmdshell",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	
	return false
}