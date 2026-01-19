<script lang="ts" setup>
const { t, locale } = useI18n();
const route = useRoute();
const { submissionsApi, filesApi } = useApi();
const { sanitizeHtml } = useSanitize();

// Cache for share URLs to avoid regenerating tokens for the same file
// Using a reactive object instead of Map for better Vue reactivity
const shareUrlCache = ref<Record<string, string>>({});

const id = route.params.id as string;

const form = ref<Form | null>(null);
const submissions = ref<Submission[]>([]);
const stats = ref<FormStats | null>(null);
const isLoading = ref(true);

// Pagination - Default 20 items per page
const pagination = usePagination(5);

// Tab state: 'summary' | 'question' | 'individual'
const activeTab = ref<"summary" | "question" | "individual">("summary");

// Question tab state
const selectedFieldId = ref<string | null>(null);

// Individual tab state
const currentSubmissionIndex = ref(0);
const sortOrder = ref<"newest" | "oldest">("newest");

const loadData = async (showLoading = true) => {
	if (showLoading) {
		isLoading.value = true;
	}
	try {
		const [submissionsData, statsData] = await Promise.all([
			submissionsApi.list(id, pagination.params.value),
			submissionsApi.stats(id),
		]);
		form.value = submissionsData.form;
		submissions.value = submissionsData.submissions?.data || [];
		if (submissionsData.submissions) {
			pagination.updateFromResponse(submissionsData.submissions);
		}
		stats.value = statsData;

		// Set default selected field for question tab
		if (form.value?.fields && form.value.fields.length > 0 && !selectedFieldId.value) {
			const firstField = form.value.fields[0];
			if (firstField) {
				selectedFieldId.value = firstField.id;
			}
		}

		useHead({
			title: t("forms.responses.title", { title: submissionsData.form.title }),
		});

		// Load share URLs for protected file fields (runs in background)
		loadProtectedFileUrls();
	} catch (error) {
		console.error("Failed to load data:", error);
	} finally {
		isLoading.value = false;
	}
};

// Load share URLs for protected files - defined here to be called from loadData
const loadProtectedFileUrls = async () => {
	// Wait for next tick to ensure formFields computed is ready
	await nextTick();

	const filePaths: string[] = [];
	const fields = form.value?.fields?.filter((f) => f.type === "file") || [];

	// Collect all protected file paths from submissions
	for (const submission of submissions.value) {
		for (const field of fields) {
			const value = submission.data[field.id];
			if (Array.isArray(value)) {
				for (const v of value) {
					if (typeof v === "string" && (v.startsWith("files/") || v.startsWith("/files/"))) {
						filePaths.push(v);
					}
				}
			} else if (typeof value === "string" && (value.startsWith("files/") || value.startsWith("/files/"))) {
				filePaths.push(value);
			}
		}
	}

	// Generate share URLs for all unique paths
	const uniquePaths = [...new Set(filePaths)];
	if (uniquePaths.length > 0) {
		// Generate share URLs inline
		for (const path of uniquePaths) {
			if (!(path in shareUrlCache.value)) {
				try {
					const cleanPath = path.startsWith("/") ? path.slice(1) : path;
					const result = await filesApi.generateShareUrl(cleanPath, 60);
					shareUrlCache.value[path] = result.url;
				} catch (error) {
					console.error("Failed to generate share URL for:", path, error);
				}
			}
		}
	}
};

// Watch for pagination changes
watch(() => pagination.params.value, () => {
	loadData(false);
}, { deep: true });

const handleDelete = async (submissionId: string) => {
	if (!confirm(t("forms.responses.confirmDelete"))) return;

	try {
		await submissionsApi.delete(id, submissionId);
		submissions.value = submissions.value.filter((s) => s.id !== submissionId);
		// Adjust index if needed
		if (currentSubmissionIndex.value >= submissions.value.length) {
			currentSubmissionIndex.value = Math.max(0, submissions.value.length - 1);
		}
	} catch (error) {
		console.error("Failed to delete submission:", error);
	}
};

const handleExportCSV = () => {
	submissionsApi.exportCSV(id);
};

const getFieldType = (fieldId: string): string => {
	const field = formFields.value.find((f) => f.id === fieldId);
	return field?.type || "text";
};

const getFieldValue = (submission: Submission, fieldId: string) => {
	const value = submission.data[fieldId];
	if (Array.isArray(value)) {
		return value.join(", ");
	}
	return value?.toString() || "-";
};

const isSignature = (fieldId: string): boolean => {
	return getFieldType(fieldId) === "signature";
};

const isFileField = (fieldId: string): boolean => {
	return getFieldType(fieldId) === "file";
};

const isRatingField = (fieldId: string): boolean => {
	return getFieldType(fieldId) === "rating";
};

const isScaleField = (fieldId: string): boolean => {
	return getFieldType(fieldId) === "scale";
};

const isRichTextField = (fieldId: string): boolean => {
	return getFieldType(fieldId) === "richtext";
};

const isBase64Image = (value: unknown): boolean => {
	if (typeof value !== "string") return false;
	return value.startsWith("data:image/");
};

// Layout field types that don't contain submission data
const layoutFieldTypes = ["section", "pagebreak", "divider", "heading", "paragraph", "image"];

// Get the field object for additional properties (like maxValue for rating)
const getField = (fieldId: string): FormField | undefined => {
	return formFields.value.find((f) => f.id === fieldId);
};

// File info object for display
interface FileInfo {
	url: string;
	filename: string;
	originalPath: string;
	isImage: boolean;
}

const getFileInfos = (value: unknown): FileInfo[] => {
	const processPath = (path: string): FileInfo => {
		const cachedUrl = shareUrlCache.value[path];
		const url = (isProtectedPath(path) && cachedUrl) ? cachedUrl : getFileUrl(path);
		return {
			url,
			filename: extractFilename(path, t("forms.responses.individual.file")),
			originalPath: path,
			isImage: isImageFile(path),
		};
	};

	if (Array.isArray(value)) {
		return value
			.filter((v) => typeof v === "string" && isFilePath(v))
			.map((v) => processPath(v));
	}
	if (typeof value === "string" && isFilePath(value)) {
		return [processPath(value)];
	}
	return [];
};

const formatDate = (dateString: string) => {
	return new Date(dateString).toLocaleString(locale.value === "de" ? "de-DE" : "en-US");
};

