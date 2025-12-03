export default defineNuxtPlugin(() => {
	const authStore = useAuthStore();
	const setupStore = useSetupStore();
	const themeStore = useThemeStore();

	// Initialize theme (sync)
	themeStore.init();

	// Load setup status and auth in parallel (non-blocking)
	Promise.all([setupStore.loadStatus(), authStore.init()]).catch((error) => {
		console.error("Failed to initialize stores:", error);
	});
});
