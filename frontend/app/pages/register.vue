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
						<input id="name" v-model="name" class="input" :placeholder="$t('auth.namePlaceholder')" required type="text" />
					</div>

					<div class="form-group">
						<label class="label" for="email">{{ $t("auth.email") }}</label>
						<input id="email" v-model="email" class="input" :placeholder="$t('auth.emailPlaceholder')" required type="email" />
					</div>

					<div class="form-group">
						<label class="label" for="password">{{ $t("auth.password") }}</label>
						<input id="password" v-model="password" class="input" :placeholder="$t('auth.minCharsPlaceholder')" required type="password" />
					</div>

					<div class="form-group">
						<label class="label" for="confirmPassword">{{ $t("auth.confirmPassword") }}</label>
						<input id="confirmPassword" v-model="confirmPassword" class="input" :placeholder="$t('auth.repeatPasswordPlaceholder')" required type="password" />
					</div>

					<button class="btn btn-primary btn-lg" :disabled="isLoading" style="width: 100%" type="submit">
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
</style>
