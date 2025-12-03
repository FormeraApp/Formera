<script lang="ts" setup>
const setupStore = useSetupStore();
</script>

<template>
	<div class="public-layout">
		<main class="public-main">
			<slot />
		</main>

		<footer v-if="setupStore.footerLinks.length > 0" class="public-footer">
			<nav class="footer-links">
				<a
					v-for="link in setupStore.footerLinks"
					:key="link.url"
					:href="link.url"
					target="_blank"
					rel="noopener noreferrer"
					class="footer-link"
				>
					{{ link.label }}
				</a>
			</nav>
		</footer>
	</div>
</template>

<style scoped>
.public-layout {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
}

.public-main {
	flex: 1;
}

.public-footer {
	position: relative;
	z-index: 10;
	padding: 1rem;
	text-align: center;
}

.footer-links {
	display: flex;
	flex-wrap: wrap;
	gap: 1.5rem;
	justify-content: center;
}

.footer-link {
	font-size: 0.8125rem;
	color: var(--text-secondary);
	text-decoration: none;
	transition: color 0.15s ease;
}

.footer-link:hover {
	color: var(--primary);
	text-decoration: underline;
}
</style>
