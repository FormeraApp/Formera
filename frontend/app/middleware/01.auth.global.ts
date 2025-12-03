export default defineNuxtRouteMiddleware((to) => {
	const authStore = useAuthStore();
	const setupStore = useSetupStore();

	// Skip middleware during loading or if setup is required (setup.global.ts handles this)
	if (authStore.isLoading || setupStore.setupRequired) {
		return;
	}

	// Guest-only routes (redirect to /forms if already logged in)
	const guestRoutes = ["/login", "/register"];
	const isGuestRoute = guestRoutes.includes(to.path);

	if (isGuestRoute && authStore.user) {
		return navigateTo("/forms");
	}

	// Protected routes
	const protectedRoutes = ["/forms", "/settings"];
	const isProtected = protectedRoutes.some((route) => to.path.startsWith(route));

	if (isProtected && !authStore.user) {
		return navigateTo("/login");
	}
});
