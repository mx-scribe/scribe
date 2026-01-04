<script lang="ts">
	import { Card } from '$lib/components/ui/card/index.js';
	import { logsStore } from '$lib/stores/index.js';
	import { ChevronDown, Sparkles } from '@lucide/svelte';

	const stats = $derived(logsStore.stats);

	// Expandable state
	let isExpanded = $state(false);

	// Known severity colors (ordered from least severe to most severe)
	const severityColors: Record<string, { color: string; order: number }> = {
		debug: { color: '#6b7280', order: 0 },
		info: { color: '#3b82f6', order: 1 },
		success: { color: '#22c55e', order: 2 },
		warning: { color: '#f59e0b', order: 3 },
		error: { color: '#ef4444', order: 4 },
		critical: { color: '#dc2626', order: 5 }
	};

	// Default color for custom severities
	const defaultColor = '#8b5cf6';

	// Get severity data sorted by order
	const severityData = $derived(() => {
		const bySeverity = stats.by_severity;
		const total = stats.total;

		const severityKeys = Object.keys(bySeverity);
		if (severityKeys.length === 0) return [];

		return severityKeys
			.map((key) => {
				const known = severityColors[key];
				const count = bySeverity[key] || 0;
				return {
					key,
					label: key.charAt(0).toUpperCase() + key.slice(1),
					color: known?.color || defaultColor,
					order: known?.order ?? 10,
					count,
					percentage: total > 0 ? Math.round((count / total) * 100) : 0
				};
			})
			.sort((a, b) => a.order - b.order);
	});
</script>

<!-- Severity Breakdown - Compact Bar -->
{#if severityData().length > 0}
	<Card class="mb-6 overflow-hidden">
		<!-- Clickable Header -->
		<button
			class="flex w-full flex-col gap-3 p-4 text-left transition-colors hover:bg-muted/50"
			onclick={() => (isExpanded = !isExpanded)}
		>
			<!-- Title Row -->
			<div class="flex w-full items-center justify-between">
				<div class="flex items-center gap-2">
					<Sparkles class="h-4 w-4 text-blue-500" />
					<span class="text-sm font-medium text-muted-foreground">Severity Breakdown</span>
				</div>
				<ChevronDown
					class="h-4 w-4 text-muted-foreground transition-transform {isExpanded ? 'rotate-180' : ''}"
				/>
			</div>

			<!-- Stacked Bar -->
			<div class="flex h-3 w-full overflow-hidden rounded-full bg-muted">
				{#each severityData() as item}
					{#if item.percentage > 0}
						<div
							class="h-full transition-all"
							style="width: {item.percentage}%; background-color: {item.color}"
							title="{item.label}: {item.count} ({item.percentage}%)"
						></div>
					{/if}
				{/each}
			</div>
		</button>

		<!-- Expanded Details -->
		{#if isExpanded}
			<div class="border-t border-border px-4 py-3">
				<div class="space-y-2">
					{#each severityData() as item}
						<div class="flex items-center gap-3">
							<span class="h-3 w-3 shrink-0 rounded-full" style="background-color: {item.color}"></span>
							<span class="w-20 shrink-0 text-sm font-medium">{item.label}</span>
							<div class="h-2 flex-1 overflow-hidden rounded-full bg-muted">
								<div
									class="h-full rounded-full transition-all"
									style="width: {item.percentage}%; background-color: {item.color}"
								></div>
							</div>
							<span class="w-10 shrink-0 text-right text-sm font-bold" style="color: {item.color}">{item.count}</span>
							<span class="w-12 shrink-0 text-right text-xs text-muted-foreground">{item.percentage}%</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	</Card>
{/if}
