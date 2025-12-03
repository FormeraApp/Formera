<script lang="ts" setup>
const { t, locale } = useI18n();
const { formsApi, submissionsApi } = useApi();
const router = useRouter();

const forms = ref<Form[]>([]);
const formStats = ref<Record<string, number>>({});
const isLoading = ref(true);
const activeMenu = ref<string | null>(null);
const searchQuery = ref("");
const filterStatus = ref<"all" | "draft" | "published" | "closed">("all");
const sortBy = ref<"updated" | "created" | "title">("updated");

const loadForms = async () => {
	try {
		const data = await formsApi.list();
		forms.value = data || [];
		isLoading.value = false;

		// Load stats for all forms in parallel (non-blocking)
		await Promise.all(
			forms.value.map(async (form) => {
				try {
					const stats = await submissionsApi.stats(form.id);
					formStats.value[form.id] = stats.total_submissions;
				} catch {
					formStats.value[form.id] = 0;
				}
			})
		);
	} catch (error) {
		console.error("Failed to load forms:", error);
		isLoading.value = false;
	}
};

const filteredForms = computed(() => {
	let result = [...forms.value];

	// Filter by search query
	if (searchQuery.value) {
		const query = searchQuery.value.toLowerCase();
		result = result.filter((form) => form.title.toLowerCase().includes(query) || form.description?.toLowerCase().includes(query));
	}

	// Filter by status
	if (filterStatus.value !== "all") {
		result = result.filter((form) => form.status === filterStatus.value);
	}

	// Sort
	result.sort((a, b) => {
		switch (sortBy.value) {
			case "title":
				return a.title.localeCompare(b.title);
			case "created":
				return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
			default:
				return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
		}
	});

	return result;
});

const stats = computed(() => ({
	total: forms.value.length,
	published: forms.value.filter((f) => f.status === "published").length,
	draft: forms.value.filter((f) => f.status === "draft").length,
	totalResponses: Object.values(formStats.value).reduce((sum, count) => sum + count, 0),
}));

const handleCreateForm = async () => {
	try {
		const newForm = await formsApi.create({
			title: t("forms.defaults.title"),
			description: "",
			fields: [],
			settings: {
				submit_button_text: t("forms.defaults.submitButton"),
				success_message: t("forms.defaults.successMessage"),
				allow_multiple: true,
				require_login: false,
				notify_on_submission: false,
			},
		});
		router.push(`/forms/${newForm.id}/edit`);
	} catch (error) {
		console.error("Failed to create form:", error);
	}
};

const handleDuplicate = async (id: string) => {
	try {
		const duplicated = await formsApi.duplicate(id);
		forms.value = [duplicated, ...forms.value];
		formStats.value[duplicated.id] = 0;
		activeMenu.value = null;
	} catch (error) {
		console.error("Failed to duplicate form:", error);
	}
};

const handleDelete = async (id: string) => {
	if (!confirm(t("forms.confirmDelete"))) {
		return;
	}
	try {
		await formsApi.delete(id);
		forms.value = forms.value.filter((f) => f.id !== id);
		delete formStats.value[id];
		activeMenu.value = null;
	} catch (error) {
		console.error("Failed to delete form:", error);
	}
};

const getStatusInfo = (status: string) => {
	switch (status) {
		case "published":
			return { class: "status-published", text: t("forms.status.published"), icon: "fa-solid fa-globe" };
		case "closed":
			return { class: "status-closed", text: t("forms.status.closed"), icon: "fa-solid fa-lock" };
		default:
			return { class: "status-draft", text: t("forms.status.draft"), icon: "fa-solid fa-file-pen" };
	}
};

const formatDate = (dateString: string) => {
	const date = new Date(dateString);
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

	if (diffDays === 0) {
		return t("forms.date.today");
	} else if (diffDays === 1) {
		return t("forms.date.yesterday");
	} else if (diffDays < 7) {
		return t("forms.date.daysAgo", { count: diffDays });
	} else {
		return date.toLocaleDateString(locale.value === "de" ? "de-DE" : "en-US", { day: "2-digit", month: "short", year: "numeric" });
	}
};

const copyFormLink = async (formId: string) => {
	const url = `${window.location.origin}/f/${formId}`;
	await navigator.clipboard.writeText(url);
	activeMenu.value = null;
};

const handleClickOutside = () => {
	activeMenu.value = null;
};

onMounted(() => {
	loadForms();
});
</script>

