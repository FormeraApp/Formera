export interface PaginationState {
	page: number;
	pageSize: number;
	totalItems: number;
	totalPages: number;
}

export const usePagination = (initialPageSize = 20) => {
	const state = reactive<PaginationState>({
		page: 1,
		pageSize: initialPageSize,
		totalItems: 0,
		totalPages: 0,
	});

	const updateFromResponse = (response: { page: number; page_size: number; total_items: number; total_pages: number }) => {
		state.page = response.page;
		state.pageSize = response.page_size;
		state.totalItems = response.total_items;
		state.totalPages = response.total_pages;
	};

	const setPage = (page: number) => {
		if (page >= 1 && page <= state.totalPages) {
			state.page = page;
		}
	};

	const setPageSize = (size: number) => {
		state.pageSize = size;
		state.page = 1; // Reset to first page when changing page size
	};

	const nextPage = () => {
		if (state.page < state.totalPages) {
			state.page++;
		}
	};

	const prevPage = () => {
		if (state.page > 1) {
			state.page--;
		}
	};

	const firstPage = () => {
		state.page = 1;
	};

	const lastPage = () => {
		state.page = state.totalPages;
	};

	const hasNextPage = computed(() => state.page < state.totalPages);
	const hasPrevPage = computed(() => state.page > 1);

	const params = computed(() => ({
		page: state.page,
		pageSize: state.pageSize,
	}));

	// Range of pages to show (e.g., [1, 2, 3, 4, 5] or [3, 4, 5, 6, 7])
	const visiblePages = computed(() => {
		const total = state.totalPages;
		const current = state.page;
		const maxVisible = 5;

		if (total <= maxVisible) {
			return Array.from({ length: total }, (_, i) => i + 1);
		}

		let start = Math.max(1, current - Math.floor(maxVisible / 2));
		const end = Math.min(total, start + maxVisible - 1);

		if (end - start + 1 < maxVisible) {
			start = Math.max(1, end - maxVisible + 1);
		}

		return Array.from({ length: end - start + 1 }, (_, i) => start + i);
	});

	return {
		state,
		params,
		updateFromResponse,
		setPage,
		setPageSize,
		nextPage,
		prevPage,
		firstPage,
		lastPage,
		hasNextPage,
		hasPrevPage,
		visiblePages,
	};
};
