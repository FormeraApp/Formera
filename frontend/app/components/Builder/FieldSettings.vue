<script lang="ts" setup>
const props = defineProps<{
	field: FormField;
}>();

const emit = defineEmits(["update:field"]);

const fieldMeta = computed(() => FIELD_META[props.field.type]);

const isLayoutField = computed(() => {
	const layoutTypes: LayoutFieldType[] = ["section", "pagebreak", "divider", "heading", "paragraph", "image"];
	return layoutTypes.includes(props.field.type as LayoutFieldType);
});

const hasOptions = computed(() => {
	return ["select", "radio", "checkbox", "dropdown"].includes(props.field.type);
});

const isRatingOrScale = computed(() => {
	return ["rating", "scale"].includes(props.field.type);
});

const supportsTextValidation = computed(() => {
	return ["text", "textarea", "richtext"].includes(props.field.type);
});

const supportsNumberValidation = computed(() => {
	return props.field.type === "number";
});

const showValidationSection = computed(() => {
	return supportsTextValidation.value || supportsNumberValidation.value;
});

const update = <K extends keyof FormField>(key: K, value: FormField[K]) => {
	emit("update:field", { ...props.field, [key]: value });
};

const updateValidation = (key: string, value: unknown) => {
	const validation = { ...props.field.validation, [key]: value };
	if (value === "" || value === undefined || value === null) {
		delete validation[key as keyof typeof validation];
	}
	emit("update:field", { ...props.field, validation });
};

const addOption = () => {
	const options = [...(props.field.options || []), `Option ${(props.field.options || []).length + 1}`];
	emit("update:field", { ...props.field, options });
};

const updateOption = (index: number, value: string) => {
	const options = [...(props.field.options || [])];
	options[index] = value;
	emit("update:field", { ...props.field, options });
};

const deleteOption = (index: number) => {
	const options = (props.field.options || []).filter((_, i) => i !== index);
	emit("update:field", { ...props.field, options });
};
</script>

