<script lang="ts" setup>
const route = useRoute();
const { formsApi, submissionsApi } = useApi();
const { validateField } = useFieldValidation();

const id = route.params.id as string;

const form = ref<Form | null>(null);
const formData = ref<Record<string, unknown>>({});
const fieldErrors = ref<Record<string, string>>({});
const touchedFields = ref<Set<string>>(new Set());
const isLoading = ref(true);
const isSubmitting = ref(false);
const error = ref<string | null>(null);
const success = ref<string | null>(null);
const currentPage = ref(0);

// Password protection state
const requiresPassword = ref(false);
const passwordInput = ref("");
const isVerifyingPassword = ref(false);
const passwordError = ref<string | null>(null);
const passwordVerified = ref(false);

// UTM/Tracking parameters
const trackingParams = ref<Record<string, string>>({});

// Countdown state
const countdownInterval = ref<ReturnType<typeof setInterval> | null>(null);
const countdown = ref({
	days: 0,
	hours: 0,
	minutes: 0,
	seconds: 0,
});
const showCountdown = ref(false);
const formExpired = ref(false);

const extractTrackingParams = () => {
	const params: Record<string, string> = {};
	const urlParams = new URLSearchParams(window.location.search);

	// Standard UTM parameters
	const utmKeys = ["utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content"];
	for (const key of utmKeys) {
		const value = urlParams.get(key);
		if (value) {
			params[key] = value;
		}
	}

	// Custom tracking parameters (any parameter starting with "ref_" or "track_")
	for (const [key, value] of urlParams.entries()) {
		if (key.startsWith("ref_") || key.startsWith("track_")) {
			params[key] = value;
		}
	}

	trackingParams.value = params;
};

// Countdown functions
const updateCountdown = (startDate: string) => {
	const now = new Date().getTime();
	const target = new Date(startDate).getTime();
	const diff = target - now;

	if (diff <= 0) {
		// Countdown finished - reload the page
		if (countdownInterval.value) {
			clearInterval(countdownInterval.value);
			countdownInterval.value = null;
		}
		window.location.reload();
		return;
	}

	countdown.value = {
		days: Math.floor(diff / (1000 * 60 * 60 * 24)),
		hours: Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)),
		minutes: Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60)),
		seconds: Math.floor((diff % (1000 * 60)) / 1000),
	};
};

const startCountdown = (startDate: string) => {
	showCountdown.value = true;
	updateCountdown(startDate);
	countdownInterval.value = setInterval(() => updateCountdown(startDate), 1000);
};

const stopCountdown = () => {
	if (countdownInterval.value) {
		clearInterval(countdownInterval.value);
		countdownInterval.value = null;
	}
	showCountdown.value = false;
};

const checkFormAvailability = (data: Form): boolean => {
	// If no settings, form is available
	if (!data.settings) {
		return true;
	}

	const now = new Date();

	// Check start date
	if (data.settings.start_date) {
		const startDate = new Date(data.settings.start_date);
		if (now < startDate) {
			startCountdown(data.settings.start_date);
			return false;
		}
	}

	// Check end date
	if (data.settings.end_date) {
		const endDate = new Date(data.settings.end_date);
		if (now > endDate) {
			formExpired.value = true;
			return false;
		}
	}

	return true;
};

const loadForm = async () => {
	try {
		const data = await formsApi.getPublic(id);
		form.value = data;

		// Check if password protection is required
		if (data.password_protected && !passwordVerified.value) {
			requiresPassword.value = true;
			isLoading.value = false;
			return;
		}

		// Check form availability (start/end date) - after password check
		if (!checkFormAvailability(data)) {
			isLoading.value = false;
			return;
		}

		// Initialize form data
		const initialData: Record<string, unknown> = {};
		(data.fields || []).forEach((field: FormField) => {
			// Skip layout fields
			if (["section", "pagebreak", "divider", "heading", "paragraph", "image"].includes(field.type)) {
				return;
			}
			if (field.type === "checkbox" || field.type === "dropdown" || field.type === "file") {
				initialData[field.id] = [];
			} else {
				initialData[field.id] = "";
			}
		});
		formData.value = initialData;
	} catch {
		error.value = "Formular nicht gefunden oder nicht verfügbar.";
	} finally {
		isLoading.value = false;
	}
};