// Filter out layout fields that don't contain submission data
const formFields = computed(() => {
	const fields = form.value?.fields || [];
	return fields.filter((f) => !layoutFieldTypes.includes(f.type));
});

// Get the currently selected field for question tab
const selectedField = computed(() => {
	if (!selectedFieldId.value) return null;
	return formFields.value.find((f) => f.id === selectedFieldId.value) || null;
});

// Group answers by value for showing duplicates
const groupedAnswers = computed(() => {
	if (!selectedFieldId.value) return [];

	const groups: Map<string, { value: unknown; displayValue: string; submissions: Submission[] }> = new Map();

	for (const submission of submissions.value) {
		const value = submission.data[selectedFieldId.value];
		const displayValue = getFieldValue(submission, selectedFieldId.value);

		if (groups.has(displayValue)) {
			groups.get(displayValue)!.submissions.push(submission);
		} else {
			groups.set(displayValue, {
				value,
				displayValue,
				submissions: [submission],
			});
		}
	}

	return Array.from(groups.values()).sort((a, b) => b.submissions.length - a.submissions.length);
});

// Sorted submissions for individual view (newest first by default)
const sortedSubmissions = computed(() => {
	const sorted = [...submissions.value];
	if (sortOrder.value === "newest") {
		sorted.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
	} else {
		sorted.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
	}
	return sorted;
});

// Current submission for individual view
const currentSubmission = computed(() => {
	if (sortedSubmissions.value.length === 0) return null;
	return sortedSubmissions.value[currentSubmissionIndex.value] || null;
});

const toggleSortOrder = () => {
	sortOrder.value = sortOrder.value === "newest" ? "oldest" : "newest";
	currentSubmissionIndex.value = 0;
};

// Navigate to specific submission from question view
const goToSubmission = (submission: Submission) => {
	const index = sortedSubmissions.value.findIndex((s) => s.id === submission.id);
	if (index !== -1) {
		currentSubmissionIndex.value = index;
		activeTab.value = "individual";
	}
};

// Navigation for individual view
const prevSubmission = () => {
	if (currentSubmissionIndex.value > 0) {
		currentSubmissionIndex.value--;
	}
};

const nextSubmission = () => {
	if (currentSubmissionIndex.value < sortedSubmissions.value.length - 1) {
		currentSubmissionIndex.value++;
	}
};

onMounted(() => {
	loadData();
});
</script>

