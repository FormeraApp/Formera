<script lang="ts" setup>
definePageMeta({
	layout: "auth",
});

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

const handleSubmit = async () => {
	error.value = "";

	if (password.value !== confirmPassword.value) {
		error.value = t("auth.passwordMismatch");
		return;
	}

	if (password.value.length < 8) {
		error.value = t("auth.passwordMinLength");
		return;
	}

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
				<input id="name" v-model="name" class="input" :placeholder="t('setup.namePlaceholder')" required type="text" />
			</div>

			<div class="form-group">
				<label class="label" for="email">{{ t("auth.email") }}</label>
				<input id="email" v-model="email" class="input" :placeholder="t('setup.emailPlaceholder')" required type="email" />
			</div>

			<div class="form-group">
				<label class="label" for="password">{{ t("auth.password") }}</label>
				<input id="password" v-model="password" class="input" :placeholder="t('setup.passwordPlaceholder')" required type="password" />
			</div>

			<div class="form-group">
				<label class="label" for="confirmPassword">{{ t("auth.confirmPassword") }}</label>
				<input id="confirmPassword" v-model="confirmPassword" class="input" :placeholder="t('setup.confirmPasswordPlaceholder')" required type="password" />
			</div>

			<div class="form-group">
				<label class="checkbox-label">
					<input v-model="allowRegistration" type="checkbox" />
					{{ t("setup.allowRegistration") }}
				</label>
				<p class="checkbox-hint">{{ t("setup.allowRegistrationHint") }}</p>
			</div>

			<button class="btn btn-primary btn-lg" :disabled="isLoading" style="width: 100%; margin-top: 0.5rem" type="submit">
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
</style>
