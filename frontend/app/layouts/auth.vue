<script lang="ts" setup>
const setupStore = useSetupStore();

const backgroundStyle = computed(() => {
	if (setupStore.loginBackgroundDisplayURL) {
		return {
			backgroundImage: `url(${setupStore.loginBackgroundDisplayURL})`,
			backgroundSize: "cover",
			backgroundPosition: "center",
			backgroundRepeat: "no-repeat",
		};
	}
	return {
		background: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
	};
});
</script>

<template>
	<div class="auth-layout" :style="backgroundStyle">
		<slot />
	</div>
</template>

<style scoped>
.auth-layout {
	position: relative;
	display: flex;
	align-items: center;
	justify-content: center;
	min-height: 100vh;
	padding: 1rem;
}

.loading-container {
	display: flex;
	align-items: center;
	justify-content: center;
}

.loading-spinner {
	width: 48px;
	height: 48px;
	border: 4px solid rgba(255, 255, 255, 0.3);
	border-top-color: white;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

@keyframes spin {
	to {
		transform: rotate(360deg);
	}
}
</style>
