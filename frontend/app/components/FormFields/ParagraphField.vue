<script lang="ts" setup>
withDefaults(
	defineProps<{
		content?: string;
	}>(),
	{
		content: "",
	}
);
</script>

<template>
	<p class="field-paragraph">{{ content }}</p>
</template>

<style scoped>
.field-paragraph {
	margin: 0.75rem 0;
	font-size: 0.9375rem;
	line-height: 1.6;
	color: var(--text-secondary);
}
</style>
