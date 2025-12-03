<script lang="ts" setup>
const authStore = useAuthStore();
const setupStore = useSetupStore();
const router = useRouter();
const route = useRoute();
const localePath = useLocalePath();

const isMobileMenuOpen = ref(false);

// Helper to extract locale prefix from path (e.g., '/de', '/fr')
function getLocalePrefix(path: string): string | null {
	const match = path.match(/^\/([a-zA-Z-]{2,5})(?=\/|$)/);
	return match ? `/${match[1]}` : null;
}

// Determine layout type based on route
const layoutType = computed(() => {
	const path = route.path;
	const localePrefix = getLocalePrefix(path);

	// Build base paths for matching
	const base = localePrefix ? localePrefix : "";

	// Auth pages: login, register, setup, index (landing)
	if (
		path === "/" + (localePrefix ? localePrefix.slice(1) : "") || // e.g., "/de"
		path === base + "/login" ||
		path === base + "/register" ||
		path === base + "/setup" ||
		path === "/" ||
		path === "/login" ||
		path === "/register" ||
		path === "/setup"
	) {
		return "auth";
	}
	// Public form pages
	if (
		path.startsWith(base + "/f/") ||
		path.startsWith("/f/")
	) {
		return "public";
	}
	// Dashboard pages (default)
	return "dashboard";
});

const isAuthLayout = computed(() => layoutType.value === "auth");
const isPublicLayout = computed(() => layoutType.value === "public");
const isDashboardLayout = computed(() => layoutType.value === "dashboard");