<template>
	<div class="forms-page" @click="handleClickOutside">
		<div v-if="isLoading" class="loading">
			<div class="loading-spinner" />
			<p>{{ $t("forms.loading") }}</p>
		</div>

		<template v-else>
			<!-- Page Header -->
			<header class="page-header">
				<div class="header-content">
					<div class="header-title">
						<h1>{{ $t("forms.title") }}</h1>
						<p class="header-subtitle">{{ $t("forms.subtitle") }}</p>
					</div>
					<button class="btn btn-primary" @click="handleCreateForm">
						<UISysIcon icon="fa-solid fa-plus" />
						<span>{{ $t("forms.newForm") }}</span>
					</button>
				</div>

				<!-- Stats Cards -->
				<div v-if="forms.length > 0" class="stats-row">
					<div class="stat-card">
						<div class="stat-icon stat-icon-total">
							<UISysIcon icon="fa-solid fa-file-lines" />
						</div>
						<div class="stat-content">
							<span class="stat-value">{{ stats.total }}</span>
							<span class="stat-label">{{ $t("forms.stats.total") }}</span>
						</div>
					</div>
					<div class="stat-card">
						<div class="stat-icon stat-icon-published">
							<UISysIcon icon="fa-solid fa-globe" />
						</div>
						<div class="stat-content">
							<span class="stat-value">{{ stats.published }}</span>
							<span class="stat-label">{{ $t("forms.stats.published") }}</span>
						</div>
					</div>
					<div class="stat-card">
						<div class="stat-icon stat-icon-draft">
							<UISysIcon icon="fa-solid fa-file-pen" />
						</div>
						<div class="stat-content">
							<span class="stat-value">{{ stats.draft }}</span>
							<span class="stat-label">{{ $t("forms.stats.drafts") }}</span>
						</div>
					</div>
					<div class="stat-card">
						<div class="stat-icon stat-icon-responses">
							<UISysIcon icon="fa-solid fa-chart-column" />
						</div>
						<div class="stat-content">
							<span class="stat-value">{{ stats.totalResponses }}</span>
							<span class="stat-label">{{ $t("forms.stats.responses") }}</span>
						</div>
					</div>
				</div>
			</header>

			<!-- Empty State -->
			<div v-if="forms.length === 0" class="empty-state">
				<div class="empty-icon">
					<UISysIcon icon="fa-solid fa-file-circle-plus" />
				</div>
				<h2>{{ $t("forms.empty.title") }}</h2>
				<p>{{ $t("forms.empty.description") }}</p>
				<button class="btn btn-primary btn-lg" @click="handleCreateForm">
					<UISysIcon icon="fa-solid fa-plus" />
					{{ $t("forms.empty.button") }}
				</button>
			</div>

			<!-- Forms List -->
			<template v-else>
				<!-- Toolbar -->
				<div class="toolbar">
					<div class="search-box">
						<UISysIcon icon="fa-solid fa-search" class="search-icon" />
						<input
							v-model="searchQuery"
							type="text"
							class="search-input"
							:placeholder="$t('forms.search.placeholder')"
						/>
					</div>
					<div class="toolbar-actions">
						<select v-model="filterStatus" class="select-input">
							<option value="all">{{ $t("forms.filter.allStatus") }}</option>
							<option value="published">{{ $t("forms.filter.published") }}</option>
							<option value="draft">{{ $t("forms.filter.draft") }}</option>
							<option value="closed">{{ $t("forms.filter.closed") }}</option>
						</select>
						<select v-model="sortBy" class="select-input">
							<option value="updated">{{ $t("forms.sort.lastEdited") }}</option>
							<option value="created">{{ $t("forms.sort.createdAt") }}</option>
							<option value="title">{{ $t("forms.sort.alphabetical") }}</option>
						</select>
					</div>
				</div>

				<!-- Results Info -->
				<div v-if="searchQuery || filterStatus !== 'all'" class="results-info">
					<span>{{ $t("forms.search.results", { count: filteredForms.length }, filteredForms.length) }}</span>
					<button v-if="searchQuery || filterStatus !== 'all'" class="clear-filters" @click="searchQuery = ''; filterStatus = 'all'">
						{{ $t("forms.search.clearFilters") }}
					</button>
				</div>

				<!-- Forms Grid -->
				<div class="forms-grid">
					<article v-for="form in filteredForms" :key="form.id" class="form-card">
						<div class="card-main">
							<div class="card-header">
								<div :class="['status-indicator', getStatusInfo(form.status).class]">
									<UISysIcon :icon="getStatusInfo(form.status).icon" />
									<span>{{ getStatusInfo(form.status).text }}</span>
								</div>
								<div class="card-menu">
									<button
										class="menu-trigger"
										@click.stop="activeMenu = activeMenu === form.id ? null : form.id"
									>
										<UISysIcon icon="fa-solid fa-ellipsis" />
									</button>
									<Transition name="dropdown">
										<div v-if="activeMenu === form.id" class="menu-dropdown" @click.stop>
											<NuxtLink :to="`/forms/${form.id}/edit`" class="menu-item">
												<UISysIcon icon="fa-solid fa-pen" />
												<span>{{ $t("forms.menu.edit") }}</span>
											</NuxtLink>
											<NuxtLink :to="`/forms/${form.id}/responses`" class="menu-item">
												<UISysIcon icon="fa-solid fa-chart-column" />
												<span>{{ $t("forms.menu.responses") }}</span>
											</NuxtLink>
											<button
												v-if="form.status === 'published'"
												class="menu-item"
												@click="copyFormLink(form.id)"
											>
												<UISysIcon icon="fa-solid fa-link" />
												<span>{{ $t("forms.menu.copyLink") }}</span>
											</button>
											<a
												v-if="form.status === 'published'"
												:href="`/f/${form.id}`"
												class="menu-item"
												target="_blank"
												rel="noopener"
											>
												<UISysIcon icon="fa-solid fa-arrow-up-right-from-square" />
												<span>{{ $t("forms.menu.open") }}</span>
											</a>
											<hr class="menu-divider" />
											<button class="menu-item" @click="handleDuplicate(form.id)">
												<UISysIcon icon="fa-solid fa-copy" />
												<span>{{ $t("forms.menu.duplicate") }}</span>
											</button>
											<button class="menu-item menu-item-danger" @click="handleDelete(form.id)">
												<UISysIcon icon="fa-solid fa-trash" />
												<span>{{ $t("forms.menu.delete") }}</span>
											</button>
										</div>
									</Transition>
								</div>
							</div>

							<NuxtLink :to="`/forms/${form.id}/edit`" class="card-body">
								<h3 class="form-title">{{ form.title || $t("forms.card.untitled") }}</h3>
								<p class="form-description">{{ form.description || $t("forms.card.noDescription") }}</p>
							</NuxtLink>

							<div class="card-meta">
								<div class="meta-item">
									<UISysIcon icon="fa-solid fa-layer-group" />
									<span>{{ $t("forms.card.fields", { count: form.fields?.length || 0 }) }}</span>
								</div>
								<div class="meta-item">
									<UISysIcon icon="fa-solid fa-inbox" />
									<span>{{ $t("forms.card.responses", { count: formStats[form.id] || 0 }) }}</span>
								</div>
							</div>
						</div>

						<div class="card-footer">
							<span class="update-time">
								<UISysIcon icon="fa-solid fa-clock" />
								{{ formatDate(form.updated_at) }}
							</span>
							<div class="quick-actions">
								<NuxtLink :to="`/forms/${form.id}/edit`" class="action-btn" :title="$t('forms.menu.edit')">
									<UISysIcon icon="fa-solid fa-pen" />
								</NuxtLink>
								<NuxtLink :to="`/forms/${form.id}/responses`" class="action-btn" :title="$t('forms.menu.responses')">
									<UISysIcon icon="fa-solid fa-chart-column" />
								</NuxtLink>
								<a
									v-if="form.status === 'published'"
									:href="`/f/${form.id}`"
									class="action-btn"
									target="_blank"
									rel="noopener"
									:title="$t('forms.menu.open')"
								>
									<UISysIcon icon="fa-solid fa-arrow-up-right-from-square" />
								</a>
							</div>
						</div>
					</article>
				</div>

				<!-- No Results -->
				<div v-if="filteredForms.length === 0 && forms.length > 0" class="no-results">
					<UISysIcon icon="fa-solid fa-search" />
					<h3>{{ $t("forms.search.noResults") }}</h3>
					<p>{{ $t("forms.search.noResultsHint") }}</p>
				</div>
			</template>
		</template>
	</div>
