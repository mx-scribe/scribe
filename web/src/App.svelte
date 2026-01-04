<script lang="ts">
	import { ModeWatcher } from 'mode-watcher';
	import { onMount, onDestroy } from 'svelte';
	import { Card } from '$lib/components/ui/card/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { LogEntry, SearchFilters, StatsOverview } from '$lib/components/dashboard/index.js';
	import { AppHeader, AppFooter } from '$lib/components/app/index.js';
	import { logsStore, uiStore } from '$lib/stores/index.js';

	// Connection state
	let isBackendConnected = $state(false);

	// Build API filters from UI state
	function buildApiFilters() {
		const dateFrom = uiStore.state.filters.dateFrom;
		// Calculate next day for "to" filter to include all logs from selected day
		let dateTo: string | undefined;
		if (dateFrom) {
			const nextDay = new Date(dateFrom);
			nextDay.setDate(nextDay.getDate() + 1);
			dateTo = nextDay.toISOString().split('T')[0];
		}
		return {
			limit: 10000, // Load all, paginate on frontend
			severity: uiStore.state.filters.severityFilter || undefined,
			source: uiStore.state.filters.sourceFilter || undefined,
			search: uiStore.state.filters.searchQuery || undefined,
			from: dateFrom || undefined,
			to: dateTo
		};
	}

	// Initialize data on mount
	onMount(async () => {
		// Try to connect to backend
		const connected = await logsStore.checkBackend();
		isBackendConnected = connected;

		if (connected) {
			// Fetch logs with filters applied on backend
			await Promise.all([logsStore.fetchLogs(buildApiFilters()), logsStore.fetchStats()]);

			// Connect to SSE for real-time updates
			logsStore.connectSSE();
		}
	});

	// Cleanup on unmount
	onDestroy(() => {
		logsStore.disconnectSSE();
	});

	// Reactive state
	const logs = $derived(logsStore.state.logs);
	const filters = $derived(uiStore.state.filters);
	const currentPage = $derived(uiStore.state.currentPage);
	const itemsPerPage = $derived(uiStore.state.itemsPerPage);

	// Refetch logs when filters change (backend filtering)
	let prevFilters = $state('');
	$effect(() => {
		const currentFilters = JSON.stringify({
			severity: filters.severityFilter,
			source: filters.sourceFilter,
			search: filters.searchQuery,
			dateFrom: filters.dateFrom
		});

		if (prevFilters !== '' && currentFilters !== prevFilters && isBackendConnected) {
			logsStore.fetchLogs(buildApiFilters());
		}
		prevFilters = currentFilters;
	});

	// Paginated logs (pagination on frontend)
	const paginatedLogs = $derived(() => {
		const start = (currentPage - 1) * itemsPerPage;
		const end = start + itemsPerPage;
		return logs.slice(start, end);
	});

	const totalPages = $derived(() => Math.ceil(logs.length / itemsPerPage));

	// Entries per page options
	const entriesOptions = [
		{ value: '10', label: '10' },
		{ value: '25', label: '25' },
		{ value: '50', label: '50' },
		{ value: '100', label: '100' }
	];

	// Local state for entries per page select
	let entriesPerPageValue = $state<string>('10');

	// Sync when itemsPerPage changes externally
	$effect(() => {
		entriesPerPageValue = String(uiStore.state.itemsPerPage);
	});

	// Track previous value to detect changes
	let prevEntriesValue = $state<string>('10');

	$effect(() => {
		if (entriesPerPageValue !== prevEntriesValue) {
			prevEntriesValue = entriesPerPageValue;
			uiStore.setItemsPerPage(Number(entriesPerPageValue));
		}
	});
</script>

<ModeWatcher defaultMode="dark" />

<div class="flex min-h-screen flex-col bg-background">
	<AppHeader isConnected={isBackendConnected} />

	<main class="flex-1 overflow-auto p-6">

		<!-- Stats Overview -->
		<StatsOverview />

		<!-- Search and Filters -->
		<div class="mb-6">
			<SearchFilters />
		</div>

		<!-- Logs List -->
		<Card>
			<div class="flex items-center justify-between border-b border-border p-4">
				<h2 class="font-semibold">Recent Logs</h2>
				<span class="text-sm text-muted-foreground">
					Showing {paginatedLogs().length} of {logs.length} logs
				</span>
			</div>
			<div>
				{#each paginatedLogs() as log (log.id)}
					<LogEntry {log} />
				{:else}
					<div class="flex h-32 items-center justify-center text-muted-foreground">
						No logs found matching your filters
					</div>
				{/each}
			</div>
			<!-- Pagination Footer -->
			<div class="flex items-center justify-between border-t border-border p-4">
				<!-- Entries per page selector -->
				<div class="flex items-center gap-2 text-sm">
					<span class="text-muted-foreground">Show</span>
					<Select.Root type="single" bind:value={entriesPerPageValue}>
						<Select.Trigger class="h-8 w-20">
							{entriesPerPageValue}
						</Select.Trigger>
						<Select.Content>
							{#each entriesOptions as option}
								<Select.Item value={option.value}>{option.label}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
					<span class="text-muted-foreground">entries</span>
				</div>

				<!-- Page info -->
				<span class="text-sm text-muted-foreground">
					Page {currentPage} of {totalPages() || 1}
				</span>

				<!-- Navigation buttons -->
				<div class="flex gap-2">
					<button
						class="rounded-md border px-3 py-1 text-sm hover:bg-muted disabled:opacity-50"
						disabled={currentPage === 1}
						onclick={() => uiStore.prevPage()}
					>
						Previous
					</button>
					<button
						class="rounded-md border px-3 py-1 text-sm hover:bg-muted disabled:opacity-50"
						disabled={currentPage >= totalPages() || totalPages() === 0}
						onclick={() => uiStore.nextPage()}
					>
						Next
					</button>
				</div>
			</div>
		</Card>
	</main>

	<AppFooter />
</div>
