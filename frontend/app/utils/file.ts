/**
 * File utility functions for handling file paths and metadata
 */

/**
 * Extract filename from a file path
 * Handles both regular paths and prefixed filenames (e.g., "abc123_document.pdf")
 */
export const extractFilename = (path: string, fallback = "file"): string => {
	const pathParts = path.split("/");
	const filename = pathParts.pop() || "";
	return filename || fallback;
};

/**
 * Check if a path points to an image file based on extension
 */
export const isImageFile = (path: string): boolean => {
	return /\.(jpg|jpeg|png|gif|webp|svg)$/i.test(path);
};

/**
 * Check if a path is a protected file path (files/ directory)
 */
export const isProtectedPath = (path: string): boolean => {
	return path.startsWith("files/") || path.includes("/files/");
};

/**
 * Check if a value looks like a file path
 */
export const isFilePath = (value: string): boolean => {
	return (
		value.startsWith("http") ||
		value.startsWith("/uploads/") ||
		value.startsWith("images/") ||
		value.startsWith("files/")
	);
};
