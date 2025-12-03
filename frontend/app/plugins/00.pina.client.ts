export default defineNuxtPlugin(async (nuxtApp) => {
	const authStore = useAuthStore();
	const setupStore = useSetupStore();
	const themeStore = useThemeStore();

	// Load setup status and auth in parallel (blocking - wait before navigation)
	await Promise.all([setupStore.loadStatus(), authStore.init()]).catch((error) => {
		console.error("Failed to initialize stores:", error);
	});

	// Initialize theme AFTER setupStore has loaded (to use backend settings)
	themeStore.init();

	// Apply language from settings
	const { setLocale, locale } = nuxtApp.$i18n;
	if (setupStore.language && setupStore.language !== locale.value) {
		setLocale(setupStore.language);
	}
});
