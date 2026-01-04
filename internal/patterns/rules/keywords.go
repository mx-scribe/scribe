package rules

// ErrorKeywords contains keywords that indicate error severity.
var ErrorKeywords = []string{
	"error", "failed", "failure", "fatal", "critical", "crash",
	"exception", "panic", "abort", "aborted", "refused", "denied",
	"reject", "rejected", "timeout", "timed out", "unavailable",
	"unreachable", "invalid", "corrupt", "corrupted", "broken",
	"violation", "exceeded", "overflow", "underflow", "leak",
	"died", "dying", "dump", "dumped", "fault", "faulted",
	"kill", "killed", "terminate", "terminated", "segfault",
	"segmentation", "core dump", "stack overflow", "out of memory",
	"oom", "cannot", "could not", "unable", "impossible",
}

// WarningKeywords contains keywords that indicate warning severity.
var WarningKeywords = []string{
	"warning", "warn", "deprecated", "deprecation", "slow",
	"slower", "delay", "delayed", "lag", "lagging", "retry",
	"retrying", "retried", "pending", "blocked", "blocking",
	"queue full", "high load", "degraded", "flaky", "unstable",
	"intermittent", "occasional", "sometimes", "timeout soon",
}

// SuccessKeywords contains keywords that indicate success.
var SuccessKeywords = []string{
	"success", "successful", "successfully", "succeeded", "complete",
	"completed", "done", "finished", "processed", "created",
	"updated", "saved", "stored", "published", "deployed",
	"approved", "accepted", "validated", "verified", "confirmed",
	"established", "connected", "ready", "available", "online",
	"restored", "recovered", "fixed", "resolved", "passed",
	"ok", "okay", "working", "operational", "healthy",
}

// DebugKeywords contains keywords that indicate debug/trace level.
var DebugKeywords = []string{
	"debug", "debugging", "trace", "tracing", "verbose",
	"entering", "entered", "exiting", "exited", "calling",
	"called", "executing", "executed", "invoking", "invoked",
	"beginning", "starting", "stopping", "ended",
}