const verifyPassword = async () => {
	if (!passwordInput.value) {
		passwordError.value = "Bitte geben Sie das Passwort ein.";
		return;
	}

	isVerifyingPassword.value = true;
	passwordError.value = null;

	try {
		const result = await formsApi.verifyPassword(id, passwordInput.value);
		if (result.valid && result.form) {
			passwordVerified.value = true;
			requiresPassword.value = false;

			// Use form data from verification response
			form.value = result.form;

			// Check form availability (start/end date) after password verified
			if (!checkFormAvailability(result.form)) {
				return;
			}

			// Initialize form data
			const initialData: Record<string, unknown> = {};
			(result.form.fields || []).forEach((field: FormField) => {
				if (["section", "pagebreak", "divider", "heading", "paragraph", "image"].includes(field.type)) {
					return;
				}
				if (field.type === "checkbox" || field.type === "dropdown" || field.type === "file") {
					initialData[field.id] = [];
				} else {
					initialData[field.id] = "";
				}
			});
			formData.value = initialData;
		} else {
			passwordError.value = "Falsches Passwort. Bitte versuchen Sie es erneut.";
		}
	} catch {
		passwordError.value = "Fehler bei der Passwort-Überprüfung.";
	} finally {
		isVerifyingPassword.value = false;
	}
};

// Dynamic page title
useHead(() => ({
	title: form.value?.title || "Formular ausfüllen",
}));

const formFields = computed(() => form.value?.fields || []);

// Split fields into pages based on pagebreak fields
const pages = computed(() => {
	const result: FormField[][] = [[]];
	let pageIndex = 0;

	for (const field of formFields.value) {
		if (field.type === "pagebreak") {
			pageIndex++;
			result[pageIndex] = [];
		} else {
			result[pageIndex]?.push(field);
		}
	}

	return result;
});

const totalPages = computed(() => pages.value.length);
const isMultiPage = computed(() => totalPages.value > 1);
const isFirstPage = computed(() => currentPage.value === 0);
const isLastPage = computed(() => currentPage.value === totalPages.value - 1);
const currentPageFields = computed(() => pages.value[currentPage.value] || []);

// Validate a single field
const validateSingleField = (field: FormField): string => {
	const value = formData.value[field.id];
	const result = validateField(value, field.type, field.required, field.validation);
	return result.valid ? "" : result.message || "";
};

const isLayoutField = (type: string) => {
	return ["section", "pagebreak", "divider", "heading", "paragraph", "image"].includes(type);
};

// Validate fields on current page
const validateCurrentPage = (): boolean => {
	let isValid = true;
	const errors: Record<string, string> = {};

	for (const field of currentPageFields.value) {
		if (isLayoutField(field.type)) continue;

		// Mark as touched
		touchedFields.value.add(field.id);

		const errorMessage = validateSingleField(field);
		if (errorMessage) {
			errors[field.id] = errorMessage;
			isValid = false;
		}
	}

	fieldErrors.value = { ...fieldErrors.value, ...errors };
	return isValid;
};

// Validate all fields
const validateAllFields = (): boolean => {
	let isValid = true;
	const errors: Record<string, string> = {};

	for (const field of formFields.value) {
		if (isLayoutField(field.type)) continue;

		const errorMessage = validateSingleField(field);
		if (errorMessage) {
			errors[field.id] = errorMessage;
			isValid = false;
		}
	}

	fieldErrors.value = errors;
	return isValid;
};

// Handle field blur (mark as touched and validate)
const handleFieldBlur = (fieldId: string) => {
	touchedFields.value.add(fieldId);
	const field = formFields.value.find((f) => f.id === fieldId);
	if (field) {
		const errorMessage = validateSingleField(field);
		if (errorMessage) {
			fieldErrors.value[fieldId] = errorMessage;
		} else {
			delete fieldErrors.value[fieldId];
		}
	}
};

// Get error for a field (only if touched)
const getFieldError = (fieldId: string): string => {
	if (!touchedFields.value.has(fieldId)) return "";
	return fieldErrors.value[fieldId] || "";
};