<template>
	<div v-if="isLoading" class="loading">
		<p>{{ $t("forms.responses.loading") }}</p>
	</div>

	<div v-else-if="!form" class="error">
		<p>{{ $t("forms.responses.formNotFound") }}</p>
	</div>

	<div v-else class="page-wrapper">
		<div class="page-content">
			<!-- Page Header -->
			<div class="page-header">
				<div class="page-header-top">
					<NuxtLink class="back-link" to="/forms">
						<UISysIcon icon="fa-solid fa-arrow-left" />
						<span>{{ $t("forms.responses.backToForms") }}</span>
					</NuxtLink>
					<div class="header-actions">
						<button class="icon-btn" :title="$t('forms.responses.refresh')" @click="() => loadData()">
							<UISysIcon icon="fa-solid fa-arrows-rotate" />
						</button>
						<button class="export-btn" @click="handleExportCSV">
							<UISysIcon icon="fa-solid fa-download" />
							<span>{{ $t("forms.responses.export") }}</span>
						</button>
					</div>
				</div>
				<div class="page-header-content">
					<h1>{{ form.title }}</h1>
					<div class="response-count">
						<UISysIcon icon="fa-solid fa-chart-simple" />
						<span>{{ $t("forms.responses.responseCount", { count: submissions.length }, submissions.length) }}</span>
					</div>
				</div>
				<!-- Tabs integrated in header -->
				<nav class="tabs">
					<button :class="['tab', { 'tab-active': activeTab === 'summary' }]" @click="activeTab = 'summary'">
						<UISysIcon icon="fa-solid fa-list" />
						<span>{{ $t("forms.responses.tabs.summary") }}</span>
					</button>
					<button :class="['tab', { 'tab-active': activeTab === 'question' }]" @click="activeTab = 'question'">
						<UISysIcon icon="fa-solid fa-circle-question" />
						<span>{{ $t("forms.responses.tabs.question") }}</span>
					</button>
					<button :class="['tab', { 'tab-active': activeTab === 'individual' }]" @click="activeTab = 'individual'">
						<UISysIcon icon="fa-solid fa-user" />
						<span>{{ $t("forms.responses.tabs.individual") }}</span>
					</button>
				</nav>
			</div>
			<!-- Empty State -->
			<div v-if="submissions.length === 0 && pagination.state.totalItems === 0" class="empty">
				<UISysIcon icon="fa-solid fa-chart-column" style="font-size: 48px" />
				<h2>{{ $t("forms.responses.empty.title") }}</h2>
				<p>{{ $t("forms.responses.empty.description") }}</p>
			</div>

			<!-- Summary Tab -->
			<div v-else-if="activeTab === 'summary'" class="summary-view">
				<div class="summary-header-card">
					<div class="summary-stat">
						<span class="summary-stat-value">{{ pagination.state.totalItems }}</span>
						<span class="summary-stat-label">{{ $t("forms.responses.summary.responses") }}</span>
					</div>
					<div class="summary-stat">
						<span class="summary-stat-value">{{ stats?.total_views || 0 }}</span>
						<span class="summary-stat-label">{{ $t("forms.responses.summary.views") }}</span>
					</div>
					<div class="summary-stat">
						<span class="summary-stat-value">{{ (stats?.conversion_rate || 0).toFixed(1) }}%</span>
						<span class="summary-stat-label">{{ $t("forms.responses.summary.conversionRate") }}</span>
					</div>
				</div>

				<div v-for="field in formFields" :key="field.id" class="summary-card">
					<div class="summary-card-header">
						<h3>{{ field.label }}</h3>
						<span class="summary-card-count">{{ $t("forms.responses.summary.answersCount", { count: submissions.length }) }}</span>
					</div>
					<div class="summary-card-content">
						<!-- For signature fields -->
						<template v-if="isSignature(field.id)">
							<div class="summary-signatures">
								<div v-for="submission in submissions.slice(0, 6)" :key="submission.id" class="summary-signature-item">
									<img
										v-if="isBase64Image(submission.data[field.id])"
										:src="submission.data[field.id] as string"
										alt="Unterschrift"
									/>
									<span v-else class="empty-value">-</span>
								</div>
								<div v-if="submissions.length > 6" class="summary-more">
									{{ $t("forms.responses.summary.more", { count: submissions.length - 6 }) }}
								</div>
							</div>
						</template>

						<!-- For file fields -->
						<template v-else-if="isFileField(field.id)">
							<div class="summary-files">
								<div
									v-for="submission in submissions.slice(0, 5)"
									:key="submission.id"
									class="summary-file-item"
								>
									<template v-if="getFileInfos(submission.data[field.id]).length > 0">
										<template v-for="file in getFileInfos(submission.data[field.id]).slice(0, 2)" :key="file.originalPath">
											<!-- Image preview for image files -->
											<div v-if="file.isImage" class="file-image-preview">
												<img :src="file.url" :alt="file.filename" />
												<a :href="file.url" target="_blank" class="file-image-overlay">
													<UISysIcon icon="fa-solid fa-expand" />
												</a>
											</div>
											<!-- Regular file link -->
											<a v-else :href="file.url" target="_blank" class="file-link">
												<UISysIcon icon="fa-solid fa-file" />
												{{ file.filename }}
											</a>
										</template>
									</template>
									<span v-else class="empty-value">-</span>
								</div>
								<div v-if="submissions.length > 5" class="summary-more">
									{{ $t("forms.responses.summary.more", { count: submissions.length - 5 }) }}
								</div>
							</div>
						</template>

						<!-- For rating fields - show stars -->
						<template v-else-if="isRatingField(field.id)">
							<div class="summary-ratings">
								<div
									v-for="submission in submissions.slice(0, 8)"
									:key="submission.id"
									class="summary-rating-item"
								>
									<div class="rating-stars">
										<UISysIcon
											v-for="i in (getField(field.id)?.maxValue || 5)"
											:key="i"
											:icon="i <= Number(submission.data[field.id] || 0) ? 'fa-solid fa-star' : 'fa-regular fa-star'"
											:class="{ 'star-filled': i <= Number(submission.data[field.id] || 0) }"
										/>
									</div>
									<span class="rating-value">{{ submission.data[field.id] || '-' }}</span>
								</div>
								<div v-if="submissions.length > 8" class="summary-more">
									{{ $t("forms.responses.summary.more", { count: submissions.length - 8 }) }}
								</div>
							</div>
						</template>

						<!-- For scale fields - show scale value with bar -->
						<template v-else-if="isScaleField(field.id)">
							<div class="summary-scales">
								<div
									v-for="submission in submissions.slice(0, 8)"
									:key="submission.id"
									class="summary-scale-item"
								>
									<div class="scale-bar">
										<div
											class="scale-bar-fill"
											:style="{
												width: `${((Number(submission.data[field.id]) - (getField(field.id)?.minValue || 1)) / ((getField(field.id)?.maxValue || 10) - (getField(field.id)?.minValue || 1))) * 100}%`
											}"
										/>
									</div>
									<span class="scale-value">{{ submission.data[field.id] || '-' }}</span>
								</div>
								<div v-if="submissions.length > 8" class="summary-more">
									{{ $t("forms.responses.summary.more", { count: submissions.length - 8 }) }}
								</div>
							</div>
						</template>

						<!-- For richtext fields - show HTML preview -->
						<template v-else-if="isRichTextField(field.id)">
							<div class="summary-richtext">
								<div
									v-for="submission in submissions.slice(0, 5)"
									:key="submission.id"
									class="summary-richtext-item"
								>
									<div
										v-if="submission.data[field.id]"
										class="richtext-content richtext-content-sm"
										v-html="sanitizeHtml(submission.data[field.id])"
									/>
									<span v-else class="empty-value">-</span>
								</div>
								<div v-if="submissions.length > 5" class="summary-more">
									{{ $t("forms.responses.summary.more", { count: submissions.length - 5 }) }}
								</div>
							</div>
						</template>

						<!-- For choice fields with stats (select, radio, checkbox, dropdown) show bar chart -->
						<template v-else-if="['select', 'radio', 'checkbox', 'dropdown'].includes(field.type) && stats?.field_stats?.[field.id] && Object.keys(stats.field_stats[field.id]!).length > 0">
							<div class="stat-bars">
								<div
									v-for="[value, count] in Object.entries(stats.field_stats[field.id]!).sort((a, b) => (b[1] as number) - (a[1] as number))"
									:key="value"
									class="stat-bar"
								>
									<div class="stat-bar-label">
										<span>{{ value || $t("forms.responses.summary.empty") }}</span>
										<span class="stat-bar-count">{{ count }} ({{ Math.round(((count as number) / submissions.length) * 100) }}%)</span>
									</div>
									<div class="stat-bar-track">
										<div
											class="stat-bar-fill"
											:style="{ width: `${((count as number) / submissions.length) * 100}%` }"
										/>
									</div>
								</div>
							</div>
						</template>

						<!-- For text/other fields show list of answers -->
						<template v-else>
							<div class="summary-answers">
								<div
									v-for="submission in submissions.slice(0, 8)"
									:key="submission.id"
									class="summary-answer-item"
								>
									{{ getFieldValue(submission, field.id) }}
								</div>
								<div v-if="submissions.length > 8" class="summary-more">
									{{ $t("forms.responses.summary.more", { count: submissions.length - 8 }) }}
								</div>
							</div>
						</template>
					</div>
				</div>

				<!-- Pagination for Summary View -->
				<UIPagination
					:page="pagination.state.page"
					:page-size="pagination.state.pageSize"
					:total-items="pagination.state.totalItems"
					:total-pages="pagination.state.totalPages"
					:visible-pages="pagination.visiblePages.value"
					:has-next-page="pagination.hasNextPage.value"
					:has-prev-page="pagination.hasPrevPage.value"
					@update:page="pagination.setPage"
					@update:page-size="pagination.setPageSize"
					@first="pagination.firstPage"
					@prev="pagination.prevPage"
					@next="pagination.nextPage"
					@last="pagination.lastPage"
				/>
			</div>

			<!-- Question Tab -->
			<div v-else-if="activeTab === 'question'" class="question-view">
				<div class="question-sidebar">
					<div class="question-list">
						<button
							v-for="field in formFields"
							:key="field.id"
							:class="['question-item', { 'question-item-active': selectedFieldId === field.id }]"
							@click="selectedFieldId = field.id"
						>
							<span class="question-item-label">{{ field.label }}</span>
							<span class="question-item-count">{{ pagination.state.totalItems }}</span>
						</button>
					</div>
				</div>

				<div class="question-content">
					<template v-if="selectedField">
						<div class="question-header">
							<h2>{{ selectedField.label }}</h2>
							<p class="question-type">{{ selectedField.type }}</p>
						</div>

						<!-- Grouped answers with duplicate count -->
						<div class="answer-groups">
							<div
								v-for="group in groupedAnswers"
								:key="group.displayValue"
								class="answer-group"
							>
								<div class="answer-group-header">
									<div class="answer-group-value">
										<!-- Signature -->
										<template v-if="isSignature(selectedFieldId!) && isBase64Image(group.value)">
											<img :src="group.value as string" alt="Unterschrift" class="answer-signature" />
										</template>
										<!-- File -->
										<template v-else-if="isFileField(selectedFieldId!)">
											<div v-if="getFileInfos(group.value).length > 0" class="answer-files">
												<template v-for="file in getFileInfos(group.value)" :key="file.originalPath">
													<!-- Image preview -->
													<div v-if="file.isImage" class="file-image-preview">
														<img :src="file.url" :alt="file.filename" />
														<a :href="file.url" target="_blank" class="file-image-overlay">
															<UISysIcon icon="fa-solid fa-expand" />
														</a>
													</div>
													<!-- Regular file -->
													<a v-else :href="file.url" target="_blank" class="file-link">
														<UISysIcon icon="fa-solid fa-file" />
														{{ file.filename }}
													</a>
												</template>
											</div>
											<span v-else class="empty-value">-</span>
										</template>
										<!-- Rating -->
										<template v-else-if="isRatingField(selectedFieldId!)">
											<div class="rating-stars">
												<UISysIcon
													v-for="i in (getField(selectedFieldId!)?.maxValue || 5)"
													:key="i"
													:icon="i <= Number(group.value || 0) ? 'fa-solid fa-star' : 'fa-regular fa-star'"
													:class="{ 'star-filled': i <= Number(group.value || 0) }"
												/>
												<span class="rating-value">{{ group.value || '-' }}</span>
											</div>
										</template>
										<!-- Scale -->
										<template v-else-if="isScaleField(selectedFieldId!)">
											<div class="scale-display">
												<div class="scale-bar">
													<div
														class="scale-bar-fill"
														:style="{
															width: `${((Number(group.value) - (getField(selectedFieldId!)?.minValue || 1)) / ((getField(selectedFieldId!)?.maxValue || 10) - (getField(selectedFieldId!)?.minValue || 1))) * 100}%`
														}"
													/>
												</div>
												<span class="scale-value">{{ group.value || '-' }}</span>
											</div>
										</template>
										<!-- Rich Text -->
										<template v-else-if="isRichTextField(selectedFieldId!)">
											<div
												v-if="group.value"
												class="richtext-content richtext-content-sm"
												v-html="sanitizeHtml(group.value)"
											/>
											<span v-else class="empty-value">-</span>
										</template>
										<!-- Text -->
										<template v-else>
											<span :class="{ 'empty-value': !group.displayValue || group.displayValue === '-' }">
												{{ group.displayValue || $t("forms.responses.summary.empty") }}
											</span>
										</template>
									</div>
									<span class="answer-group-count" :class="{ 'has-duplicates': group.submissions.length > 1 }">
										{{ group.submissions.length }}x
									</span>
								</div>

								<!-- Show submissions when there are duplicates or always show for navigation -->
								<div class="answer-group-submissions">
									<button
										v-for="submission in group.submissions"
										:key="submission.id"
										class="submission-link"
										@click="goToSubmission(submission)"
									>
										<UISysIcon icon="fa-solid fa-user" />
										<span>{{ formatDate(submission.created_at) }}</span>
										<UISysIcon icon="fa-solid fa-arrow-right" class="arrow-icon" />
									</button>
								</div>
							</div>
						</div>
					</template>
				</div>

				<!-- Pagination for Question View -->
				<UIPagination
					:page="pagination.state.page"
					:page-size="pagination.state.pageSize"
					:total-items="pagination.state.totalItems"
					:total-pages="pagination.state.totalPages"
					:visible-pages="pagination.visiblePages.value"
					:has-next-page="pagination.hasNextPage.value"
					:has-prev-page="pagination.hasPrevPage.value"
					class="question-pagination"
					@update:page="pagination.setPage"
					@update:page-size="pagination.setPageSize"
					@first="pagination.firstPage"
					@prev="pagination.prevPage"
					@next="pagination.nextPage"
					@last="pagination.lastPage"
				/>
			</div>

			<!-- Individual Tab -->
			<div v-else-if="activeTab === 'individual'" class="individual-view">
				<div class="individual-nav">
					<button
						class="nav-btn"
						:disabled="currentSubmissionIndex === 0"
						@click="prevSubmission"
					>
						<UISysIcon icon="fa-solid fa-chevron-left" />
					</button>
					<span class="nav-counter">
						{{ currentSubmissionIndex + 1 }} {{ $t("forms.responses.individual.of") }} {{ sortedSubmissions.length }}
					</span>
					<button
						class="nav-btn"
						:disabled="currentSubmissionIndex === sortedSubmissions.length - 1"
						@click="nextSubmission"
					>
						<UISysIcon icon="fa-solid fa-chevron-right" />
					</button>
					<button
						class="sort-btn"
						:title="sortOrder === 'newest' ? $t('forms.responses.individual.newestFirst') : $t('forms.responses.individual.oldestFirst')"
						@click="toggleSortOrder"
					>
						<UISysIcon :icon="sortOrder === 'newest' ? 'fa-solid fa-arrow-down-wide-short' : 'fa-solid fa-arrow-up-wide-short'" />
						<span>{{ sortOrder === 'newest' ? $t('forms.responses.individual.newest') : $t('forms.responses.individual.oldest') }}</span>
					</button>
				</div>

				<div v-if="currentSubmission" class="individual-card">
					<div class="individual-header">
						<div class="individual-meta">
							<UISysIcon icon="fa-solid fa-clock" />
							<span>{{ formatDate(currentSubmission.created_at) }}</span>
						</div>
						<button class="btn btn-danger btn-sm" @click="handleDelete(currentSubmission.id)">
							<UISysIcon icon="fa-solid fa-trash" />
							{{ $t("forms.responses.individual.delete") }}
						</button>
					</div>

					<div class="individual-fields">
						<div v-for="field in formFields" :key="field.id" class="individual-field">
							<label>{{ field.label }}</label>

							<!-- Signature -->
							<template v-if="isSignature(field.id) && isBase64Image(currentSubmission.data[field.id])">
								<div class="signature-preview">
									<img :src="currentSubmission.data[field.id] as string" alt="Unterschrift" />
								</div>
							</template>

							<!-- File -->
							<template v-else-if="isFileField(field.id)">
								<div v-if="getFileInfos(currentSubmission.data[field.id]).length > 0" class="file-list">
									<template v-for="file in getFileInfos(currentSubmission.data[field.id])" :key="file.originalPath">
										<!-- Image preview -->
										<div v-if="file.isImage" class="file-image-large">
											<img :src="file.url" :alt="file.filename" />
											<a :href="file.url" target="_blank" class="file-download-btn">
												<UISysIcon icon="fa-solid fa-download" />
												{{ file.filename }}
											</a>
										</div>
										<!-- Regular file -->
										<a v-else :href="file.url" target="_blank" class="file-link">
											<UISysIcon icon="fa-solid fa-file" />
											{{ file.filename }}
										</a>
									</template>
								</div>
								<p v-else class="empty-value">-</p>
							</template>

							<!-- Rating -->
							<template v-else-if="isRatingField(field.id)">
								<div class="rating-display">
									<div class="rating-stars-large">
										<UISysIcon
											v-for="i in (getField(field.id)?.maxValue || 5)"
											:key="i"
											:icon="i <= Number(currentSubmission.data[field.id] || 0) ? 'fa-solid fa-star' : 'fa-regular fa-star'"
											:class="{ 'star-filled': i <= Number(currentSubmission.data[field.id] || 0) }"
										/>
									</div>
									<span class="rating-text">{{ $t("forms.responses.individual.ofValue", { value: currentSubmission.data[field.id] || '-', max: getField(field.id)?.maxValue || 5 }) }}</span>
								</div>
							</template>

							<!-- Scale -->
							<template v-else-if="isScaleField(field.id)">
								<div class="scale-display-large">
									<div class="scale-labels">
										<span>{{ getField(field.id)?.minLabel || getField(field.id)?.minValue || 1 }}</span>
										<span>{{ getField(field.id)?.maxLabel || getField(field.id)?.maxValue || 10 }}</span>
									</div>
									<div class="scale-bar-large">
										<div
											class="scale-bar-fill"
											:style="{
												width: `${((Number(currentSubmission.data[field.id]) - (getField(field.id)?.minValue || 1)) / ((getField(field.id)?.maxValue || 10) - (getField(field.id)?.minValue || 1))) * 100}%`
											}"
										/>
										<span class="scale-marker" :style="{
											left: `${((Number(currentSubmission.data[field.id]) - (getField(field.id)?.minValue || 1)) / ((getField(field.id)?.maxValue || 10) - (getField(field.id)?.minValue || 1))) * 100}%`
										}">
											{{ currentSubmission.data[field.id] || '-' }}
										</span>
									</div>
								</div>
							</template>

							<!-- Rich Text -->
							<template v-else-if="isRichTextField(field.id)">
								<div
									v-if="currentSubmission.data[field.id]"
									class="richtext-content"
									v-html="sanitizeHtml(currentSubmission.data[field.id])"
								/>
								<p v-else class="empty-value">-</p>
							</template>

							<!-- Text -->
							<template v-else>
								<p :class="{ 'empty-value': !currentSubmission.data[field.id] }">
									{{ getFieldValue(currentSubmission, field.id) }}
								</p>
							</template>
						</div>
					</div>

					<div v-if="currentSubmission.metadata" class="individual-metadata">
						<h4>{{ $t("forms.responses.individual.metadata") }}</h4>
						<div class="metadata-grid">
							<div v-if="currentSubmission.metadata.ip">
								<span class="metadata-label">{{ $t("forms.responses.individual.ipAddress") }}</span>
								<span>{{ currentSubmission.metadata.ip }}</span>
							</div>
							<div v-if="currentSubmission.metadata.user_agent">
								<span class="metadata-label">{{ $t("forms.responses.individual.userAgent") }}</span>
								<span class="metadata-ua">{{ currentSubmission.metadata.user_agent }}</span>
							</div>
						</div>
					</div>
				</div>

				<!-- Pagination for Individual View -->
				<UIPagination
					:page="pagination.state.page"
					:page-size="pagination.state.pageSize"
					:total-items="pagination.state.totalItems"
					:total-pages="pagination.state.totalPages"
					:visible-pages="pagination.visiblePages.value"
					:has-next-page="pagination.hasNextPage.value"
					:has-prev-page="pagination.hasPrevPage.value"
					@update:page="pagination.setPage"
					@update:page-size="pagination.setPageSize"
					@first="pagination.firstPage"
					@prev="pagination.prevPage"
					@next="pagination.nextPage"
					@last="pagination.lastPage"
				/>
			</div>
		</div>
	</div>
