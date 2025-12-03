import process from "node:process";
import tailwindcss from "@tailwindcss/vite";
import pkg from "./package.json";

export default defineNuxtConfig({
	appConfig: {
		buildDate: new Date().toISOString(),
	},

	compatibilityDate: "2025-05-15",

	devServer: {
		cors: {
			origin: "*",
		},
	},
	devtools: {
		enabled: true,

		timeline: {
			enabled: true,
		},
	},

	experimental: {
		asyncContext: true,
		inlineRouteRules: true,
		lazyHydration: true,
		payloadExtraction: true,
		sharedPrerenderData: true,
		viewTransition: true,
		writeEarlyHints: true,
	},

	future: {
		compatibilityVersion: 5,
		typescriptBundlerResolution: true,
	},

	typescript: {
		strict: true,
		typeCheck: false,
	},

	nitro: {
		preset: "static",
		esbuild: {
			options: {
				target: "esnext",
			},
		},
		routeRules: {
			"/_ipx/**": { headers: { "cache-control": `public,max-age=691200,s-maxage=691200` } },
		},
	},

	sourcemap: { client: "hidden" },

	vite: {
		build: {
			chunkSizeWarningLimit: 800,
			sourcemap: "hidden",
		},
		plugins: [tailwindcss()],
	},

	modules: ["@pinia/nuxt", "pinia-plugin-persistedstate/nuxt", "@nuxt/image", "@nuxtjs/sitemap", "@nuxtjs/robots", "@nuxtjs/i18n"],
	css: ["~/assets/css/main.css", "~/assets/css/style.scss"],

	i18n: {
		locales: [
			{ code: "en", name: "English", file: "en.json" },
			{ code: "de", name: "Deutsch", file: "de.json" },
		],
		defaultLocale: "en",
		langDir: "locales/",
		strategy: "prefix_except_default",
		detectBrowserLanguage: {
			useCookie: true,
			cookieKey: "i18n_locale",
			redirectOn: "root",
		},
	},

	image: {
		domains: ["localhost"],
		format: ["webp", "avif"],
		ipx: {
			maxAge: 2592000, // 30 days in seconds
		},
		provider: "ipx",
		quality: 80,
		screens: {
			lg: 1024,
			md: 768,
			sm: 640,
			xl: 1280,
			xxl: 1536,
		},
	},

	pinia: {
		storesDirs: ["~/store"],
	},

	imports: {
		autoImport: true,
		dirs: ["shared/**/**", "composables/**/**", "store/**/**"],
		scan: true,
	},

	runtimeConfig: {
		public: {
			apiUrl: process.env.BASE_URL ? `${process.env.BASE_URL}/api` : "http://localhost:8080/api",
			siteUrl: process.env.BASE_URL || "http://localhost:3000",
			indexable: process.env.NUXT_PUBLIC_INDEXABLE !== "false",
			VERSION: pkg.version,
			defaults: {
				title: "Formera",
				description: "Self-hosted form builder. Privacy-friendly alternative to Google Forms.",
				keywords: "form builder, self-hosted, open source, formera",
				favicon: "/favicon.ico",
			},
		},
	},

	app: {
		head: {
			title: "Formera",
			meta: [{ name: "description", content: "Formera - Self-Hosted Form Builder" }],
			script: [
				{
					innerHTML: `(function(){var t=localStorage.getItem('theme');if(t){document.documentElement.setAttribute('data-theme',t)}else if(window.matchMedia('(prefers-color-scheme:dark)').matches){document.documentElement.setAttribute('data-theme','dark')}})()`,
					tagPosition: "head",
				},
			],
		},
	},

	site: {
		url: process.env.BASE_URL || "http://localhost:3000",
		name: "Formera",
		indexable: process.env.NUXT_PUBLIC_INDEXABLE !== "false",
	},

	robots: {
		disallow: process.env.NUXT_PUBLIC_INDEXABLE === "false" ? ["/"] : ["/login", "/register", "/setup", "/settings"],
	},

	sitemap: {
		enabled: process.env.NUXT_PUBLIC_INDEXABLE !== "false",
		exclude: ["/login", "/register", "/setup", "/settings", "/forms/**"],
	},
});
