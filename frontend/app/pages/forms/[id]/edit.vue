<script lang="ts" setup>
import { VueDraggable } from "vue-draggable-plus";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const formsStore = useFormsStore();
const toastStore = useToastStore();

const id = route.params.id as string;

const form = ref<Form | null>(null);
const fields = ref<FormField[]>([]);
const selectedFieldId = ref<string | null>(null);
const activeView = ref<"designer" | "settings">("designer");
const showFieldSettingsModal = ref(false);
const showDeleteDialog = ref(false);
const fieldToDelete = ref<string | null>(null);
const isLoading = ref(true);

// Slug & Password settings
const slugInput = ref("");
const slugStatus = ref<"idle" | "checking" | "available" | "taken" | "invalid">("idle");
const passwordEnabled = ref(false);
const passwordInput = ref("");

// Debounced slug check
let slugCheckTimeout: ReturnType<typeof setTimeout> | null = null;

const checkSlug = async (value: string) => {
	if (slugCheckTimeout) clearTimeout(slugCheckTimeout);

	if (!value) {
		slugStatus.value = "idle";
		return;
	}

	// Basic validation
	const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
	if (value.length < 3 || value.length > 100 || !slugPattern.test(value)) {
		slugStatus.value = "invalid";
		return;
	}

	slugStatus.value = "checking";

	slugCheckTimeout = setTimeout(async () => {
		const result = await formsStore.checkSlugAvailability(value, form.value?.id);
		slugStatus.value = result.available ? "available" : "taken";
	}, 500);
};

const normalizeSlug = (value: string) => {
	return value.toLowerCase().trim().replace(/\s+/g, "-").replace(/_/g, "-").replace(/--+/g, "-").replace(/^-|-$/g, "");
};

const onSlugInput = (event: Event) => {
	const input = event.target as HTMLInputElement;
	const normalized = normalizeSlug(input.value);
	slugInput.value = normalized;
	checkSlug(normalized);
	markDirty();
};

const selectedField = computed(() => fields.value.find((f) => f.id === selectedFieldId.value));

// Auto-Save
const { isSaving, isDirty, statusText, markDirty, saveNow } = useAutoSave({
	delay: 2000,
	onSave: async () => {
		if (!form.value) return;
		const updatedFields = fields.value.map((f, i) => ({ ...f, order: i }));

		// Build update request
		const updateData: UpdateFormRequest = {
			...form.value,
			fields: updatedFields,
			slug: slugInput.value || undefined,
			password_protected: passwordEnabled.value,
		};

		// Only include password if it was changed
		if (passwordInput.value) {
			updateData.password = passwordInput.value;
		}

		await formsStore.updateForm(form.value.id, updateData);
		fields.value = updatedFields;

		// Clear password input after save
		passwordInput.value = "";

		toastStore.success(t("forms.editor.saved"), t("forms.editor.savedMessage"));
	},
	onError: () => {
		toastStore.error(t("forms.editor.saveFailed"), t("forms.editor.saveFailedMessage"));
	},
	onNoChanges: () => {
		toastStore.info(t("forms.editor.noChanges"), t("forms.editor.noChangesMessage"));
	},
});

const loadForm = async () => {
	try {
		const data = await formsStore.fetchForm(id);
		// Ensure design object exists with defaults
		if (!data.settings.design) {
			data.settings.design = {
				maxWidth: "lg",
				primaryColor: "#6366f1",
				backgroundColor: "#f3f4f6",
				formBackgroundColor: "#ffffff",
				textColor: "#111827",
				borderRadius: "lg",
				headerStyle: "default",
				buttonStyle: "filled",
				fontFamily: "default",
				backgroundSize: "cover",
			};
		} else if (!data.settings.design.backgroundSize) {
			data.settings.design.backgroundSize = "cover";
		}
		form.value = data;
		fields.value = data.fields || [];

		// Initialize slug and password settings
		slugInput.value = data.slug || "";
		passwordEnabled.value = data.password_protected || false;
		if (data.slug) {
			slugStatus.value = "available";
		}

		useHead({
			title: t("forms.editor.editTitle", { title: data.title }),
		});
	} catch {
		router.push("/forms");
	} finally {
		isLoading.value = false;
	}
};

