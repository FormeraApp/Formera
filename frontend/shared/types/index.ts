// Input field types
export type InputFieldType = "text" | "textarea" | "number" | "email" | "phone" | "date" | "time" | "url" | "richtext";

// Choice field types
export type ChoiceFieldType = "select" | "radio" | "checkbox" | "dropdown";

// Special field types
export type SpecialFieldType = "file" | "rating" | "scale" | "signature";

// Layout field types (not for data, only for structure)
export type LayoutFieldType = "section" | "pagebreak" | "divider" | "heading" | "paragraph" | "image";

// All field types
export type FieldType = InputFieldType | ChoiceFieldType | SpecialFieldType | LayoutFieldType;

// Field categories for builder
export const FIELD_CATEGORIES = {
	input: ["text", "textarea", "number", "email", "phone", "date", "time", "url", "richtext"] as InputFieldType[],
	choice: ["select", "radio", "checkbox", "dropdown"] as ChoiceFieldType[],
	special: ["file", "rating", "scale", "signature"] as SpecialFieldType[],
	layout: ["section", "pagebreak", "divider", "heading", "paragraph", "image"] as LayoutFieldType[],
} as const;

// Field metadata for UI - use i18n keys `fields.${type}.label` and `fields.${type}.description` for labels/descriptions
export const FIELD_META: Record<FieldType, { icon: string }> = {
	// Input
	text: { icon: "fa-solid fa-font" },
	textarea: { icon: "fa-solid fa-align-left" },
	number: { icon: "fa-solid fa-hashtag" },
	email: { icon: "fa-solid fa-envelope" },
	phone: { icon: "fa-solid fa-phone" },
	date: { icon: "fa-solid fa-calendar" },
	time: { icon: "fa-solid fa-clock" },
	url: { icon: "fa-solid fa-link" },
	richtext: { icon: "fa-solid fa-bold" },
	// Choice
	select: { icon: "fa-solid fa-caret-down" },
	radio: { icon: "fa-solid fa-circle-dot" },
	checkbox: { icon: "fa-solid fa-square-check" },
	dropdown: { icon: "fa-solid fa-list-check" },
	// Special
	file: { icon: "fa-solid fa-upload" },
	rating: { icon: "fa-solid fa-star" },
	scale: { icon: "fa-solid fa-sliders" },
	signature: { icon: "fa-solid fa-signature" },
	// Layout
	section: { icon: "fa-solid fa-layer-group" },
	pagebreak: { icon: "fa-solid fa-file-lines" },
	divider: { icon: "fa-solid fa-minus" },
	heading: { icon: "fa-solid fa-heading" },
	paragraph: { icon: "fa-solid fa-paragraph" },
	image: { icon: "fa-solid fa-image" },
};

// Validation rules for fields
export interface FieldValidation {
	// Text validation
	minLength?: number;
	maxLength?: number;
	pattern?: string; // Regex pattern
	patternMessage?: string; // Error message for regex
	// Number validation
	min?: number;
	max?: number;
	// File validation
	maxFileSize?: number; // in MB
	allowedTypes?: string[];
	// Custom error messages
	requiredMessage?: string;
	invalidMessage?: string;
}

export interface FormField {
	id: string;
	type: FieldType;
	label: string;
	placeholder?: string;
	required: boolean;
	options?: string[];
	validation?: FieldValidation;
	order: number;
	// Additional fields for extended features
	description?: string;
	// Section-specific
	sectionTitle?: string;
	sectionDescription?: string;
	collapsible?: boolean;
	collapsed?: boolean;
	// Layout-specific
	content?: string; // For paragraph, heading, image
	headingLevel?: 1 | 2 | 3 | 4;
	imageUrl?: string;
	imageAlt?: string;
	// Rich Text
	richTextContent?: string;
	// Rating/Scale
	minValue?: number;
	maxValue?: number;
	minLabel?: string;
	maxLabel?: string;
	// File Upload
	allowedTypes?: string[];
	maxFileSize?: number;
	multiple?: boolean;
}

