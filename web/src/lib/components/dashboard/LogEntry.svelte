<script lang="ts">
	import * as Collapsible from '$lib/components/ui/collapsible/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { ChevronDown } from '@lucide/svelte';
	import type { LogEntry as LogEntryType } from '$lib/stores/logs.svelte.js';

	interface Props {
		log: LogEntryType;
	}

	let { log }: Props = $props();
	let isOpen = $state(false);

	// Severity colors - supports custom severities with default purple
	const severityColors: Record<string, string> = {
		fatal: '#991b1b',
		critical: '#dc2626',
		error: '#ef4444',
		warning: '#f59e0b',
		success: '#22c55e',
		info: '#3b82f6',
		debug: '#6b7280',
		trace: '#9ca3af'
	};

	const defaultColor = '#8b5cf6'; // Purple for custom severities

	const severityColor = $derived(severityColors[log.header.type] ?? defaultColor);

	// List view format: MMM dd, yyyy HH:mm:ss (e.g., "Jan 04, 2026 04:19:16")
	function formatTime(timestamp: string): string {
		const date = new Date(timestamp);
		const month = date.toLocaleString('en-US', { month: 'short' });
		const day = String(date.getDate()).padStart(2, '0');
		const year = date.getFullYear();
		const hours = String(date.getHours()).padStart(2, '0');
		const minutes = String(date.getMinutes()).padStart(2, '0');
		const seconds = String(date.getSeconds()).padStart(2, '0');
		return `${month} ${day}, ${year} ${hours}:${minutes}:${seconds}`;
	}

	// ISO format for JSON display: yyyy-mm-dd HH:mm:ss
	function formatISODateTime(timestamp: string): string {
		const date = new Date(timestamp);
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		const hours = String(date.getHours()).padStart(2, '0');
		const minutes = String(date.getMinutes()).padStart(2, '0');
		const seconds = String(date.getSeconds()).padStart(2, '0');
		return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
	}

	// Syntax highlight JSON with formatted dates
	function highlightJson(obj: Record<string, unknown>): string {
		// Format date fields before stringifying
		const formatted = {
			...obj,
			created_at: typeof obj.created_at === 'string' ? formatISODateTime(obj.created_at) : obj.created_at,
			updated_at: typeof obj.updated_at === 'string' ? formatISODateTime(obj.updated_at) : obj.updated_at
		};
		const json = JSON.stringify(formatted, null, 2);
		return json
			// Keys (must be before strings to avoid conflicts)
			.replace(/"([^"]+)":/g, '<span class="text-purple-400">"$1"</span>:')
			// String values
			.replace(/: "([^"]*)"/g, ': <span class="text-green-400">"$1"</span>')
			// Numbers
			.replace(/: (\d+\.?\d*)/g, ': <span class="text-amber-400">$1</span>')
			// Booleans
			.replace(/: (true|false)/g, ': <span class="text-blue-400">$1</span>')
			// Null
			.replace(/: (null)/g, ': <span class="text-red-400">$1</span>');
	}
</script>

<Collapsible.Root bind:open={isOpen} class="log-entry border-b border-border">
	<Collapsible.Trigger class="w-full cursor-pointer px-4 py-3 text-left hover:bg-muted/50">
		<div class="flex items-start gap-3">
			<!-- Severity Circle -->
			<span
				class="mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full"
				style="background-color: {severityColor}"
			></span>

			<!-- Main Content -->
			<div class="min-w-0 flex-1">
				<!-- First row: Timestamp, Badge, Source -->
				<div class="mb-1 flex items-center gap-2 text-xs text-muted-foreground">
					<span class="font-mono">{formatTime(log.created_at)}</span>
					<Badge
						variant="outline"
						class="shrink-0 px-1.5 py-0 text-[10px] font-medium uppercase"
						style="border-color: {severityColor}; color: {severityColor}"
					>
						{log.header.type}
					</Badge>
					{#if log.header.source}
						<span class="text-muted-foreground/70">{log.header.source}</span>
					{/if}
				</div>

				<!-- Second row: Title -->
				<div class="font-medium">{log.header.title}</div>

				<!-- Third row: Description -->
				{#if log.header.description}
					<p class="mt-0.5 text-sm text-muted-foreground">
						{log.header.description}
					</p>
				{/if}
			</div>

			<!-- Chevron -->
			<ChevronDown
				class="mt-1 h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200 {isOpen ? 'rotate-180' : ''}"
			/>
		</div>
	</Collapsible.Trigger>

	<Collapsible.Content class="border-t border-border/50 bg-muted/30">
		<div class="p-4">
			<!-- Full Log JSON -->
			<pre class="overflow-x-auto rounded-md bg-card p-3 text-sm"><code class="text-foreground">{@html highlightJson(log as unknown as Record<string, unknown>)}</code></pre>
		</div>
	</Collapsible.Content>
</Collapsible.Root>
