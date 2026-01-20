<script setup lang="ts">
import { ref, computed, watch } from "vue";
import type { FormTemplate, TemplateCategoryInfo } from "~~/shared/types";
import { FIELD_META } from "~~/shared/types";

const props = defineProps<{
	show: boolean;
}>();

const emit = defineEmits<{
	close: [];
	select: [templateId: string];
}>();

const { t } = useI18n();
const { templatesApi } = useApi();

// State
const templates = ref<FormTemplate[]>([]);
const categories = ref<TemplateCategoryInfo[]>([]);
const selectedCategory = ref<string>("all");
const searchQuery = ref("");
const isLoading = ref(false);
const selectedTemplate = ref<FormTemplate | null>(null);

// Load data
const loadData = async () => {
	isLoading.value = true;
	try {
		[templates.value, categories.value] = await Promise.all([
			templatesApi.list(),
			templatesApi.categories(),
		]);
		// Add "all" category
		categories.value.unshift({
			category: "all",
			count: templates.value.length,
			label: t("templates.categories.all"),
			icon: "grid-2",
		});
	} catch (error) {
		console.error("Failed to load templates:", error);
	} finally {
		isLoading.value = false;
	}
};

// Filtered templates
const filteredTemplates = computed(() => {
	let result = templates.value;

	// Filter by category
	if (selectedCategory.value && selectedCategory.value !== "all") {
		result = result.filter((t) => t.template_category === selectedCategory.value);
	}

	// Filter by search
	if (searchQuery.value) {
		const query = searchQuery.value.toLowerCase();
		result = result.filter(
			(t) =>
				t.title.toLowerCase().includes(query) ||
				t.description.toLowerCase().includes(query)
		);
	}

	return result;
});

// Watch show prop
watch(
	() => props.show,
	(newValue) => {
		if (newValue) {
			loadData();
			selectedTemplate.value = null;
		}
	}
);

const handleSelectTemplate = (template: FormTemplate) => {
	selectedTemplate.value = template;
};

const handleUseTemplate = () => {
	if (selectedTemplate.value) {
		emit("select", selectedTemplate.value.id);
		emit("close");
	}
};

const handleBack = () => {
	selectedTemplate.value = null;
};
</script>

<template>
	<Teleport to="body">
		<Transition name="modal">
			<div v-if="show" class="modal-overlay" @click.self="emit('close')">
				<div class="modal-container">
					<!-- Header -->
					<div class="modal-header">
						<button v-if="selectedTemplate" class="back-button" @click="handleBack">
							<i class="fa-solid fa-arrow-left"></i>
						</button>
						<h2>{{ selectedTemplate ? selectedTemplate.title : $t("templates.browse") }}</h2>
						<button class="close-button" @click="emit('close')">
							<i class="fa-solid fa-xmark"></i>
						</button>
					</div>

					<!-- Template Preview (when selected) -->
					<div v-if="selectedTemplate" class="template-preview">
						<div class="preview-content">
							<div class="preview-header">
								<div class="template-category-badge">
									<i :class="`fa-solid fa-${categories.find((c) => c.category === selectedTemplate?.template_category)?.icon || 'folder'}`"></i>
									{{ $t(`templates.categories.${selectedTemplate.template_category}`) }}
								</div>
								<p class="preview-description">{{ selectedTemplate.description }}</p>
							</div>

							<div class="preview-fields">
								<h3>{{ $t("templates.fields", { count: selectedTemplate.fields.length }) }}</h3>
								<div class="fields-list">
									<div
										v-for="field in selectedTemplate.fields"
										:key="field.id"
										class="field-item"
									>
										<i :class="FIELD_META[field.type]?.icon || 'fa-solid fa-circle'"></i>
										<span class="field-label">{{ field.label }}</span>
										<span v-if="field.required" class="required-badge">{{ $t("builder.fieldItem.required") }}</span>
									</div>
								</div>
							</div>

							<div class="preview-actions">
								<button class="btn btn-secondary" @click="handleBack">
									{{ $t("common.back") }}
								</button>
								<button class="btn btn-primary" @click="handleUseTemplate">
									{{ $t("templates.useTemplate") }}
								</button>
							</div>
						</div>
					</div>

					<!-- Template Gallery (default view) -->
					<div v-else class="gallery-content">
						<!-- Search Bar -->
						<div class="search-bar">
							<i class="fa-solid fa-search"></i>
							<input
								v-model="searchQuery"
								type="text"
								:placeholder="$t('templates.search')"
								class="search-input"
							/>
						</div>

						<div class="gallery-layout">
							<!-- Sidebar with categories -->
							<aside class="categories-sidebar">
								<button
									v-for="category in categories"
									:key="category.category"
									class="category-item"
									:class="{ active: selectedCategory === category.category }"
									@click="selectedCategory = category.category"
								>
									<i :class="`fa-solid fa-${category.icon}`"></i>
									<span>{{ category.label }}</span>
									<span class="count">{{ category.count }}</span>
								</button>
							</aside>

							<!-- Templates Grid -->
							<div class="templates-grid">
								<div v-if="isLoading" class="loading-state">
									<i class="fa-solid fa-spinner fa-spin"></i>
									<p>{{ $t("templates.loading") }}</p>
								</div>

								<div v-else-if="filteredTemplates.length === 0" class="empty-state">
									<i class="fa-solid fa-folder-open"></i>
									<p>{{ $t("templates.noResults") }}</p>
								</div>

								<div
									v-for="template in filteredTemplates"
									v-else
									:key="template.id"
									class="template-card"
									@click="handleSelectTemplate(template)"
								>
									<div class="template-icon">
										<i :class="`fa-solid fa-${categories.find((c) => c.category === template.template_category)?.icon || 'folder'}`"></i>
									</div>
									<div class="template-info">
										<h3>{{ template.title }}</h3>
										<p>{{ template.description }}</p>
										<div class="template-meta">
											<span class="field-count">
												<i class="fa-solid fa-list"></i>
												{{ template.fields.length }} {{ $t("templates.fields", { count: template.fields.length }) }}
											</span>
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</Transition>
	</Teleport>
