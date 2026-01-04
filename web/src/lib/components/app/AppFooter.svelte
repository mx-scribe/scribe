<script lang="ts">
	import { logsStore } from '$lib/stores/index.js';

	const version = $derived(logsStore.state.version || '1.0.0');
</script>

<footer class="flex h-10 shrink-0 items-center justify-between border-t border-border bg-background px-6 text-xs text-muted-foreground">
	<span>SCRIBE v{version}</span>
	<span>
		Built with ❤️ by <a href="https://mendex.io" target="_blank" rel="noopener noreferrer" class="font-medium text-foreground hover:underline">Mendex</a>
	</span>
</footer>
