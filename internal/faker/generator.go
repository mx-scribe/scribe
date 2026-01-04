package faker

import (
	"fmt"
	"math/rand/v2"
)

// Generator creates random log entries.
type Generator struct {
	rng   *rand.Rand
	chaos bool
}

// NewGenerator creates a new log generator.
func NewGenerator(seed int64, chaos bool) *Generator {
	var rng *rand.Rand
	if seed != 0 {
		rng = rand.New(rand.NewPCG(uint64(seed), uint64(seed+1))) //nolint:gosec // Not for cryptographic use
	} else {
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())) //nolint:gosec // Not for cryptographic use
	}
	return &Generator{rng: rng, chaos: chaos}
}

// Generate returns a random log based on category distribution.
func (g *Generator) Generate() LogEntry {
	roll := g.rng.IntN(100)

	switch {
	case roll < WeightHTTP:
		return g.GenerateHTTP()
	case roll < WeightHTTP+WeightApplication:
		return g.GenerateApplication()
	case roll < WeightHTTP+WeightApplication+WeightDatabase:
		return g.GenerateDatabase()
	case roll < WeightHTTP+WeightApplication+WeightDatabase+WeightSecurity:
		return g.GenerateSecurity()
	case roll < WeightHTTP+WeightApplication+WeightDatabase+WeightSecurity+WeightSystem:
		return g.GenerateSystem()
	case roll < WeightHTTP+WeightApplication+WeightDatabase+WeightSecurity+WeightSystem+WeightBusiness:
		return g.GenerateBusiness()
	default:
		return g.GenerateChaos()
	}
}

// GenerateCategory returns a log from a specific category.
func (g *Generator) GenerateCategory(category string) LogEntry {
	switch category {
	case "http":
		return g.GenerateHTTP()
	case "application":
		return g.GenerateApplication()
	case "database":
		return g.GenerateDatabase()
	case "security":
		return g.GenerateSecurity()
	case "system":
		return g.GenerateSystem()
	case "business":
		return g.GenerateBusiness()
	case "chaos":
		return g.GenerateChaos()
	default:
		return g.Generate()
	}
}

// GenerateHTTP creates an HTTP access log.
func (g *Generator) GenerateHTTP() LogEntry {
	method := randomPick(g.rng, httpMethods)
	path := randomPick(g.rng, httpPaths)
	var status int
	if g.chaos {
		status = randomPick(g.rng, httpStatusesChaos)
	} else {
		status = randomPick(g.rng, httpStatusesNormal)
	}
	responseTime := randomDuration(g.rng, 5, 500)
	if g.chaos && g.rng.IntN(10) < 3 {
		responseTime = randomDuration(g.rng, 1000, 5000) // slow response
	}

	return LogEntry{
		Header: LogHeader{
			Title:  fmt.Sprintf("%s %s HTTP/1.1", method, path),
			Source: "nginx",
		},
		Body: map[string]any{
			"remote_addr":      randomIP(g.rng),
			"method":           method,
			"path":             path,
			"status":           status,
			"bytes":            randomDuration(g.rng, 100, 50000),
			"response_time_ms": responseTime,
			"user_agent":       randomPick(g.rng, userAgents),
		},
	}
}

// GenerateApplication creates an application log.
func (g *Generator) GenerateApplication() LogEntry {
	// Pick type: auth, job, notification, or error with stack trace
	logType := g.rng.IntN(10)

	switch {
	case logType < 3: // Auth events
		return g.generateAuthLog()
	case logType < 5: // Job events
		return g.generateJobLog()
	case logType < 7: // Notification events
		return g.generateNotificationLog()
	default: // Error with stack trace
		return g.generateStackTraceLog()
	}
}

func (g *Generator) generateAuthLog() LogEntry {
	success := !g.chaos || g.rng.IntN(2) == 0

	if success {
		return LogEntry{
			Header: LogHeader{
				Title:    "User authentication successful",
				Source:   "auth-service",
				Severity: "success",
			},
			Body: map[string]any{
				"user_id":    randomID(g.rng, "usr_"),
				"method":     "password",
				"ip":         randomIP(g.rng),
				"session_id": randomID(g.rng, "sess_"),
			},
		}
	}

	return LogEntry{
		Header: LogHeader{
			Title:    "Authentication failed: invalid credentials",
			Source:   "auth-service",
			Severity: "warning",
		},
		Body: map[string]any{
			"email":    randomEmail(g.rng),
			"ip":       randomIP(g.rng),
			"attempts": g.rng.IntN(5) + 1,
		},
	}
}