const addField = (type: FieldType) => {
	const newField: FormField = {
		id: `field_${Date.now()}`,
		type,
		label: t(`fields.${type}.label`),
		description: t(`fields.${type}.description`),
		required: false,
		order: fields.value.length,
		options: ["select", "radio", "checkbox", "dropdown"].includes(type) ? ["Option 1", "Option 2"] : undefined,
		minValue: type === "rating" ? 1 : type === "scale" ? 1 : undefined,
		maxValue: type === "rating" ? 5 : type === "scale" ? 10 : undefined,
		headingLevel: type === "heading" ? 2 : undefined,
	};
	fields.value.push(newField);
	selectedFieldId.value = newField.id;
	markDirty();
};

const updateField = (updatedField: FormField) => {
	const index = fields.value.findIndex((f) => f.id === updatedField.id);
	if (index !== -1) {
		fields.value[index] = updatedField;
		markDirty();
	}
};

const confirmDeleteField = (fieldId: string) => {
	fieldToDelete.value = fieldId;
	showDeleteDialog.value = true;
};

const deleteField = () => {
	if (!fieldToDelete.value) return;
	fields.value = fields.value.filter((f) => f.id !== fieldToDelete.value);
	if (selectedFieldId.value === fieldToDelete.value) {
		selectedFieldId.value = null;
	}
	fieldToDelete.value = null;
	showDeleteDialog.value = false;
	markDirty();
};

const handleFieldsReorder = () => {
	markDirty();
};

watch(
	() => form.value?.title,
	() => markDirty()
);

watch(
	() => form.value?.description,
	() => markDirty()
);

watch(
	() => form.value?.settings,
	() => markDirty(),
	{ deep: true }
);

onMounted(() => {
	loadForm();
});

onBeforeUnmount(() => {
	formsStore.clearCurrentForm();
});

// Computed form link
const config = useRuntimeConfig();
const formLink = computed(() => {
	if (!form.value) return "";
	const baseUrl = config.public.siteUrl || window.location.origin;
	const identifier = slugInput.value || form.value.id;
	return `${baseUrl}/f/${identifier}`;
});

const openFieldSettings = (fieldId: string) => {
	selectedFieldId.value = fieldId;
	showFieldSettingsModal.value = true;
};

const updateFormFromSettings = (updatedForm: Form) => {
	form.value = updatedForm;
};
</script>

