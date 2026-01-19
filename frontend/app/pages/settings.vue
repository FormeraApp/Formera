<script lang="ts" setup>
import type { FooterLink, UserRole } from "~~/shared/types";

const { t, locale, setLocale, locales } = useI18n();
const { settingsApi, usersApi, uploadApi } = useApi();
const authStore = useAuthStore();
const setupStore = useSetupStore();
const themeStore = useThemeStore();
const localePath = useLocalePath();

// Available locales for the dropdown
const availableLocales = computed(() =>
	locales.value.map((l) => {
		if (typeof l === "string") {
			return { code: l, name: l };
		}
		return { code: l.code, name: l.name || l.code };
	})
);

// Active tab
const activeTab = ref<"general" | "design" | "footer" | "users">("general");

// Settings state
const settings = ref<Settings | null>(null);
const isLoading = ref(true);
const isSaving = ref(false);
const message = ref<{ type: "success" | "error"; text: string } | null>(null);
const loadError = ref<"permission" | "generic" | null>(null);

// Footer links state
const footerLinks = ref<FooterLink[]>([]);

// Design state
const primaryColor = ref("#6366f1");
const logoURL = ref("");
const logoShowText = ref(true);
const faviconURL = ref("");
const loginBackgroundURL = ref("");
const isUploadingLogo = ref(false);
const isUploadingFavicon = ref(false);
const isUploadingLoginBackground = ref(false);

// Language and Theme state
const selectedLanguage = ref<"en" | "de">("en");
const selectedTheme = ref<"light" | "dark" | "system">("system");

// Users state
const users = ref<User[]>([]);
const isLoadingUsers = ref(false);
const usersPagination = usePagination(5);
const showUserModal = ref(false);
const editingUser = ref<User | null>(null);
const userForm = ref({
	name: "",
	email: "",
	password: "",
	role: "user" as UserRole,
});
const userFormError = ref<string | null>(null);
const isSavingUser = ref(false);

const loadSettings = async () => {
	try {
		const data = await settingsApi.get();
		settings.value = data;
		footerLinks.value = data.footer_links || [];
		primaryColor.value = data.primary_color || "#6366f1";
		logoURL.value = data.logo_url || "";
		logoShowText.value = data.logo_show_text ?? true;
		faviconURL.value = data.favicon_url || "";
		loginBackgroundURL.value = data.login_background_url || "";
		selectedLanguage.value = data.language || "en";
		selectedTheme.value = data.theme || "system";
		loadError.value = null;
	} catch (error) {
		console.error("Failed to load settings:", error);
		// Check if it's a permission error (403 Forbidden)
		if (error instanceof Error && error.message.toLowerCase().includes("forbidden")) {
			loadError.value = "permission";
		} else {
			loadError.value = "generic";
		}
	} finally {
		isLoading.value = false;
	}
};

const loadUsers = async () => {
	isLoadingUsers.value = true;
	try {
		const response = await usersApi.list(usersPagination.params.value);
		users.value = response.data || [];
		usersPagination.updateFromResponse(response);
	} catch (error) {
		console.error("Failed to load users:", error);
	} finally {
		isLoadingUsers.value = false;
	}
};

// Watch for pagination changes
watch(() => usersPagination.params.value, () => {
	loadUsers();
}, { deep: true });

const handleSave = async () => {
	if (!settings.value) return;
	isSaving.value = true;
	message.value = null;

	try {
		const updated = await settingsApi.update({
			allow_registration: settings.value.allow_registration,
			app_name: settings.value.app_name,
			footer_links: footerLinks.value,
			primary_color: primaryColor.value,
			logo_url: logoURL.value,
			logo_show_text: logoShowText.value,
			favicon_url: faviconURL.value,
			login_background_url: loginBackgroundURL.value,
			language: selectedLanguage.value,
			theme: selectedTheme.value,
		});
		// Apply language change immediately
		if (selectedLanguage.value !== locale.value) {
			setLocale(selectedLanguage.value);
		}
		// Apply theme change immediately
		themeStore.setPreference(selectedTheme.value);
		settings.value = updated;
		message.value = { type: "success", text: t("settings.saved") };
		// Refresh setup store to update globally
		await setupStore.refresh();
	} catch {
		message.value = { type: "error", text: t("settings.saveError") };
	} finally {
		isSaving.value = false;
	}
};

// Footer links management
const addFooterLink = () => {
	footerLinks.value.push({ label: "", url: "" });
};

const removeFooterLink = (index: number) => {
	footerLinks.value.splice(index, 1);
};

// User management
const openAddUserModal = () => {
	editingUser.value = null;
	userForm.value = { name: "", email: "", password: "", role: "user" };
	userFormError.value = null;
	showUserModal.value = true;
};

