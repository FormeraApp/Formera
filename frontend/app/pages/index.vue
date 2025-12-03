<script lang="ts" setup>
definePageMeta({
	layout: "auth",
});

const authStore = useAuthStore();

onMounted(async () => {
	if (authStore.isLoading) {
		await authStore.init();
	}

	if (authStore.token) {
		navigateTo("/forms", { replace: true });
	} else {
		navigateTo("/login", { replace: true });
	}
});
</script>

<template>
	<div class="loading-container">
		<div class="spinner" />
	</div>
</template>

<style scoped>
.loading-container {
	display: flex;
	justify-content: center;
	align-items: center;
	height: 100vh;
}

.spinner {
	width: 40px;
	height: 40px;
	border: 3px solid var(--border);
	border-top-color: var(--primary);
	border-radius: 50%;
	animation: spin 1s linear infinite;
}

@keyframes spin {
	to {
		transform: rotate(360deg);
	}
}
</style>