// Navigate to next page
const nextPage = () => {
	error.value = null;
	if (!validateCurrentPage()) {
		error.value = "Bitte überprüfen Sie die markierten Felder.";
		return;
	}
	if (!isLastPage.value) {
		currentPage.value++;
		window.scrollTo({ top: 0, behavior: "smooth" });
	}
};

// Navigate to previous page
const prevPage = () => {
	error.value = null;
	if (!isFirstPage.value) {
		currentPage.value--;
		window.scrollTo({ top: 0, behavior: "smooth" });
	}
};

const handleSubmit = async () => {
	if (!form.value) return;

	// Validate current page first
	if (!validateCurrentPage()) {
		error.value = "Bitte überprüfen Sie die markierten Felder.";
		return;
	}

	// Mark all fields as touched
	formFields.value.forEach((field) => {
		if (!isLayoutField(field.type)) {
			touchedFields.value.add(field.id);
		}
	});

	// Validate all fields
	if (!validateAllFields()) {
		error.value = "Bitte überprüfen Sie die markierten Felder.";
		return;
	}

	isSubmitting.value = true;
	error.value = null;

	try {
		// Include tracking parameters if present
		const metadata = Object.keys(trackingParams.value).length > 0 ? trackingParams.value : undefined;
		const response = await submissionsApi.submit(form.value.id, formData.value, metadata);
		success.value = response.message || form.value.settings.success_message || "Vielen Dank für Ihre Antwort!";
	} catch (err: unknown) {
		const errorMessage = err instanceof Error ? err.message : "Fehler beim Absenden";
		error.value = errorMessage;
	} finally {
		isSubmitting.value = false;
	}
};

// Design computed styles
const design = computed(() => form.value?.settings?.design);
const hasCustomDesign = computed(() => {
	const d = design.value;
	return d && (d.primaryColor || d.backgroundColor || d.formBackgroundColor || d.textColor);
});

const containerStyle = computed(() => {
	const d = design.value;
	if (!d) return {};

	// Custom design overrides global theme variables within the form container
	const styles: Record<string, string | undefined> = {};

	// Only set CSS variables if custom design has values
	if (d.primaryColor) {
		styles["--form-primary"] = d.primaryColor;
		styles["--primary"] = d.primaryColor;
		styles["--border-focus"] = d.primaryColor;
	}
	if (d.backgroundColor) {
		styles["--form-bg"] = d.backgroundColor;
		styles.backgroundColor = d.backgroundColor;
	}
	if (d.formBackgroundColor) {
		styles["--form-surface"] = d.formBackgroundColor;
		styles["--surface"] = d.formBackgroundColor;
	}
	if (d.textColor) {
		styles["--form-text"] = d.textColor;
		styles["--text"] = d.textColor;
	}

	return styles;
});

const backgroundImageUrl = computed(() => {
	const img = design.value?.backgroundImage;
	return img ? getFileUrl(img) : null;
});
const backgroundImageSize = computed(() => design.value?.backgroundSize || "cover");

const formClass = computed(() => {
	const d = design.value;
	const classes = ["form"];
	if (hasCustomDesign.value) classes.push("custom-design");
	if (d?.maxWidth) classes.push(`form-width-${d.maxWidth}`);
	if (d?.borderRadius) classes.push(`form-radius-${d.borderRadius}`);
	if (d?.headerStyle) classes.push(`form-header-${d.headerStyle}`);
	if (d?.buttonStyle) classes.push(`form-button-${d.buttonStyle}`);
	if (d?.fontFamily && d.fontFamily !== "default") classes.push(`form-font-${d.fontFamily}`);
	return classes;
});

onMounted(() => {
	extractTrackingParams();
	loadForm();
});

onUnmounted(() => {
	stopCountdown();
});
</script>