</template>

<style scoped>
.forms-page {
	max-width: 1200px;
	margin: 0 auto;
}

/* Loading State */
.loading {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 1rem;
	min-height: 400px;
	color: var(--text-secondary);
}

.loading-spinner {
	width: 40px;
	height: 40px;
	border: 3px solid var(--border);
	border-top-color: var(--primary);
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

@keyframes spin {
	to { transform: rotate(360deg); }
}

/* Page Header */
.page-header {
	margin-bottom: 1rem;
}

.header-content {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 1rem;
	margin-bottom: 1.5rem;
}

.header-title h1 {
	font-size: 1.75rem;
	font-weight: 700;
	color: var(--text);
	margin-bottom: 0.25rem;
}

.header-subtitle {
	font-size: 0.9375rem;
	color: var(--text-secondary);
}

/* Stats Row */
.stats-row {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: 1rem;
}

.stat-card {
	display: flex;
	align-items: center;
	gap: 1rem;
	padding: 1rem 1.25rem;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
}

.stat-icon {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 44px;
	height: 44px;
	border-radius: var(--radius);
	font-size: 1.125rem;
}

.stat-icon-total {
	background: rgba(99, 102, 241, 0.1);
	color: var(--primary);
}

.stat-icon-published {
	background: rgba(34, 197, 94, 0.1);
	color: var(--success);
}

.stat-icon-draft {
	background: rgba(245, 158, 11, 0.1);
	color: var(--warning);
}

.stat-icon-responses {
	background: rgba(59, 130, 246, 0.1);
	color: #3b82f6;
}

.stat-content {
	display: flex;
	flex-direction: column;
}

.stat-value {
	font-size: 1.5rem;
	font-weight: 700;
	color: var(--text);
	line-height: 1.2;
}

.stat-label {
	font-size: 0.8125rem;
	color: var(--text-secondary);
}

/* Empty State */
.empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	text-align: center;
	padding: 4rem 2rem;
	background: var(--surface);
	border: 2px dashed var(--border);
	border-radius: var(--radius-lg);
}