</template>

<style scoped>
.modal-overlay {
	position: fixed;
	inset: 0;
	background: rgba(0, 0, 0, 0.5);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
	padding: 1rem;
}

.modal-container {
	background: var(--surface);
	border-radius: 0.75rem;
	box-shadow: var(--shadow-lg);
	max-width: 80rem;
	width: 100%;
	max-height: 90vh;
	overflow: hidden;
	display: flex;
	flex-direction: column;
}

.modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 1.5rem;
	border-bottom: 1px solid var(--border);
}

.modal-header h2 {
	font-size: 1.25rem;
	font-weight: 600;
	color: var(--text);
	margin: 0;
	flex: 1;
	text-align: center;
}

.back-button,
.close-button {
	background: none;
	border: none;
	font-size: 1.25rem;
	color: var(--text-secondary);
	cursor: pointer;
	padding: 0.5rem;
	display: flex;
	align-items: center;
	justify-content: center;
	width: 2.5rem;
	height: 2.5rem;
	border-radius: 0.375rem;
	transition: all 0.15s;
}

.back-button:hover,
.close-button:hover {
	background: var(--surface-hover);
	color: var(--text);
}

/* Template Preview */
.template-preview {
	flex: 1;
	overflow-y: auto;
	padding: 2rem;
}

.preview-content {
	max-width: 48rem;
	margin: 0 auto;
}

.preview-header {
	margin-bottom: 2rem;
}

.template-category-badge {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem 1rem;
	background: color-mix(in srgb, var(--primary) 20%, transparent);
	color: var(--primary);
	border-radius: 0.5rem;
	font-size: 0.875rem;
	font-weight: 500;
	margin-bottom: 1rem;
}

.preview-description {
	font-size: 1.125rem;
	color: var(--text-secondary);
	margin: 0;
}

.preview-fields {
	margin-bottom: 2rem;
}

.preview-fields h3 {
	font-size: 1rem;
	font-weight: 600;
	color: var(--text);
	margin: 0 0 1rem 0;
}

