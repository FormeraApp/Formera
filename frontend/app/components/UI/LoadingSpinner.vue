<script lang="ts" setup>
const props = withDefaults(defineProps<{
	size?: "sm" | "md" | "lg";
	label?: string;
}>(), {
	size: "md",
	label: "",
});

const { t } = useI18n();
const labelText = computed(() => props.label || t("common.loading"));
</script>

<template>
	<div class="spinner-container" :class="`spinner-${size}`" role="status" :aria-label="labelText">
		<div class="spinner" aria-hidden="true">
			<div class="spinner-circle"></div>
		</div>
		<span class="sr-only">{{ labelText }}</span>
	</div>
</template>

<style scoped>
.spinner-container {
	display: inline-flex;
	align-items: center;
	justify-content: center;
}

.spinner {
	animation: rotate 1s linear infinite;
}

.spinner-sm .spinner {
	width: 1rem;
	height: 1rem;
}

.spinner-md .spinner {
	width: 1.5rem;
	height: 1.5rem;
}

.spinner-lg .spinner {
	width: 2.5rem;
	height: 2.5rem;
}

.spinner-circle {
	width: 100%;
	height: 100%;
	border: 2px solid var(--border);
	border-top-color: var(--primary);
	border-radius: 50%;
}

.spinner-sm .spinner-circle {
	border-width: 2px;
}

.spinner-lg .spinner-circle {
	border-width: 3px;
}

@keyframes rotate {
	to {
		transform: rotate(360deg);
	}
}

.sr-only {
	position: absolute;
	width: 1px;
	height: 1px;
	padding: 0;
	margin: -1px;
	overflow: hidden;
	clip: rect(0, 0, 0, 0);
	white-space: nowrap;
	border: 0;
}
</style>