<template>
	<div class="form-container" :style="containerStyle">
		<!-- Background image as real img element for lazy loading and better performance -->
		<img
			v-if="backgroundImageUrl"
			:src="backgroundImageUrl"
			:class="['background-image', `bg-size-${backgroundImageSize}`]"
			alt=""
			loading="lazy"
			aria-hidden="true"
		/>

		<div v-if="isLoading" class="loading">Formular wird geladen...</div>

		<!-- Countdown before form starts -->
		<div v-else-if="showCountdown" class="countdown-gate">
			<div class="countdown-card">
				<UISysIcon icon="fa-solid fa-clock" extra-classes="countdown-icon" />
				<h2>{{ form?.title || 'Formular' }}</h2>
				<p>Dieses Formular ist noch nicht verfügbar. Es startet in:</p>

				<div class="countdown-timer">
					<div v-if="countdown.days > 0" class="countdown-unit">
						<span class="countdown-value">{{ countdown.days }}</span>
						<span class="countdown-label">{{ countdown.days === 1 ? 'Tag' : 'Tage' }}</span>
					</div>
					<div class="countdown-unit">
						<span class="countdown-value">{{ String(countdown.hours).padStart(2, '0') }}</span>
						<span class="countdown-label">{{ countdown.hours === 1 ? 'Stunde' : 'Stunden' }}</span>
					</div>
					<div class="countdown-unit">
						<span class="countdown-value">{{ String(countdown.minutes).padStart(2, '0') }}</span>
						<span class="countdown-label">{{ countdown.minutes === 1 ? 'Minute' : 'Minuten' }}</span>
					</div>
					<div class="countdown-unit">
						<span class="countdown-value">{{ String(countdown.seconds).padStart(2, '0') }}</span>
						<span class="countdown-label">{{ countdown.seconds === 1 ? 'Sekunde' : 'Sekunden' }}</span>
					</div>
				</div>

				<p class="countdown-hint">Die Seite wird automatisch aktualisiert, sobald das Formular verfügbar ist.</p>
			</div>
		</div>

		<!-- Form Expired -->
		<div v-else-if="formExpired" class="expired-gate">
			<div class="expired-card">
				<UISysIcon icon="fa-solid fa-calendar-xmark" extra-classes="expired-icon" />
				<h2>{{ form?.title || 'Formular' }}</h2>
				<p>Dieses Formular ist leider nicht mehr verfügbar.</p>
				<p class="expired-hint">Der Einreichungszeitraum ist abgelaufen.</p>
			</div>
		</div>

		<!-- Password Protection Gate -->
		<div v-else-if="requiresPassword" class="password-gate">
			<div class="password-card">
				<UISysIcon icon="fa-solid fa-lock" extra-classes="password-icon" />
				<h2>Passwort erforderlich</h2>
				<p>Dieses Formular ist passwortgeschützt. Bitte geben Sie das Passwort ein, um fortzufahren.</p>

				<form class="password-form" @submit.prevent="verifyPassword">
					<div v-if="passwordError" class="password-error">
						<UISysIcon icon="fa-solid fa-circle-exclamation" />
						{{ passwordError }}
					</div>

					<input
						v-model="passwordInput"
						type="password"
						class="input password-input"
						placeholder="Passwort eingeben"
						autofocus
					/>
					<button
						type="submit"
						class="btn btn-primary password-submit"
						:disabled="isVerifyingPassword"
					>
						{{ isVerifyingPassword ? "Prüfen..." : "Entsperren" }}
					</button>
				</form>
			</div>
		</div>

		<div v-else-if="error && !form" class="error-state">
			<UISysIcon icon="fa-solid fa-circle-exclamation" style="font-size: 48px" />
			<h2>Fehler</h2>
			<p>{{ error }}</p>
		</div>

		<div v-else-if="success" class="success-state">
			<UISysIcon icon="fa-solid fa-circle-check" style="font-size: 48px" />
			<h2>Vielen Dank!</h2>
			<p>{{ success }}</p>
		</div>

		<form v-else-if="form" :class="formClass" @submit.prevent="handleSubmit" novalidate>
			<div class="header">
				<h1>{{ form.title }}</h1>
				<p v-if="form.description">{{ form.description }}</p>
				<!-- Progress indicator for multi-page forms -->
				<div v-if="isMultiPage" class="progress-indicator">
					<div class="progress-bar">
						<div
							class="progress-fill"
							:style="{ width: `${((currentPage + 1) / totalPages) * 100}%` }"
						/>
					</div>
					<span class="progress-text">Seite {{ currentPage + 1 }} von {{ totalPages }}</span>
				</div>
			</div>

			<div v-if="error" class="error">
				<UISysIcon icon="fa-solid fa-circle-exclamation" />
				{{ error }}
			</div>

			<div class="fields">
				<template v-for="field in currentPageFields" :key="field.id">
					<!-- Layout: Section -->
					<FormFieldsSectionField
						v-if="field.type === 'section'"
						:label="field.label"
						:description="field.sectionDescription"
					/>

					<!-- Layout: Divider -->
					<FormFieldsDividerField v-else-if="field.type === 'divider'" />

					<!-- Layout: Heading -->
					<FormFieldsHeadingField
						v-else-if="field.type === 'heading'"
						:label="field.label"
						:level="field.headingLevel || 2"
					/>

					<!-- Layout: Paragraph -->
					<FormFieldsParagraphField
						v-else-if="field.type === 'paragraph'"
						:content="field.content"
					/>


					<!-- Regular fields -->
					<FormFieldsFieldWrapper
						v-else-if="!isLayoutField(field.type)"
						:label="field.label"
						:description="field.description"
						:required="field.required"
						:error="getFieldError(field.id)"
						:field-id="field.id"
					>
						<!-- Text, Email, Phone, URL -->
						<FormFieldsTextField
							v-if="field.type === 'text' || field.type === 'email' || field.type === 'phone' || field.type === 'url'"
							:id="field.id"
							v-model="formData[field.id] as string"
							:type="field.type"
							:placeholder="field.placeholder"
							:required="field.required"
							@blur="handleFieldBlur(field.id)"
						/>

						<!-- Number -->
						<FormFieldsNumberField
							v-else-if="field.type === 'number'"
							:id="field.id"
							v-model="formData[field.id] as string"
							:placeholder="field.placeholder"
							:required="field.required"
							:min="field.validation?.min"
							:max="field.validation?.max"
							@blur="handleFieldBlur(field.id)"
						/>

						<!-- Textarea -->
						<FormFieldsTextareaField
							v-else-if="field.type === 'textarea'"
							:id="field.id"
							v-model="formData[field.id] as string"
							:placeholder="field.placeholder"
							:required="field.required"
							:rows="4"
							@blur="handleFieldBlur(field.id)"
						/>

						<!-- Rich Text Editor -->
						<FormFieldsRichTextField
							v-else-if="field.type === 'richtext'"
							:id="field.id"
							v-model="formData[field.id] as string"
							:placeholder="field.placeholder"
							:required="field.required"
							@blur="handleFieldBlur(field.id)"
						/>

						<!-- Date / Time -->
						<FormFieldsDateTimeField
							v-else-if="field.type === 'date' || field.type === 'time'"
							:id="field.id"
							v-model="formData[field.id] as string"
							:type="field.type"
							:required="field.required"
							@blur="handleFieldBlur(field.id)"
						/>

						<!-- Select (single) -->
						<FormFieldsSelectField
							v-else-if="field.type === 'select'"
							:id="field.id"
							v-model="formData[field.id] as string"
							:options="field.options || []"
							:required="field.required"
							@blur="handleFieldBlur(field.id)"
						/>

						<!-- Dropdown (multi-select) -->
						<FormFieldsMultiSelectField
							v-else-if="field.type === 'dropdown'"
							:id="field.id"
							v-model="formData[field.id] as string[]"
							:options="field.options || []"
							:required="field.required"
							@blur="handleFieldBlur(field.id)"
						/>

						<!-- Radio -->
						<FormFieldsRadioField
							v-else-if="field.type === 'radio'"
							v-model="formData[field.id] as string"
							:name="field.id"
							:options="field.options || []"
							:required="field.required"
							@change="handleFieldBlur(field.id)"
						/>

						<!-- Checkbox -->
						<FormFieldsCheckboxField
							v-else-if="field.type === 'checkbox'"
							v-model="formData[field.id] as string[]"
							:options="field.options || []"
							@change="handleFieldBlur(field.id)"
						/>

						<!-- Rating -->
						<FormFieldsRatingField
							v-else-if="field.type === 'rating'"
							v-model="formData[field.id] as number"
							:min-value="field.minValue || 1"
							:max-value="field.maxValue || 5"
							@change="handleFieldBlur(field.id)"
						/>

						<!-- Scale -->
						<FormFieldsScaleField
							v-else-if="field.type === 'scale'"
							v-model="formData[field.id] as number"
							:min-value="field.minValue || 1"
							:max-value="field.maxValue || 10"
							:min-label="field.minLabel"
							:max-label="field.maxLabel"
							@change="handleFieldBlur(field.id)"
						/>

						<!-- File Upload -->
						<FormFieldsFileUploadField
							v-else-if="field.type === 'file'"
							v-model="formData[field.id] as string[]"
							:field-id="field.id"
							:required="field.required"
							:multiple="field.multiple"
							:allowed-types="field.allowedTypes || []"
							:max-file-size="field.maxFileSize || 10"
							@change="handleFieldBlur(field.id)"
						/>

						<!-- Signature -->
						<FormFieldsSignatureField
							v-else-if="field.type === 'signature'"
							v-model="formData[field.id] as string"
							@change="handleFieldBlur(field.id)"
						/>
					</FormFieldsFieldWrapper>
				</template>
			</div>

			<div class="footer">
				<div v-if="isMultiPage" class="footer-nav">
					<button
						v-if="!isFirstPage"
						class="btn btn-secondary"
						type="button"
						@click="prevPage"
					>
						<UISysIcon icon="fa-solid fa-arrow-left" />
						Zurück
					</button>
					<div v-else />
					<button
						v-if="!isLastPage"
						class="btn btn-primary"
						type="button"
						@click="nextPage"
					>
						Weiter
						<UISysIcon icon="fa-solid fa-arrow-right" />
					</button>
					<button
						v-else
						class="btn btn-primary"
						:disabled="isSubmitting"
						type="submit"
					>
						{{ isSubmitting ? "Wird gesendet..." : form.settings.submit_button_text || "Absenden" }}
					</button>
				</div>
				<button
					v-else
					class="btn btn-primary btn-lg btn-full"
					:disabled="isSubmitting"
					type="submit"
				>
					{{ isSubmitting ? "Wird gesendet..." : form.settings.submit_button_text || "Absenden" }}
				</button>
			</div>
		</form>
	</div>