const openEditUserModal = (user: User) => {
	editingUser.value = user;
	userForm.value = { name: user.name, email: user.email, password: "", role: user.role };
	userFormError.value = null;
	showUserModal.value = true;
};

const closeUserModal = () => {
	showUserModal.value = false;
	editingUser.value = null;
	userForm.value = { name: "", email: "", password: "", role: "user" };
	userFormError.value = null;
};

const handleSaveUser = async () => {
	userFormError.value = null;

	if (!userForm.value.name || !userForm.value.email) {
		userFormError.value = t("settings.users.nameEmailRequired");
		return;
	}

	if (!editingUser.value && !userForm.value.password) {
		userFormError.value = t("settings.users.passwordRequired");
		return;
	}

	isSavingUser.value = true;

	try {
		if (editingUser.value) {
			// Update existing user
			const updateData: Partial<User & { password?: string }> = {
				name: userForm.value.name,
				email: userForm.value.email,
				role: userForm.value.role,
			};
			if (userForm.value.password) {
				updateData.password = userForm.value.password;
			}
			await usersApi.update(editingUser.value.id, updateData);
		} else {
			// Create new user
			await usersApi.create({
				name: userForm.value.name,
				email: userForm.value.email,
				password: userForm.value.password,
				role: userForm.value.role,
			});
		}
		closeUserModal();
		await loadUsers();
		message.value = { type: "success", text: editingUser.value ? t("settings.users.userUpdated") : t("settings.users.userCreated") };
	} catch (error) {
		userFormError.value = error instanceof Error ? error.message : t("settings.saveError");
	} finally {
		isSavingUser.value = false;
	}
};

const handleDeleteUser = async (user: User) => {
	if (user.id === authStore.user?.id) {
		message.value = { type: "error", text: t("settings.users.cannotDeleteSelf") };
		return;
	}

	if (!confirm(t("settings.users.confirmDelete", { name: user.name }))) {
		return;
	}

	try {
		await usersApi.delete(user.id);
		await loadUsers();
		message.value = { type: "success", text: t("settings.users.userDeleted") };
	} catch (error) {
		message.value = { type: "error", text: error instanceof Error ? error.message : t("common.error") };
	}
};

const formatDate = (dateString: string) => {
	const dateLocale = locale.value === "de" ? "de-DE" : "en-US";
	return new Date(dateString).toLocaleDateString(dateLocale, {
		day: "2-digit",
		month: "2-digit",
		year: "numeric",
	});
};

// Computed display URLs for previews
const logoDisplayURL = computed(() => getFileUrl(logoURL.value));
const faviconDisplayURL = computed(() => getFileUrl(faviconURL.value));
const loginBackgroundDisplayURL = computed(() => getFileUrl(loginBackgroundURL.value));

// Upload handlers
const handleLogoUpload = async (event: Event) => {
	const input = event.target as HTMLInputElement;
	const file = input.files?.[0];
	if (!file) return;

	isUploadingLogo.value = true;
	try {
		const result = await uploadApi.uploadImage(file);
		// Store the path, not the URL
		logoURL.value = result.path;
	} catch (error) {
		message.value = { type: "error", text: error instanceof Error ? error.message : t("settings.design.uploadError") };
	} finally {
		isUploadingLogo.value = false;
		input.value = "";
	}
};

const handleFaviconUpload = async (event: Event) => {
	const input = event.target as HTMLInputElement;
	const file = input.files?.[0];
	if (!file) return;

	isUploadingFavicon.value = true;
	try {
		const result = await uploadApi.uploadImage(file);
		// Store the path, not the URL
		faviconURL.value = result.path;
	} catch (error) {
		message.value = { type: "error", text: error instanceof Error ? error.message : t("settings.design.uploadError") };
	} finally {
		isUploadingFavicon.value = false;
		input.value = "";
	}
};

const handleLoginBackgroundUpload = async (event: Event) => {
	const input = event.target as HTMLInputElement;
	const file = input.files?.[0];
	if (!file) return;

	isUploadingLoginBackground.value = true;
	try {
		const result = await uploadApi.uploadImage(file);
		// Store the path, not the URL
		loginBackgroundURL.value = result.path;
	} catch (error) {
		message.value = { type: "error", text: error instanceof Error ? error.message : t("settings.design.uploadError") };
	} finally {
		isUploadingLoginBackground.value = false;
		input.value = "";
	}
};

// Load users when switching to users tab
watch(activeTab, (tab) => {
	if (tab === "users" && users.value.length === 0) {
		loadUsers();
	}
});

onMounted(() => {
	loadSettings();
});
</script>

