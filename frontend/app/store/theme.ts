import { defineStore } from "pinia";

export type Theme = "light" | "dark";

export const useThemeStore = defineStore("theme", () => {
	const theme = ref<Theme>("light");

	const init = () => {
		if (import.meta.client) {
			const stored = localStorage.getItem("theme") as Theme | null;
			if (stored) {
				theme.value = stored;
			} else if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
				theme.value = "dark";
			}
			applyTheme();
		}
	};

	const applyTheme = () => {
		if (import.meta.client) {
			document.documentElement.setAttribute("data-theme", theme.value);
		}
	};

	const toggleTheme = () => {
		theme.value = theme.value === "light" ? "dark" : "light";
		if (import.meta.client) {
			localStorage.setItem("theme", theme.value);
		}
		applyTheme();
	};

	const setTheme = (newTheme: Theme) => {
		theme.value = newTheme;
		if (import.meta.client) {
			localStorage.setItem("theme", newTheme);
		}
		applyTheme();
	};

	return {
		theme,
		init,
		toggleTheme,
		setTheme,
	};
});
