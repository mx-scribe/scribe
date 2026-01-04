// API Types for SCRIBE Backend

// Log Header
export interface LogHeader {
	title: string;
	severity: string;
	source?: string;
	color?: string;
	description?: string;
}

// Log Metadata
export interface LogMetadata {
	derived_severity?: string;
	derived_source?: string;
	derived_category?: string;
}

// Log Entry from API
export interface LogEntry {
	id: number;
	header: LogHeader;
	body: Record<string, unknown>;
	metadata?: LogMetadata;
	created_at: string;
}

// Create Log Request
export interface CreateLogRequest {
	header: {
		title: string;
		severity?: string;
		source?: string;
		color?: string;
		description?: string;
	};
	body?: Record<string, unknown>;
}

// Create Log Response
export interface CreateLogResponse {
	id: number;
	title: string;
	severity: string;
	created_at: string;
}

// List Logs Response
export interface ListLogsResponse {
	logs: LogEntry[];
	total: number;
	limit: number;
	page: number;
}

// List Logs Filters
export interface LogFilters {
	limit?: number;
	page?: number;
	severity?: string;
	source?: string;
	search?: string;
	from?: string;
	to?: string;
}

// Stats Response
export interface StatsResponse {
	total: number;
	last_24_hours: number;
	by_severity: Record<string, number>;
	by_source: Record<string, number>;
}

// Health Response
export interface HealthResponse {
	status: string;
	version?: string;
}

// API Error
export interface ApiError {
	error: string;
}