<template>
	<div class="settings-page">
		<div v-if="isLoading" class="loading">
			<p>{{ $t("settings.loadingSettings") }}</p>
		</div>

		<div v-else-if="loadError" class="error-state">
			<UISysIcon :icon="loadError === 'permission' ? 'fa-solid fa-lock' : 'fa-solid fa-exclamation-circle'" />
			<p>{{ loadError === 'permission' ? $t("settings.noPermission") : $t("settings.loadError") }}</p>
			<NuxtLink v-if="loadError === 'permission'" class="btn btn-secondary" :to="localePath('/forms')">
				{{ $t("common.back") }}
			</NuxtLink>
		</div>

		<template v-else>
			<!-- Header -->
			<div class="page-header">
				<div class="page-header-top">
					<NuxtLink class="back-link" :to="localePath('/forms')">
						<UISysIcon icon="fa-solid fa-arrow-left" />
						<span>{{ $t("common.back") }}</span>
					</NuxtLink>
				</div>
				<div class="page-header-content">
					<h1>{{ $t("settings.title") }}</h1>
					<p class="page-description">{{ $t("settings.description") }}</p>
				</div>
				<nav class="tabs">
					<button :class="['tab', { 'tab-active': activeTab === 'general' }]" @click="activeTab = 'general'">
						<UISysIcon icon="fa-solid fa-gear" />
						<span>{{ $t("settings.tabs.general") }}</span>
					</button>
					<button :class="['tab', { 'tab-active': activeTab === 'design' }]" @click="activeTab = 'design'">
						<UISysIcon icon="fa-solid fa-palette" />
						<span>{{ $t("settings.tabs.design") }}</span>
					</button>
					<button :class="['tab', { 'tab-active': activeTab === 'footer' }]" @click="activeTab = 'footer'">
						<UISysIcon icon="fa-solid fa-link" />
						<span>{{ $t("settings.tabs.footer") }}</span>
					</button>
					<button :class="['tab', { 'tab-active': activeTab === 'users' }]" @click="activeTab = 'users'">
						<UISysIcon icon="fa-solid fa-users" />
						<span>{{ $t("settings.tabs.users") }}</span>
					</button>
				</nav>
			</div>

			<!-- Message -->
			<div v-if="message" :class="['message', message.type]">
				<UISysIcon :icon="message.type === 'success' ? 'fa-solid fa-check-circle' : 'fa-solid fa-exclamation-circle'" />
				{{ message.text }}
			</div>

			<!-- General Tab -->
			<div v-if="activeTab === 'general'" class="tab-content">
				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-palette" />
						<h2>{{ $t("settings.general.application") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-group">
							<label class="label" for="appName">{{ $t("settings.general.appName") }}</label>
							<input id="appName" v-model="settings!.app_name" class="input" type="text" />
							<p class="form-hint">{{ $t("settings.general.appNameHint") }}</p>
						</div>
					</div>
				</div>

				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-globe" />
						<h2>{{ $t("settings.general.localization") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-row">
							<div class="form-group">
								<label class="label" for="language">{{ $t("settings.general.language") }}</label>
								<select id="language" v-model="selectedLanguage" class="input">
									<option v-for="loc in availableLocales" :key="loc.code" :value="loc.code">
										{{ loc.name }}
									</option>
								</select>
								<p class="form-hint">{{ $t("settings.general.languageHint") }}</p>
							</div>
							<div class="form-group">
								<label class="label" for="theme">{{ $t("settings.general.theme") }}</label>
								<select id="theme" v-model="selectedTheme" class="input">
									<option value="system">{{ $t("settings.general.themeSystem") }}</option>
									<option value="light">{{ $t("settings.general.themeLight") }}</option>
									<option value="dark">{{ $t("settings.general.themeDark") }}</option>
								</select>
								<p class="form-hint">{{ $t("settings.general.themeHint") }}</p>
							</div>
						</div>
					</div>
				</div>

				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-shield" />
						<h2>{{ $t("settings.general.security") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-group">
							<label class="toggle-label">
								<input v-model="settings!.allow_registration" type="checkbox" class="toggle-input" />
								<span class="toggle-switch" />
								<span class="toggle-text">{{ $t("settings.general.allowRegistration") }}</span>
							</label>
							<p class="form-hint">{{ $t("settings.general.allowRegistrationHint") }}</p>
						</div>
					</div>
				</div>

				<div class="actions">
					<button :disabled="isSaving" class="btn btn-primary" @click="handleSave">
						<UISysIcon icon="fa-solid fa-floppy-disk" />
						{{ isSaving ? $t("common.loading") : $t("common.save") }}
					</button>
				</div>
			</div>

			<!-- Design Tab -->
			<div v-if="activeTab === 'design'" class="tab-content">
				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-swatchbook" />
						<h2>{{ $t("settings.design.colors") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-group">
							<label class="label" for="primaryColor">{{ $t("settings.design.primaryColor") }}</label>
							<div class="color-picker-wrapper">
								<input
									id="primaryColor"
									v-model="primaryColor"
									type="color"
									class="color-picker"
								/>
								<input
									v-model="primaryColor"
									type="text"
									class="input color-input"
									placeholder="#6366f1"
								/>
							</div>
							<p class="form-hint">{{ $t("settings.design.primaryColorHint") }}</p>
						</div>
						<div class="color-preview">
							<span class="preview-label">{{ $t("settings.design.preview") }}:</span>
							<div class="preview-buttons">
								<button class="preview-btn" :style="{ background: primaryColor }">
									{{ $t("settings.design.button") }}
								</button>
								<span class="preview-link" :style="{ color: primaryColor }">{{ $t("settings.design.link") }}</span>
							</div>
						</div>
					</div>
				</div>

				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-image" />
						<h2>{{ $t("settings.design.branding") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-group">
							<label class="label">{{ $t("settings.design.logo") }}</label>
							<p class="form-hint upload-hint">{{ $t("settings.design.logoHint") }}</p>
							<div class="upload-area">
								<div v-if="logoURL" class="image-preview logo-preview">
									<img :src="logoDisplayURL" alt="Logo" />
									<button class="remove-image" :title="$t('settings.design.removeLogo')" @click="logoURL = ''">
										<UISysIcon icon="fa-solid fa-xmark" />
									</button>
								</div>
								<label v-else class="upload-placeholder" :class="{ uploading: isUploadingLogo }">
									<input
										type="file"
										accept="image/*"
										class="file-input"
										:disabled="isUploadingLogo"
										@change="handleLogoUpload"
									/>
									<UISysIcon v-if="!isUploadingLogo" icon="fa-solid fa-cloud-arrow-up" />
									<UISysIcon v-else icon="fa-solid fa-spinner fa-spin" />
									<span>{{ isUploadingLogo ? $t("settings.design.uploading") : $t("settings.design.uploadLogo") }}</span>
								</label>
							</div>
							<label v-if="logoURL" class="toggle-label logo-text-toggle">
								<input v-model="logoShowText" type="checkbox" class="toggle-input" />
								<span class="toggle-switch" />
								<span class="toggle-text">{{ $t("settings.design.showAppName") }}</span>
							</label>
						</div>

						<div class="form-group">
							<label class="label">{{ $t("settings.design.favicon") }}</label>
							<p class="form-hint upload-hint">{{ $t("settings.design.faviconHint") }}</p>
							<div class="upload-area">
								<div v-if="faviconURL" class="image-preview favicon-preview">
									<img :src="faviconDisplayURL" alt="Favicon" />
									<button class="remove-image" :title="$t('settings.design.removeFavicon')" @click="faviconURL = ''">
										<UISysIcon icon="fa-solid fa-xmark" />
									</button>
								</div>
								<label v-else class="upload-placeholder" :class="{ uploading: isUploadingFavicon }">
									<input
										type="file"
										accept="image/*,.ico"
										class="file-input"
										:disabled="isUploadingFavicon"
										@change="handleFaviconUpload"
									/>
									<UISysIcon v-if="!isUploadingFavicon" icon="fa-solid fa-cloud-arrow-up" />
									<UISysIcon v-else icon="fa-solid fa-spinner fa-spin" />
									<span>{{ isUploadingFavicon ? $t("settings.design.uploading") : $t("settings.design.uploadFavicon") }}</span>
								</label>
							</div>
						</div>
					</div>
				</div>

				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-right-to-bracket" />
						<h2>{{ $t("settings.design.loginPage") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-group">
							<label class="label">{{ $t("settings.design.backgroundImage") }}</label>
							<p class="form-hint upload-hint">{{ $t("settings.design.backgroundImageHint") }}</p>
							<div class="upload-area">
								<div v-if="loginBackgroundURL" class="image-preview background-preview">
									<img :src="loginBackgroundDisplayURL" alt="Login Background" />
									<button class="remove-image" :title="$t('settings.design.removeBackground')" @click="loginBackgroundURL = ''">
										<UISysIcon icon="fa-solid fa-xmark" />
									</button>
								</div>
								<label v-else class="upload-placeholder background-upload" :class="{ uploading: isUploadingLoginBackground }">
									<input
										type="file"
										accept="image/*"
										class="file-input"
										:disabled="isUploadingLoginBackground"
										@change="handleLoginBackgroundUpload"
									/>
									<UISysIcon v-if="!isUploadingLoginBackground" icon="fa-solid fa-cloud-arrow-up" />
									<UISysIcon v-else icon="fa-solid fa-spinner fa-spin" />
									<span>{{ isUploadingLoginBackground ? $t("settings.design.uploading") : $t("settings.design.uploadBackground") }}</span>
								</label>
							</div>
							<p v-if="!loginBackgroundURL" class="form-hint">{{ $t("settings.design.noBackgroundHint") }}</p>
						</div>
					</div>
				</div>

				<div class="actions">
					<button :disabled="isSaving" class="btn btn-primary" @click="handleSave">
						<UISysIcon icon="fa-solid fa-floppy-disk" />
						{{ isSaving ? $t("common.loading") : $t("common.save") }}
					</button>
				</div>
			</div>

			<!-- Footer Links Tab -->
			<div v-if="activeTab === 'footer'" class="tab-content">
				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-link" />
						<h2>{{ $t("settings.footer.title") }}</h2>
					</div>
					<div class="card-body">
						<p class="section-description">
							{{ $t("settings.footer.description") }}
						</p>

						<div v-if="footerLinks.length === 0" class="empty-links">
							<UISysIcon icon="fa-solid fa-link-slash" />
							<p>{{ $t("settings.footer.noLinks") }}</p>
						</div>

						<div v-else class="footer-links-list">
							<div v-for="(link, index) in footerLinks" :key="index" class="footer-link-item">
								<div class="link-inputs">
									<input
										v-model="link.label"
										class="input"
										type="text"
										:placeholder="$t('settings.footer.labelPlaceholder')"
									/>
									<input
										v-model="link.url"
										class="input"
										type="url"
										:placeholder="$t('settings.footer.urlPlaceholder')"
									/>
								</div>
								<button class="btn-icon-danger" :title="$t('settings.footer.removeLink')" @click="removeFooterLink(index)">
									<UISysIcon icon="fa-solid fa-trash" />
								</button>
							</div>
						</div>

						<button class="btn btn-secondary add-link-btn" @click="addFooterLink">
							<UISysIcon icon="fa-solid fa-plus" />
							{{ $t("settings.footer.addLink") }}
						</button>
					</div>
				</div>

				<div class="actions">
					<button :disabled="isSaving" class="btn btn-primary" @click="handleSave">
						<UISysIcon icon="fa-solid fa-floppy-disk" />
						{{ isSaving ? $t("common.loading") : $t("common.save") }}
					</button>
				</div>
			</div>

			<!-- Users Tab -->
			<div v-if="activeTab === 'users'" class="tab-content">
				<div class="card">
					<div class="card-header">
						<div class="card-header-left">
							<UISysIcon icon="fa-solid fa-users" />
							<h2>{{ $t("settings.users.title") }}</h2>
						</div>
						<button class="btn btn-primary btn-sm" @click="openAddUserModal">
							<UISysIcon icon="fa-solid fa-plus" />
							{{ $t("settings.users.addUser") }}
						</button>
					</div>
					<div class="card-body">
						<div v-if="isLoadingUsers" class="loading-users">
							<p>{{ $t("settings.users.loadingUsers") }}</p>
						</div>

						<div v-else-if="users.length === 0" class="empty-users">
							<UISysIcon icon="fa-solid fa-user-slash" />
							<p>{{ $t("settings.users.noUsers") }}</p>
						</div>

						<div v-else class="users-list">
							<div v-for="user in users" :key="user.id" class="user-item">
								<div class="user-avatar">
									{{ user.name.charAt(0).toUpperCase() }}
								</div>
								<div class="user-info">
									<div class="user-name">
										{{ user.name }}
										<span v-if="user.id === authStore.user?.id" class="badge-you">{{ $t("common.you") }}</span>
										<span :class="['badge-role', user.role === 'admin' ? 'badge-admin' : 'badge-user']">
											{{ user.role === "admin" ? $t("settings.users.admin") : $t("settings.users.user") }}
										</span>
									</div>
									<div class="user-email">{{ user.email }}</div>
									<div class="user-meta">{{ $t("settings.users.createdAt") }} {{ formatDate(user.created_at) }}</div>
								</div>
								<div class="user-actions">
									<button class="btn-icon" :title="$t('common.edit')" @click="openEditUserModal(user)">
										<UISysIcon icon="fa-solid fa-pen" />
									</button>
									<button
										v-if="user.id !== authStore.user?.id"
										class="btn-icon-danger"
										:title="$t('common.delete')"
										@click="handleDeleteUser(user)"
									>
										<UISysIcon icon="fa-solid fa-trash" />
									</button>
								</div>
							</div>

							<!-- Pagination -->
							<UIPagination
								:page="usersPagination.state.page"
								:page-size="usersPagination.state.pageSize"
								:total-items="usersPagination.state.totalItems"
								:total-pages="usersPagination.state.totalPages"
								:visible-pages="usersPagination.visiblePages.value"
								:has-next-page="usersPagination.hasNextPage.value"
								:has-prev-page="usersPagination.hasPrevPage.value"
								@update:page="usersPagination.setPage"
								@update:page-size="usersPagination.setPageSize"
								@first="usersPagination.firstPage"
								@prev="usersPagination.prevPage"
								@next="usersPagination.nextPage"
								@last="usersPagination.lastPage"
							/>
						</div>
					</div>
				</div>
			</div>

			<!-- User Modal -->
			<Teleport to="body">
				<div v-if="showUserModal" class="modal-overlay" @click.self="closeUserModal">
					<div class="modal">
						<div class="modal-header">
							<h3>{{ editingUser ? $t("settings.users.editUserTitle") : $t("settings.users.createUser") }}</h3>
							<button class="modal-close" @click="closeUserModal">
								<UISysIcon icon="fa-solid fa-xmark" />
							</button>
						</div>
						<div class="modal-body">
							<div v-if="userFormError" class="form-error">
								<UISysIcon icon="fa-solid fa-exclamation-circle" />
								{{ userFormError }}
							</div>

							<div class="form-group">
								<label class="label" for="userName">{{ $t("auth.name") }}</label>
								<input id="userName" v-model="userForm.name" class="input" type="text" placeholder="Max Mustermann" />
							</div>

							<div class="form-group">
								<label class="label" for="userEmail">{{ $t("auth.email") }}</label>
								<input id="userEmail" v-model="userForm.email" class="input" type="email" placeholder="max@example.com" />
							</div>

							<div class="form-group">
								<label class="label" for="userPassword">
									{{ $t("auth.password") }}
									<span v-if="editingUser" class="label-hint">{{ $t("settings.users.leaveEmptyToKeep") }}</span>
								</label>
								<input id="userPassword" v-model="userForm.password" class="input" type="password" placeholder="••••••••" />
							</div>

							<div class="form-group">
								<label class="label" for="userRole">{{ $t("settings.users.role") }}</label>
								<select id="userRole" v-model="userForm.role" class="input">
									<option value="user">{{ $t("settings.users.user") }}</option>
									<option value="admin">{{ $t("settings.users.administrator") }}</option>
								</select>
								<p class="form-hint">{{ $t("settings.users.roleHint") }}</p>
							</div>
						</div>
						<div class="modal-footer">
							<button class="btn btn-secondary" @click="closeUserModal">{{ $t("common.cancel") }}</button>
							<button class="btn btn-primary" :disabled="isSavingUser" @click="handleSaveUser">
								<UISysIcon icon="fa-solid fa-floppy-disk" />
								{{ isSavingUser ? $t("common.loading") : $t("common.save") }}
							</button>
						</div>
					</div>
				</div>
			</Teleport>
		</template>
	</div>
</template>

<style scoped>
.settings-page {
	max-width: 800px;
	margin: 0 auto;
}

.loading,
.error-state {
	display: flex;
	flex-direction: column;
	gap: 1rem;
	justify-content: center;
	align-items: center;
	min-height: 200px;
	color: var(--text-secondary);
	text-align: center;
	padding: 2rem;
}

.error-state i {
	font-size: 2.5rem;
	opacity: 0.6;
}

.error-state p {
	max-width: 400px;
	line-height: 1.5;
}

/* Page Header */
.page-header {
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	margin-bottom: 1.5rem;
	overflow: hidden;
}

.page-header-top {
	padding: 0.75rem 1.25rem;
	border-bottom: 1px solid var(--border);
}

.back-link {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem 0.75rem;
	font-size: 0.8125rem;
	font-weight: 500;
	color: var(--text-secondary);
	text-decoration: none;
	border-radius: var(--radius);
	transition: all 0.15s ease;
}

.back-link:hover {
	color: var(--text);
	background: var(--surface-hover);
	text-decoration: none;
}

.page-header-content {
	padding: 1.25rem;
}

.page-header-content h1 {
	font-size: 1.5rem;
	font-weight: 600;
	color: var(--text);
	margin-bottom: 0.25rem;
}

.page-description {
	font-size: 0.875rem;
	color: var(--text-secondary);
}

/* Tabs */
.tabs {
	display: flex;
	gap: 0;
	padding: 0 0rem;
	border-top: 1px solid var(--border);
	background: var(--surface-hover);
}

.tab {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.875rem 1.25rem;
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text-secondary);
	background: transparent;
	border: none;
	border-radius: 0 0 0 0;
	cursor: pointer;
	transition: all 0.15s ease;
}

.tab:hover {
	color: var(--text);
	background: var(--surface);
}

.tab-active {
	color: var(--primary);
	background: var(--surface);
}

/* Message */
.message {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.875rem 1rem;
	margin-bottom: 1.5rem;
	font-size: 0.875rem;
	border-radius: var(--radius);
}

.message.success {
	color: var(--success);
	background: rgba(34, 197, 94, 0.1);
}

.message.error {
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
}

/* Tab Content */
.tab-content {
	display: flex;
	flex-direction: column;
	gap: 1.5rem;
}

/* Card */
.card {
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	overflow: hidden;
}

.card-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.625rem;
	padding: 1rem 1.25rem;
	border-bottom: 1px solid var(--border);
	background: var(--surface-hover);
}

.card-header-left {
	display: flex;
	align-items: center;
	gap: 0.625rem;
}

.card-header h2 {
	font-size: 0.9375rem;
	font-weight: 600;
}

.card-header i {
	color: var(--text-secondary);
}

.card-body {
	padding: 1.5rem;
}

.section-description {
	font-size: 0.875rem;
	color: var(--text-secondary);
	margin-bottom: 1.25rem;
}

/* Form */
.form-row {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 1rem;
}

.form-group {
	margin-bottom: 1.25rem;
}

.form-group:last-child {
	margin-bottom: 0;
}

.form-row .form-group {
	margin-bottom: 0;
}

/* Toggle Switch */
.toggle-label {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	cursor: pointer;
}

.toggle-input {
	display: none;
}

.toggle-switch {
	position: relative;
	width: 44px;
	height: 24px;
	background: var(--border);
	border-radius: 12px;
	transition: background 0.2s ease;
}

.toggle-switch::after {
	content: "";
	position: absolute;
	top: 2px;
	left: 2px;
	width: 20px;
	height: 20px;
	background: white;
	border-radius: 50%;
	transition: transform 0.2s ease;
}

.toggle-input:checked + .toggle-switch {
	background: var(--primary);
}

.toggle-input:checked + .toggle-switch::after {
	transform: translateX(20px);
}

.toggle-text {
	font-size: 0.9375rem;
	font-weight: 500;
}

/* Footer Links */
.empty-links {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.5rem;
	padding: 2rem;
	color: var(--text-secondary);
	text-align: center;
}

.empty-links i {
	font-size: 2rem;
	opacity: 0.5;
}

.footer-links-list {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
	margin-bottom: 1rem;
}

.footer-link-item {
	display: flex;
	align-items: center;
	gap: 0.75rem;
}

.link-inputs {
	display: flex;
	flex: 1;
	gap: 0.5rem;
}

.link-inputs .input {
	flex: 1;
}

.add-link-btn {
	margin-top: 0.5rem;
}

/* Users */
.loading-users,
.empty-users {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.5rem;
	padding: 2rem;
	color: var(--text-secondary);
	text-align: center;
}

.empty-users i {
	font-size: 2rem;
	opacity: 0.5;
}

.users-list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.user-item {
	display: flex;
	align-items: center;
	gap: 1rem;
	padding: 1rem;
	background: var(--background);
	border-radius: var(--radius);
	transition: background 0.15s ease;
}

.user-item:hover {
	background: var(--surface-hover);
}

.user-avatar {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 40px;
	height: 40px;
	font-size: 1rem;
	font-weight: 600;
	color: white;
	background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
	border-radius: 50%;
	flex-shrink: 0;
}

.user-info {
	flex: 1;
	min-width: 0;
}

.user-name {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	font-weight: 600;
	color: var(--text);
}

.badge-you {
	font-size: 0.6875rem;
	font-weight: 500;
	padding: 0.125rem 0.375rem;
	background: rgba(99, 102, 241, 0.1);
	color: var(--primary);
	border-radius: var(--radius);
}

.badge-role {
	font-size: 0.6875rem;
	font-weight: 500;
	padding: 0.125rem 0.375rem;
	border-radius: var(--radius);
}

.badge-admin {
	background: rgba(245, 158, 11, 0.1);
	color: var(--warning);
}

.badge-user {
	background: rgba(107, 114, 128, 0.1);
	color: var(--text-secondary);
}

.user-email {
	font-size: 0.875rem;
	color: var(--text-secondary);
}

.user-meta {
	font-size: 0.75rem;
	color: var(--text-secondary);
	margin-top: 0.25rem;
}

.user-actions {
	display: flex;
	gap: 0.25rem;
}

/* Icon Buttons */
.btn-icon,
.btn-icon-danger {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 32px;
	height: 32px;
	color: var(--text-secondary);
	background: none;
	border: none;
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s ease;
}

.btn-icon:hover {
	color: var(--text);
	background: var(--surface);
}

.btn-icon-danger:hover {
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
}

/* Actions */
.actions {
	display: flex;
	justify-content: flex-end;
}

/* Modal */
.modal-overlay {
	position: fixed;
	inset: 0;
	z-index: 1000;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 1rem;
	background: rgba(0, 0, 0, 0.5);
	backdrop-filter: blur(4px);
}

.modal {
	width: 100%;
	max-width: 480px;
	background: var(--surface);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-lg);
	overflow: hidden;
}

.modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 1rem 1.25rem;
	border-bottom: 1px solid var(--border);
}

.modal-header h3 {
	font-size: 1.125rem;
	font-weight: 600;
}

.modal-close {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 32px;
	height: 32px;
	color: var(--text-secondary);
	background: none;
	border: none;
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s ease;
}

.modal-close:hover {
	color: var(--text);
	background: var(--surface-hover);
}

.modal-body {
	padding: 1.5rem;
}

.modal-footer {
	display: flex;
	justify-content: flex-end;
	gap: 0.75rem;
	padding: 1rem 1.25rem;
	border-top: 1px solid var(--border);
	background: var(--background);
}

.form-error {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.75rem 1rem;
	margin-bottom: 1rem;
	font-size: 0.875rem;
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
	border-radius: var(--radius);
}

.label-hint {
	font-weight: 400;
	color: var(--text-secondary);
}

/* Design Tab - Color Picker */
.color-picker-wrapper {
	display: flex;
	gap: 0.75rem;
	align-items: center;
}

.color-picker {
	width: 48px;
	height: 40px;
	padding: 0;
	border: 1px solid var(--border);
	border-radius: var(--radius);
	cursor: pointer;
	background: transparent;
}

.color-picker::-webkit-color-swatch-wrapper {
	padding: 4px;
}

.color-picker::-webkit-color-swatch {
	border: none;
	border-radius: 4px;
}

.color-input {
	flex: 1;
	max-width: 140px;
	font-family: monospace;
	text-transform: uppercase;
}

.color-preview {
	display: flex;
	align-items: center;
	gap: 1rem;
	padding: 1rem;
	margin-top: 1rem;
	background: var(--background);
	border-radius: var(--radius);
}

.preview-label {
	font-size: 0.8125rem;
	color: var(--text-secondary);
}

.preview-buttons {
	display: flex;
	align-items: center;
	gap: 1rem;
}

.preview-btn {
	padding: 0.5rem 1rem;
	font-size: 0.875rem;
	font-weight: 500;
	color: white;
	border: none;
	border-radius: var(--radius);
	cursor: default;
}

.preview-link {
	font-size: 0.875rem;
	font-weight: 500;
	cursor: default;
}

/* Design Tab - Upload */
.upload-hint {
	margin-bottom: 0.75rem;
}

.upload-area {
	display: flex;
	gap: 1rem;
}

.upload-placeholder {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.5rem;
	width: 200px;
	height: 100px;
	padding: 1rem;
	color: var(--text-secondary);
	font-size: 0.8125rem;
	background: var(--background);
	border: 2px dashed var(--border);
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s ease;
}

.upload-placeholder:hover {
	color: var(--primary);
	border-color: var(--primary);
	background: rgba(99, 102, 241, 0.05);
}

.upload-placeholder.uploading {
	cursor: wait;
	opacity: 0.7;
}

.upload-placeholder i {
	font-size: 1.5rem;
}

.file-input {
	display: none;
}

.image-preview {
	position: relative;
	display: flex;
	align-items: center;
	justify-content: center;
	background: var(--background);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	overflow: hidden;
}

.logo-preview {
	width: 200px;
	height: 100px;
	padding: 0.5rem;
}

.logo-preview img {
	max-width: 100%;
	max-height: 100%;
	object-fit: contain;
}

.favicon-preview {
	width: 80px;
	height: 80px;
	padding: 0.5rem;
}

.favicon-preview img {
	max-width: 100%;
	max-height: 100%;
	object-fit: contain;
}

.remove-image {
	position: absolute;
	top: 4px;
	right: 4px;
	display: flex;
	align-items: center;
	justify-content: center;
	width: 24px;
	height: 24px;
	color: white;
	background: rgba(0, 0, 0, 0.5);
	border: none;
	border-radius: 50%;
	cursor: pointer;
	transition: background 0.15s ease;
}

.remove-image:hover {
	background: var(--error);
}

.logo-text-toggle {
	margin-top: 1rem;
}

.background-preview {
	width: 100%;
	max-width: 320px;
	height: 180px;
}

.background-preview img {
	width: 100%;
	height: 100%;
	object-fit: cover;
}

.background-upload {
	width: 100%;
	max-width: 320px;
	height: 120px;
}

/* Responsive */
@media (max-width: 640px) {
	.tabs {
		overflow-x: auto;
	}

	.tab span {
		display: none;
	}

	.link-inputs {
		flex-direction: column;
	}

	.user-item {
		flex-wrap: wrap;
	}

	.user-actions {
		width: 100%;
		justify-content: flex-end;
		margin-top: 0.5rem;
		padding-top: 0.5rem;
		border-top: 1px solid var(--border);
	}
}
</style>
