<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Sun, Moon, Download, FileCode, FileSpreadsheet, Activity, Clock } from '@lucide/svelte';
	import { toggleMode } from 'mode-watcher';
	import { logsStore } from '$lib/stores/index.js';
	import logo from '$lib/assets/logo.svg';

	// Props
	interface Props {
		isConnected?: boolean;
	}
	let { isConnected = false }: Props = $props();

	// Check if there are logs to export
	const hasLogs = $derived(logsStore.state.logs.length > 0);

	// Stats
	const stats = $derived(logsStore.stats);

	// Current datetime state
	let currentDateTime = $state('');
	let dateTimeInterval: ReturnType<typeof setInterval>;

	function updateDateTime() {
		const now = new Date();
		const month = now.toLocaleString('en-US', { month: 'short' });
		const day = String(now.getDate()).padStart(2, '0');
		const year = now.getFullYear();
		const hours = String(now.getHours()).padStart(2, '0');
		const minutes = String(now.getMinutes()).padStart(2, '0');
		const seconds = String(now.getSeconds()).padStart(2, '0');
		currentDateTime = `${month} ${day}, ${year} ${hours}:${minutes}:${seconds}`;
	}

	onMount(() => {
		updateDateTime();
		dateTimeInterval = setInterval(updateDateTime, 1000);
	});

	onDestroy(() => {
		if (dateTimeInterval) clearInterval(dateTimeInterval);
	});

	function exportToCSV() {
		const logs = logsStore.state.logs;
		if (logs.length === 0) return;

		const headers = ['ID', 'Timestamp', 'Severity', 'Title', 'Source', 'Description'];
		const rows = logs.map(log => [
			log.id,
			log.created_at,
			log.header.type,
			log.header.title,
			log.header.source ?? '',
			log.header.description ?? ''
		]);

		const csv = [headers.join(','), ...rows.map(r => r.map(c => `"${String(c).replace(/"/g, '""')}"`).join(','))].join('\n');
		downloadFile(csv, 'scribe-logs.csv', 'text/csv');
	}

	function exportToJSON() {
		const logs = logsStore.state.logs;
		if (logs.length === 0) return;

		const json = JSON.stringify(logs, null, 2);
		downloadFile(json, 'scribe-logs.json', 'application/json');
	}

	function downloadFile(content: string, filename: string, type: string) {
		const blob = new Blob([content], { type });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = filename;
		a.click();
		URL.revokeObjectURL(url);
	}
</script>

<header class="sticky top-0 z-10 shrink-0 border-b border-border bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
	<!-- Main Header Row -->
	<div class="flex h-12 items-center px-6">
		<!-- Logo and Title (Left) -->
		<div class="flex items-center gap-3">
			<img src={logo} alt="SCRIBE" class="h-7 w-7" />
			<span class="text-lg font-semibold">SCRIBE</span>
		</div>

		<!-- Right side controls -->
		<div class="ml-auto flex items-center gap-2">
			<!-- Export Dropdown -->
			<DropdownMenu.Root>
				<DropdownMenu.Trigger disabled={!hasLogs}>
					<Button variant="ghost" size="icon" class="h-8 w-8" disabled={!hasLogs}>
						<Download class="h-4 w-4" />
						<span class="sr-only">Export</span>
					</Button>
				</DropdownMenu.Trigger>
				<DropdownMenu.Content align="end">
					<DropdownMenu.Item onSelect={exportToCSV}>
						<FileSpreadsheet class="mr-2 h-4 w-4" />
						Export as CSV
					</DropdownMenu.Item>
					<DropdownMenu.Item onSelect={exportToJSON}>
						<FileCode class="mr-2 h-4 w-4" />
						Export as JSON
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>

			<!-- Theme Toggle -->
			<Button variant="ghost" size="icon" class="h-8 w-8" onclick={toggleMode}>
				<Sun class="h-4 w-4 scale-100 rotate-0 transition-all dark:scale-0 dark:-rotate-90" />
				<Moon
					class="absolute h-4 w-4 scale-0 rotate-90 transition-all dark:scale-100 dark:rotate-0"
				/>
				<span class="sr-only">Toggle theme</span>
			</Button>
		</div>
	</div>

	<!-- Status Bar -->
	<div class="flex h-8 items-center gap-6 border-t border-border/50 bg-muted/30 px-6 text-xs">
		<!-- Connection Status -->
		<div class="flex items-center gap-1.5">
			<span class="relative flex h-2 w-2">
				<span
					class="absolute inline-flex h-full w-full animate-ping rounded-full opacity-75 {isConnected ? 'bg-green-500' : 'bg-red-500'}"
				></span>
				<span
					class="relative inline-flex h-2 w-2 rounded-full {isConnected ? 'bg-green-500' : 'bg-red-500'}"
				></span>
			</span>
			<span class="{isConnected ? 'text-green-500' : 'text-red-500'}">{isConnected ? 'Online' : 'Offline'}</span>
		</div>

		<span class="text-border">|</span>

		<!-- Total Logs -->
		<div class="flex items-center gap-1.5">
			<Activity class="h-3.5 w-3.5 text-muted-foreground" />
			<span class="text-muted-foreground">Total Logs:</span>
			<span class="font-semibold">{stats.total.toLocaleString()}</span>
		</div>

		<span class="text-border">|</span>

		<!-- Last 24h -->
		<div class="flex items-center gap-1.5">
			<Clock class="h-3.5 w-3.5 text-muted-foreground" />
			<span class="text-muted-foreground">Last 24h:</span>
			<span class="font-semibold">{stats.recent.toLocaleString()}</span>
		</div>

		<!-- DateTime (Right) -->
		<div class="ml-auto font-mono text-muted-foreground">
			{currentDateTime}
		</div>
	</div>
</header>
