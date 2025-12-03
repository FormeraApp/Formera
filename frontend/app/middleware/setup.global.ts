export default defineNuxtRouteMiddleware((to) => {
	const setupStore = useSetupStore();

	// Skip middleware during loading
	if (setupStore.isLoading) {
		return;
	}

	// If setup is required, redirect all routes to setup
	if (setupStore.setupRequired && to.path !== "/setup") {
		return navigateTo("/setup");
	}

	// If setup is complete, redirect setup page to login
	if (!setupStore.setupRequired && to.path === "/setup") {
		return navigateTo("/login");
	}
});
