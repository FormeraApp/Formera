import type { FieldValidation } from "~~/shared/types";

export interface ValidationResult {
	valid: boolean;
	message: string | null;
}

// Vordefinierte Regex-Patterns
export const VALIDATION_PATTERNS = {
	email: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
	phone: /^[\d\s\-+()]{6,20}$/,
	url: /^https?:\/\/.+/,
	postalCode: /^\d{5}$/,
	alphanumeric: /^[a-zA-Z0-9]+$/,
} as const;

export function useFieldValidation() {
	const { t } = useI18n();

	// Vordefinierte Fehlermeldungen (mit i18n)
	const VALIDATION_MESSAGES = {
		required: t("validation.required"),
		email: t("validation.email"),
		phone: t("validation.phone"),
		url: t("validation.url"),
		minLength: (min: number) => t("validation.minLength", { min }),
		maxLength: (max: number) => t("validation.maxLength", { max }),
		min: (min: number) => t("validation.min", { min }),
		max: (max: number) => t("validation.max", { max }),
		pattern: t("validation.pattern"),
		fileSize: (max: number) => t("validation.fileSize", { max }),
		fileType: (types: string[]) => t("validation.fileType", { types: types.join(", ") }),
		invalidNumber: t("validation.invalidNumber"),
	};
	/**
	 * Validiert einen Text-Wert
	 */
	const validateText = (
		value: string,
		options: {
			required?: boolean;
			minLength?: number;
			maxLength?: number;
			pattern?: string | RegExp;
			patternMessage?: string;
			requiredMessage?: string;
		}
	): ValidationResult => {
		const trimmedValue = value?.trim() || "";

		// Required check
		if (options.required && !trimmedValue) {
			return {
				valid: false,
				message: options.requiredMessage || VALIDATION_MESSAGES.required,
			};
		}

		// Skip other validations if empty and not required
		if (!trimmedValue) {
			return { valid: true, message: null };
		}

		// Min length
		if (options.minLength && trimmedValue.length < options.minLength) {
			return {
				valid: false,
				message: VALIDATION_MESSAGES.minLength(options.minLength),
			};
		}

		// Max length
		if (options.maxLength && trimmedValue.length > options.maxLength) {
			return {
				valid: false,
				message: VALIDATION_MESSAGES.maxLength(options.maxLength),
			};
		}

		// Pattern
		if (options.pattern) {
			const regex = typeof options.pattern === "string" ? new RegExp(options.pattern) : options.pattern;
			if (!regex.test(trimmedValue)) {
				return {
					valid: false,
					message: options.patternMessage || VALIDATION_MESSAGES.pattern,
				};
			}
		}

		return { valid: true, message: null };
	};

	/**
	 * Validiert eine E-Mail-Adresse
	 */
	const validateEmail = (value: string, options: { required?: boolean; requiredMessage?: string } = {}): ValidationResult => {
		const trimmedValue = value?.trim() || "";

		if (options.required && !trimmedValue) {
			return {
				valid: false,
				message: options.requiredMessage || VALIDATION_MESSAGES.required,
			};
		}

		if (!trimmedValue) {
			return { valid: true, message: null };
		}

		if (!VALIDATION_PATTERNS.email.test(trimmedValue)) {
			return {
				valid: false,
				message: VALIDATION_MESSAGES.email,
			};
		}

		return { valid: true, message: null };
	};

	/**
	 * Validiert eine Telefonnummer
	 */
	const validatePhone = (value: string, options: { required?: boolean; requiredMessage?: string } = {}): ValidationResult => {
		const trimmedValue = value?.trim() || "";

		if (options.required && !trimmedValue) {
			return {
				valid: false,
				message: options.requiredMessage || VALIDATION_MESSAGES.required,
			};
		}

		if (!trimmedValue) {
			return { valid: true, message: null };
		}

		if (!VALIDATION_PATTERNS.phone.test(trimmedValue)) {
			return {
				valid: false,
				message: VALIDATION_MESSAGES.phone,
			};
		}

		return { valid: true, message: null };
	};

	/**
	 * Validiert eine URL
	 */
	const validateUrl = (value: string, options: { required?: boolean; requiredMessage?: string } = {}): ValidationResult => {
		const trimmedValue = value?.trim() || "";

		if (options.required && !trimmedValue) {
			return {
				valid: false,
				message: options.requiredMessage || VALIDATION_MESSAGES.required,
			};
		}

		if (!trimmedValue) {
			return { valid: true, message: null };
		}

		if (!VALIDATION_PATTERNS.url.test(trimmedValue)) {
			return {
				valid: false,
				message: VALIDATION_MESSAGES.url,
			};
		}

		return { valid: true, message: null };
	};

	/**
	 * Validiert eine Zahl
	 */
	const validateNumber = (
		value: string | number,
		options: {
			required?: boolean;
			min?: number;
			max?: number;
			requiredMessage?: string;
		} = {}
	): ValidationResult => {
		const stringValue = String(value ?? "").trim();

		if (options.required && !stringValue) {
			return {
				valid: false,
				message: options.requiredMessage || VALIDATION_MESSAGES.required,
			};
		}

		if (!stringValue) {
			return { valid: true, message: null };
		}

		const numValue = Number(stringValue);
		if (isNaN(numValue)) {
			return {
				valid: false,
				message: VALIDATION_MESSAGES.invalidNumber,
			};
		}

		if (options.min !== undefined && numValue < options.min) {
			return {
				valid: false,
				message: VALIDATION_MESSAGES.min(options.min),
			};
		}

		if (options.max !== undefined && numValue > options.max) {
			return {
				valid: false,
				message: VALIDATION_MESSAGES.max(options.max),
			};
		}

		return { valid: true, message: null };
	};

	/**
	 * Validiert eine Datei
	 */
	const validateFile = (
		files: File[] | null,
		options: {
			required?: boolean;
			maxFileSize?: number; // in MB
			allowedTypes?: string[];
			requiredMessage?: string;
		} = {}
	): ValidationResult => {
		if (options.required && (!files || files.length === 0)) {
			return {
				valid: false,
				message: options.requiredMessage || VALIDATION_MESSAGES.required,
			};
		}

		if (!files || files.length === 0) {
			return { valid: true, message: null };
		}

		for (const file of files) {
			// File size check
			if (options.maxFileSize) {
				const maxBytes = options.maxFileSize * 1024 * 1024;
				if (file.size > maxBytes) {
					return {
						valid: false,
						message: VALIDATION_MESSAGES.fileSize(options.maxFileSize),
					};
				}
			}

			// File type check
			if (options.allowedTypes && options.allowedTypes.length > 0) {
				const fileExtension = `.${file.name.split(".").pop()?.toLowerCase()}`;
				const mimeType = file.type;

				const isAllowed = options.allowedTypes.some((type) => {
					const normalizedType = type.toLowerCase().trim();
					return normalizedType === fileExtension || normalizedType === mimeType || mimeType.startsWith(normalizedType.replace("*", ""));
				});

				if (!isAllowed) {
					return {
						valid: false,
						message: VALIDATION_MESSAGES.fileType(options.allowedTypes),
					};
				}
			}
		}

		return { valid: true, message: null };
	};

	/**
	 * Validiert ein Feld basierend auf seinem Typ und Validierungsregeln
	 */
	const validateField = (value: unknown, type: string, required: boolean, validation?: FieldValidation): ValidationResult => {
		const requiredMessage = validation?.requiredMessage;

		switch (type) {
			case "email":
				return validateEmail(String(value || ""), { required, requiredMessage });

			case "phone":
				return validatePhone(String(value || ""), { required, requiredMessage });

			case "url":
				return validateUrl(String(value || ""), { required, requiredMessage });

			case "number":
				return validateNumber(value as string | number, {
					required,
					min: validation?.min,
					max: validation?.max,
					requiredMessage,
				});

			case "text":
			case "textarea":
			case "richtext":
				return validateText(String(value || ""), {
					required,
					minLength: validation?.minLength,
					maxLength: validation?.maxLength,
					pattern: validation?.pattern,
					patternMessage: validation?.patternMessage,
					requiredMessage,
				});

			case "file":
				return validateFile(value as File[] | null, {
					required,
					maxFileSize: validation?.maxFileSize,
					allowedTypes: validation?.allowedTypes,
					requiredMessage,
				});

			default:
				// Für andere Feldtypen: nur required check
				if (required) {
					const hasValue = Array.isArray(value) ? value.length > 0 : !!value;
					if (!hasValue) {
						return {
							valid: false,
							message: requiredMessage || VALIDATION_MESSAGES.required,
						};
					}
				}
				return { valid: true, message: null };
		}
	};

	return {
		validateText,
		validateEmail,
		validatePhone,
		validateUrl,
		validateNumber,
		validateFile,
		validateField,
		VALIDATION_PATTERNS,
		VALIDATION_MESSAGES,
	};
}
