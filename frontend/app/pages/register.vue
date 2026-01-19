<script lang="ts" setup>
const { t } = useI18n();
const authStore = useAuthStore();
const setupStore = useSetupStore();
const router = useRouter();
const localePath = useLocalePath();

const name = ref("");
const email = ref("");
const password = ref("");
const confirmPassword = ref("");
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
		await authStore.register(email.value, password.value, name.value);
		router.push(localePath("/forms"));
	} catch {
		error.value = t("auth.registrationFailed");
	} finally {
		isLoading.value = false;
	}
};
</script>

<template>
	<ClientOnly>
		<div class="card">
			<!-- Registration disabled -->
			<template v-if="!setupStore.allowRegistration">
				<div class="header">
					<UISysIcon icon="fa-solid fa-shield-halved" style="color: var(--text-secondary); font-size: 40px" />
					<h1>{{ $t("auth.registrationDisabled") }}</h1>
					<p>{{ $t("auth.registrationNotAvailable") }}</p>
				</div>
				<div class="footer">
					<p>
						{{ $t("auth.hasAccount") }}
						<NuxtLink :to="localePath('/login')">{{ $t("auth.loginNow") }}</NuxtLink>
					</p>
				</div>
			</template>

			<!-- Registration form -->
			<template v-else>
				<div class="header">
					<UISysIcon icon="fa-solid fa-file-lines" class="logo" style="font-size: 40px" />
					<h1>{{ $t("auth.register") }}</h1>
					<p>{{ $t("auth.createAccount") }}</p>
				</div>

				<form class="form" @submit.prevent="handleSubmit">
					<div v-if="error" class="error">{{ error }}</div>

					<div class="form-group">
						<label class="label" for="name">{{ $t("auth.name") }}</label>
						<input
							id="name"
							v-model="name"
							class="input"
							:class="{ 'input-error': touched.name && errors.name }"
							:placeholder="$t('auth.namePlaceholder')"
							type="text"
							@blur="touched.name = true; validateName()"
							@input="validateName()"
						/>
						<span v-if="touched.name && errors.name" class="field-error">{{ errors.name }}</span>
					</div>

					<div class="form-group">
						<label class="label" for="email">{{ $t("auth.email") }}</label>
						<input
							id="email"
							v-model="email"
							class="input"
							:class="{ 'input-error': touched.email && errors.email }"
							:placeholder="$t('auth.emailPlaceholder')"
							type="email"
							@blur="touched.email = true; validateEmail()"
							@input="validateEmail()"
						/>
						<span v-if="touched.email && errors.email" class="field-error">{{ errors.email }}</span>
					</div>

					<div class="form-group">
						<label class="label" for="password">{{ $t("auth.password") }}</label>
						<input
							id="password"
							v-model="password"
							class="input"
							:class="{ 'input-error': touched.password && errors.password }"
							:placeholder="$t('auth.minCharsPlaceholder')"
							type="password"
							@blur="touched.password = true; validatePassword()"
							@input="validatePassword()"
						/>
						<span v-if="touched.password && errors.password" class="field-error">{{ errors.password }}</span>
					</div>

					<div class="form-group">
						<label class="label" for="confirmPassword">{{ $t("auth.confirmPassword") }}</label>
						<input
							id="confirmPassword"
							v-model="confirmPassword"
							class="input"
							:class="{ 'input-error': touched.confirmPassword && errors.confirmPassword }"
							:placeholder="$t('auth.repeatPasswordPlaceholder')"
							type="password"
							@blur="touched.confirmPassword = true; validateConfirmPassword()"
							@input="validateConfirmPassword()"
						/>
						<span v-if="touched.confirmPassword && errors.confirmPassword" class="field-error">{{ errors.confirmPassword }}</span>
					</div>

					<button class="btn btn-primary btn-lg" :disabled="isLoading || !isFormValid" style="width: 100%" type="submit">
						{{ isLoading ? $t("auth.registering") : $t("auth.register") }}
					</button>
				</form>

				<div class="footer">
					<p>
						{{ $t("auth.hasAccount") }}
						<NuxtLink :to="localePath('/login')">{{ $t("auth.loginNow") }}</NuxtLink>
					</p>
				</div>
			</template>
		</div>
		<template #fallback>
			<div class="card">
				<div class="header">
					<UISysIcon icon="fa-solid fa-file-lines" class="logo" style="font-size: 40px" />
					<h1>{{ $t("auth.register") }}</h1>
					<p>{{ $t("common.loading") }}</p>
				</div>
			</div>
		</template>
	</ClientOnly>
</template>

<style scoped>
.card {
	width: 100%;
	max-width: 420px;
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

.footer {
	padding: 1.5rem 2rem;
	text-align: center;
	background-color: var(--background);
	border-top: 1px solid var(--border);
}

.footer p {
	font-size: 0.875rem;
	color: var(--text-secondary);
}

.footer a {
	font-weight: 500;
	color: var(--primary);
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
