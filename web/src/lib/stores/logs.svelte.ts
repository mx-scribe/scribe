// Logs Store - Log data management using Svelte 5 runes with SSE support

import { getClient, getSSEClient, type LogFilters as ApiLogFilters, type SSELogCreatedData } from '$lib/api/index.js';

export interface LogHeader {
	type: string;
	title: string;
	color?: string;
	description?: string;
	source?: string;
}

export interface LogEntry {
	id: string;
	header: LogHeader;
	body: Record<string, unknown>;
	created_at: string;
	updated_at: string;
}

export interface LogStats {
	total: number;
	recent: number;
	error_rate: number;
	by_severity: Record<string, number>;
	by_source: Record<string, number>;
}

interface LogsState {
	logs: LogEntry[];
	stats: LogStats | null;
	isLoading: boolean;
	error: string | null;
	sseConnected: boolean;
	version: string;
}

const defaultStats: LogStats = {
	total: 0,
	recent: 0,
	error_rate: 0,
	by_severity: {},
	by_source: {}
};

function createLogsStore() {
	let state = $state<LogsState>({
		logs: [],
		stats: null,
		isLoading: false,
		error: null,
		sseConnected: false,
		version: ''
	});

	// SSE client reference
	let sseClient: ReturnType<typeof getSSEClient> | null = null;

	return {
		get state() {
			return state;
		},

		get stats(): LogStats {
			return state.stats ?? defaultStats;
		},

		// Actions
		setLogs(logs: LogEntry[]) {
			state.logs = logs;
		},

		setStats(stats: LogStats) {
			state.stats = stats;
		},

		setLoading(loading: boolean) {
			state.isLoading = loading;
		},

		setError(error: string | null) {
			state.error = error;
		},

		// Add a single log to the beginning of the list (for SSE)
		addLog(log: LogEntry) {
			state.logs = [log, ...state.logs];
		},

		// Remove a log by ID (for SSE)
		removeLog(id: string) {
			state.logs = state.logs.filter((l) => l.id !== id);
		},

		// Connect to SSE for real-time updates
		connectSSE() {
			if (sseClient) return; // Already connected

			sseClient = getSSEClient();
			sseClient.connect({
				onConnected: () => {
					state.sseConnected = true;
				},
				onDisconnected: () => {
					state.sseConnected = false;
				},
				onLogCreated: (data: SSELogCreatedData) => {
					// Add new log to the top of the list
					const newLog: LogEntry = {
						id: String(data.id),
						header: {
							type: data.header.severity,
							title: data.header.title,
							color: data.header.color,
							description: data.header.description,
							source: data.header.source
						},
						body: data.body,
						created_at: data.created_at,
						updated_at: data.created_at
					};
					state.logs = [newLog, ...state.logs];

					// Update stats
					if (state.stats) {
						const severity = data.header.severity;
						const newBySeverity = { ...state.stats.by_severity };
						newBySeverity[severity] = (newBySeverity[severity] || 0) + 1;

						const newTotal = state.stats.total + 1;
						const errorCount = (newBySeverity['error'] || 0) + (newBySeverity['critical'] || 0);

						state.stats = {
							...state.stats,
							total: newTotal,
							recent: state.stats.recent + 1,
							error_rate: newTotal > 0 ? (errorCount / newTotal) * 100 : 0,
							by_severity: newBySeverity
						};
					}
				},
				onLogDeleted: (data) => {
					state.logs = state.logs.filter((l) => l.id !== String(data.id));
					if (state.stats) {
						state.stats = {
							...state.stats,
							total: Math.max(0, state.stats.total - 1)
						};
					}
				},
				onStatsUpdated: (data) => {
					// Full stats refresh from server
					if (state.stats) {
						const errorCount = (data.by_severity['error'] || 0) + (data.by_severity['critical'] || 0);
						state.stats = {
							total: data.total,
							recent: data.last_24_hours,
							error_rate: data.total > 0 ? (errorCount / data.total) * 100 : 0,
							by_severity: data.by_severity,
							by_source: data.by_source
						};
					}
				},
				onError: (error) => {
					state.error = error.message;
				}
			});
		},

		// Disconnect SSE
		disconnectSSE() {
			if (sseClient) {
				sseClient.disconnect();
				sseClient = null;
				state.sseConnected = false;
			}
		},

		// Fetch logs from API
		async fetchLogs(filters?: ApiLogFilters) {
			state.isLoading = true;
			state.error = null;

			try {
				const client = getClient();
				const response = await client.listLogs(filters);

				state.logs = response.logs.map((log) => ({
					id: String(log.id),
					header: {
						type: log.header.severity,
						title: log.header.title,
						color: log.header.color,
						description: log.header.description,
						source: log.header.source
					},
					body: log.body,
					created_at: log.created_at,
					updated_at: log.created_at
				}));

				return response;
			} catch (error) {
				state.error = error instanceof Error ? error.message : 'Failed to fetch logs';
				throw error;
			} finally {
				state.isLoading = false;
			}
		},

		// Fetch stats from API
		async fetchStats() {
			try {
				const client = getClient();
				const apiStats = await client.getStats();

				const errorCount = (apiStats.by_severity['error'] || 0) + (apiStats.by_severity['critical'] || 0);

				state.stats = {
					total: apiStats.total,
					recent: apiStats.last_24_hours,
					error_rate: apiStats.total > 0 ? (errorCount / apiStats.total) * 100 : 0,
					by_severity: apiStats.by_severity,
					by_source: apiStats.by_source
				};

				return state.stats;
			} catch (error) {
				state.error = error instanceof Error ? error.message : 'Failed to fetch stats';
				throw error;
			}
		},

		// Check if backend is available
		async checkBackend(): Promise<boolean> {
			try {
				const client = getClient();
				const health = await client.health();
				state.version = health.version ?? '';
				return true;
			} catch {
				return false;
			}
		}
	};
}

export const logsStore = createLogsStore();
