<script lang="ts" setup>
const { t } = useI18n();

const props = defineProps<{
	page: number;
	pageSize: number;
	totalItems: number;
	totalPages: number;
	visiblePages: number[];
	hasNextPage: boolean;
	hasPrevPage: boolean;
	pageSizeOptions?: number[];
}>();

const emit = defineEmits<{
	(e: "update:page", page: number): void;
	(e: "update:pageSize", size: number): void;
	(e: "first"): void;
	(e: "prev"): void;
	(e: "next"): void;
	(e: "last"): void;
}>();

const pageSizes = computed(() => props.pageSizeOptions || [10, 20, 50, 100]);

const startItem = computed(() => (props.page - 1) * props.pageSize + 1);
const endItem = computed(() => Math.min(props.page * props.pageSize, props.totalItems));
</script>

<template>
	<div class="pagination">
		<div class="pagination-info">
			<span>{{ t("pagination.showing", { start: startItem, end: endItem, total: totalItems }) }}</span>
		</div>

		<div class="pagination-controls">
			<button
				class="pagination-btn"
				:disabled="!hasPrevPage"
				:title="t('pagination.first')"
				@click="emit('first')"
			>
				<UISysIcon icon="fa-solid fa-angles-left" />
			</button>
			<button
				class="pagination-btn"
				:disabled="!hasPrevPage"
				:title="t('pagination.previous')"
				@click="emit('prev')"
			>
				<UISysIcon icon="fa-solid fa-chevron-left" />
			</button>

			<div class="pagination-pages">
				<button
					v-for="pageNum in visiblePages"
					:key="pageNum"
					:class="['pagination-page', { 'pagination-page-active': pageNum === page }]"
					@click="emit('update:page', pageNum)"
				>
					{{ pageNum }}
				</button>
			</div>

			<button
				class="pagination-btn"
				:disabled="!hasNextPage"
				:title="t('pagination.next')"
				@click="emit('next')"
			>
				<UISysIcon icon="fa-solid fa-chevron-right" />
			</button>
			<button
				class="pagination-btn"
				:disabled="!hasNextPage"
				:title="t('pagination.last')"
				@click="emit('last')"
			>
				<UISysIcon icon="fa-solid fa-angles-right" />
			</button>
		</div>

		<div class="pagination-size">
			<label>{{ t("pagination.perPage") }}</label>
			<select
				:value="pageSize"
				class="pagination-select"
				@change="emit('update:pageSize', Number(($event.target as HTMLSelectElement).value))"
			>
				<option v-for="size in pageSizes" :key="size" :value="size">
					{{ size }}
				</option>
			</select>
		</div>
	</div>
</template>

<style scoped>
.pagination {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
	padding: 1rem;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	flex-wrap: wrap;
}

.pagination-info {
	font-size: 0.875rem;
	color: var(--text-secondary);
}

.pagination-controls {
	display: flex;
	align-items: center;
	gap: 0.25rem;
}

.pagination-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 36px;
	height: 36px;
	color: var(--text-secondary);
	background: var(--surface-hover);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s ease;
}

.pagination-btn:hover:not(:disabled) {
	color: var(--text);
	background: var(--surface);
	border-color: var(--primary);
}

.pagination-btn:disabled {
	opacity: 0.4;
	cursor: not-allowed;
}

.pagination-pages {
	display: flex;
	gap: 0.25rem;
	margin: 0 0.5rem;
}

.pagination-page {
	display: flex;
	align-items: center;
	justify-content: center;
	min-width: 36px;
	height: 36px;
	padding: 0 0.5rem;
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text);
	background: var(--surface-hover);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s ease;
}

.pagination-page:hover {
	background: var(--surface);
	border-color: var(--primary);
}

.pagination-page-active {
	color: white;
	background: var(--primary);
	border-color: var(--primary);
}

.pagination-page-active:hover {
	background: var(--primary-dark);
}

.pagination-size {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.pagination-size label {
	font-size: 0.875rem;
	color: var(--text-secondary);
}

.pagination-select {
	padding: 0.5rem 2rem 0.5rem 0.75rem;
	font-size: 0.875rem;
	color: var(--text);
	background: var(--surface-hover);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	cursor: pointer;
	appearance: none;
	background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
	background-position: right 0.5rem center;
	background-repeat: no-repeat;
	background-size: 1.25em 1.25em;
}

.pagination-select:focus {
	outline: none;
	border-color: var(--primary);
}

@media (max-width: 768px) {
	.pagination {
		flex-direction: column;
		gap: 0.75rem;
	}

	.pagination-info,
	.pagination-size {
		width: 100%;
		justify-content: center;
	}

	.pagination-pages {
		display: none;
	}
}
</style>
