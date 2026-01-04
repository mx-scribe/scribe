package rules

// BusinessPatterns maps business logic patterns to severity levels.
var BusinessPatterns = map[string]string{
	// Payment patterns
	"payment failed":     "error",
	"payment successful": "success",
	"payment pending":    "info",
	"payment declined":   "error",
	"payment processed":  "success",

	// Order patterns
	"order completed": "success",
	"order canceled":  "warning",
	"order failed":    "error",
	"order placed":    "success",
	"order refunded":  "info",

	// Subscription patterns
	"subscription expired":   "warning",
	"subscription renewed":   "success",
	"subscription canceled":  "warning",
	"subscription activated": "success",

	// Trial patterns
	"trial expired": "info",
	"trial started": "success",

	// Invoice patterns
	"invoice overdue": "warning",
	"invoice paid":    "success",
	"invoice sent":    "info",

	// Refund patterns
	"refund processed": "info",
	"refund failed":    "error",
	"refund requested": "info",

	// User patterns
	"user registered":   "success",
	"user deleted":      "warning",
	"user deactivated":  "warning",
	"user activated":    "success",
	"login successful":  "success",
	"login failed":      "warning",
	"logout successful": "info",
	"password changed":  "info",
	"password reset":    "info",

	// Cart patterns
	"cart abandoned":     "warning",
	"cart updated":       "info",
	"checkout started":   "info",
	"checkout completed": "success",
}