</template>

<style scoped>
.form-container {
	position: relative;
	display: flex;
	justify-content: center;
	width: 100%;
	min-height: 100vh;
	padding: 2rem 1rem;
}

/* Fixed background image */
.background-image {
	position: fixed;
	inset: 0;
	z-index: 0;
	width: 100%;
	height: 100%;
	pointer-events: none;
}

/* Background size variants */
.bg-size-contain {
	object-fit: contain;
	object-position: center;
}

.bg-size-cover {
	object-fit: cover;
	object-position: center;
}

.bg-size-auto {
	object-fit: none;
	object-position: center;
}

/* Ensure content is above background */
.form-container > *:not(.background-image) {
	position: relative;
	z-index: 1;
}

.loading {
	padding: 4rem;
	color: var(--form-text, var(--text));
	text-align: center;
}

.error-state,
.success-state {
	width: 100%;
	max-width: 400px;
	padding: 3rem;
	color: var(--form-text, var(--text));
	text-align: center;
	background: var(--form-surface, var(--surface));
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-lg);
}

.error-state i {
	margin-bottom: 1rem;
	color: var(--error);
}

.success-state i {
	margin-bottom: 1rem;
	color: var(--form-primary, var(--success));
}

.error-state h2,
.success-state h2 {
	margin-bottom: 0.5rem;
	font-size: 1.5rem;
}