.empty-icon {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 80px;
	height: 80px;
	margin-bottom: 1.5rem;
	font-size: 2rem;
	color: var(--primary);
	background: rgba(99, 102, 241, 0.1);
	border-radius: 50%;
}

.empty-state h2 {
	font-size: 1.25rem;
	font-weight: 600;
	color: var(--text);
	margin-bottom: 0.5rem;
}

.empty-state p {
	font-size: 0.9375rem;
	color: var(--text-secondary);
	margin-bottom: 1.5rem;
	max-width: 320px;
}

/* Toolbar */
.toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
	margin-bottom: 1rem;
	padding: 1rem;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
}

.search-box {
	position: relative;
	flex: 1;
	max-width: 400px;
}

.search-box :deep(.search-icon) {
	position: absolute;
	left: 0.875rem;
	top: 50%;
	transform: translateY(-50%);
	color: var(--text-secondary);
	font-size: 0.875rem;
	pointer-events: none;
	z-index: 1;
}

.search-input {
	width: 100%;
	padding: 0.625rem 0.875rem 0.625rem 2.5rem;
	font-size: 0.875rem;
	color: var(--text);
	background: var(--background);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.search-input:focus {
	outline: none;
	border-color: var(--primary);
	box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.toolbar-actions {
	display: flex;
	gap: 0.5rem;
}

.select-input {
	padding: 0.625rem 2rem 0.625rem 0.875rem;
	font-size: 0.875rem;
	color: var(--text);
	background: var(--background);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	cursor: pointer;
	appearance: none;
	background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
	background-position: right 0.5rem center;
	background-repeat: no-repeat;
	background-size: 1.25em 1.25em;
}

.select-input:focus {
	outline: none;
	border-color: var(--primary);
}

/* Results Info */
.results-info {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.5rem 0;
	margin-bottom: 1rem;
	font-size: 0.875rem;
	color: var(--text-secondary);
}

.clear-filters {
	font-size: 0.8125rem;
	color: var(--primary);
	background: none;
	border: none;
	cursor: pointer;
	text-decoration: underline;
}

.clear-filters:hover {
	color: var(--primary-dark);
}

/* Forms Grid */
.forms-grid {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
	gap: 1.25rem;
}

/* Form Card */
.form-card {
	display: flex;
	flex-direction: column;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	overflow: hidden;
	transition: border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
}

.form-card:hover {
	border-color: var(--primary-light);
	box-shadow: var(--shadow-lg);
	transform: translateY(-2px);
}

.card-main {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
	padding: 1rem;
	flex: 1;
}

.card-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 0 0.75rem 0; 
	margin: 0;
}

/* Status Indicator */
.status-indicator {
	display: inline-flex;
	align-items: center;
	gap: 0.25rem;
	padding: 0.1875rem 0.5rem;
	font-size: 0.6875rem;
	font-weight: 500;
	border-radius: var(--radius);
}

.status-published {
	background: rgba(34, 197, 94, 0.1);
	color: var(--success);
}

.status-draft {
	background: rgba(245, 158, 11, 0.1);
	color: var(--warning);
}

.status-closed {
	background: rgba(239, 68, 68, 0.1);
	color: var(--error);
}

/* Card Menu */
.card-menu {
	position: relative;
}

.menu-trigger {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 28px;
	height: 28px;
	color: var(--text-secondary);
	background: none;
	border: none;
	border-radius: var(--radius);
	cursor: pointer;
	transition: background-color 0.15s ease, color 0.15s ease;
}

.menu-trigger:hover {
	color: var(--text);
	background: var(--surface-hover);
}

.menu-dropdown {
	position: absolute;
	top: calc(100% + 4px);
	right: 0;
	z-index: 50;
	min-width: 180px;
	padding: 0.375rem;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-lg);
}

