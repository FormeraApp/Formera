import { defineStore } from "pinia";
import type { FooterLink } from "~~/shared/types";

export const useSetupStore = defineStore("setup", () => {
	const setupRequired = ref(false);
	const allowRegistration = ref(true);
	const appName = ref("Formera");
	const footerLinks = ref<FooterLink[]>([]);
	const primaryColor = ref("#6366f1");
	const logoURL = ref("");
	const logoShowText = ref(true);
	const faviconURL = ref("");
	const loginBackgroundURL = ref("");
	const language = ref("en");
	const theme = ref<"light" | "dark" | "system">("system");
	const isLoading = ref(true); // Start with loading true to prevent hydration mismatch

	const logoDisplayURL = computed(() => getFileUrl(logoURL.value));
	const faviconDisplayURL = computed(() => getFileUrl(faviconURL.value));
	const loginBackgroundDisplayURL = computed(() => getFileUrl(loginBackgroundURL.value));

	const { setupApi } = useApi();

	const applyPrimaryColor = (color: string) => {
		if (import.meta.client && color) {
			document.documentElement.style.setProperty("--primary", color);
			const darkerColor = adjustColor(color, -20);
			const lighterColor = adjustColor(color, 30);
			document.documentElement.style.setProperty("--primary-dark", darkerColor);
			document.documentElement.style.setProperty("--primary-light", lighterColor);
		}
	};

	const adjustColor = (hex: string, percent: number): string => {
		const num = parseInt(hex.replace("#", ""), 16);
		const amt = Math.round(2.55 * percent);
		const R = Math.max(0, Math.min(255, (num >> 16) + amt));
		const G = Math.max(0, Math.min(255, ((num >> 8) & 0x00ff) + amt));
		const B = Math.max(0, Math.min(255, (num & 0x0000ff) + amt));
		return `#${((1 << 24) | (R << 16) | (G << 8) | B).toString(16).slice(1)}`;
	};

	const applyTheme = (themeValue: "light" | "dark" | "system") => {
		if (import.meta.client) {
			let effectiveTheme: "light" | "dark" = "light";
			if (themeValue === "system") {
				effectiveTheme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
			} else {
				effectiveTheme = themeValue;
			}
			// Only apply if different from current (prevents flash during hydration)
			const currentTheme = document.documentElement.getAttribute("data-theme");
			if (currentTheme !== effectiveTheme) {
				document.documentElement.setAttribute("data-theme", effectiveTheme);
			}
		}
	};

	const loadStatus = async () => {
		try {
			const data = await setupApi.getStatus();
			setupRequired.value = data.setup_required;
			allowRegistration.value = data.allow_registration;
			appName.value = data.app_name;
			footerLinks.value = data.footer_links || [];
			primaryColor.value = data.primary_color || "#6366f1";
			logoURL.value = data.logo_url || "";
			logoShowText.value = data.logo_show_text ?? true;
			faviconURL.value = data.favicon_url || "";
			loginBackgroundURL.value = data.login_background_url || "";
			language.value = data.language || "en";
			theme.value = data.theme || "system";
			applyPrimaryColor(primaryColor.value);
			applyTheme(theme.value);
		} catch (error) {
			console.error("Failed to load setup status:", error);
			setupRequired.value = false;
			allowRegistration.value = true;
			appName.value = "Formera";
			footerLinks.value = [];
			primaryColor.value = "#6366f1";
			logoURL.value = "";
			logoShowText.value = true;
			faviconURL.value = "";
			loginBackgroundURL.value = "";
			language.value = "en";
			theme.value = "system";
		} finally {
			isLoading.value = false;
		}
	};

	const refresh = async () => {
		isLoading.value = true;
		await loadStatus();
	};

	watch(primaryColor, (newColor) => {
		applyPrimaryColor(newColor);
	});

	return {
		setupRequired,
		allowRegistration,
		appName,
		footerLinks,
		primaryColor,
		logoURL,
		logoShowText,
		faviconURL,
		loginBackgroundURL,
		language,
		theme,
		logoDisplayURL,
		faviconDisplayURL,
		loginBackgroundDisplayURL,
		isLoading,
		loadStatus,
		refresh,
		applyPrimaryColor,
		applyTheme,
	};
});