.error-state p,
.success-state p {
	color: var(--text-secondary);
}

/* Base form styles */
.form {
	width: 100%;
	max-width: 768px; /* Default to lg */
	overflow: hidden;
	color: var(--form-text, var(--text));
	background: var(--form-surface, var(--surface));
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-lg);
}

/* Width variants */
.form-width-sm { max-width: 480px; }
.form-width-md { max-width: 640px; }
.form-width-lg { max-width: 768px; }
.form-width-xl { max-width: 896px; }

/* Border radius variants */
.form-radius-none { border-radius: 0; }
.form-radius-sm { border-radius: 4px; }
.form-radius-md { border-radius: 8px; }
.form-radius-lg { border-radius: 16px; }

/* Font variants */
.form-font-serif { font-family: Georgia, "Times New Roman", serif; }
.form-font-mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }

/* Header styles */
.header {
	padding: 2rem 2rem 1.5rem;
	border-bottom: 1px solid var(--border);
}

.form-header-colored .header {
	color: white;
	background: var(--form-primary, var(--primary));
	border-bottom: none;
}

.form-header-colored .header p {
	color: rgba(255, 255, 255, 0.85);
}

.form-header-minimal .header {
	padding-bottom: 1rem;
	border-bottom: none;
}

.header h1 {
	margin-bottom: 0.5rem;
	font-size: 1.75rem;
}

