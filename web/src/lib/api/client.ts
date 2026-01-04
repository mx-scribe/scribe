// SCRIBE API Client

import type {
	LogEntry,
	CreateLogRequest,
	CreateLogResponse,
	ListLogsResponse,
	LogFilters,
	StatsResponse,
	HealthResponse,
	ApiError
} from './types.js';

export class ScribeApiError extends Error {
	constructor(
		public status: number,
		message: string
	) {
		super(message);
		this.name = 'ScribeApiError';
	}
}

export interface ScribeClientConfig {
	baseUrl: string;
	timeout?: number;
}

export class ScribeClient {
	private baseUrl: string;
	private timeout: number;

	constructor(config: ScribeClientConfig) {
		this.baseUrl = config.baseUrl.replace(/\/$/, '');
		this.timeout = config.timeout ?? 10000;
	}

	private async request<T>(
		method: string,
		path: string,
		options?: {
			body?: unknown;
			params?: Record<string, string | number | undefined>;
		}
	): Promise<T> {
		const url = new URL(`${this.baseUrl}${path}`);

		if (options?.params) {
			Object.entries(options.params).forEach(([key, value]) => {
				if (value !== undefined && value !== '') {
					url.searchParams.set(key, String(value));
				}
			});
		}

		const controller = new AbortController();
		const timeoutId = setTimeout(() => controller.abort(), this.timeout);

		try {
			const response = await fetch(url.toString(), {
				method,
				headers: {
					'Content-Type': 'application/json'
				},
				body: options?.body ? JSON.stringify(options.body) : undefined,
				signal: controller.signal
			});

			clearTimeout(timeoutId);

			if (!response.ok) {
				const error: ApiError = await response.json().catch(() => ({ error: 'Unknown error' }));
				throw new ScribeApiError(response.status, error.error);
			}

			return response.json();
		} catch (error) {
			clearTimeout(timeoutId);
			if (error instanceof ScribeApiError) {
				throw error;
			}
			if (error instanceof Error && error.name === 'AbortError') {
				throw new ScribeApiError(0, 'Request timeout');
			}
			throw new ScribeApiError(0, error instanceof Error ? error.message : 'Network error');
		}
	}

	// Health check
	async health(): Promise<HealthResponse> {
		return this.request<HealthResponse>('GET', '/health');
	}

	// Logs
	async createLog(log: CreateLogRequest): Promise<CreateLogResponse> {
		return this.request<CreateLogResponse>('POST', '/api/logs', { body: log });
	}

	async listLogs(filters?: LogFilters): Promise<ListLogsResponse> {
		return this.request<ListLogsResponse>('GET', '/api/logs', {
			params: {
				limit: filters?.limit,
				page: filters?.page,
				severity: filters?.severity,
				source: filters?.source,
				search: filters?.search,
				from: filters?.from,
				to: filters?.to
			}
		});
	}

	async getLog(id: number): Promise<LogEntry> {
		return this.request<LogEntry>('GET', `/api/logs/${id}`);
	}

	// Stats
	async getStats(): Promise<StatsResponse> {
		return this.request<StatsResponse>('GET', '/api/stats');
	}

	// Export
	async exportJson(filters?: LogFilters): Promise<LogEntry[]> {
		return this.request<LogEntry[]>('GET', '/api/export/json', {
			params: {
				severity: filters?.severity,
				source: filters?.source,
				search: filters?.search,
				from: filters?.from,
				to: filters?.to
			}
		});
	}

	exportCsvUrl(filters?: LogFilters): string {
		const url = new URL(`${this.baseUrl}/api/export/csv`);
		if (filters) {
			Object.entries(filters).forEach(([key, value]) => {
				if (value !== undefined && value !== '') {
					url.searchParams.set(key, String(value));
				}
			});
		}
		return url.toString();
	}
}

// Default client instance
let defaultClient: ScribeClient | null = null;

export function getClient(): ScribeClient {
	if (!defaultClient) {
		// In development, use the Go backend on port 8080
		// In production, use same origin
		const baseUrl = import.meta.env.DEV ? 'http://localhost:8080' : window.location.origin;
		defaultClient = new ScribeClient({ baseUrl });
	}
	return defaultClient;
}

export function setClient(client: ScribeClient): void {
	defaultClient = client;
}
