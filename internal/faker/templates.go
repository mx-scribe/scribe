package faker

import (
	"fmt"
	"math/rand/v2"
)

// LogEntry represents a log to be sent to the API.
type LogEntry struct {
	Header LogHeader `json:"header"`
	Body   any       `json:"body,omitempty"`
}

// LogHeader represents the header of a log entry.
type LogHeader struct {
	Title    string `json:"title"`
	Source   string `json:"source,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// HTTP paths for realistic logs.
var httpPaths = []string{
	"/api/users", "/api/users/{id}", "/api/auth/login", "/api/auth/logout",
	"/api/products", "/api/products/{id}", "/api/orders", "/api/orders/{id}",
	"/api/payments", "/api/webhooks", "/api/health", "/api/metrics",
	"/static/js/app.js", "/static/css/main.css", "/favicon.ico",
}

// HTTP methods with realistic distribution.
var httpMethods = []string{"GET", "GET", "GET", "GET", "POST", "POST", "PUT", "DELETE", "PATCH"}

// HTTP status codes with realistic distribution.
var httpStatusesNormal = []int{200, 200, 200, 200, 200, 201, 204, 301, 302, 400, 401, 404, 500}
var httpStatusesChaos = []int{200, 200, 400, 401, 403, 404, 500, 500, 502, 503}

// User agents.
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15",
	"curl/8.1.0",
	"PostmanRuntime/7.32.0",
}

// Service names (reserved for future use).
var _ = []string{
	"api-server", "auth-service", "payment-service", "notification-service",
	"worker", "scheduler", "gateway", "frontend", "backend",
}

// Database queries.
var dbQueries = []string{
	"SELECT * FROM users WHERE id = $1",
	"INSERT INTO orders (user_id, total) VALUES ($1, $2)",
	"UPDATE products SET stock = stock - 1 WHERE id = $1",
	"DELETE FROM sessions WHERE expires_at < NOW()",
	"SELECT COUNT(*) FROM log_entries WHERE created_at > $1",
}

// Error messages.
var errorMessages = []string{
	"connection refused", "timeout exceeded", "invalid request",
	"unauthorized access", "resource not found", "internal server error",
	"database connection failed", "rate limit exceeded", "invalid token",
}

// Success messages (reserved for future use).
var _ = []string{
	"User authentication successful", "Payment processed successfully",
	"Order created successfully", "Email sent successfully",
	"Job completed: daily-report", "Deployment completed: v1.2.3",
	"Cache warmed successfully", "Backup completed",
}

// Warning messages (reserved for future use).
var _ = []string{
	"High memory usage detected: 85%", "Slow query detected: 1234ms",
	"Rate limit approaching: 80% used", "Certificate expiring in 7 days",
	"Connection pool at 90% capacity", "Retry attempt 2/3",
	"Deprecated API endpoint called", "Disk usage at 75%",
}

// Security events.
var securityEvents = []string{
	"Brute force attack detected", "SQL injection attempt blocked",
	"Invalid authentication attempt", "Rate limit exceeded for IP",
	"Suspicious login from new location", "CSRF token mismatch",
	"XSS attempt blocked", "IP blocked: multiple failed logins",
}

// System events.
var systemEvents = []string{
	"Service started", "Service stopped", "Health check passed",
	"Container restarted", "Memory limit reached",
	"Process crashed and recovered", "Disk cleanup completed",
}

// Stack traces by language.
var goStackTrace = `goroutine 1 [running]:
main.processItems(0xc0000b4000, 0x3)
	/app/handlers/items.go:45 +0x123
main.main()
	/app/main.go:23 +0x456`

var jsStackTrace = `TypeError: Cannot read property 'id' of undefined
    at UserProfile (components/UserProfile.jsx:23:15)
    at renderWithHooks (react-dom.js:1234:22)
    at mountIndeterminateComponent (react-dom.js:5678:13)`

var pythonStackTrace = `Traceback (most recent call last):
  File "/app/worker.py", line 45, in process
    user_id = data['user_id']
KeyError: 'user_id'`

var javaStackTrace = `java.lang.NullPointerException
	at com.app.service.UserService.getUser(UserService.java:42)
	at com.app.controller.UserController.show(UserController.java:28)
	at sun.reflect.NativeMethodAccessorImpl.invoke(Native Method)`

// Helper functions for random data generation.

func randomIP(rng *rand.Rand) string {
	return fmt.Sprintf("192.168.%d.%d", rng.IntN(256), rng.IntN(256))
}

func randomID(rng *rand.Rand, prefix string) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, 8)
	for i := range id {
		id[i] = chars[rng.IntN(len(chars))]
	}
	return prefix + string(id)
}

func randomEmail(rng *rand.Rand) string {
	names := []string{"john", "jane", "bob", "alice", "user", "admin", "test"}
	domains := []string{"example.com", "test.io", "mail.org", "company.co"}
	return names[rng.IntN(len(names))] + "@" + domains[rng.IntN(len(domains))]
}

func randomPick[T any](rng *rand.Rand, items []T) T {
	return items[rng.IntN(len(items))]
}

func randomDuration(rng *rand.Rand, min, max int) int {
	if min >= max {
		return min
	}
	return min + rng.IntN(max-min)
}

func randomAmount(rng *rand.Rand) float64 {
	amounts := []float64{9.99, 19.99, 29.99, 49.99, 99.99, 149.99, 199.99, 299.99}
	return amounts[rng.IntN(len(amounts))]
}

func randomCurrency(rng *rand.Rand) string {
	currencies := []string{"EUR", "USD", "GBP"}
	return currencies[rng.IntN(len(currencies))]
}

func randomHost(rng *rand.Rand) string {
	hosts := []string{"server-01", "server-02", "web-01", "api-01", "worker-01"}
	return hosts[rng.IntN(len(hosts))]
}

func randomContainerID(rng *rand.Rand) string {
	const hex = "0123456789abcdef"
	id := make([]byte, 12)
	for i := range id {
		id[i] = hex[rng.IntN(len(hex))]
	}
	return string(id)
}