.header p {
	font-size: 0.9375rem;
	color: var(--text-secondary);
}

/* Button styles */
.form :deep(.btn-primary) {
	background: var(--form-primary, var(--primary));
	border-color: var(--form-primary, var(--primary));
}

.form :deep(.btn-primary:hover) {
	filter: brightness(1.1);
}

.form-button-outline :deep(.btn-primary) {
	color: var(--form-primary, var(--primary));
	background: transparent;
	border: 2px solid var(--form-primary, var(--primary));
}

.form-button-outline :deep(.btn-primary:hover) {
	color: white;
	background: var(--form-primary, var(--primary));
}

/* Progress bar uses primary color */
.progress-fill {
	background: var(--form-primary, var(--primary));
}

.error {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	padding: 0.75rem 1rem;
	margin: 1rem 2rem 0;
	font-size: 0.875rem;
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
	border-radius: var(--radius);
}

.fields {
	padding: 1.5rem 2rem;
}

.footer {
	padding: 1.5rem 2rem;
	background: rgba(0, 0, 0, 0.02);
	border-top: 1px solid var(--border);
}

.footer-nav {
	display: flex;
	gap: 1rem;
	justify-content: space-between;
}

.footer-nav button {
	display: flex;
	gap: 0.5rem;
	align-items: center;
}

.btn-full {
	width: 100%;
}

/* Progress indicator */
.progress-indicator {
	margin-top: 1rem;
}

.progress-bar {
	height: 4px;
	overflow: hidden;
	background: var(--border);
	border-radius: 2px;
}

.progress-fill {
	height: 100%;
	border-radius: 2px;
	transition: width 0.3s ease;
}

.progress-text {
	display: block;
	margin-top: 0.5rem;
	font-size: 0.8125rem;
	color: var(--text-secondary);
}

.form-header-colored .progress-bar {
	background: rgba(255, 255, 255, 0.3);
}

.form-header-colored .progress-fill {
	background: white;
}

.form-header-colored .progress-text {
	color: rgba(255, 255, 255, 0.85);
}

/* Custom design isolation - prevent dark mode from overriding custom colors */
.custom-design {
	/* Set derived colors based on main text/surface colors */
	--text-secondary: color-mix(in srgb, var(--text) 70%, transparent);
	--border: color-mix(in srgb, var(--text) 20%, transparent);
	--surface-hover: color-mix(in srgb, var(--surface) 95%, var(--text));
	--background: color-mix(in srgb, var(--surface) 97%, var(--text));
}

/* Override input styles within custom design forms */
.custom-design :deep(.input),
.custom-design :deep(input),
.custom-design :deep(textarea),
.custom-design :deep(select) {
	color: var(--text);
	background-color: var(--surface);
	border-color: var(--border);
}

.custom-design :deep(.input:focus),
.custom-design :deep(input:focus),
.custom-design :deep(textarea:focus),
.custom-design :deep(select:focus) {
	border-color: var(--form-primary, var(--primary));
	box-shadow: 0 0 0 3px color-mix(in srgb, var(--form-primary, var(--primary)) 20%, transparent);
}

.custom-design :deep(.label),
.custom-design :deep(label) {
	color: var(--text);
}

/* Ensure buttons respect custom primary color */
.custom-design :deep(.btn-secondary) {
	color: var(--text);
	background-color: var(--surface);
	border-color: var(--border);
}

.custom-design :deep(.btn-secondary:hover) {
	background-color: var(--surface-hover);
}