// Auth layout background style
const backgroundStyle = computed(() => {
	if (!isAuthLayout.value) return {};
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

const handleLogout = () => {
	authStore.logout();
	router.push(localePath("/login"));
};

const toggleMobileMenu = () => {
	isMobileMenuOpen.value = !isMobileMenuOpen.value;
};

const closeMobileMenu = () => {
	isMobileMenuOpen.value = false;
};

// Close mobile menu on route change
watch(
	() => route.path,
	() => {
		closeMobileMenu();
	}
);
</script>

<template>
	<!-- Auth Layout (login, register, setup) -->
	<div v-if="isAuthLayout" class="auth-layout" :style="backgroundStyle">
		<NuxtPage />
	</div>

	<!-- Public Layout (public forms) -->
	<div v-else-if="isPublicLayout" class="public-layout">
		<main class="public-main">
			<NuxtPage />
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

	<!-- Dashboard Layout (authenticated pages) -->
	<div v-else class="layout">
		<header class="header">
			<div class="header-container">
				<!-- Logo -->
				<NuxtLink class="logo" :to="localePath('/forms')">
					<template v-if="setupStore.logoURL">
						<img :src="setupStore.logoDisplayURL" :alt="setupStore.appName" class="logo-image" />
						<span v-if="setupStore.logoShowText" class="logo-text">{{ setupStore.appName }}</span>
					</template>
					<template v-else>
						<div class="logo-icon">
							<UISysIcon icon="fa-solid fa-file-lines" />
						</div>
						<span class="logo-text">{{ setupStore.appName }}</span>
					</template>
				</NuxtLink>

				<!-- Desktop Navigation -->
				<nav v-if="authStore.user" class="nav-desktop">
					<NuxtLink class="nav-link" :to="localePath('/forms')">
						<UISysIcon icon="fa-solid fa-folder" />
						<span>{{ $t("nav.forms") }}</span>
					</NuxtLink>
				</nav>

				<!-- Desktop Actions -->
				<div v-if="authStore.user" class="header-actions">
					<div class="user-info">
						<div class="user-avatar">
							{{ authStore.user.name.charAt(0).toUpperCase() }}
						</div>
						<span class="user-name">{{ authStore.user.name }}</span>
					</div>

					<div class="action-buttons">
						<ThemeToggle />
						<NuxtLink class="action-btn" :to="localePath('/settings')" :title="$t('nav.settings')">
							<UISysIcon icon="fa-solid fa-gear" />
						</NuxtLink>
						<button class="action-btn logout" :title="$t('auth.logout')" @click="handleLogout">
							<UISysIcon icon="fa-solid fa-right-from-bracket" />
						</button>
					</div>
				</div>

				<!-- Mobile Menu Button -->
				<button v-if="authStore.user" class="mobile-menu-btn" @click="toggleMobileMenu">
					<UISysIcon :icon="isMobileMenuOpen ? 'fa-solid fa-xmark' : 'fa-solid fa-bars'" />
				</button>
			</div>

			<!-- Mobile Menu -->
			<Transition name="slide">
				<div v-if="isMobileMenuOpen && authStore.user" class="mobile-menu">
					<div class="mobile-user">
						<div class="user-avatar">
							{{ authStore.user.name.charAt(0).toUpperCase() }}
						</div>
						<div class="mobile-user-info">
							<span class="user-name">{{ authStore.user.name }}</span>
							<span class="user-email">{{ authStore.user.email }}</span>
						</div>
					</div>

					<nav class="mobile-nav">
						<NuxtLink class="mobile-nav-link" :to="localePath('/forms')" @click="closeMobileMenu">
							<UISysIcon icon="fa-solid fa-folder" />
							<span>{{ $t("nav.forms") }}</span>
						</NuxtLink>
						<NuxtLink class="mobile-nav-link" :to="localePath('/settings')" @click="closeMobileMenu">
							<UISysIcon icon="fa-solid fa-gear" />
							<span>{{ $t("nav.settings") }}</span>
						</NuxtLink>
					</nav>

					<div class="mobile-footer">
						<ThemeToggle />
						<button class="mobile-logout" @click="handleLogout">
							<UISysIcon icon="fa-solid fa-right-from-bracket" />
							<span>{{ $t("auth.logout") }}</span>
						</button>
					</div>
				</div>
			</Transition>
		</header>

		<main class="main">
			<NuxtPage />
		</main>

		<footer class="footer">
			<div class="footer-container">
				<span class="footer-app-name">{{ setupStore.appName }}</span>

				<nav v-if="setupStore.footerLinks.length > 0" class="footer-links">
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
			</div>
		</footer>
	</div>
</template>

<style scoped>
/* ========================================
   Auth Layout Styles
   ======================================== */
.auth-layout {
	position: relative;
	display: flex;
	align-items: center;
	justify-content: center;
	min-height: 100vh;
	padding: 1rem;
}

/* ========================================
   Public Layout Styles
   ======================================== */
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

/* ========================================
   Dashboard Layout Styles
   ======================================== */
.layout {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
}

/* Header */
.header {
	position: sticky;
	top: 0;
	z-index: 100;
	background: var(--surface);
	border-bottom: 1px solid var(--border);
}

.header-container {
	display: flex;
	align-items: center;
	justify-content: space-between;
	max-width: 1200px;
	height: 64px;
	padding: 0 1.5rem;
	margin: 0 auto;
}

/* Logo */
.logo {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	text-decoration: none;
	transition: opacity 0.15s ease;
}

.logo:hover {
	opacity: 0.8;
	text-decoration: none;
}

.logo-icon {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 36px;
	height: 36px;
	font-size: 1rem;
	color: white;
	background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
	border-radius: var(--radius);
}

.logo-text {
	font-size: 1.125rem;
	font-weight: 600;
	color: var(--text);
}

.logo-image {
	max-height: 36px;
	width: auto;
	object-fit: contain;
}

/* Desktop Navigation */
.nav-desktop {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	margin-left: 2rem;
}

.nav-link {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem 1rem;
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text-secondary);
	text-decoration: none;
	border-radius: var(--radius);
	transition: all 0.15s ease;
}

.nav-link:hover {
	color: var(--text);
	background: var(--surface-hover);
	text-decoration: none;
}

.nav-link.router-link-active {
	color: var(--primary);
	background: rgba(99, 102, 241, 0.1);
}

/* Header Actions */
.header-actions {
	display: flex;
	align-items: center;
	gap: 1.5rem;
	margin-left: auto;
}

.user-info {
	display: flex;
	align-items: center;
	gap: 0.75rem;
}

.user-avatar {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 32px;
	height: 32px;
	font-size: 0.875rem;
	font-weight: 600;
	color: white;
	background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
	border-radius: 50%;
}

.user-name {
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text);
}

