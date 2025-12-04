const getServerFileUrl = (pathOrUrl: string, apiUrl: string): string => {
	if (!pathOrUrl) return "";

	// Already a full URL or data URL
	if (pathOrUrl.startsWith("http://") || pathOrUrl.startsWith("https://") || pathOrUrl.startsWith("data:")) {
		return pathOrUrl;
	}

	// Use /uploads/ endpoint for public files (images/, logos, favicons)
	const cleanPath = pathOrUrl.startsWith("/") ? pathOrUrl.slice(1) : pathOrUrl;
	return `${apiUrl}/uploads/${cleanPath}`;
};

export default defineNuxtPlugin(async () => {
	const config = useRuntimeConfig();
	const siteUrl = config.public.siteUrl as string;
	const apiUrl = config.public.apiUrl as string;
	const defaults = config.public.defaults as {
		title: string;
		description: string;
		keywords: string;
		favicon: string;
	};

	let faviconUrl = defaults.favicon;
	let logoUrl = "";
	let appName = defaults.title;
	let language = "en";
	let theme: "light" | "dark" | "system" = "system";

	try {
		const response = await $fetch<{
			app_name: string;
			favicon_url: string;
			logo_url: string;
			language: string;
			theme: "light" | "dark" | "system";
		}>(`${apiUrl}/api/setup/status`);

		if (response.favicon_url) {
			faviconUrl = getServerFileUrl(response.favicon_url, apiUrl);
		}
		if (response.logo_url) {
			logoUrl = getServerFileUrl(response.logo_url, apiUrl);
		}
		if (response.app_name) {
			appName = response.app_name;
		}
		if (response.language) {
			language = response.language;
		}
		if (response.theme) {
			theme = response.theme;
		}
	} catch {
		// API not available during build, use defaults
	}

	// For SSR, we can only apply light/dark directly
	// "system" preference requires client-side JS, so default to light on server
	const effectiveTheme = theme === "system" ? "light" : theme;

	// Inline script to immediately apply system theme preference before paint
	// This prevents flash when theme is set to "system"
	const themeScript =
		theme === "system"
			? `(function(){var d=document.documentElement;if(window.matchMedia('(prefers-color-scheme:dark)').matches){d.setAttribute('data-theme','dark')}})();`
			: "";

	useSeoMeta({
		description: defaults.description,
		ogTitle: appName,
		ogDescription: defaults.description,
		ogType: "website",
		ogImage: logoUrl,
		ogUrl: siteUrl,
		twitterTitle: appName,
		twitterDescription: defaults.description,
		twitterImage: logoUrl,
		twitterCard: "summary",
	});

	useHead({
		htmlAttrs: {
			lang: language,
			"data-theme": effectiveTheme,
		},
		title: appName,
		link: [
			{ rel: "stylesheet", href: "https://rsms.me/inter/inter.css" },
			{ rel: "icon", href: faviconUrl },
			{ rel: "canonical", href: siteUrl },
		],
		meta: [
			{ name: "title", content: appName },
			{ name: "keywords", content: defaults.keywords },
		],
		script: [
			// Inline script to apply system theme before paint (prevents flash)
			...(themeScript ? [{ innerHTML: themeScript, tagPosition: "head" as const }] : []),
			{
				src: "https://kit.fontawesome.com/b0b0028fa2.js",
				crossorigin: "anonymous",
			},
		],
	});
});
