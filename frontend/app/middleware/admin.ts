export default defineNuxtRouteMiddleware(() => {
	const authStore = useAuthStore();

	// Check if user is admin
	if (!authStore.user || authStore.user.role !== "admin") {
		return navigateTo("/forms");
	}
});