.menu-item {
	display: flex;
	align-items: center;
	gap: 0.625rem;
	width: 100%;
	padding: 0.5rem 0.75rem;
	font-size: 0.875rem;
	color: var(--text);
	text-decoration: none;
	background: none;
	border: none;
	border-radius: var(--radius);
	cursor: pointer;
	transition: background-color 0.15s ease;
}

.menu-item:hover {
	background: var(--surface-hover);
	text-decoration: none;
}

.menu-item-danger {
	color: var(--error);
}

.menu-item-danger:hover {
	background: rgba(239, 68, 68, 0.1);
}

.menu-divider {
	margin: 0.375rem 0;
	border: none;
	border-top: 1px solid var(--border);
}

/* Card Body */
.card-body {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
	text-decoration: none;
	padding: 0;
}

.card-body:hover {
	text-decoration: none;
}

.form-title {
	font-size: 1rem;
	font-weight: 600;
	color: var(--text);
	line-height: 1.3;
	transition: color 0.15s ease;
}

.card-body:hover .form-title {
	color: var(--primary);
}

.form-description {
	font-size: 0.8125rem;
	color: var(--text-secondary);
	line-height: 1.4;
	display: -webkit-box;
	line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
}

/* Card Meta */
.card-meta {
	display: flex;
	gap: 1rem;
}

.meta-item {
	display: flex;
	align-items: center;
	gap: 0.375rem;
	font-size: 0.75rem;
	color: var(--text-secondary);
}

.meta-item :deep(i),
.meta-item :deep(svg) {
	font-size: 0.6875rem;
	opacity: 0.7;
}

/* Card Footer */
.card-footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.75rem 1rem;
	background: var(--background);
	border-top: 1px solid var(--border);
}

.update-time {
	display: flex;
	align-items: center;
	gap: 0.375rem;
	font-size: 0.75rem;
	color: var(--text-secondary);
}

.update-time :deep(i),
.update-time :deep(svg) {
	font-size: 0.6875rem;
	opacity: 0.7;
}

.quick-actions {
	display: flex;
	gap: 0.25rem;
}

.action-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 28px;
	height: 28px;
	color: var(--text-secondary);
	text-decoration: none;
	border-radius: var(--radius);
	transition: background-color 0.15s ease, color 0.15s ease;
}

.action-btn:hover {
	color: var(--primary);
	background: rgba(99, 102, 241, 0.1);
	text-decoration: none;
}

/* No Results */
.no-results {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	text-align: center;
	padding: 3rem;
	color: var(--text-secondary);
}

.no-results i {
	font-size: 2rem;
	margin-bottom: 1rem;
	opacity: 0.5;
}

.no-results h3 {
	font-size: 1.125rem;
	font-weight: 600;
	color: var(--text);
	margin-bottom: 0.25rem;
}

.no-results p {
	font-size: 0.875rem;
}

/* Dropdown Animation */
.dropdown-enter-active,
.dropdown-leave-active {
	transition: opacity 0.15s ease, transform 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
	opacity: 0;
	transform: translateY(-8px);
}

/* Responsive */
@media (max-width: 1024px) {
	.stats-row {
		grid-template-columns: repeat(2, 1fr);
	}
}

@media (max-width: 768px) {
	.header-content {
		flex-direction: column;
		align-items: stretch;
	}

	.header-content .btn {
		justify-content: center;
	}

	.stats-row {
		grid-template-columns: repeat(2, 1fr);
		gap: 0.75rem;
	}

	.stat-card {
		padding: 0.875rem;
	}

	.stat-icon {
		width: 40px;
		height: 40px;
	}

	.stat-value {
		font-size: 1.25rem;
	}

	.toolbar {
		flex-direction: column;
		align-items: stretch;
	}

	.search-box {
		max-width: none;
	}

	.toolbar-actions {
		flex-wrap: wrap;
	}

	.select-input {
		flex: 1;
		min-width: 140px;
	}

	.forms-grid {
		grid-template-columns: 1fr;
	}
}

@media (max-width: 480px) {
	.stats-row {
		grid-template-columns: 1fr 1fr;
	}

	.stat-card {
		flex-direction: column;
		align-items: flex-start;
		gap: 0.5rem;
	}

	.quick-actions {
		display: none;
	}
}
</style>
