<script lang="ts" setup>
const { t } = useI18n();
const { setupApi } = useApi();

const name = ref("");
const email = ref("");
const password = ref("");
const confirmPassword = ref("");
const appName = ref("Formera");
const allowRegistration = ref(false);
const error = ref("");
const isLoading = ref(false);

// Field-level validation errors
const errors = reactive({
	name: "",
	email: "",
	password: "",
	confirmPassword: "",
});

// Touched state for showing errors only after interaction
const touched = reactive({
	name: false,
	email: false,
	password: false,
	confirmPassword: false,
});

const isValidEmail = (emailValue: string) => {
	const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
	return emailRegex.test(emailValue);
};

const validateName = () => {
	if (!name.value.trim()) {
		errors.name = t("validation.required");
	} else if (name.value.trim().length < 2) {
		errors.name = t("validation.minLength", { min: 2 });
	} else {
		errors.name = "";
	}
};

const validateEmail = () => {
	if (!email.value.trim()) {
		errors.email = t("validation.required");
	} else if (!isValidEmail(email.value)) {
		errors.email = t("validation.email");
	} else {
		errors.email = "";
	}
};

const validatePassword = () => {
	if (!password.value) {
		errors.password = t("validation.required");
	} else if (password.value.length < 8) {
		errors.password = t("auth.passwordMinLength");
	} else {
		errors.password = "";
	}
	// Re-validate confirm password when password changes
	if (touched.confirmPassword) {
		validateConfirmPassword();
	}
};

const validateConfirmPassword = () => {
	if (!confirmPassword.value) {
		errors.confirmPassword = t("validation.required");
	} else if (confirmPassword.value !== password.value) {
		errors.confirmPassword = t("auth.passwordMismatch");
	} else {
		errors.confirmPassword = "";
	}
};

const isFormValid = computed(() => {
	return (
		name.value.trim().length >= 2 &&
		email.value.trim() &&
		isValidEmail(email.value) &&
		password.value.length >= 8 &&
		confirmPassword.value === password.value
	);
});

const handleSubmit = async () => {
	// Mark all fields as touched
	touched.name = true;
	touched.email = true;
	touched.password = true;
	touched.confirmPassword = true;

	// Validate all fields
	validateName();
	validateEmail();
	validatePassword();
	validateConfirmPassword();

	// Check for errors
	if (errors.name || errors.email || errors.password || errors.confirmPassword) {
		return;
	}

	error.value = "";
	isLoading.value = true;

	try {
		const response = await setupApi.complete({
			email: email.value,
			password: password.value,
			name: name.value,
			app_name: appName.value,
			allow_registration: allowRegistration.value,
		});

		localStorage.setItem("token", response.token);
		window.location.href = "/forms";
	} catch {
		error.value = t("setup.failed");
	} finally {
		isLoading.value = false;
	}
};
</script>

