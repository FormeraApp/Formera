<script lang="ts" setup>
const props = defineProps<{
	field: FormField;
	selected?: boolean;
}>();

const emit = defineEmits(["select", "delete", "update:field", "openSettings"]);

const fieldMeta = computed(() => FIELD_META[props.field.type]);

const isLayoutField = computed(() => {
	const layoutTypes: LayoutFieldType[] = ["section", "pagebreak", "divider", "heading", "paragraph", "image"];
	return layoutTypes.includes(props.field.type as LayoutFieldType);
});

const handleLabelChange = (event: Event) => {
	const target = event.target as HTMLInputElement;
	emit("update:field", { ...props.field, label: target.value });
};

const handleSectionDescriptionChange = (event: Event) => {
	const target = event.target as HTMLTextAreaElement;
	emit("update:field", { ...props.field, sectionDescription: target.value });
};

const handleContentChange = (event: Event) => {
	const target = event.target as HTMLTextAreaElement;
	emit("update:field", { ...props.field, content: target.value });
};
</script>

<template>
	<div
		:class="['field-item', { 'field-item-selected': selected, 'field-item-layout': isLayoutField }]"
		@click="emit('select')"
	>
		<div class="field-drag-handle">
			<UISysIcon icon="fa-solid fa-grip-vertical" />
		</div>

		<!-- Layout Fields: Section -->
		<div v-if="field.type === 'section'" class="field-content field-section">
			<div class="section-header">
				<UISysIcon :icon="fieldMeta.icon" class="section-icon" />
				<input
					:value="field.label"
					class="field-label section-title"
					:placeholder="$t('builder.fieldItem.sectionTitle')"
					type="text"
					@click.stop
					@input="handleLabelChange"
				/>
			</div>
			<textarea
				:value="field.sectionDescription || ''"
				class="section-description"
				:placeholder="$t('builder.fieldItem.sectionDescription')"
				rows="2"
				@click.stop
				@input="handleSectionDescriptionChange"
			></textarea>
		</div>

		<!-- Layout Fields: Divider -->
		<div v-else-if="field.type === 'divider'" class="field-content field-divider">
			<hr class="divider-line" />
		</div>

		<!-- Layout Fields: Pagebreak -->
		<div v-else-if="field.type === 'pagebreak'" class="field-content field-pagebreak">
			<div class="pagebreak-indicator">
				<UISysIcon icon="fa-solid fa-file-lines" />
				<span>{{ $t("builder.fieldItem.pagebreak") }}</span>
			</div>
		</div>

		<!-- Layout Fields: Heading -->
		<div v-else-if="field.type === 'heading'" class="field-content field-heading">
			<input
				:value="field.label"
				:class="['heading-input', `heading-${field.headingLevel || 2}`]"
				:placeholder="$t('builder.fieldItem.headingPlaceholder')"
				type="text"
				@click.stop
				@input="handleLabelChange"
			/>
		</div>

		<!-- Layout Fields: Paragraph -->
		<div v-else-if="field.type === 'paragraph'" class="field-content field-paragraph">
			<textarea
				:value="field.content || ''"
				class="paragraph-input"
				:placeholder="$t('builder.fieldItem.paragraphPlaceholder')"
				rows="3"
				@click.stop
				@input="handleContentChange"
			></textarea>
		</div>

		<!-- Regular Fields -->
		<div v-else class="field-content">
			<div class="field-header">
				<span class="field-icon">
					<UISysIcon :icon="fieldMeta.icon" />
				</span>
				<input
					:value="field.label"
					class="field-label"
					:placeholder="$t('builder.fieldItem.fieldLabel')"
					type="text"
					@click.stop
					@input="handleLabelChange"
				/>
			</div>
			<div class="field-meta">
				<span class="field-type">{{ $t(`fields.${field.type}.label`) }}</span>
				<span v-if="field.required" class="field-required">{{ $t("builder.fieldItem.required") }}</span>
				<span v-if="field.description" class="field-has-description">
					<UISysIcon icon="fa-solid fa-info-circle" />
				</span>
			</div>
		</div>

		<div class="field-actions">
			<button class="field-action" :aria-label="$t('builder.fieldItem.settings')" @click.stop="emit('openSettings')">
				<UISysIcon icon="fa-solid fa-gear" />
			</button>
			<button class="field-action field-action-delete" :aria-label="$t('builder.fieldItem.deleteField')" @click.stop="emit('delete')">
				<UISysIcon icon="fa-solid fa-trash" />
			</button>
		</div>
	</div>
