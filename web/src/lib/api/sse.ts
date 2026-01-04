// SSE (Server-Sent Events) client for real-time updates

export interface SSEEvent {
	type: string;
	data: unknown;
}

export interface SSELogCreatedData {
	id: number;
	header: {
		title: string;
		severity: string;
		source: string;
		color: string;
		description?: string;
	};
	body: Record<string, unknown>;
	metadata: {
		derived_severity: string;
		derived_source: string;
		derived_category: string;
	};
	created_at: string;
}

export interface SSELogDeletedData {
	id: number;
}

export interface SSEStatsUpdatedData {
	total: number;
	last_24_hours: number;
	by_severity: Record<string, number>;
	by_source: Record<string, number>;
}

export type SSEEventHandler = {
	onLogCreated?: (data: SSELogCreatedData) => void;
	onLogDeleted?: (data: SSELogDeletedData) => void;
	onStatsUpdated?: (data: SSEStatsUpdatedData) => void;
	onConnected?: () => void;
	onDisconnected?: () => void;
	onError?: (error: Error) => void;
};

export class SSEClient {
	private eventSource: EventSource | null = null;
	private baseUrl: string;
	private handlers: SSEEventHandler = {};
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;
	private reconnectDelay = 1000;

	constructor(baseUrl: string) {
		this.baseUrl = baseUrl.replace(/\/$/, '');
	}

	connect(handlers: SSEEventHandler): void {
		this.handlers = handlers;
		this.doConnect();
	}

	private doConnect(): void {
		if (this.eventSource) {
			this.eventSource.close();
		}

		const url = `${this.baseUrl}/api/events`;
		this.eventSource = new EventSource(url);

		this.eventSource.onopen = () => {
			this.reconnectAttempts = 0;
			this.handlers.onConnected?.();
		};

		this.eventSource.onerror = () => {
			this.handlers.onDisconnected?.();
			this.handleReconnect();
		};

		// Listen for specific event types
		this.eventSource.addEventListener('connected', (event: MessageEvent) => {
			try {
				JSON.parse(event.data);
				this.handlers.onConnected?.();
			} catch {
				// Ignore parse errors
			}
		});

		this.eventSource.addEventListener('log_created', (event: MessageEvent) => {
			try {
				const parsed = JSON.parse(event.data) as SSEEvent;
				this.handlers.onLogCreated?.(parsed.data as SSELogCreatedData);
			} catch {
				// Ignore parse errors
			}
		});

		this.eventSource.addEventListener('log_deleted', (event: MessageEvent) => {
			try {
				const parsed = JSON.parse(event.data) as SSEEvent;
				this.handlers.onLogDeleted?.(parsed.data as SSELogDeletedData);
			} catch {
				// Ignore parse errors
			}
		});

		this.eventSource.addEventListener('stats_updated', (event: MessageEvent) => {
			try {
				const parsed = JSON.parse(event.data) as SSEEvent;
				this.handlers.onStatsUpdated?.(parsed.data as SSEStatsUpdatedData);
			} catch {
				// Ignore parse errors
			}
		});

		// Ping events just keep the connection alive, no action needed
		this.eventSource.addEventListener('ping', () => {
			// Connection is alive
		});
	}

	private handleReconnect(): void {
		if (this.reconnectAttempts >= this.maxReconnectAttempts) {
			this.handlers.onError?.(new Error('Max reconnection attempts reached'));
			return;
		}

		this.reconnectAttempts++;
		const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

		setTimeout(() => {
			this.doConnect();
		}, delay);
	}

	disconnect(): void {
		if (this.eventSource) {
			this.eventSource.close();
			this.eventSource = null;
		}
	}

	isConnected(): boolean {
		return this.eventSource?.readyState === EventSource.OPEN;
	}
}

// Default SSE client instance
let defaultSSEClient: SSEClient | null = null;

export function getSSEClient(): SSEClient {
	if (!defaultSSEClient) {
		const baseUrl = import.meta.env.DEV ? 'http://localhost:8080' : window.location.origin;
		defaultSSEClient = new SSEClient(baseUrl);
	}
	return defaultSSEClient;
}
