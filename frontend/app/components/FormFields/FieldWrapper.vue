<script lang="ts" setup>
withDefaults(defineProps<{
	label: string;
	description?: string;
	required?: boolean;
	error?: string;
	fieldId?: string;
}>(), {
	description: "",
	required: false,
	error: "",
	fieldId: "",
});
</script>

<template>
	<div class="field" :class="{ 'field-error': error }">
		<label class="field-label" :for="fieldId">
			{{ label }}
			<span v-if="required" class="required">*</span>
		</label>
		<p v-if="description" class="field-description">{{ description }}</p>
		<slot />
		<p v-if="error" class="field-error-message" role="alert">
			<UISysIcon icon="fa-solid fa-circle-exclamation" />
			{{ error }}
		</p>
	</div>
</template>

<style scoped>
.field {
	margin-bottom: 1.5rem;
}

.field:last-child {
	margin-bottom: 0;
}

.field-label {
	display: block;
	margin-bottom: 0.5rem;
	font-weight: 500;
}

.required {
	margin-left: 0.25rem;
	color: var(--error);
}

.field-description {
	margin: 0 0 0.5rem;
	font-size: 0.8125rem;
	color: var(--text-secondary);
}

.field-error-message {
	display: flex;
	gap: 0.375rem;
	align-items: center;
	margin: 0.375rem 0 0;
	font-size: 0.8125rem;
	color: var(--error);
}

.field-error :deep(.input) {
	border-color: var(--error);
}

.field-error :deep(.input:focus) {
	border-color: var(--error);
	box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
}
</style>
