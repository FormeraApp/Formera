<script lang="ts" setup>
withDefaults(
	defineProps<{
		label: string;
		level?: 1 | 2 | 3 | 4;
	}>(),
	{
		level: 2,
	}
);
</script>

<template>
	<component :is="`h${level}`" class="field-heading">
		{{ label }}
	</component>
</template>

<style scoped>
.field-heading {
	margin: 1.5rem 0 0.75rem;
	font-weight: 600;
	color: var(--text);
}
</style>
