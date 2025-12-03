<script lang="ts" setup>
withDefaults(defineProps<{
	icon?: string;
	title: string;
	description?: string;
}>(), {
	icon: "fa-solid fa-inbox",
});
</script>

<template>
	<div class="empty-state">
		<div class="empty-icon">
			<UISysIcon :icon="icon" aria-hidden="true" />
		</div>
		<h2 class="empty-title">{{ title }}</h2>
		<p v-if="description" class="empty-description">{{ description }}</p>
		<div v-if="$slots.default" class="empty-actions">
			<slot />
		</div>
	</div>
</template>

<style scoped>
.empty-state {
	padding: 3rem 1.5rem;
	text-align: center;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
}

.empty-icon {
	margin-bottom: 1rem;
	font-size: 3rem;
	color: var(--text-secondary);
}

.empty-title {
	margin-bottom: 0.5rem;
	font-size: 1.25rem;
	font-weight: 600;
	color: var(--text);
}

.empty-description {
	margin-bottom: 1.5rem;
	font-size: 0.9375rem;
	color: var(--text-secondary);
}

.empty-actions {
	display: flex;
	gap: 0.75rem;
	justify-content: center;
}
</style>