<template>
	<div v-if="isLoading" class="loading">
		<UILoadingSpinner size="lg" :label="$t('forms.editor.loading')" />
	</div>

	<div v-else-if="form" class="builder">
		<!-- Header -->
		<header class="header">
			<div class="header-left">
				<button class="btn btn-ghost" :aria-label="$t('forms.editor.backToOverview')" @click="router.push('/forms')">
					<UISysIcon icon="fa-solid fa-arrow-left" />
				</button>
				<input v-model="form.title" class="title-input" :placeholder="$t('forms.editor.titlePlaceholder')" type="text" />
				<span v-if="statusText" class="save-status" :class="{ 'save-status-dirty': isDirty, 'save-status-saving': isSaving }">
					{{ statusText }}
				</span>
			</div>

			<!-- Navigation Tabs -->
			<nav class="header-tabs">
				<button
					:class="['header-tab', { 'header-tab-active': activeView === 'designer' }]"
					@click="activeView = 'designer'"
				>
					<UISysIcon icon="fa-solid fa-pen-ruler" />
					<span>{{ $t("forms.editor.tabs.designer") }}</span>
				</button>
				<button
					:class="['header-tab', { 'header-tab-active': activeView === 'settings' }]"
					@click="activeView = 'settings'"
				>
					<UISysIcon icon="fa-solid fa-gear" />
					<span>{{ $t("forms.editor.tabs.settings") }}</span>
				</button>
			</nav>

			<div class="header-right">
				<span :class="['status-badge', `status-${form.status}`]">
					<UISysIcon :icon="form.status === 'draft' ? 'fa-solid fa-file-pen' : form.status === 'published' ? 'fa-solid fa-globe' : 'fa-solid fa-lock'" />
					{{ $t(`forms.status.${form.status}`) }}
				</span>
				<a v-if="form.status === 'published'" :href="`/f/${slugInput || form.id}`" class="btn btn-secondary" rel="noopener noreferrer" target="_blank">
					<UISysIcon icon="fa-solid fa-eye" />
					<span>{{ $t("forms.editor.preview") }}</span>
				</a>
				<button :disabled="isSaving" class="btn btn-secondary" @click="saveNow">
					<UISysIcon icon="fa-solid fa-floppy-disk" />
					<span>{{ isSaving ? $t("forms.editor.saving") : $t("forms.editor.save") }}</span>
				</button>
			</div>
		</header>

		<!-- Designer View -->
		<div v-if="activeView === 'designer'" class="content">
			<!-- Field Types Sidebar -->
			<aside class="sidebar">
				<BuilderFieldTypePicker @select="addField" />
			</aside>

			<!-- Form Preview -->
			<main id="main-content" class="main">
				<div class="form-preview">
					<div class="form-header">
						<input v-model="form.title" class="form-title" :placeholder="$t('forms.editor.titlePlaceholder')" type="text" />
						<textarea v-model="form.description" class="form-description" :placeholder="$t('forms.editor.addDescription')" rows="2" />
					</div>

					<div v-if="fields.length === 0" class="empty-state">
						<UISysIcon icon="fa-solid fa-plus" style="font-size: 32px" />
						<p>{{ $t("forms.editor.emptyState") }}</p>
					</div>

					<VueDraggable
						v-else
						v-model="fields"
						target=".fields-container"
						handle=".field-drag-handle"
						:animation="150"
						@end="handleFieldsReorder"
					>
						<div class="fields-list">
							<div class="fields-container">
								<BuilderFieldItem
									v-for="field in fields"
									:key="field.id"
									:field="field"
									:selected="selectedFieldId === field.id"
									@select="selectedFieldId = field.id"
									@delete="confirmDeleteField(field.id)"
									@update:field="updateField"
									@open-settings="openFieldSettings(field.id)"
								/>
							</div>
						</div>
					</VueDraggable>
				</div>
			</main>
		</div>

		<!-- Settings View -->
		<div v-else-if="activeView === 'settings'" class="settings-view">
			<BuilderFormSettings
				:form="form"
				:slug-input="slugInput"
				:slug-status="slugStatus"
				:password-enabled="passwordEnabled"
				:password-input="passwordInput"
				:form-link="formLink"
				@update:form="updateFormFromSettings"
				@update:password-enabled="passwordEnabled = $event; markDirty()"
				@update:password-input="passwordInput = $event; markDirty()"
				@slug-input="onSlugInput"
				@mark-dirty="markDirty"
			/>
		</div>
	</div>

	<!-- Field Settings Modal -->
	<BuilderFieldSettingsModal
		v-model:open="showFieldSettingsModal"
		:field="selectedField"
		@update:field="updateField"
	/>

	<!-- Delete Confirmation Dialog -->
	<DialogConfirmDialog
		v-model:open="showDeleteDialog"
		:title="$t('forms.editor.deleteField.title')"
		:message="$t('forms.editor.deleteField.message')"
		:confirm-text="$t('forms.editor.deleteField.confirm')"
		variant="danger"
		@confirm="deleteField"
	/>
</template>

<style scoped>
.builder {
	display: flex;
	flex-direction: column;
	height: calc(100vh - 64px);
	margin: -2rem -1rem;
}

.loading {
	display: flex;
	align-items: center;
	justify-content: center;
	height: calc(100vh - 64px);
}

.header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
	padding: 0.75rem 1rem;
	background: var(--surface);
	border-bottom: 1px solid var(--border);
}

.header-left {
	display: flex;
	gap: 0.75rem;
	align-items: center;
}

.header-tabs {
	display: flex;
	gap: 0.25rem;
	padding: 0.25rem;
	background: var(--background);
	border-radius: var(--radius);
}

.header-tab {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	padding: 0.5rem 1rem;
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text-secondary);
	background: transparent;
	border: none;
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s ease;
}

.header-tab:hover {
	color: var(--text);
}

.header-tab-active {
	color: var(--primary);
	background: var(--surface);
	box-shadow: var(--shadow-sm);
}

.title-input {
	min-width: 200px;
	padding: 0.375rem;
	font-size: 1.125rem;
	font-weight: 600;
	color: var(--text);
	background: none;
	border: none;
}

.title-input:focus {
	outline: none;
	background: var(--background);
	border-radius: var(--radius);
}

.save-status {
	padding: 0.25rem 0.5rem;
	font-size: 0.75rem;
	color: var(--text-secondary);
	background: var(--background);
	border-radius: var(--radius);
}

.save-status-dirty {
	color: var(--warning);
}

.save-status-saving {
	color: var(--primary);
}