/* Radio and Checkbox option hover styles */
.custom-design :deep(.radio-label),
.custom-design :deep(.checkbox-label) {
	color: var(--text);
}

.custom-design :deep(.radio-label:hover),
.custom-design :deep(.checkbox-label:hover) {
	background-color: var(--background);
}

/* Scale field styles */
.custom-design :deep(.scale-btn) {
	color: var(--text);
	background: var(--background);
	border-color: var(--border);
}

.custom-design :deep(.scale-btn:hover) {
	border-color: var(--form-primary, var(--primary));
}

.custom-design :deep(.scale-btn-active) {
	color: white;
	background: var(--form-primary, var(--primary));
	border-color: var(--form-primary, var(--primary));
}

.custom-design :deep(.scale-label) {
	color: var(--text-secondary);
}

/* Rating field styles */
.custom-design :deep(.rating-star) {
	color: var(--border);
}

.custom-design :deep(.rating-star-active) {
	color: #fbbf24;
}

/* Countdown Gate Styles */
.countdown-gate {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 100%;
	min-height: 100vh;
}

.countdown-card {
	width: 100%;
	max-width: 500px;
	padding: 2.5rem;
	text-align: center;
	background: white;
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-lg);
}

:deep(.countdown-icon) {
	display: block;
	margin-bottom: 1rem;
	font-size: 48px;
	color: var(--form-primary, var(--primary));
}

.countdown-card h2 {
	margin-bottom: 0.5rem;
	font-size: 1.5rem;
	color: #111827;
}

.countdown-card > p {
	margin-bottom: 1.5rem;
	font-size: 0.9375rem;
	color: #6b7280;
}

.countdown-timer {
	display: flex;
	gap: 1rem;
	justify-content: center;
	margin-bottom: 1.5rem;
}

.countdown-unit {
	display: flex;
	flex-direction: column;
	align-items: center;
	min-width: 70px;
	padding: 1rem;
	background: #f3f4f6;
	border-radius: var(--radius);
}

.countdown-value {
	font-size: 2rem;
	font-weight: 700;
	line-height: 1;
	color: var(--form-primary, var(--primary));
	font-variant-numeric: tabular-nums;
}

.countdown-label {
	margin-top: 0.25rem;
	font-size: 0.75rem;
	color: #6b7280;
	text-transform: uppercase;
	letter-spacing: 0.05em;
}

.countdown-hint {
	margin-top: 1rem;
	font-size: 0.8125rem;
	color: #9ca3af;
}

/* Expired Gate Styles */
.expired-gate {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 100%;
	min-height: 100vh;
}

.expired-card {
	width: 100%;
	max-width: 450px;
	padding: 2.5rem;
	text-align: center;
	background: white;
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-lg);
}

:deep(.expired-icon) {
	display: block;
	margin-bottom: 1rem;
	font-size: 48px;
	color: #ef4444;
}

.expired-card h2 {
	margin-bottom: 0.5rem;
	font-size: 1.5rem;
	color: #111827;
}

.expired-card > p {
	font-size: 0.9375rem;
	color: #6b7280;
}

.expired-hint {
	margin-top: 0.5rem;
	font-size: 0.8125rem;
	color: #9ca3af;
}

/* Password Gate Styles */
.password-gate {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 100%;
	min-height: 100vh;
}

.password-card {
	width: 100%;
	max-width: 400px;
	padding: 2.5rem;
	text-align: center;
	background: var(--surface);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-lg);
}

:deep(.password-icon) {
	display: block;
	margin-bottom: 1rem;
	font-size: 48px;
	color: var(--primary);
}

.password-card h2 {
	margin-bottom: 0.5rem;
	font-size: 1.5rem;
	color: var(--text);
}

.password-card > p {
	margin-bottom: 1.5rem;
	font-size: 0.9375rem;
	color: var(--text-secondary);
}

.password-form {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

.password-error {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	justify-content: center;
	padding: 0.75rem;
	font-size: 0.875rem;
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
	border-radius: var(--radius);
}

.password-input {
	width: 100%;
	margin-bottom: 1rem;
}

.password-submit {
	width: 100%;
}
</style>
