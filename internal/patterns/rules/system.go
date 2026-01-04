package rules

// SystemErrorCodes maps system error codes to severity levels.
var SystemErrorCodes = map[string]string{
	"ECONNREFUSED": "error",
	"ETIMEDOUT":    "error",
	"ENOTFOUND":    "error",
	"ECONNRESET":   "error",
	"EPIPE":        "error",
	"EACCES":       "error",
	"ENOENT":       "warning",
	"EISDIR":       "error",
	"EMFILE":       "critical",
	"ENOMEM":       "critical",
	"ENOSPC":       "critical",
	"EIO":          "critical",
	"EROFS":        "error",
}