.status-badge {
	display: flex;
	gap: 0.375rem;
	align-items: center;
	padding: 0.375rem 0.75rem;
	font-size: 0.75rem;
	font-weight: 500;
	border-radius: var(--radius);
}

.status-badge.status-draft {
	color: #92400e;
	background: #fef3c7;
}

.status-badge.status-published {
	color: #166534;
	background: #dcfce7;
}

.status-badge.status-closed {
	color: #991b1b;
	background: #fee2e2;
}

.header-right {
	display: flex;
	gap: 0.5rem;
	align-items: center;
}

.content {
	display: flex;
	flex: 1;
	overflow: hidden;
}

.settings-view {
	flex: 1;
	overflow-y: auto;
}

.sidebar {
	width: 240px;
	padding: 1rem;
	overflow-y: auto;
	background: var(--surface);
	border-right: 1px solid var(--border);
}

.main {
	flex: 1;
	padding: 2rem;
	overflow-y: auto;
	background: var(--background);
}

.form-preview {
	max-width: 640px;
	margin: 0 auto;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow);
}

.form-header {
	padding: 2rem 2rem 1rem;
	border-bottom: 1px solid var(--border);
}

.form-title {
	width: 100%;
	padding: 0;
	margin-bottom: 0.5rem;
	font-size: 1.5rem;
	font-weight: 600;
	color: var(--text);
	background: none;
	border: none;
}

.form-title:focus {
	outline: none;
}

.form-description {
	width: 100%;
	padding: 0;
	font-size: 0.9375rem;
	color: var(--text-secondary);
	resize: none;
	background: none;
	border: none;
}

.form-description:focus {
	outline: none;
}

.empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 4rem 2rem;
	color: var(--text-secondary);
	text-align: center;
}

.empty-state i {
	margin-bottom: 1rem;
	opacity: 0.5;
}

.fields-list {
	padding: 1.5rem;
}

/* Tablet */
@media (max-width: 1024px) {
	.sidebar {
		width: 200px;
	}

	.main {
		padding: 1.5rem;
	}

	.form-header {
		padding: 1.5rem 1.5rem 1rem;
	}

	.fields-list {
		padding: 1rem;
	}
}

/* Mobile */
@media (max-width: 768px) {
	.builder {
		height: auto;
		min-height: calc(100vh - 64px);
	}

	.header {
		flex-wrap: wrap;
		gap: 0.5rem;
		padding: 0.5rem;
	}

	.header-left {
		order: 1;
		flex: 1;
		min-width: 0;
	}

	.header-tabs {
		order: 3;
		width: 100%;
		justify-content: center;
	}

	.header-tab {
		flex: 1;
		justify-content: center;
		padding: 0.5rem;
	}

	.header-tab span {
		display: none;
	}

	.header-right {
		order: 2;
	}

	.title-input {
		min-width: 100px;
		font-size: 1rem;
	}

	.save-status {
		display: none;
	}

	.content {
		flex-direction: column;
	}

	.sidebar {
		width: 100%;
		max-height: none;
		padding: 0.75rem;
		overflow: visible;
		border-right: none;
		border-bottom: 1px solid var(--border);
	}

	.main {
		flex: 1;
		padding: 1rem;
		overflow: visible;
	}

	.form-preview {
		border-radius: var(--radius);
	}

	.form-header {
		padding: 1rem;
	}

	.form-title {
		font-size: 1.25rem;
	}

	.form-description {
		font-size: 0.875rem;
	}

	.empty-state {
		padding: 2rem 1rem;
	}

	.fields-list {
		padding: 0.75rem;
	}

	.header-right .btn {
		padding: 0.5rem;
	}

	.header-right .btn span,
	.header-right .btn:not(:has(i)) {
		font-size: 0.75rem;
	}

	/* Hide button text on mobile, show only icons */
	.header-right .btn-secondary span {
		display: none;
	}
}

/* Small mobile */
@media (max-width: 480px) {
	.header-left .btn {
		padding: 0.375rem;
	}

	.title-input {
		min-width: 80px;
		font-size: 0.875rem;
	}

	.header-tab {
		padding: 0.375rem;
		font-size: 0.8125rem;
	}

	.form-preview {
		margin: 0 -0.5rem;
		border-radius: 0;
		border-left: none;
		border-right: none;
	}

	.main {
		padding: 0.75rem 0.5rem;
	}
}
</style>
