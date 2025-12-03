<script lang="ts" setup>
withDefaults(defineProps<{
	label: string;
	description?: string;
}>(), {
	description: "",
});
</script>

<template>
	<div class="field-section">
		<h3 class="section-title">{{ label }}</h3>
		<p v-if="description" class="section-description">{{ description }}</p>
	</div>
</template>

<style scoped>
.field-section {
	margin: 1.5rem 0 1rem;
	padding-bottom: 0.75rem;
	border-bottom: 2px solid var(--border);
}

.section-title {
	margin: 0 0 0.25rem;
	font-size: 1.125rem;
	font-weight: 600;
	color: var(--text);
}

.section-description {
	margin: 0;
	font-size: 0.875rem;
	color: var(--text-secondary);
}
</style>
