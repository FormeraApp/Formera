export const useCaptcha = () => {
	const loadedScripts = ref<Set<string>>(new Set());

	const loadTurnstile = async (siteKey: string): Promise<void> => {
		if (loadedScripts.value.has("turnstile")) return;

		return new Promise((resolve, reject) => {
			const script = document.createElement("script");
			script.src = "https://challenges.cloudflare.com/turnstile/v0/api.js";
			script.async = true;
			script.defer = true;
			script.onload = () => {
				loadedScripts.value.add("turnstile");
				resolve();
			};
			script.onerror = reject;
			document.head.appendChild(script);
		});
	};

	const loadRecaptcha = async (siteKey: string): Promise<void> => {
		if (loadedScripts.value.has("recaptcha")) return;

		return new Promise((resolve, reject) => {
			const script = document.createElement("script");
			script.src = `https://www.google.com/recaptcha/api.js?render=${siteKey}`;
			script.async = true;
			script.defer = true;
			script.onload = () => {
				loadedScripts.value.add("recaptcha");
				resolve();
			};
			script.onerror = reject;
			document.head.appendChild(script);
		});
	};

	const loadHCaptcha = async (siteKey: string): Promise<void> => {
		if (loadedScripts.value.has("hcaptcha")) return;

		return new Promise((resolve, reject) => {
			const script = document.createElement("script");
			script.src = "https://js.hcaptcha.com/1/api.js";
			script.async = true;
			script.defer = true;
			script.onload = () => {
				loadedScripts.value.add("hcaptcha");
				resolve();
			};
			script.onerror = reject;
			document.head.appendChild(script);
		});
	};

	const executeRecaptcha = async (
		siteKey: string,
		action: string = "submit"
	): Promise<string> => {
		return new Promise((resolve, reject) => {
			(window as any).grecaptcha.ready(() => {
				(window as any).grecaptcha
					.execute(siteKey, { action })
					.then((token: string) => resolve(token))
					.catch(reject);
			});
		});
	};

	const renderTurnstile = (containerId: string, siteKey: string): string | null => {
		if (!(window as any).turnstile) {
			return null;
		}

		const container = document.getElementById(containerId);
		if (!container) {
			return null;
		}

		try {
			// Turnstile's render() expects either the container element or just the ID (without #)
			const widgetId = (window as any).turnstile.render(container, {
				sitekey: siteKey,
			});
			return widgetId;
		} catch (e) {
			return null;
		}
	};

	const renderHCaptcha = (containerId: string, siteKey: string): string | null => {
		if (!(window as any).hcaptcha) {
			return null;
		}

		const container = document.getElementById(containerId);
		if (!container) {
			return null;
		}

		try {
			// hCaptcha's render() expects the container element
			const widgetId = (window as any).hcaptcha.render(container, {
				sitekey: siteKey,
			});
			return widgetId;
		} catch (e) {
			return null;
		}
	};

	const getTurnstileResponse = (widgetId?: string): string | null => {
		if (!(window as any).turnstile) return null;
		return (window as any).turnstile.getResponse(widgetId);
	};

	const getHCaptchaResponse = (widgetId?: string): string | null => {
		if (!(window as any).hcaptcha) return null;
		return (window as any).hcaptcha.getResponse(widgetId);
	};

	return {
		loadTurnstile,
		loadRecaptcha,
		loadHCaptcha,
		executeRecaptcha,
		renderTurnstile,
		renderHCaptcha,
		getTurnstileResponse,
		getHCaptchaResponse,
	};
};