func (g *Generator) generateJobLog() LogEntry {
	jobs := []string{"daily-report", "cleanup-old-data", "sync-inventory", "send-reminders", "generate-invoices"}
	job := randomPick(g.rng, jobs)
	success := !g.chaos || g.rng.IntN(3) != 0

	if success {
		return LogEntry{
			Header: LogHeader{
				Title:    fmt.Sprintf("Job completed: %s", job),
				Source:   "job-worker",
				Severity: "success",
			},
			Body: map[string]any{
				"job_id":            randomID(g.rng, "job_"),
				"job_type":          job,
				"duration_ms":       randomDuration(g.rng, 500, 30000),
				"records_processed": g.rng.IntN(10000) + 100,
			},
		}
	}

	return LogEntry{
		Header: LogHeader{
			Title:    fmt.Sprintf("Job failed: %s", job),
			Source:   "job-worker",
			Severity: "error",
		},
		Body: map[string]any{
			"job_id":   randomID(g.rng, "job_"),
			"job_type": job,
			"error":    randomPick(g.rng, errorMessages),
			"attempt":  g.rng.IntN(3) + 1,
		},
	}
}

func (g *Generator) generateNotificationLog() LogEntry {
	types := []string{"email", "sms", "push"}
	notifType := randomPick(g.rng, types)
	success := !g.chaos || g.rng.IntN(4) != 0

	if success {
		return LogEntry{
			Header: LogHeader{
				Title:    fmt.Sprintf("%s sent successfully", notifType),
				Source:   "notification-service",
				Severity: "success",
			},
			Body: map[string]any{
				"type":       notifType,
				"recipient":  randomEmail(g.rng),
				"message_id": randomID(g.rng, "msg_"),
			},
		}
	}

	return LogEntry{
		Header: LogHeader{
			Title:    fmt.Sprintf("%s delivery failed", notifType),
			Source:   "notification-service",
			Severity: "error",
		},
		Body: map[string]any{
			"type":      notifType,
			"recipient": randomEmail(g.rng),
			"error":     "recipient unreachable",
		},
	}
}

func (g *Generator) generateStackTraceLog() LogEntry {
	lang := g.rng.IntN(4)

	switch lang {
	case 0:
		return LogEntry{
			Header: LogHeader{
				Title:    "panic: runtime error: index out of range",
				Source:   "api-server",
				Severity: "critical",
			},
			Body: map[string]any{
				"error": "index out of range [5] with length 3",
				"stack": goStackTrace,
			},
		}
	case 1:
		return LogEntry{
			Header: LogHeader{
				Title:    "TypeError: Cannot read property 'id' of undefined",
				Source:   "frontend",
				Severity: "error",
			},
			Body: map[string]any{
				"error": "TypeError: Cannot read property 'id' of undefined",
				"stack": jsStackTrace,
			},
		}
	case 2:
		return LogEntry{
			Header: LogHeader{
				Title:    "KeyError: 'user_id'",
				Source:   "worker",
				Severity: "error",
			},
			Body: map[string]any{
				"error": "KeyError: 'user_id'",
				"stack": pythonStackTrace,
			},
		}
	default:
		return LogEntry{
			Header: LogHeader{
				Title:    "java.lang.NullPointerException",
				Source:   "backend",
				Severity: "error",
			},
			Body: map[string]any{
				"error": "NullPointerException",
				"stack": javaStackTrace,
			},
		}
	}
}

// GenerateDatabase creates a database log.
func (g *Generator) GenerateDatabase() LogEntry {
	query := randomPick(g.rng, dbQueries)
	duration := randomDuration(g.rng, 1, 100)

	// Sometimes generate slow query or error
	if g.chaos && g.rng.IntN(5) == 0 {
		return LogEntry{
			Header: LogHeader{
				Title:    "Slow query detected",
				Source:   "postgresql",
				Severity: "warning",
			},
			Body: map[string]any{
				"query":       query,
				"duration_ms": randomDuration(g.rng, 1000, 5000),
				"threshold":   1000,
			},
		}
	}

	if g.chaos && g.rng.IntN(10) == 0 {
		errors := []string{"connection refused", "deadlock detected", "connection pool exhausted", "query timeout"}
		return LogEntry{
			Header: LogHeader{
				Title:    fmt.Sprintf("Database error: %s", randomPick(g.rng, errors)),
				Source:   "postgresql",
				Severity: "error",
			},
			Body: map[string]any{
				"query": query,
				"error": randomPick(g.rng, errors),
			},
		}
	}

	return LogEntry{
		Header: LogHeader{
			Title:  fmt.Sprintf("Query executed in %dms", duration),
			Source: "postgresql",
		},
		Body: map[string]any{
			"query":         query,
			"duration_ms":   duration,
			"rows_affected": g.rng.IntN(100) + 1,
			"connection_id": randomID(g.rng, "conn_"),
		},
	}
}

// GenerateSecurity creates a security log.
func (g *Generator) GenerateSecurity() LogEntry {
	if g.chaos || g.rng.IntN(3) == 0 {
		event := randomPick(g.rng, securityEvents)
		return LogEntry{
			Header: LogHeader{
				Title:    event,
				Source:   "security",
				Severity: "critical",
			},
			Body: map[string]any{
				"event_type": "security_alert",
				"ip":         randomIP(g.rng),
				"blocked":    true,
			},
		}
	}

	// Normal security events
	events := []string{
		"Session validated successfully",
		"Token refreshed",
		"Password reset requested",
		"MFA challenge completed",
	}
	return LogEntry{
		Header: LogHeader{
			Title:    randomPick(g.rng, events),
			Source:   "security",
			Severity: "info",
		},
		Body: map[string]any{
			"user_id": randomID(g.rng, "usr_"),
			"ip":      randomIP(g.rng),
		},
	}
}

