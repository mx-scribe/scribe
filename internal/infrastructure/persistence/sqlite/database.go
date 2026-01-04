package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Database represents the SQLite database connection.
type Database struct {
	conn *sql.DB
	path string
}

// NewDatabase creates a new database connection with WAL mode.
func NewDatabase(dbPath string) (*Database, error) {
	// Connection string with pragmas for WAL mode
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", dbPath)

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// SQLite works best with single connection
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	// Test connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &Database{
		conn: conn,
		path: dbPath,
	}

	return db, nil
}

// Conn returns the underlying database connection.
func (db *Database) Conn() *sql.DB {
	return db.conn
}

// Close closes the database connection with WAL checkpoint.
func (db *Database) Close() error {
	if db.conn == nil {
		return nil
	}

	// Checkpoint WAL before closing (critical for data integrity)
	if _, err := db.conn.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		// Log but don't fail - we still want to close the connection
		fmt.Printf("warning: WAL checkpoint failed: %v\n", err)
	}

	return db.conn.Close()
}

// Path returns the database file path.
func (db *Database) Path() string {
	return db.path
}
