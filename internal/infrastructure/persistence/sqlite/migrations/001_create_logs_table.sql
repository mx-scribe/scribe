-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS logs (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    title            TEXT NOT NULL,
    severity         TEXT NOT NULL DEFAULT 'info',
    source           TEXT,
    color            TEXT,
    description      TEXT,
    body             TEXT,
    derived_severity TEXT,
    derived_source   TEXT,
    derived_category TEXT,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_logs_severity ON logs(severity);
CREATE INDEX IF NOT EXISTS idx_logs_source ON logs(source);
CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at);
CREATE INDEX IF NOT EXISTS idx_logs_derived_severity ON logs(derived_severity);
CREATE INDEX IF NOT EXISTS idx_logs_derived_source ON logs(derived_source);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS logs;
-- +goose StatementEnd