<template>
	<div class="card" style="max-width: 480px">
		<div class="header">
			<UISysIcon icon="fa-solid fa-file-lines" class="logo" style="font-size: 40px" />
			<h1>{{ t("setup.welcome") }}</h1>
			<p>{{ t("setup.createAdmin") }}</p>
		</div>

		<form class="form" @submit.prevent="handleSubmit">
			<div v-if="error" class="error">{{ error }}</div>

			<div class="info-box">
				<UISysIcon icon="fa-solid fa-shield" style="color: var(--primary); flex-shrink: 0; margin-top: 2px; font-size: 20px" />
				<div>
					<strong>{{ t("setup.firstStart") }}</strong>
					<p>{{ t("setup.firstStartInfo") }}</p>
				</div>
			</div>

			<div class="form-group">
				<label class="label" for="appName">{{ t("setup.appName") }}</label>
				<input id="appName" v-model="appName" class="input" placeholder="Formera" type="text" />
			</div>

			<div class="form-group">
				<label class="label" for="name">{{ t("setup.yourName") }}</label>
				<input
					id="name"
					v-model="name"
					class="input"
					:class="{ 'input-error': touched.name && errors.name }"
					:placeholder="t('setup.namePlaceholder')"
					type="text"
					@blur="touched.name = true; validateName()"
					@input="validateName()"
				/>
				<span v-if="touched.name && errors.name" class="field-error">{{ errors.name }}</span>
			</div>

			<div class="form-group">
				<label class="label" for="email">{{ t("auth.email") }}</label>
				<input
					id="email"
					v-model="email"
					class="input"
					:class="{ 'input-error': touched.email && errors.email }"
					:placeholder="t('setup.emailPlaceholder')"
					type="email"
					@blur="touched.email = true; validateEmail()"
					@input="validateEmail()"
				/>
				<span v-if="touched.email && errors.email" class="field-error">{{ errors.email }}</span>
			</div>

			<div class="form-group">
				<label class="label" for="password">{{ t("auth.password") }}</label>
				<input
					id="password"
					v-model="password"
					class="input"
					:class="{ 'input-error': touched.password && errors.password }"
					:placeholder="t('setup.passwordPlaceholder')"
					type="password"
					@blur="touched.password = true; validatePassword()"
					@input="validatePassword()"
				/>
				<span v-if="touched.password && errors.password" class="field-error">{{ errors.password }}</span>
			</div>

			<div class="form-group">
				<label class="label" for="confirmPassword">{{ t("auth.confirmPassword") }}</label>
				<input
					id="confirmPassword"
					v-model="confirmPassword"
					class="input"
					:class="{ 'input-error': touched.confirmPassword && errors.confirmPassword }"
					:placeholder="t('setup.confirmPasswordPlaceholder')"
					type="password"
					@blur="touched.confirmPassword = true; validateConfirmPassword()"
					@input="validateConfirmPassword()"
				/>
				<span v-if="touched.confirmPassword && errors.confirmPassword" class="field-error">{{ errors.confirmPassword }}</span>
			</div>

			<div class="form-group">
				<label class="checkbox-label">
					<input v-model="allowRegistration" type="checkbox" />
					{{ t("setup.allowRegistration") }}
				</label>
				<p class="checkbox-hint">{{ t("setup.allowRegistrationHint") }}</p>
			</div>

			<button class="btn btn-primary btn-lg" :disabled="isLoading || !isFormValid" style="width: 100%; margin-top: 0.5rem" type="submit">
				{{ isLoading ? t("setup.running") : t("setup.complete") }}
			</button>
		</form>
	</div>
</template>

<style scoped>
.card {
	width: 100%;
	overflow: hidden;
	background: var(--surface);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-lg);
}

.header {
	padding: 2rem 2rem 1rem;
	text-align: center;
}

.logo {
	margin-bottom: 1rem;
	color: var(--primary);
}

.header h1 {
	margin-bottom: 0.5rem;
	font-size: 1.5rem;
	color: var(--text);
}

.header p {
	font-size: 0.9375rem;
	color: var(--text-secondary);
}

.form {
	padding: 1rem 2rem 2rem;
}

.error {
	padding: 0.75rem 1rem;
	margin-bottom: 1rem;
	font-size: 0.875rem;
	color: var(--error);
	background-color: rgba(239, 68, 68, 0.1);
	border-radius: var(--radius);
}

.info-box {
	display: flex;
	gap: 0.75rem;
	align-items: flex-start;
	padding: 1rem;
	margin-bottom: 1rem;
	background: rgba(99, 102, 241, 0.1);
	border-radius: 8px;
}

.info-box strong {
	display: block;
	font-size: 0.875rem;
}

.info-box p {
	margin: 0.25rem 0 0;
	font-size: 0.8125rem;
	color: var(--text-secondary);
}

.checkbox-label {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	font-size: 0.9375rem;
	cursor: pointer;
}

.checkbox-label input {
	width: 1rem;
	height: 1rem;
}

.checkbox-hint {
	margin-top: 0.375rem;
	margin-left: 1.5rem;
	font-size: 0.8125rem;
	color: var(--text-secondary);
}

.input-error {
	border-color: var(--error);
}

.input-error:focus {
	box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.2);
}

.field-error {
	display: block;
	margin-top: 0.375rem;
	font-size: 0.8125rem;
	color: var(--error);
}
</style>
