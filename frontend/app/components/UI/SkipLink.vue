<script lang="ts" setup>
const props = withDefaults(
	defineProps<{
		target?: string;
		label?: string;
	}>(),
	{
		target: "#main-content",
		label: "",
	}
);

const { t } = useI18n();
const labelText = computed(() => props.label || t("skipLink.label"));
</script>

<template>
	<a :href="target" class="skip-link">
		{{ labelText }}
	</a>
</template>

<style scoped>
.skip-link {
	position: fixed;
	top: -100%;
	left: 50%;
	z-index: 10000;
	padding: 0.75rem 1.5rem;
	font-weight: 500;
	color: white;
	background: var(--primary);
	border-radius: var(--radius);
	transform: translateX(-50%);
	transition: top 0.2s;
}

.skip-link:focus {
	top: 1rem;
	outline: 2px solid white;
	outline-offset: 2px;
}
</style>