</template>

<style scoped>
.field-item {
	display: flex;
	gap: 0.75rem;
	align-items: flex-start;
	padding: 1rem;
	margin-bottom: 0.75rem;
	cursor: pointer;
	background: var(--background);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	transition: all 0.2s;
}

.field-item:hover {
	border-color: var(--primary-light);
}

.field-item-selected {
	border-color: var(--primary);
	box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.field-item-layout {
	background: var(--surface);
}

.field-drag-handle {
	padding: 0.25rem;
	margin-top: 0.25rem;
	color: var(--text-secondary);
	cursor: grab;
}

.field-drag-handle:active {
	cursor: grabbing;
}

.field-content {
	flex: 1;
	min-width: 0;
}

.field-header {
	display: flex;
	gap: 0.5rem;
	align-items: center;
}

.field-icon {
	color: var(--text-secondary);
}

.field-label {
	flex: 1;
	padding: 0;
	font-size: 0.9375rem;
	font-weight: 500;
	color: var(--text);
	background: none;
	border: none;
}

.field-label:focus {
	outline: none;
}

.field-meta {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	margin-top: 0.25rem;
}

.field-type {
	font-size: 0.75rem;
	color: var(--text-secondary);
}

.field-required {
	padding: 0.125rem 0.375rem;
	font-size: 0.6875rem;
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
	border-radius: 9999px;
}

.field-has-description {
	color: var(--primary);
	font-size: 0.75rem;
}

.field-actions {
	display: flex;
	gap: 0.25rem;
	margin-top: 0.25rem;
}

.field-action {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 28px;
	height: 28px;
	padding: 0;
	color: var(--text-secondary);
	cursor: pointer;
	background: none;
	border: none;
	border-radius: var(--radius);
	transition: all 0.15s ease;
}

.field-action:hover {
	color: var(--primary);
	background: rgba(99, 102, 241, 0.1);
}

.field-action-delete:hover {
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
}

/* Section */
.field-section {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.section-header {
	display: flex;
	gap: 0.5rem;
	align-items: center;
}

.section-icon {
	color: var(--primary);
}

.section-title {
	font-size: 1.125rem;
	font-weight: 600;
}

.section-description {
	padding: 0.5rem;
	font-size: 0.875rem;
	color: var(--text-secondary);
	resize: none;
	background: var(--background);
	border: 1px solid var(--border);
	border-radius: var(--radius);
}

.section-description:focus {
	outline: none;
	border-color: var(--border-focus);
}

/* Divider */
.field-divider {
	padding: 0.5rem 0;
}

.divider-line {
	border: none;
	border-top: 2px solid var(--border);
}

/* Pagebreak */
.field-pagebreak {
	padding: 0.5rem 0;
}

.pagebreak-indicator {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	justify-content: center;
	padding: 0.75rem;
	font-size: 0.875rem;
	color: var(--text-secondary);
	background: var(--background);
	border: 2px dashed var(--border);
	border-radius: var(--radius);
}

/* Heading */
.heading-input {
	width: 100%;
	padding: 0;
	font-weight: 600;
	color: var(--text);
	background: none;
	border: none;
}

.heading-input:focus {
	outline: none;
}

.heading-1 {
	font-size: 1.5rem;
}

.heading-2 {
	font-size: 1.25rem;
}

.heading-3 {
	font-size: 1.125rem;
}

.heading-4 {
	font-size: 1rem;
}

/* Paragraph */
.paragraph-input {
	width: 100%;
	padding: 0.5rem;
	font-size: 0.9375rem;
	color: var(--text);
	resize: none;
	background: var(--background);
	border: 1px solid var(--border);
	border-radius: var(--radius);
}

.paragraph-input:focus {
	outline: none;
	border-color: var(--border-focus);
}
</style>