</template>

<style scoped>
/* Page Layout */
.page-wrapper {
	min-height: 100vh;
	background: var(--background);
}

.loading,
.error {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	min-height: 50vh;
	color: var(--text-secondary);
}

/* Page Header */
.page-header {
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow);
	margin-bottom: 1rem;
	overflow: hidden;
}

.page-header-top {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0.75rem 1.25rem;
	border-bottom: 1px solid var(--border);
}

.back-link {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem 0.75rem;
	font-size: 0.8125rem;
	font-weight: 500;
	color: var(--text-secondary);
	text-decoration: none;
	border-radius: var(--radius);
	transition: all 0.15s ease;
}

.back-link:hover {
	color: var(--text);
	background: var(--surface-hover);
	text-decoration: none;
}

.header-actions {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.icon-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 36px;
	height: 36px;
	color: var(--text-secondary);
	background: transparent;
	border: none;
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s ease;
}

.icon-btn:hover {
	color: var(--text);
	background: var(--surface-hover);
}

.export-btn {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem 1rem;
	font-size: 0.8125rem;
	font-weight: 500;
	color: var(--text);
	background: var(--surface-hover);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s ease;
}

.export-btn:hover {
	background: var(--primary);
	color: white;
	border-color: var(--primary);
}

.page-header-content {
	padding: 1.25rem 1.25rem;
}