<template>
	<div class="field-settings">
		<div class="settings-header">
			<UISysIcon :icon="fieldMeta.icon" />
			<span>{{ $t(`fields.${field.type}.label`) }}</span>
		</div>

		<div class="settings-form">
			<!-- Common: Label -->
			<div v-if="!['divider', 'pagebreak'].includes(field.type)" class="form-group">
				<label class="label">{{ $t("builder.fieldSettings.label") }}</label>
				<input
					:value="field.label"
					class="input"
					type="text"
					@input="update('label', ($event.target as HTMLInputElement).value)"
				/>
			</div>

			<!-- Common: Description (not for layout) -->
			<div v-if="!isLayoutField" class="form-group">
				<label class="label">{{ $t("builder.fieldSettings.helpText") }}</label>
				<input
					:value="field.description || ''"
					class="input"
					:placeholder="$t('builder.fieldSettings.helpTextPlaceholder')"
					type="text"
					@input="update('description', ($event.target as HTMLInputElement).value)"
				/>
			</div>

			<!-- Common: Placeholder (for input fields) -->
			<div v-if="['text', 'textarea', 'number', 'email', 'phone', 'url', 'date', 'time'].includes(field.type)" class="form-group">
				<label class="label">{{ $t("builder.fieldSettings.placeholder") }}</label>
				<input
					:value="field.placeholder || ''"
					class="input"
					type="text"
					@input="update('placeholder', ($event.target as HTMLInputElement).value)"
				/>
			</div>

			<!-- Common: Required (not for layout) -->
			<div v-if="!isLayoutField" class="form-group">
				<label class="checkbox-label">
					<input
						:checked="field.required"
						type="checkbox"
						@change="update('required', ($event.target as HTMLInputElement).checked)"
					/>
					{{ $t("builder.fieldSettings.required") }}
				</label>
			</div>

			<!-- Options for select/radio/checkbox -->
			<div v-if="hasOptions" class="form-group">
				<label class="label">{{ $t("builder.fieldSettings.options") }}</label>
				<div v-for="(option, index) in field.options || []" :key="index" class="option-row">
					<input
						:value="option"
						class="input"
						type="text"
						@input="updateOption(index, ($event.target as HTMLInputElement).value)"
					/>
					<button
						class="option-delete"
						:aria-label="$t('builder.fieldSettings.removeOption')"
						@click="deleteOption(index)"
					>
						<UISysIcon icon="fa-solid fa-trash" />
					</button>
				</div>
				<button class="btn btn-sm btn-secondary" @click="addOption">
					<UISysIcon icon="fa-solid fa-plus" />
					{{ $t("builder.fieldSettings.addOption") }}
				</button>
			</div>

			<!-- Rating/Scale Settings -->
			<div v-if="isRatingOrScale" class="form-group">
				<label class="label">{{ $t("builder.fieldSettings.range") }}</label>
				<div class="range-inputs">
					<div>
						<span class="range-label">{{ $t("builder.fieldSettings.min") }}</span>
						<input
							:value="field.minValue || 1"
							class="input input-sm"
							min="0"
							type="number"
							@input="update('minValue', Number(($event.target as HTMLInputElement).value))"
						/>
					</div>
					<div>
						<span class="range-label">{{ $t("builder.fieldSettings.max") }}</span>
						<input
							:value="field.maxValue || (field.type === 'rating' ? 5 : 10)"
							class="input input-sm"
							min="1"
							type="number"
							@input="update('maxValue', Number(($event.target as HTMLInputElement).value))"
						/>
					</div>
				</div>
			</div>

			<!-- Scale Labels -->
			<div v-if="field.type === 'scale'" class="form-group">
				<label class="label">{{ $t("builder.fieldSettings.labels") }}</label>
				<div class="scale-labels">
					<input
						:value="field.minLabel || ''"
						class="input"
						:placeholder="$t('builder.fieldSettings.minLabelPlaceholder')"
						type="text"
						@input="update('minLabel', ($event.target as HTMLInputElement).value)"
					/>
					<input
						:value="field.maxLabel || ''"
						class="input"
						:placeholder="$t('builder.fieldSettings.maxLabelPlaceholder')"
						type="text"
						@input="update('maxLabel', ($event.target as HTMLInputElement).value)"
					/>
				</div>
			</div>

			<!-- Heading Level -->
			<div v-if="field.type === 'heading'" class="form-group">
				<label class="label">{{ $t("builder.fieldSettings.size") }}</label>
				<select
					:value="field.headingLevel || 2"
					class="input"
					@change="update('headingLevel', Number(($event.target as HTMLSelectElement).value) as 1 | 2 | 3 | 4)"
				>
					<option :value="1">{{ $t("builder.fieldSettings.headingLarge") }}</option>
					<option :value="2">{{ $t("builder.fieldSettings.headingMedium") }}</option>
					<option :value="3">{{ $t("builder.fieldSettings.headingSmall") }}</option>
					<option :value="4">{{ $t("builder.fieldSettings.headingVerySmall") }}</option>
				</select>
			</div>

			<!-- Section Settings -->
			<div v-if="field.type === 'section'" class="form-group">
				<label class="checkbox-label">
					<input
						:checked="field.collapsible"
						type="checkbox"
						@change="update('collapsible', ($event.target as HTMLInputElement).checked)"
					/>
					{{ $t("builder.fieldSettings.collapsible") }}
				</label>
			</div>

			<!-- File Upload Settings -->
			<div v-if="field.type === 'file'" class="form-group">
				<label class="label">{{ $t("builder.fieldSettings.allowedFileTypes") }}</label>
				<input
					:value="field.allowedTypes?.join(', ') || ''"
					class="input"
					:placeholder="$t('builder.fieldSettings.allowedFileTypesPlaceholder')"
					type="text"
					@input="update('allowedTypes', ($event.target as HTMLInputElement).value.split(',').map(t => t.trim()).filter(Boolean))"
				/>
				<p class="form-hint">{{ $t("builder.fieldSettings.allowedFileTypesHint") }}</p>
			</div>

			<div v-if="field.type === 'file'" class="form-group">
				<label class="label">{{ $t("builder.fieldSettings.maxFileSize") }}</label>
				<input
					:value="field.maxFileSize || 10"
					class="input"
					min="1"
					type="number"
					@input="update('maxFileSize', Number(($event.target as HTMLInputElement).value))"
				/>
			</div>

			<div v-if="field.type === 'file'" class="form-group">
				<label class="checkbox-label">
					<input
						:checked="field.multiple"
						type="checkbox"
						@change="update('multiple', ($event.target as HTMLInputElement).checked)"
					/>
					{{ $t("builder.fieldSettings.multipleFiles") }}
				</label>
			</div>

			<!-- Validation Section -->
			<div v-if="showValidationSection" class="validation-section">
				<div class="section-title">
					<UISysIcon icon="fa-solid fa-shield-check" />
					{{ $t("builder.fieldSettings.validation") }}
				</div>

				<!-- Text Validation: Min/Max Length -->
				<template v-if="supportsTextValidation">
					<div class="form-group">
						<label class="label">{{ $t("builder.fieldSettings.charLength") }}</label>
						<div class="range-inputs">
							<div>
								<span class="range-label">{{ $t("builder.fieldSettings.min") }}</span>
								<input
									:value="field.validation?.minLength || ''"
									class="input input-sm"
									min="0"
									placeholder="0"
									type="number"
									@input="updateValidation('minLength', ($event.target as HTMLInputElement).value ? Number(($event.target as HTMLInputElement).value) : undefined)"
								/>
							</div>
							<div>
								<span class="range-label">{{ $t("builder.fieldSettings.max") }}</span>
								<input
									:value="field.validation?.maxLength || ''"
									class="input input-sm"
									min="1"
									placeholder="∞"
									type="number"
									@input="updateValidation('maxLength', ($event.target as HTMLInputElement).value ? Number(($event.target as HTMLInputElement).value) : undefined)"
								/>
							</div>
						</div>
					</div>

					<div class="form-group">
						<label class="label">{{ $t("builder.fieldSettings.regexPattern") }}</label>
						<input
							:value="field.validation?.pattern || ''"
							class="input input-mono"
							:placeholder="$t('builder.fieldSettings.regexPatternPlaceholder')"
							type="text"
							@input="updateValidation('pattern', ($event.target as HTMLInputElement).value)"
						/>
						<p class="form-hint">{{ $t("builder.fieldSettings.regexPatternHint") }}</p>
					</div>

					<div v-if="field.validation?.pattern" class="form-group">
						<label class="label">{{ $t("builder.fieldSettings.patternErrorMessage") }}</label>
						<input
							:value="field.validation?.patternMessage || ''"
							class="input"
							:placeholder="$t('builder.fieldSettings.patternErrorPlaceholder')"
							type="text"
							@input="updateValidation('patternMessage', ($event.target as HTMLInputElement).value)"
						/>
					</div>
				</template>

				<!-- Number Validation: Min/Max Value -->
				<template v-if="supportsNumberValidation">
					<div class="form-group">
						<label class="label">{{ $t("builder.fieldSettings.valueRange") }}</label>
						<div class="range-inputs">
							<div>
								<span class="range-label">{{ $t("builder.fieldSettings.min") }}</span>
								<input
									:value="field.validation?.min ?? ''"
									class="input input-sm"
									placeholder="-∞"
									type="number"
									@input="updateValidation('min', ($event.target as HTMLInputElement).value !== '' ? Number(($event.target as HTMLInputElement).value) : undefined)"
								/>
							</div>
							<div>
								<span class="range-label">{{ $t("builder.fieldSettings.max") }}</span>
								<input
									:value="field.validation?.max ?? ''"
									class="input input-sm"
									placeholder="∞"
									type="number"
									@input="updateValidation('max', ($event.target as HTMLInputElement).value !== '' ? Number(($event.target as HTMLInputElement).value) : undefined)"
								/>
							</div>
						</div>
					</div>
				</template>

				<!-- Custom Error Messages -->
				<div class="form-group">
					<label class="label">{{ $t("builder.fieldSettings.requiredErrorMessage") }}</label>
					<input
						:value="field.validation?.requiredMessage || ''"
						class="input"
						:placeholder="$t('builder.fieldSettings.requiredErrorPlaceholder')"
						type="text"
						@input="updateValidation('requiredMessage', ($event.target as HTMLInputElement).value)"
					/>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.field-settings {
	display: flex;
	flex-direction: column;
	height: 100%;
}