// GenerateSystem creates a system log.
func (g *Generator) GenerateSystem() LogEntry {
	// Resource alerts in chaos mode
	if g.chaos && g.rng.IntN(3) == 0 {
		metrics := []struct {
			name      string
			value     int
			threshold int
		}{
			{"memory_usage", randomDuration(g.rng, 85, 98), 85},
			{"cpu_usage", randomDuration(g.rng, 80, 99), 80},
			{"disk_usage", randomDuration(g.rng, 90, 99), 90},
		}
		m := randomPick(g.rng, metrics)
		return LogEntry{
			Header: LogHeader{
				Title:    fmt.Sprintf("High %s detected: %d%%", m.name, m.value),
				Source:   "monitor",
				Severity: "warning",
			},
			Body: map[string]any{
				"metric":    m.name,
				"value":     m.value,
				"threshold": m.threshold,
				"host":      randomHost(g.rng),
			},
		}
	}

	// Container events
	if g.rng.IntN(2) == 0 {
		actions := []string{"started", "stopped", "restarted"}
		action := randomPick(g.rng, actions)
		containers := []string{"scribe-web-1", "scribe-api-1", "scribe-worker-1", "postgres-1", "redis-1"}
		container := randomPick(g.rng, containers)
		return LogEntry{
			Header: LogHeader{
				Title:    fmt.Sprintf("Container %s: %s", action, container),
				Source:   "docker",
				Severity: "info",
			},
			Body: map[string]any{
				"container_id": randomContainerID(g.rng),
				"image":        "scribe:latest",
				"action":       action,
			},
		}
	}

	// Normal system events
	event := randomPick(g.rng, systemEvents)
	return LogEntry{
		Header: LogHeader{
			Title:  event,
			Source: "system",
		},
		Body: map[string]any{
			"host": randomHost(g.rng),
		},
	}
}

// GenerateBusiness creates a business log.
func (g *Generator) GenerateBusiness() LogEntry {
	// Payment events
	if g.rng.IntN(2) == 0 {
		success := !g.chaos || g.rng.IntN(3) != 0

		if success {
			return LogEntry{
				Header: LogHeader{
					Title:    "Payment processed successfully",
					Source:   "payment-service",
					Severity: "success",
				},
				Body: map[string]any{
					"transaction_id": randomID(g.rng, "txn_"),
					"amount":         randomAmount(g.rng),
					"currency":       randomCurrency(g.rng),
					"method":         "card",
					"customer_id":    randomID(g.rng, "cus_"),
				},
			}
		}

		reasons := []string{"card declined", "insufficient funds", "invalid card", "fraud detected"}
		return LogEntry{
			Header: LogHeader{
				Title:    "Payment failed",
				Source:   "payment-service",
				Severity: "error",
			},
			Body: map[string]any{
				"transaction_id": randomID(g.rng, "txn_"),
				"amount":         randomAmount(g.rng),
				"currency":       randomCurrency(g.rng),
				"error":          randomPick(g.rng, reasons),
				"customer_id":    randomID(g.rng, "cus_"),
			},
		}
	}

	// Deployment events
	versions := []string{"1.2.3", "1.2.4", "1.3.0", "2.0.0-rc1"}
	return LogEntry{
		Header: LogHeader{
			Title:    fmt.Sprintf("Deployment completed: v%s", randomPick(g.rng, versions)),
			Source:   "deploy",
			Severity: "success",
		},
		Body: map[string]any{
			"version":     randomPick(g.rng, versions),
			"environment": "production",
			"duration_s":  randomDuration(g.rng, 30, 120),
			"commit":      randomID(g.rng, ""),
		},
	}
}

// GenerateChaos creates unstructured/messy logs.
func (g *Generator) GenerateChaos() LogEntry {
	chaosType := g.rng.IntN(6)

	switch chaosType {
	case 0: // Plain text error
		return LogEntry{
			Header: LogHeader{
				Title: "ERROR 2026-01-03 10:15:32 - Failed to connect to database after 3 retries",
			},
		}
	case 1: // Mixed format
		return LogEntry{
			Header: LogHeader{
				Title: fmt.Sprintf("[WARN] disk usage %d%% on /dev/sda1 - consider cleanup", randomDuration(g.rng, 80, 95)),
			},
		}
	case 2: // Legacy format
		return LogEntry{
			Header: LogHeader{
				Title: "***CRITICAL*** System overload detected - immediate action required",
			},
		}
	case 3: // Minimal
		words := []string{"timeout", "error", "failed", "retry", "disconnect"}
		return LogEntry{
			Header: LogHeader{
				Title: randomPick(g.rng, words),
			},
		}
	case 4: // JSON inside title
		return LogEntry{
			Header: LogHeader{
				Title: `{"level":"error","msg":"connection refused","port":5432}`,
			},
		}
	default: // Multi-line
		return LogEntry{
			Header: LogHeader{
				Title: "Error processing request\nDetails: connection timeout\nRetry in 30s",
			},
		}
	}
}