.page-header-content h1 {
	font-size: 1.5rem;
	font-weight: 600;
	color: var(--text);
	margin-bottom: 0.5rem;
	line-height: 1.3;
}

.response-count {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	font-size: 0.875rem;
	color: var(--text-secondary);
}

.response-count i {
	color: var(--primary);
}

/* Tabs */
.tabs {
	display: flex;
	gap: 0;
	padding: 0 0;
	border-top: 1px solid var(--border);
	background: var(--surface-hover);
}

.tab {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.875rem 1.25rem;
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text-secondary);
	background: transparent;
	border: none;
	border-radius: 0 0 0 0;
	cursor: pointer;
	transition: all 0.15s ease;
}

.tab:hover {
	color: var(--text);
	background: var(--surface);
}

.tab-active {
	color: var(--primary);
	background: var(--surface);
}

.tab-active:hover {
	background: var(--surface);
}

/* Page Content */
.page-content {
	max-width: 1100px;
	margin: 0 auto;
	padding: 1.5rem;
	min-height: calc(100vh - 200px);
}

/* Empty State */
.empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 4rem 2rem;
	text-align: center;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
}

.empty i {
	margin-bottom: 1.5rem;
	color: var(--text-secondary);
	opacity: 0.5;
}

.empty h2 {
	margin-bottom: 0.5rem;
	font-size: 1.25rem;
	font-weight: 600;
	color: var(--text);
}