.fields-list {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.field-item {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	padding: 0.75rem;
	background: var(--background);
	border-radius: 0.5rem;
	border: 1px solid var(--border);
}

.field-item i {
	color: var(--text-secondary);
	width: 1.25rem;
}

.field-label {
	flex: 1;
	font-weight: 500;
	color: var(--text);
}

.required-badge {
	font-size: 0.75rem;
	padding: 0.25rem 0.5rem;
	background: color-mix(in srgb, var(--warning) 30%, transparent);
	color: var(--warning);
	border-radius: 0.25rem;
	font-weight: 500;
}

.preview-actions {
	display: flex;
	gap: 1rem;
	justify-content: flex-end;
}

/* Gallery Content */
.gallery-content {
	flex: 1;
	overflow: hidden;
	display: flex;
	flex-direction: column;
}

.search-bar {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	padding: 1.5rem;
	border-bottom: 1px solid var(--border);
}

.search-bar i {
	color: var(--text-secondary);
}

.search-input {
	flex: 1;
	border: none;
	outline: none;
	font-size: 1rem;
	color: var(--text);
	background: transparent;
}

.search-input::placeholder {
	color: var(--text-secondary);
}

.gallery-layout {
	flex: 1;
	display: grid;
	grid-template-columns: 16rem 1fr;
	overflow: hidden;
}

@media (max-width: 768px) {
	.gallery-layout {
		grid-template-columns: 1fr;
	}

	.categories-sidebar {
		display: none;
	}
}

.categories-sidebar {
	border-right: 1px solid var(--border);
	overflow-y: auto;
	padding: 1rem;
}

.category-item {
	width: 100%;
	display: flex;
	align-items: center;
	gap: 0.75rem;
	padding: 0.75rem;
	border: none;
	background: var(--surface);
	border-radius: 0.5rem;
	cursor: pointer;
	transition: all 0.15s;
	font-size: 0.875rem;
	color: var(--text);
	margin-bottom: 0.25rem;
}

.category-item:hover {
	background: var(--surface-hover);
}

.category-item.active {
	background: color-mix(in srgb, var(--primary) 20%, transparent);
	color: var(--primary);
	font-weight: 500;
}

.category-item i {
	width: 1.25rem;
}

.category-item span:first-of-type {
	flex: 1;
	text-align: left;
}

.category-item .count {
	font-size: 0.75rem;
	padding: 0.125rem 0.5rem;
	background: var(--border);
	border-radius: 0.375rem;
	font-weight: 500;
	color: var(--text-secondary);
}

.category-item.active .count {
	background: var(--primary);
	color: white;
}

.templates-grid {
	overflow-y: auto;
	padding: 1.5rem;
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(18rem, 1fr));
	gap: 1.5rem;
	align-content: start;
}

.loading-state,
.empty-state {
	grid-column: 1 / -1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 4rem 2rem;
	color: var(--text-secondary);
	text-align: center;
}

.loading-state i,
.empty-state i {
	font-size: 3rem;
	margin-bottom: 1rem;
}

.template-card {
	background: var(--surface);
	border: 2px solid var(--border);
	border-radius: 0.75rem;
	padding: 1.5rem;
	cursor: pointer;
	transition: all 0.2s;
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

.template-card:hover {
	border-color: var(--primary);
	transform: translateY(-2px);
	box-shadow: var(--shadow-lg);
}

.template-icon {
	width: 3rem;
	height: 3rem;
	border-radius: 0.75rem;
	background: color-mix(in srgb, var(--primary) 20%, transparent);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 1.5rem;
	color: var(--primary);
}

.template-info h3 {
	font-size: 1.125rem;
	font-weight: 600;
	color: var(--text);
	margin: 0;
}

.template-info p {
	font-size: 0.875rem;
	color: var(--text-secondary);
	margin: 0;
}

.template-meta {
	display: flex;
	align-items: center;
	gap: 1rem;
	font-size: 0.75rem;
	color: var(--text-secondary);
}

.field-count {
	display: flex;
	align-items: center;
	gap: 0.375rem;
}

/* Buttons */
.btn {
	padding: 0.625rem 1.25rem;
	border-radius: 0.5rem;
	font-weight: 500;
	font-size: 0.875rem;
	cursor: pointer;
	transition: all 0.15s;
	border: none;
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
}

.btn-primary {
	background: var(--primary);
	color: white;
}

.btn-primary:hover {
	background: var(--primary-dark);
}

.btn-secondary {
	background: var(--surface);
	color: var(--text);
	border: 1px solid var(--border);
}

.btn-secondary:hover {
	background: var(--surface-hover);
}

/* Modal transition */
.modal-enter-active,
.modal-leave-active {
	transition: opacity 0.2s;
}

.modal-enter-from,
.modal-leave-to {
	opacity: 0;
}

.modal-enter-active .modal-container,
.modal-leave-active .modal-container {
	transition: transform 0.2s;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
	transform: scale(0.95);
}
</style>
