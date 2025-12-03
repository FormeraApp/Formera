const getServerFileUrl = (pathOrUrl: string, apiUrl: string): string => {
	if (!pathOrUrl) return "";

	if (pathOrUrl.startsWith("http://") || pathOrUrl.startsWith("https://") || pathOrUrl.startsWith("/")) {
		return pathOrUrl;
	}

	const cleanPath = pathOrUrl.startsWith("/") ? pathOrUrl.slice(1) : pathOrUrl;
	return `${apiUrl}/files/${cleanPath}`;
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

	try {
		const response = await $fetch<{
			app_name: string;
			favicon_url: string;
			logo_url: string;
		}>(`${apiUrl}/setup/status`);

		if (response.favicon_url) {
			faviconUrl = getServerFileUrl(response.favicon_url, apiUrl);
		}
		if (response.logo_url) {
			logoUrl = getServerFileUrl(response.logo_url, apiUrl);
		}
		if (response.app_name) {
			appName = response.app_name;
		}
	} catch {
		// API not available during build, use defaults
	}

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
			lang: "en",
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
			{
				src: "https://kit.fontawesome.com/b0b0028fa2.js",
				crossorigin: "anonymous",
			},
		],
	});
});
