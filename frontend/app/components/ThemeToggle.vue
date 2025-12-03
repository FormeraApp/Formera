<script lang="ts" setup>
const { t } = useI18n();
const themeStore = useThemeStore();
</script>

<template>
	<button class="toggle" :aria-label="themeStore.theme === 'light' ? t('theme.enableDark') : t('theme.enableLight')" :title="themeStore.theme === 'light' ? t('theme.dark') : t('theme.light')" @click="themeStore.toggleTheme">
		<span class="sr-only">Theme Toggle</span>

		<UISysIcon icon="fa-solid fa-moon" v-if="themeStore.theme === 'light'"></UISysIcon>
		<UISysIcon icon="fa-solid fa-sun" v-else></UISysIcon>
	</button>
</template>

<style scoped>
.toggle {
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0.375rem;
	color: var(--text-secondary);
	cursor: pointer;
	background: none;
	border: none;
	border-radius: var(--radius);
	transition: all 0.2s;
}

.toggle:hover {
	color: var(--primary);
	background-color: var(--surface-hover);
}
</style>