.empty p {
	color: var(--text-secondary);
	max-width: 300px;
}

/* Summary View */
.summary-view {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

.summary-header-card {
	display: flex;
	justify-content: center;
	gap: 3rem;
	padding: 2.5rem;
	background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow);
}

.summary-stat {
	display: flex;
	flex-direction: column;
	align-items: center;
}

.summary-stat-value {
	font-size: 2.5rem;
	font-weight: 700;
	color: white;
	line-height: 1;
}

.summary-stat-label {
	margin-top: 0.5rem;
	font-size: 0.875rem;
	color: rgba(255, 255, 255, 0.85);
	font-weight: 500;
}

.summary-card {
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	overflow: hidden;
	box-shadow: var(--shadow);
}

.summary-card-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 1rem 1.25rem;
	background: var(--surface-hover);
	border-bottom: 1px solid var(--border);
}

.summary-card-header h3 {
	font-size: 0.9375rem;
	font-weight: 600;
	color: var(--text);
}

.summary-card-count {
	font-size: 0.75rem;
	color: var(--text-secondary);
	background: var(--surface);
	padding: 0.25rem 0.625rem;
	border-radius: 999px;
}

.summary-card-content {
	padding: 1.25rem;
}

/* Stat Bars */
.stat-bars {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.stat-bar {
	display: flex;
	flex-direction: column;
	gap: 0.375rem;
}

.stat-bar-label {
	display: flex;
	justify-content: space-between;
	font-size: 0.875rem;
}

.stat-bar-count {
	color: var(--text-secondary);
}

.stat-bar-track {
	height: 10px;
	background: var(--surface-hover);
	border-radius: 5px;
	overflow: hidden;
}

.stat-bar-fill {
	height: 100%;
	background: linear-gradient(90deg, var(--primary) 0%, var(--primary-light) 100%);
	border-radius: 5px;
	transition: width 0.4s ease;
}

/* Summary Answers */
.summary-answers {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.summary-answer-item {
	padding: 0.875rem 1rem;
	font-size: 0.875rem;
	color: var(--text);
	background: var(--surface-hover);
	border-radius: var(--radius);
	border-left: 3px solid var(--primary);
}

.summary-more {
	padding: 0.75rem 1rem;
	font-size: 0.8125rem;
	color: var(--primary);
	text-align: center;
	cursor: pointer;
	border-radius: var(--radius);
	transition: background 0.15s ease;
}

.summary-more:hover {
	background: var(--surface-hover);
}

/* Summary Signatures */
.summary-signatures {
	display: flex;
	flex-wrap: wrap;
	gap: 0.75rem;
}

.summary-signature-item {
	padding: 0.5rem;
	background: #ffffff;
	border-radius: var(--radius);
	border: 1px solid var(--border);
}

.summary-signature-item img {
	max-width: 120px;
	max-height: 60px;
	display: block;
}

/* Summary Files */
.summary-files {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.summary-file-item {
	display: flex;
	flex-wrap: wrap;
	gap: 0.5rem;
}

/* Summary Richtext */
.summary-richtext {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.summary-richtext-item {
	padding: 0.75rem;
	background: var(--background);
	border-radius: var(--radius);
	border: 1px solid var(--border);
}

/* Question View */
.question-view {
	display: flex;
	flex-wrap: wrap;
	gap: 1.5rem;
	min-height: 500px;
}

.question-pagination {
	width: 100%;
}

.question-sidebar {
	width: 280px;
	flex-shrink: 0;
}

.question-list {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	padding: 0.5rem;
	box-shadow: var(--shadow);
	position: sticky;
	top: 220px;
}

.question-item {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0.875rem 1rem;
	font-size: 0.875rem;
	text-align: left;
	color: var(--text);
	cursor: pointer;
	background: none;
	border: none;
	border-radius: var(--radius);
	transition: all 0.15s ease;
}

.question-item:hover {
	background: var(--surface-hover);
}

.question-item-active {
	background: var(--primary);
	color: white;
}

.question-item-label {
	flex: 1;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.question-item-count {
	font-size: 0.75rem;
	color: var(--text-secondary);
	background: var(--surface-hover);
	padding: 0.25rem 0.5rem;
	border-radius: 999px;
	font-weight: 500;
}

.question-item-active .question-item-count {
	background: rgba(255, 255, 255, 0.2);
	color: white;
}

.question-content {
	flex: 1;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	overflow: hidden;
	box-shadow: var(--shadow);
}

.question-header {
	padding: 1.5rem;
	background: var(--surface-hover);
	border-bottom: 1px solid var(--border);
}

.question-header h2 {
	font-size: 1.125rem;
	font-weight: 600;
	color: var(--text);
	margin-bottom: 0.375rem;
}

.question-type {
	display: inline-flex;
	align-items: center;
	padding: 0.25rem 0.5rem;
	font-size: 0.6875rem;
	font-weight: 600;
	color: var(--primary);
	background: rgba(99, 102, 241, 0.1);
	border-radius: var(--radius);
	text-transform: uppercase;
	letter-spacing: 0.025em;
}

/* Answer Groups */
.answer-groups {
	max-height: 500px;
	overflow-y: auto;
}

.answer-group {
	border-bottom: 1px solid var(--border);
}

.answer-group:last-child {
	border-bottom: none;
}

.answer-group-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
	padding: 1rem 1.25rem;
	background: var(--background);
}

.answer-group-value {
	flex: 1;
	font-size: 0.9375rem;
	word-break: break-word;
}

.answer-group-count {
	font-size: 0.75rem;
	color: var(--text-secondary);
	background: var(--surface);
	padding: 0.25rem 0.5rem;
	border-radius: var(--radius);
	margin-left: 1rem;
}

.answer-group-count.has-duplicates {
	background: rgba(99, 102, 241, 0.1);
	color: var(--primary);
	font-weight: 600;
}

.answer-group-submissions {
	display: flex;
	flex-direction: column;
}

.submission-link {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.75rem 1.25rem;
	font-size: 0.8125rem;
	color: var(--text-secondary);
	cursor: pointer;
	background: none;
	border: none;
	text-align: left;
	transition: all 0.2s;
}

.submission-link:hover {
	background: var(--background);
	color: var(--primary);
}

.submission-link .arrow-icon {
	margin-left: auto;
	opacity: 0;
	transition: opacity 0.2s;
}

.submission-link:hover .arrow-icon {
	opacity: 1;
}

.answer-signature {
	max-width: 200px;
	max-height: 80px;
	padding: 0.5rem;
	background: #ffffff;
	border: 1px solid var(--border);
	border-radius: var(--radius);
}

.answer-files {
	display: flex;
	flex-wrap: wrap;
	gap: 0.5rem;
}

/* Individual View */
.individual-view {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

.individual-nav {
	display: flex;
	justify-content: center;
	align-items: center;
	gap: 1.5rem;
	padding: 1rem 1.5rem;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow);
}

.nav-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 40px;
	height: 40px;
	color: var(--text);
	cursor: pointer;
	background: var(--surface-hover);
	border: 1px solid var(--border);
	border-radius: 50%;
	transition: all 0.15s ease;
}

.nav-btn:hover:not(:disabled) {
	background: var(--primary);
	color: white;
	border-color: var(--primary);
	transform: scale(1.05);
}

.nav-btn:disabled {
	opacity: 0.3;
	cursor: not-allowed;
}

.nav-counter {
	font-size: 0.9375rem;
	font-weight: 600;
	color: var(--text);
	min-width: 100px;
	text-align: center;
}

.sort-btn {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem 1rem;
	margin-left: 1rem;
	font-size: 0.8125rem;
	font-weight: 500;
	color: var(--text-secondary);
	background: var(--surface-hover);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s ease;
}

.sort-btn:hover {
	color: var(--text);
	background: var(--surface);
	border-color: var(--primary);
}

.individual-card {
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	overflow: hidden;
	box-shadow: var(--shadow);
}

.individual-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 1.25rem 1.5rem;
	background: var(--surface-hover);
	border-bottom: 1px solid var(--border);
}

.individual-meta {
	display: flex;
	align-items: center;
	gap: 0.625rem;
	font-size: 0.875rem;
	color: var(--text-secondary);
}

.individual-meta i {
	color: var(--primary);
}

.individual-fields {
	padding: 1.5rem;
	display: flex;
	flex-direction: column;
	gap: 1.75rem;
}

.individual-field {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
	padding-bottom: 1.5rem;
	border-bottom: 1px solid var(--border);
}

.individual-field:last-child {
	padding-bottom: 0;
	border-bottom: none;
}

.individual-field label {
	font-size: 0.8125rem;
	font-weight: 600;
	color: var(--text-secondary);
}

.individual-field p {
	font-size: 1rem;
	color: var(--text);
	word-break: break-word;
	line-height: 1.5;
}

.individual-metadata {
	padding: 1.25rem 1.5rem;
	background: var(--surface-hover);
	border-top: 1px solid var(--border);
}

.individual-metadata h4 {
	font-size: 0.75rem;
	font-weight: 600;
	color: var(--text-secondary);
	text-transform: uppercase;
	letter-spacing: 0.05em;
	margin-bottom: 0.75rem;
}

.metadata-grid {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	font-size: 0.8125rem;
}

.metadata-label {
	color: var(--text-secondary);
	margin-right: 0.5rem;
}

.metadata-ua {
	word-break: break-all;
	color: var(--text-secondary);
	font-family: monospace;
	font-size: 0.75rem;
}

/* Shared Styles */
.empty-value {
	color: var(--text-secondary);
}

.signature-preview {
	padding: 0.75rem;
	background: #ffffff;
	border-radius: var(--radius);
	border: 1px solid var(--border);
	display: inline-block;
}

.signature-preview img {
	max-width: 100%;
	max-height: 150px;
	display: block;
}

.file-list {
	display: flex;
	flex-direction: column;
	gap: 0.375rem;
}

.file-link {
	display: inline-flex;
	align-items: center;
	gap: 0.375rem;
	padding: 0.5rem 0.75rem;
	font-size: 0.8125rem;
	color: var(--primary);
	text-decoration: none;
	background: rgba(99, 102, 241, 0.1);
	border-radius: var(--radius);
	transition: all 0.2s;
}

.file-link:hover {
	background: rgba(99, 102, 241, 0.2);
}

/* File Image Preview */
.file-image-preview {
	position: relative;
	display: inline-block;
	border-radius: var(--radius);
	overflow: hidden;
}

.file-image-preview img {
	max-width: 120px;
	max-height: 80px;
	display: block;
	object-fit: cover;
}

.file-image-overlay {
	position: absolute;
	inset: 0;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(0, 0, 0, 0.5);
	color: white;
	opacity: 0;
	transition: opacity 0.2s;
}

.file-image-preview:hover .file-image-overlay {
	opacity: 1;
}

.file-image-large {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	padding: 0.75rem;
	background: var(--background);
	border-radius: var(--radius);
}

.file-image-large img {
	max-width: 100%;
	max-height: 300px;
	object-fit: contain;
	border-radius: var(--radius);
}

.file-download-btn {
	display: inline-flex;
	align-items: center;
	gap: 0.375rem;
	padding: 0.5rem 0.75rem;
	font-size: 0.8125rem;
	color: var(--primary);
	text-decoration: none;
	background: rgba(99, 102, 241, 0.1);
	border-radius: var(--radius);
	transition: all 0.2s;
	align-self: flex-start;
}

.file-download-btn:hover {
	background: rgba(99, 102, 241, 0.2);
}

/* Rating Styles */
.summary-ratings {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.summary-rating-item {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	padding: 0.5rem 0.75rem;
	background: var(--background);
	border-radius: var(--radius);
}

.rating-stars {
	display: flex;
	align-items: center;
	gap: 0.125rem;
}

.rating-stars i {
	color: var(--text-secondary);
	font-size: 0.875rem;
}

.rating-stars .star-filled {
	color: #fbbf24;
}

.rating-value {
	font-size: 0.8125rem;
	color: var(--text-secondary);
	margin-left: 0.5rem;
}

.rating-display {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.rating-stars-large {
	display: flex;
	align-items: center;
	gap: 0.25rem;
}

.rating-stars-large i {
	color: var(--text-secondary);
	font-size: 1.25rem;
}

.rating-stars-large .star-filled {
	color: #fbbf24;
}

.rating-text {
	font-size: 0.875rem;
	color: var(--text-secondary);
}

/* Scale Styles */
.summary-scales {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.summary-scale-item {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	padding: 0.5rem 0.75rem;
	background: var(--background);
	border-radius: var(--radius);
}

.scale-bar {
	flex: 1;
	height: 8px;
	background: var(--surface);
	border-radius: 4px;
	overflow: hidden;
}

.scale-bar-fill {
	height: 100%;
	background: var(--primary);
	border-radius: 4px;
	transition: width 0.3s ease;
}

.scale-value {
	font-size: 0.875rem;
	font-weight: 500;
	min-width: 2rem;
	text-align: right;
}

.scale-display {
	display: flex;
	align-items: center;
	gap: 0.75rem;
}

.scale-display-large {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.scale-labels {
	display: flex;
	justify-content: space-between;
	font-size: 0.75rem;
	color: var(--text-secondary);
}

.scale-bar-large {
	position: relative;
	height: 12px;
	background: var(--background);
	border-radius: 6px;
	overflow: visible;
}

.scale-bar-large .scale-bar-fill {
	height: 100%;
	background: var(--primary);
	border-radius: 6px;
}

.scale-marker {
	position: absolute;
	top: -24px;
	transform: translateX(-50%);
	font-size: 0.75rem;
	font-weight: 600;
	color: var(--primary);
	background: var(--surface);
	padding: 0.125rem 0.375rem;
	border-radius: var(--radius);
	border: 1px solid var(--border);
}

/* Responsive */
@media (max-width: 768px) {
	.page-header-top {
		padding: 0.75rem 1rem;
	}

	.page-header-content {
		padding: 1rem;
	}

	.page-header-content h1 {
		font-size: 1.25rem;
	}

	.tabs {
		padding: 0;
		overflow-x: auto;
	}

	.tab {
		padding: 0.75rem 1rem;
		font-size: 0.8125rem;
		border-radius: 0;
	}

	.tab span {
		display: none;
	}

	.page-content {
		padding: 1rem;
	}

	.question-view {
		flex-direction: column;
	}

	.question-sidebar {
		width: 100%;
	}

	.question-list {
		flex-direction: row;
		flex-wrap: wrap;
		position: static;
	}

	.question-item {
		flex: 1;
		min-width: 120px;
	}

	.individual-nav {
		padding: 0.75rem 1rem;
		gap: 1rem;
	}

	.nav-btn {
		width: 36px;
		height: 36px;
	}

	.individual-header {
		flex-direction: column;
		gap: 1rem;
		align-items: flex-start;
	}

	.summary-header-card {
		padding: 1.5rem 1rem;
		gap: 1.5rem;
		flex-wrap: wrap;
	}

	.summary-stat-value {
		font-size: 1.75rem;
	}

	.export-btn span {
		display: none;
	}

	.sort-btn {
		padding: 0.5rem;
		margin-left: 0.5rem;
	}

	.sort-btn span {
		display: none;
	}
}
</style>

<!-- Unscoped styles for v-html content -->
<style>
.richtext-content {
	line-height: 1.6;
	color: var(--text);
}

.richtext-content h1,
.richtext-content h2,
.richtext-content h3 {
	margin: 0.5em 0;
	font-weight: 600;
	color: var(--text);
}

.richtext-content h1 {
	font-size: 1.5rem;
}

.richtext-content h2 {
	font-size: 1.25rem;
}

.richtext-content h3 {
	font-size: 1.1rem;
}

.richtext-content p {
	margin: 0.5em 0;
	color: var(--text);
}

.richtext-content ul,
.richtext-content ol {
	margin: 0.5em 0;
	padding-left: 1.5em;
	color: var(--text);
}

.richtext-content ul {
	list-style-type: disc;
}

.richtext-content ol {
	list-style-type: decimal;
}

.richtext-content li {
	margin: 0.25em 0;
	color: var(--text);
}

.richtext-content li::marker {
	color: var(--text);
}

.richtext-content blockquote {
	margin: 0.5em 0;
	padding: 0.5em 1em;
	border-left: 3px solid var(--primary);
	background: var(--background);
	font-style: italic;
	color: var(--text);
}

.richtext-content a {
	color: var(--primary);
	text-decoration: underline;
}

.richtext-content strong,
.richtext-content b {
	font-weight: 600;
}

.richtext-content em,
.richtext-content i:not([class]) {
	font-style: italic;
}

.richtext-content u {
	text-decoration: underline;
}

.richtext-content s,
.richtext-content strike {
	text-decoration: line-through;
}

.richtext-content-sm {
	font-size: 0.875rem;
	max-height: 150px;
	overflow: hidden;
	position: relative;
}

.richtext-content-sm::after {
	content: "";
	position: absolute;
	bottom: 0;
	left: 0;
	right: 0;
	height: 30px;
	background: linear-gradient(transparent, var(--background));
	pointer-events: none;
}

.richtext-content-sm h1,
.richtext-content-sm h2,
.richtext-content-sm h3 {
	font-size: 1rem;
}
</style>