.action-buttons {
	display: flex;
	align-items: center;
	gap: 0.25rem;
}

.action-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 36px;
	height: 36px;
	color: var(--text-secondary);
	text-decoration: none;
	cursor: pointer;
	background: none;
	border: none;
	border-radius: var(--radius);
	transition: all 0.15s ease;
}

.action-btn:hover {
	color: var(--text);
	background: var(--surface-hover);
}

.action-btn.logout:hover {
	color: var(--error);
}

/* Mobile Menu Button */
.mobile-menu-btn {
	display: none;
	align-items: center;
	justify-content: center;
	width: 40px;
	height: 40px;
	font-size: 1.25rem;
	color: var(--text);
	cursor: pointer;
	background: none;
	border: none;
	border-radius: var(--radius);
}

.mobile-menu-btn:hover {
	background: var(--surface-hover);
}

/* Mobile Menu */
.mobile-menu {
	display: none;
	flex-direction: column;
	padding: 1rem;
	background: var(--surface);
	border-top: 1px solid var(--border);
}

.mobile-user {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	padding: 1rem;
	margin-bottom: 0.5rem;
	background: var(--surface-hover);
	border-radius: var(--radius-lg);
}

.mobile-user .user-avatar {
	width: 40px;
	height: 40px;
	font-size: 1rem;
}

.mobile-user-info {
	display: flex;
	flex-direction: column;
}

.mobile-user-info .user-name {
	font-weight: 600;
}

.mobile-user-info .user-email {
	font-size: 0.8125rem;
	color: var(--text-secondary);
}

.mobile-nav {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
	padding: 0.5rem 0;
}

.mobile-nav-link {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	padding: 0.875rem 1rem;
	font-size: 0.9375rem;
	font-weight: 500;
	color: var(--text);
	text-decoration: none;
	border-radius: var(--radius);
	transition: background 0.15s ease;
}

.mobile-nav-link:hover {
	background: var(--surface-hover);
	text-decoration: none;
}

.mobile-nav-link.router-link-active {
	color: var(--primary);
	background: rgba(99, 102, 241, 0.1);
}

.mobile-footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-top: 1rem;
	margin-top: 0.5rem;
	border-top: 1px solid var(--border);
}

.mobile-logout {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.625rem 1rem;
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--error);
	cursor: pointer;
	background: none;
	border: none;
	border-radius: var(--radius);
	transition: background 0.15s ease;
}

.mobile-logout:hover {
	background: rgba(239, 68, 68, 0.1);
}

/* Main */
.main {
	flex: 1;
	padding: 2rem 1rem;
}

/* Footer (shared) */
.footer {
	padding: 1rem 1.5rem;
	background: var(--surface);
	border-top: 1px solid var(--border);
}

.footer-container {
	display: flex;
	gap: 1.5rem;
	align-items: center;
	justify-content: center;
	max-width: 1200px;
	margin: 0 auto;
}

.footer-app-name {
	font-size: 0.8125rem;
	font-weight: 500;
	color: var(--text-secondary);
}

.footer-links {
	display: flex;
	flex-wrap: wrap;
	gap: 1.5rem;
	align-items: center;
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

/* Transitions */
.slide-enter-active,
.slide-leave-active {
	transition: all 0.2s ease;
}

.slide-enter-from,
.slide-leave-to {
	opacity: 0;
	transform: translateY(-10px);
}

/* Responsive */
@media (max-width: 768px) {
	.header-container {
		padding: 0 1rem;
	}

	.nav-desktop,
	.header-actions {
		display: none;
	}

	.mobile-menu-btn {
		display: flex;
	}

	.mobile-menu {
		display: flex;
	}

	.main {
		padding: 1.5rem 1rem;
	}

	.footer-container {
		gap: 0.75rem;
	}

	.footer-links {
		gap: 1rem;
	}
}

@media (min-width: 769px) {
	.footer-container {
		flex-direction: row;
		justify-content: center;
	}

	.footer-container:not(.footer-centered) {
		justify-content: space-between;
	}
}
</style>
