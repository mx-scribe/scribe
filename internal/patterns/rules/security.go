package rules

// SecurityPatterns contains keywords that indicate security issues (critical).
var SecurityPatterns = []string{
	"unauthorized", "forbidden", "auth failed", "authentication failed",
	"permission denied", "access denied", "invalid token", "token expired",
	"session expired", "breach", "leaked", "exposed", "vulnerability",
	"injection", "xss", "csrf", "compromised", "malicious", "exploit",
	"brute force", "ddos", "flooding", "suspicious", "intrusion",
	"sql injection", "code injection", "command injection",
}
