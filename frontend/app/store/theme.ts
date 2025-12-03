import { defineStore } from "pinia";

export type Theme = "light" | "dark";
export type ThemePreference = "light" | "dark" | "system";

export const useThemeStore = defineStore("theme", () => {
	const theme = ref<Theme>("light");
	const preference = ref<ThemePreference>("system");

	const getEffectiveTheme = (pref: ThemePreference): Theme => {
		if (pref === "system") {
			return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
		}
		return pref;
	};

	const init = () => {
		if (import.meta.client) {
			// Get preference from setupStore (loaded from backend)
			const setupStore = useSetupStore();
			preference.value = setupStore.theme || "system";
			theme.value = getEffectiveTheme(preference.value);

			// Only apply theme if it differs from what SSR already set
			// This prevents flash during hydration
			const currentTheme = document.documentElement.getAttribute("data-theme");
			if (currentTheme !== theme.value) {
				applyTheme();
			}

			// Listen for system preference changes
			window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", (e) => {
				if (preference.value === "system") {
					theme.value = e.matches ? "dark" : "light";
					applyTheme();
				}
			});
		}
	};

	const applyTheme = () => {
		if (import.meta.client) {
			document.documentElement.setAttribute("data-theme", theme.value);
		}
	};

	const toggleTheme = () => {
		theme.value = theme.value === "light" ? "dark" : "light";
		preference.value = theme.value;
		applyTheme();
	};

	const setTheme = (newTheme: Theme) => {
		theme.value = newTheme;
		preference.value = newTheme;
		applyTheme();
	};

	const setPreference = (newPref: ThemePreference) => {
		preference.value = newPref;
		theme.value = getEffectiveTheme(newPref);
		applyTheme();
	};

	return {
		theme,
		preference,
		init,
		toggleTheme,
		setTheme,
		setPreference,
	};
});