.settings-header {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	padding-bottom: 1rem;
	margin-bottom: 1rem;
	font-weight: 600;
	color: var(--text);
	border-bottom: 1px solid var(--border);
}

.settings-header i {
	color: var(--primary);
}

.settings-form {
	display: flex;
	flex-direction: column;
	gap: 1rem;
	overflow-y: auto;
}

.form-group {
	display: flex;
	flex-direction: column;
	gap: 0.375rem;
}

.label {
	font-size: 0.8125rem;
	font-weight: 500;
	color: var(--text);
}

.checkbox-label {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	font-size: 0.875rem;
	cursor: pointer;
}

.checkbox-label input {
	width: 1rem;
	height: 1rem;
	cursor: pointer;
}

.option-row {
	display: flex;
	gap: 0.5rem;
	margin-bottom: 0.5rem;
}

.option-delete {
	flex-shrink: 0;
	padding: 0.375rem 0.5rem;
	color: var(--text-secondary);
	cursor: pointer;
	background: none;
	border: 1px solid var(--border);
	border-radius: var(--radius);
}

.option-delete:hover {
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
	border-color: var(--error);
}

.range-inputs {
	display: flex;
	gap: 1rem;
}

.range-inputs > div {
	flex: 1;
}

.range-label {
	display: block;
	margin-bottom: 0.25rem;
	font-size: 0.75rem;
	color: var(--text-secondary);
}

.input-sm {
	padding: 0.375rem 0.5rem;
	font-size: 0.875rem;
}

.scale-labels {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.form-hint {
	font-size: 0.75rem;
	color: var(--text-secondary);
}

/* Validation Section */
.validation-section {
	padding-top: 1rem;
	margin-top: 0.5rem;
	border-top: 1px solid var(--border);
}

.section-title {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	margin-bottom: 1rem;
	font-size: 0.8125rem;
	font-weight: 600;
	color: var(--text-secondary);
}

.section-title i {
	color: var(--primary);
}

.input-mono {
	font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
	font-size: 0.8125rem;
}
</style>
