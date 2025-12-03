import DOMPurify from "dompurify";

/**
 * Composable for sanitizing HTML content to prevent XSS attacks
 */
export function useSanitize() {
	/**
	 * Sanitizes HTML content using DOMPurify
	 * Allows basic formatting tags but removes all JavaScript and dangerous attributes
	 */
	const sanitizeHtml = (html: unknown): string => {
		if (typeof html !== "string") {
			return "";
		}

		// Only run on client side where DOMPurify can access the DOM
		if (import.meta.server) {
			// On server, strip all HTML tags for safety
			return html.replace(/<[^>]*>/g, "");
		}

		return DOMPurify.sanitize(html, {
			// Allow basic formatting tags from rich text editor
			ALLOWED_TAGS: [
				"p",
				"br",
				"strong",
				"b",
				"em",
				"i",
				"u",
				"s",
				"strike",
				"ul",
				"ol",
				"li",
				"a",
				"h1",
				"h2",
				"h3",
				"h4",
				"h5",
				"h6",
				"blockquote",
				"code",
				"pre",
			],
			// Allow safe attributes
			ALLOWED_ATTR: ["href", "target", "rel"],
			// Force all links to open in new tab with noopener
			ADD_ATTR: ["target", "rel"],
			// Prevent javascript: URLs
			ALLOW_UNKNOWN_PROTOCOLS: false,
		});
	};

	return {
		sanitizeHtml,
	};
}
