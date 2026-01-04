<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Search, X, ListFilter, Calendar } from '@lucide/svelte';
	import { uiStore, logsStore } from '$lib/stores/index.js';

	const filters = $derived(uiStore.state.filters);
	const stats = $derived(logsStore.stats);

	// Dynamic severity options from backend data
	const severityOptions = $derived(() => {
		const options = [{ value: '', label: 'All Severities' }];
		const severities = Object.keys(stats.by_severity);

		// Add each severity from the database
		severities.forEach((sev) => {
			options.push({
				value: sev,
				label: sev.charAt(0).toUpperCase() + sev.slice(1)
			});
		});

		return options;
	});

	// Dynamic source options from backend data
	const sourceOptions = $derived(() => {
		const options = [{ value: '', label: 'All Sources' }];
		const sources = Object.keys(stats.by_source);

		// Add each source from the database
		sources.forEach((src) => {
			options.push({
				value: src,
				label: src
			});
		});

		return options;
	});

	// Local state for form inputs - synced from store via $effect
	let searchValue = $state('');
	let severityValue = $state<string | undefined>('');
	let sourceValue = $state<string | undefined>('');
	let dateFromValue = $state('');

	// Sync local state from store when filters change externally
	$effect(() => {
		searchValue = filters.searchQuery;
	});

	$effect(() => {
		severityValue = filters.severityFilter || undefined;
	});

	$effect(() => {
		sourceValue = filters.sourceFilter || undefined;
	});

	$effect(() => {
		dateFromValue = filters.dateFrom ?? '';
	});

	// Track previous values to detect changes from Select
	let prevSeverity = $state<string | undefined>(undefined);
	let prevSource = $state<string | undefined>(undefined);

	// Sync Select changes to store
	$effect(() => {
		if (severityValue !== prevSeverity) {
			prevSeverity = severityValue;
			uiStore.setSeverityFilter(severityValue ?? '');
		}
	});

	$effect(() => {
		if (sourceValue !== prevSource) {
			prevSource = sourceValue;
			uiStore.setSourceFilter(sourceValue ?? '');
		}
	});

	const severityLabel = $derived(
		severityOptions().find((o) => o.value === (severityValue ?? ''))?.label ?? 'All Severities'
	);

	const sourceLabel = $derived(
		sourceOptions().find((o) => o.value === (sourceValue ?? ''))?.label ?? 'All Sources'
	);

	function handleSearch() {
		uiStore.setSearchQuery(searchValue);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleSearch();
		}
	}

	function handleDateChange(event: Event) {
		const target = event.target as HTMLInputElement;
		dateFromValue = target.value;
		uiStore.setDateRange(target.value || null, null);
	}

	function clearFilters() {
		searchValue = '';
		severityValue = undefined;
		sourceValue = undefined;
		dateFromValue = '';
		uiStore.clearFilters();
	}

	const hasActiveFilters = $derived(
		filters.searchQuery !== '' ||
			filters.severityFilter !== '' ||
			filters.sourceFilter !== '' ||
			filters.dateFrom !== null
	);
</script>

<div class="flex flex-col gap-4 lg:flex-row lg:items-center">
	<!-- Search Input -->
	<div class="relative flex-1">
		<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
		<Input
			placeholder="Search logs by title, description, or body..."
			class="pl-9 pr-9"
			bind:value={searchValue}
			onkeydown={handleKeydown}
		/>
		{#if searchValue}
			<Button
				variant="ghost"
				size="icon"
				class="absolute right-1 top-1/2 h-7 w-7 -translate-y-1/2"
				onclick={() => {
					searchValue = '';
					uiStore.setSearchQuery('');
				}}
			>
				<X class="h-4 w-4" />
			</Button>
		{/if}
	</div>

	<!-- Filters -->
	<div class="flex flex-wrap items-center gap-2">
		<div class="flex items-center gap-1 text-sm text-muted-foreground">
			<ListFilter class="h-4 w-4" />
			<span class="hidden sm:inline">Filters:</span>
		</div>

		<Select.Root type="single" bind:value={severityValue}>
			<Select.Trigger class="h-8 w-36">
				{severityLabel}
			</Select.Trigger>
			<Select.Content>
				{#each severityOptions() as option}
					<Select.Item value={option.value}>{option.label}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>

		<Select.Root type="single" bind:value={sourceValue}>
			<Select.Trigger class="h-8 w-36">
				{sourceLabel}
			</Select.Trigger>
			<Select.Content>
				{#each sourceOptions() as option}
					<Select.Item value={option.value}>{option.label}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>

		<!-- Date Filter -->
		<div class="relative">
			<Calendar class="absolute left-2 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground pointer-events-none" />
			<input
				type="date"
				class="h-8 w-36 rounded-md border border-input bg-background pl-8 pr-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
				value={dateFromValue}
				onchange={handleDateChange}
				title="Filter by date"
			/>
		</div>

		{#if hasActiveFilters}
			<Button variant="ghost" size="sm" class="h-8" onclick={clearFilters}>
				<X class="mr-1 h-3 w-3" />
				Clear
			</Button>
		{/if}
	</div>
</div>