// Design settings for the public form
export interface FormDesign {
	// Colors
	primaryColor?: string;
	backgroundColor?: string;
	formBackgroundColor?: string;
	textColor?: string;
	// Background image
	backgroundImage?: string;
	backgroundSize?: "cover" | "contain" | "auto";
	backgroundPosition?: "center" | "top" | "bottom";
	// Layout
	maxWidth?: "sm" | "md" | "lg" | "xl"; // sm=480, md=640, lg=768, xl=896
	borderRadius?: "none" | "sm" | "md" | "lg";
	// Header style
	headerStyle?: "default" | "colored" | "minimal";
	// Button style
	buttonStyle?: "filled" | "outline";
	// Font family
	fontFamily?: "default" | "serif" | "mono";
}

export interface FormSettings {
	submit_button_text: string;
	success_message: string;
	allow_multiple: boolean;
	require_login: boolean;
	notify_on_submission: boolean;
	notification_email?: string;
	max_submissions?: number;
	start_date?: string;
	end_date?: string;
	// Design settings
	design?: FormDesign;
}

export type FormStatus = "draft" | "published" | "closed";

export interface Form {
	id: string;
	user_id: string;
	title: string;
	description: string;
	slug?: string;
	fields: FormField[];
	settings: FormSettings;
	status: FormStatus;
	password_protected: boolean;
	created_at: string;
	updated_at: string;
}

// Request type for updating forms with password
export interface UpdateFormRequest extends Partial<Form> {
	password?: string; // Only sent when setting a new password
}

export type UserRole = "admin" | "user";

export interface User {
	id: string;
	email: string;
	name: string;
	role: UserRole;
	created_at: string;
}

export interface SubmissionMetadata {
	ip?: string;
	user_agent?: string;
	referrer?: string;
	// UTM/Tracking parameters
	utm_source?: string;
	utm_medium?: string;
	utm_campaign?: string;
	utm_term?: string;
	utm_content?: string;
	// Custom tracking parameters
	tracking?: Record<string, string>;
}

export interface Submission {
	id: string;
	form_id: string;
	data: Record<string, unknown>;
	metadata: SubmissionMetadata;
	created_at: string;
}

export interface AuthResponse {
	token: string;
	user: User;
}

// Generic pagination response from backend
export interface PaginatedResponse<T> {
	data: T;
	page: number;
	page_size: number;
	total_items: number;
	total_pages: number;
}

// Pagination parameters for API requests
export interface PaginationParams {
	page?: number;
	pageSize?: number;
}

export interface SubmissionsResponse {
	form: Form;
	submissions: PaginatedResponse<Submission[]>;
}

export interface FormStats {
	total_submissions: number;
	total_views: number;
	conversion_rate: number;
	field_stats: Record<string, Record<string, number>>;
}

export interface FooterLink {
	label: string;
	url: string;
}

export interface SetupStatus {
	setup_required: boolean;
	allow_registration: boolean;
	app_name: string;
	footer_links?: FooterLink[];
	primary_color: string;
	logo_url: string;
	logo_show_text: boolean;
	favicon_url: string;
	login_background_url: string;
	language: string;
	theme: "light" | "dark" | "system";
}

export interface Settings {
	id: number;
	allow_registration: boolean;
	setup_completed: boolean;
	app_name: string;
	footer_links?: FooterLink[];
	primary_color: string;
	logo_url: string;
	logo_show_text: boolean;
	favicon_url: string;
	login_background_url: string;
	language: "en" | "de";
	theme: "light" | "dark" | "system";
	created_at: string;
	updated_at: string;
}

export interface UploadResult {
	id: string;
	path: string; // Relative path for storage (e.g., "images/2025/12/abc123.png")
	url: string; // Immediate URL (presigned for S3, direct for local)
	filename: string;
	size: number;
	mimeType: string;
}
