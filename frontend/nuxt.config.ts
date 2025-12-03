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
		preset: "node-server",
		esbuild: {
			options: {
				target: "esnext",
			},
		},
		routeRules: {
			"/_ipx/**": { headers: { "cache-control": `public,max-age=691200,s-maxage=691200` } },
		},
	},

	sourcemap: { client: false, server: false },

	vite: {
		build: {
			chunkSizeWarningLimit: 800,
			sourcemap: false,
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
			// These can be overridden at runtime via NUXT_PUBLIC_* env vars
			apiUrl: "/api",
			siteUrl: "http://localhost:3000",
			indexable: true,
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
		url: "http://localhost:3000",
		name: "Formera",
		indexable: true,
	},

	robots: {
		disallow: process.env.NUXT_PUBLIC_INDEXABLE === "false" ? ["/"] : ["/login", "/register", "/setup", "/settings"],
	},

	sitemap: {
		enabled: process.env.NUXT_PUBLIC_INDEXABLE !== "false",
		exclude: ["/login", "/register", "/setup", "/settings", "/forms/**"],
	},
});
