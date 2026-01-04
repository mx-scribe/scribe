// UI Store - Global UI state management using Svelte 5 runes

interface Filters {
	searchQuery: string;
	severityFilter: string;
	sourceFilter: string;
	dateFrom: string | null;
	dateTo: string | null;
}

interface UIState {
	// Filters
	filters: Filters;

	// Pagination
	currentPage: number;
	itemsPerPage: number;
}

const defaultFilters: Filters = {
	searchQuery: '',
	severityFilter: '',
	sourceFilter: '',
	dateFrom: null,
	dateTo: null
};

function createUIStore() {
	let state = $state<UIState>({
		filters: { ...defaultFilters },
		currentPage: 1,
		itemsPerPage: 10
	});

	return {
		get state() {
			return state;
		},

		// Filter actions
		setSearchQuery(query: string) {
			state.filters.searchQuery = query;
			state.currentPage = 1;
		},
		setSeverityFilter(severity: string) {
			state.filters.severityFilter = severity;
			state.currentPage = 1;
		},
		setSourceFilter(source: string) {
			state.filters.sourceFilter = source;
			state.currentPage = 1;
		},
		setDateRange(from: string | null, to: string | null) {
			state.filters.dateFrom = from;
			state.filters.dateTo = to;
			state.currentPage = 1;
		},
		clearFilters() {
			state.filters = { ...defaultFilters };
			state.currentPage = 1;
		},

		// Pagination actions
		setPage(page: number) {
			state.currentPage = page;
		},
		nextPage() {
			state.currentPage++;
		},
		prevPage() {
			if (state.currentPage > 1) state.currentPage--;
		},
		setItemsPerPage(count: number) {
			state.itemsPerPage = count;
			state.currentPage = 1;
		}
	};
}

export const uiStore = createUIStore();
