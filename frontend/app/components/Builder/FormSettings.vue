<script lang="ts" setup>
const { t } = useI18n();

const props = defineProps<{
	form: Form;
	slugInput?: string;
	slugStatus?: "idle" | "checking" | "available" | "taken" | "invalid";
	passwordEnabled?: boolean;
	passwordInput?: string;
	formLink?: string;
}>();

const emit = defineEmits(["update:form", "update:passwordEnabled", "update:passwordInput", "slugInput", "copyLink", "markDirty"]);

const activeTab = ref<"general" | "design" | "access">("general");
const showPassword = ref(false);
const toastStore = useToastStore();

const updateStatus = (newStatus: string) => {
	emit("update:form", { ...props.form, status: newStatus });
	emit("markDirty");
};

const statusInfo = computed(() => {
	const now = new Date();
	const startDate = props.form.settings.start_date ? new Date(props.form.settings.start_date) : null;
	const endDate = props.form.settings.end_date ? new Date(props.form.settings.end_date) : null;

	if (props.form.status === "published") {
		if (startDate && now < startDate) {
			return {
				type: "warning" as const,
				message: t("builder.formSettings.statusWarningFuture"),
			};
		}
		if (endDate && now > endDate) {
			return {
				type: "warning" as const,
				message: t("builder.formSettings.statusWarningExpired"),
			};
		}
	}

	if (props.form.status === "closed") {
		return {
			type: "info" as const,
			message: t("builder.formSettings.statusInfoClosed"),
		};
	}

	return null;
});

const updateSettings = (key: string, value: unknown) => {
	const newSettings = { ...props.form.settings, [key]: value };
	emit("update:form", { ...props.form, settings: newSettings });
	emit("markDirty");
};

const updateDesign = (key: string, value: unknown) => {
	const newDesign = { ...props.form.settings.design, [key]: value };
	const newSettings = { ...props.form.settings, design: newDesign };
	emit("update:form", { ...props.form, settings: newSettings });
	emit("markDirty");
};

const handleSlugInput = (event: Event) => {
	const input = event.target as HTMLInputElement;
	emit("slugInput", input.value);
};

const handleCopyLink = async () => {
	try {
		await navigator.clipboard.writeText(props.formLink || "");
		toastStore.success(t("builder.formSettings.linkCopied"), t("builder.formSettings.linkCopiedMessage"));
	} catch {
		toastStore.error(t("builder.formSettings.linkCopyError"), t("builder.formSettings.linkCopyErrorMessage"));
	}
};
</script>

<template>
	<div class="form-settings">
		<nav class="settings-tabs">
			<button
				:class="['tab', { 'tab-active': activeTab === 'general' }]"
				@click="activeTab = 'general'"
			>
				<UISysIcon icon="fa-solid fa-sliders" />
				<span>{{ $t("builder.formSettings.tabs.general") }}</span>
			</button>
			<button
				:class="['tab', { 'tab-active': activeTab === 'design' }]"
				@click="activeTab = 'design'"
			>
				<UISysIcon icon="fa-solid fa-palette" />
				<span>{{ $t("builder.formSettings.tabs.design") }}</span>
			</button>
			<button
				:class="['tab', { 'tab-active': activeTab === 'access' }]"
				@click="activeTab = 'access'"
			>
				<UISysIcon icon="fa-solid fa-link" />
				<span>{{ $t("builder.formSettings.tabs.access") }}</span>
			</button>
		</nav>

		<div class="settings-content">
			<!-- General Tab -->
			<div v-if="activeTab === 'general'" class="tab-content">
				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-pen" />
						<h2>{{ $t("builder.formSettings.form") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-group">
							<label class="label">{{ $t("builder.formSettings.buttonText") }}</label>
							<input
								:value="form.settings.submit_button_text"
								class="input"
								type="text"
								@input="updateSettings('submit_button_text', ($event.target as HTMLInputElement).value)"
							/>
							<p class="form-hint">{{ $t("builder.formSettings.buttonTextHint") }}</p>
						</div>

						<div class="form-group">
							<label class="label">{{ $t("builder.formSettings.successMessage") }}</label>
							<textarea
								:value="form.settings.success_message"
								class="input"
								rows="3"
								@input="updateSettings('success_message', ($event.target as HTMLTextAreaElement).value)"
							/>
							<p class="form-hint">{{ $t("builder.formSettings.successMessageHint") }}</p>
						</div>

						<div class="form-group">
							<label class="checkbox-label">
								<input
									:checked="form.settings.allow_multiple"
									type="checkbox"
									@change="updateSettings('allow_multiple', ($event.target as HTMLInputElement).checked)"
								/>
								{{ $t("builder.formSettings.allowMultiple") }}
							</label>
						</div>
					</div>
				</div>

				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-clock" />
						<h2>{{ $t("builder.formSettings.availability") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-group">
							<label class="label">{{ $t("builder.formSettings.status") }}</label>
							<div class="status-selector">
								<button
									:class="['status-option', { active: form.status === 'draft' }]"
									type="button"
									@click="updateStatus('draft')"
								>
									<UISysIcon icon="fa-solid fa-file-pen" />
									<span>{{ $t("builder.formSettings.draft") }}</span>
								</button>
								<button
									:class="['status-option', { active: form.status === 'published' }]"
									type="button"
									@click="updateStatus('published')"
								>
									<UISysIcon icon="fa-solid fa-globe" />
									<span>{{ $t("builder.formSettings.published") }}</span>
								</button>
								<button
									:class="['status-option', { active: form.status === 'closed' }]"
									type="button"
									@click="updateStatus('closed')"
								>
									<UISysIcon icon="fa-solid fa-lock" />
									<span>{{ $t("builder.formSettings.closed") }}</span>
								</button>
							</div>
							<div v-if="statusInfo" :class="['status-info', statusInfo.type]">
								<UISysIcon :icon="statusInfo.type === 'warning' ? 'fa-solid fa-triangle-exclamation' : 'fa-solid fa-info-circle'" />
								<span>{{ statusInfo.message }}</span>
							</div>
						</div>

						<div class="form-row">
							<div class="form-group">
								<label class="label">{{ $t("builder.formSettings.startDateTime") }}</label>
								<div class="input-with-reset">
									<input
										:value="form.settings.start_date"
										:class="['input', { 'has-value': form.settings.start_date }]"
										type="datetime-local"
										@input="updateSettings('start_date', ($event.target as HTMLInputElement).value)"
									/>
									<button
										v-if="form.settings.start_date"
										class="btn-reset"
										type="button"
										:title="$t('builder.formSettings.reset')"
										@click="updateSettings('start_date', '')"
									>
										<UISysIcon icon="fa-solid fa-xmark" />
									</button>
								</div>
							</div>

							<div class="form-group">
								<label class="label">{{ $t("builder.formSettings.endDateTime") }}</label>
								<div class="input-with-reset">
									<input
										:value="form.settings.end_date"
										:class="['input', { 'has-value': form.settings.end_date }]"
										type="datetime-local"
										@input="updateSettings('end_date', ($event.target as HTMLInputElement).value)"
									/>
									<button
										v-if="form.settings.end_date"
										class="btn-reset"
										type="button"
										:title="$t('builder.formSettings.reset')"
										@click="updateSettings('end_date', '')"
									>
										<UISysIcon icon="fa-solid fa-xmark" />
									</button>
								</div>
							</div>
						</div>
						<p class="form-hint inline-hint">{{ $t("builder.formSettings.noTimeRestriction") }}</p>

						<div class="form-group">
							<label class="label">{{ $t("builder.formSettings.maxSubmissions") }}</label>
							<div class="input-with-reset" style="max-width: 200px">
								<input
									:value="form.settings.max_submissions"
									:class="['input', { 'has-value': form.settings.max_submissions && form.settings.max_submissions > 0 }]"
									min="0"
									type="number"
									@input="updateSettings('max_submissions', Number(($event.target as HTMLInputElement).value))"
								/>
								<button
									v-if="form.settings.max_submissions && form.settings.max_submissions > 0"
									class="btn-reset"
									type="button"
									:title="$t('builder.formSettings.reset')"
									@click="updateSettings('max_submissions', 0)"
								>
									<UISysIcon icon="fa-solid fa-xmark" />
								</button>
							</div>
							<p class="form-hint">{{ $t("builder.formSettings.maxSubmissionsHint") }}</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Design Tab -->
			<div v-if="activeTab === 'design'" class="tab-content">
				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-swatchbook" />
						<h2>{{ $t("builder.formSettings.colors") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-row">
							<div class="form-group">
								<label class="label">{{ $t("builder.formSettings.primaryColor") }}</label>
								<div class="color-input-row">
									<input
										:value="form.settings.design?.primaryColor || '#6366f1'"
										class="color-input"
										type="color"
										@input="updateDesign('primaryColor', ($event.target as HTMLInputElement).value)"
									/>
									<input
										:value="form.settings.design?.primaryColor || '#6366f1'"
										class="input color-text"
										type="text"
										@input="updateDesign('primaryColor', ($event.target as HTMLInputElement).value)"
									/>
								</div>
							</div>

							<div class="form-group">
								<label class="label">{{ $t("builder.formSettings.textColor") }}</label>
								<div class="color-input-row">
									<input
										:value="form.settings.design?.textColor || '#111827'"
										class="color-input"
										type="color"
										@input="updateDesign('textColor', ($event.target as HTMLInputElement).value)"
									/>
									<input
										:value="form.settings.design?.textColor || '#111827'"
										class="input color-text"
										type="text"
										@input="updateDesign('textColor', ($event.target as HTMLInputElement).value)"
									/>
								</div>
							</div>
						</div>

						<div class="form-row">
							<div class="form-group">
								<label class="label">{{ $t("builder.formSettings.pageBackground") }}</label>
								<div class="color-input-row">
									<input
										:value="form.settings.design?.backgroundColor || '#f3f4f6'"
										class="color-input"
										type="color"
										@input="updateDesign('backgroundColor', ($event.target as HTMLInputElement).value)"
									/>
									<input
										:value="form.settings.design?.backgroundColor || '#f3f4f6'"
										class="input color-text"
										type="text"
										@input="updateDesign('backgroundColor', ($event.target as HTMLInputElement).value)"
									/>
								</div>
							</div>

							<div class="form-group">
								<label class="label">{{ $t("builder.formSettings.formBackground") }}</label>
								<div class="color-input-row">
									<input
										:value="form.settings.design?.formBackgroundColor || '#ffffff'"
										class="color-input"
										type="color"
										@input="updateDesign('formBackgroundColor', ($event.target as HTMLInputElement).value)"
									/>
									<input
										:value="form.settings.design?.formBackgroundColor || '#ffffff'"
										class="input color-text"
										type="text"
										@input="updateDesign('formBackgroundColor', ($event.target as HTMLInputElement).value)"
									/>
								</div>
							</div>
						</div>
					</div>
				</div>

				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-ruler-combined" />
						<h2>{{ $t("builder.formSettings.layout") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-row">
							<div class="form-group">
								<label class="label">{{ $t("builder.formSettings.formWidth") }}</label>
								<select
									:value="form.settings.design?.maxWidth || 'lg'"
									class="input"
									@change="updateDesign('maxWidth', ($event.target as HTMLSelectElement).value)"
								>
									<option value="sm">{{ $t("builder.formSettings.widthNarrow") }}</option>
									<option value="md">{{ $t("builder.formSettings.widthStandard") }}</option>
									<option value="lg">{{ $t("builder.formSettings.widthWide") }}</option>
									<option value="xl">{{ $t("builder.formSettings.widthVeryWide") }}</option>
								</select>
							</div>

							<div class="form-group">
								<label class="label">{{ $t("builder.formSettings.borderRadius") }}</label>
								<select
									:value="form.settings.design?.borderRadius || 'lg'"
									class="input"
									@change="updateDesign('borderRadius', ($event.target as HTMLSelectElement).value)"
								>
									<option value="none">{{ $t("builder.formSettings.radiusNone") }}</option>
									<option value="sm">{{ $t("builder.formSettings.radiusSmall") }}</option>
									<option value="md">{{ $t("builder.formSettings.radiusMedium") }}</option>
									<option value="lg">{{ $t("builder.formSettings.radiusLarge") }}</option>
								</select>
							</div>
						</div>
					</div>
				</div>

				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-wand-magic-sparkles" />
						<h2>{{ $t("builder.formSettings.style") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-row">
							<div class="form-group">
								<label class="label">{{ $t("builder.formSettings.headerStyle") }}</label>
								<select
									:value="form.settings.design?.headerStyle || 'default'"
									class="input"
									@change="updateDesign('headerStyle', ($event.target as HTMLSelectElement).value)"
								>
									<option value="default">{{ $t("builder.formSettings.headerDefault") }}</option>
									<option value="colored">{{ $t("builder.formSettings.headerColored") }}</option>
									<option value="minimal">{{ $t("builder.formSettings.headerMinimal") }}</option>
								</select>
							</div>

							<div class="form-group">
								<label class="label">{{ $t("builder.formSettings.buttonStyle") }}</label>
								<select
									:value="form.settings.design?.buttonStyle || 'filled'"
									class="input"
									@change="updateDesign('buttonStyle', ($event.target as HTMLSelectElement).value)"
								>
									<option value="filled">{{ $t("builder.formSettings.buttonFilled") }}</option>
									<option value="outline">{{ $t("builder.formSettings.buttonOutline") }}</option>
								</select>
							</div>
						</div>

						<div class="form-group">
							<label class="label">{{ $t("builder.formSettings.fontFamily") }}</label>
							<select
								:value="form.settings.design?.fontFamily || 'default'"
								class="input"
								style="max-width: 300px"
								@change="updateDesign('fontFamily', ($event.target as HTMLSelectElement).value)"
							>
								<option value="default">{{ $t("builder.formSettings.fontDefault") }}</option>
								<option value="serif">{{ $t("builder.formSettings.fontSerif") }}</option>
								<option value="mono">{{ $t("builder.formSettings.fontMono") }}</option>
							</select>
						</div>
					</div>
				</div>

				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-image" />
						<h2>{{ $t("builder.formSettings.backgroundImage") }}</h2>
					</div>
					<div class="card-body">
						<UIImageUpload
							:model-value="form.settings.design?.backgroundImage"
							:label="$t('builder.formSettings.uploadBackgroundImage')"
							@update:model-value="updateDesign('backgroundImage', $event)"
						/>

						<div v-if="form.settings.design?.backgroundImage" class="form-group" style="margin-top: 1rem">
							<label class="label">{{ $t("builder.formSettings.imageSize") }}</label>
							<select
								:value="form.settings.design?.backgroundSize || 'cover'"
								class="input"
								style="max-width: 300px"
								@change="updateDesign('backgroundSize', ($event.target as HTMLSelectElement).value)"
							>
								<option value="cover">{{ $t("builder.formSettings.sizeCover") }}</option>
								<option value="contain">{{ $t("builder.formSettings.sizeContain") }}</option>
								<option value="auto">{{ $t("builder.formSettings.sizeAuto") }}</option>
							</select>
						</div>
					</div>
				</div>
			</div>

			<!-- Access Tab -->
			<div v-if="activeTab === 'access'" class="tab-content">
				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-link" />
						<h2>{{ $t("builder.formSettings.formLink") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-group">
							<label class="label">{{ $t("builder.formSettings.currentLink") }}</label>
							<div class="link-display">
								<input
									:value="formLink"
									class="input link-input"
									readonly
									type="text"
								/>
								<button
									class="btn btn-secondary btn-sm"
									:title="$t('builder.formSettings.copyLink')"
									type="button"
									@click="handleCopyLink"
								>
									<UISysIcon icon="fa-solid fa-copy" />
								</button>
							</div>
						</div>

						<div class="form-group">
							<label class="label">
								{{ $t("builder.formSettings.customUrl") }}
								<span class="label-hint">{{ $t("builder.formSettings.customUrlOptional") }}</span>
							</label>
							<div class="slug-input-wrapper">
								<span class="slug-prefix">/f/</span>
								<input
									:value="slugInput"
									class="input slug-input"
									:placeholder="$t('builder.formSettings.customUrlPlaceholder')"
									type="text"
									@input="handleSlugInput"
								/>
								<span v-if="slugStatus === 'checking'" class="slug-status checking">
									<UISysIcon icon="fa-solid fa-spinner fa-spin" />
								</span>
								<span v-else-if="slugStatus === 'available'" class="slug-status available">
									<UISysIcon icon="fa-solid fa-check" />
								</span>
								<span v-else-if="slugStatus === 'taken'" class="slug-status taken">
									<UISysIcon icon="fa-solid fa-xmark" />
								</span>
								<span v-else-if="slugStatus === 'invalid'" class="slug-status invalid">
									<UISysIcon icon="fa-solid fa-exclamation" />
								</span>
							</div>
							<p v-if="slugStatus === 'taken'" class="form-hint error">
								{{ $t("builder.formSettings.urlTaken") }}
							</p>
							<p v-else-if="slugStatus === 'invalid'" class="form-hint error">
								{{ $t("builder.formSettings.urlInvalid") }}
							</p>
							<p v-else class="form-hint">
								{{ $t("builder.formSettings.urlAutomatic") }}
							</p>
						</div>
					</div>
				</div>

				<div class="card">
					<div class="card-header">
						<UISysIcon icon="fa-solid fa-lock" />
						<h2>{{ $t("builder.formSettings.passwordProtection") }}</h2>
					</div>
					<div class="card-body">
						<div class="form-group">
							<label class="checkbox-label">
								<input
									:checked="passwordEnabled"
									type="checkbox"
									@change="$emit('update:passwordEnabled', ($event.target as HTMLInputElement).checked); $emit('markDirty')"
								/>
								{{ $t("builder.formSettings.enablePassword") }}
							</label>
							<p class="form-hint">{{ $t("builder.formSettings.passwordHint") }}</p>
						</div>

						<div v-if="passwordEnabled" class="form-group">
							<label class="label">
								{{ form.password_protected ? $t("builder.formSettings.setNewPassword") : $t("builder.formSettings.setPassword") }}
							</label>
							<div class="password-input-wrapper">
								<input
									:value="passwordInput"
									:type="showPassword ? 'text' : 'password'"
									class="input"
									:placeholder="form.password_protected ? $t('builder.formSettings.keepExistingPassword') : $t('builder.formSettings.passwordPlaceholder')"
									@input="$emit('update:passwordInput', ($event.target as HTMLInputElement).value); $emit('markDirty')"
								/>
								<button
									class="btn btn-ghost btn-sm"
									type="button"
									@click="showPassword = !showPassword"
								>
									<UISysIcon :icon="showPassword ? 'fa-solid fa-eye-slash' : 'fa-solid fa-eye'" />
								</button>
							</div>
							<p v-if="form.password_protected && !passwordInput" class="form-hint">
								{{ $t("builder.formSettings.passwordAlreadySet") }}
							</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.form-settings {
	display: flex;
	flex-direction: column;
	height: 100%;
}

.settings-tabs {
	display: flex;
	gap: 0;
	padding: 0 1rem;
	background: var(--surface);
	border-bottom: 1px solid var(--border);
}

.tab {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	padding: 0.875rem 1.25rem;
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text-secondary);
	background: transparent;
	border: none;
	border-bottom: 2px solid transparent;
	cursor: pointer;
	transition: all 0.15s ease;
}

.tab:hover {
	color: var(--text);
}

.tab-active {
	color: var(--primary);
	border-bottom-color: var(--primary);
}

.settings-content {
	flex: 1;
	padding: 1.5rem;
	overflow-y: auto;
	background: var(--background);
}

.tab-content {
	display: flex;
	flex-direction: column;
	gap: 1.5rem;
	max-width: 800px;
	margin: 0 auto;
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
	gap: 0.625rem;
	padding: 1rem 1.25rem;
	border-bottom: 1px solid var(--border);
	background: var(--surface-hover);
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

/* Form */
.form-group {
	margin-bottom: 1.25rem;
}

.form-group:last-child {
	margin-bottom: 0;
}

.form-row {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 1rem;
	margin-bottom: 1.25rem;
}

.form-row:last-child {
	margin-bottom: 0;
}

.form-row .form-group {
	margin-bottom: 0;
}

.label {
	display: block;
	margin-bottom: 0.375rem;
	font-size: 0.8125rem;
	font-weight: 500;
	color: var(--text);
}

.label-hint {
	font-weight: 400;
	color: var(--text-secondary);
}

.form-hint {
	margin: 0.375rem 0 0;
	font-size: 0.75rem;
	color: var(--text-secondary);
}

.form-hint.inline-hint {
	margin-top: -0.75rem;
	margin-bottom: 1rem;
}

.form-hint.error {
	color: var(--error);
}

/* Input with Reset Button */
.input-with-reset {
	position: relative;
	display: flex;
	align-items: center;
}

.input-with-reset .input {
	flex: 1;
}

.input-with-reset .input.has-value {
	padding-right: 2.25rem;
}

.btn-reset {
	position: absolute;
	right: 0.5rem;
	display: flex;
	align-items: center;
	justify-content: center;
	width: 1.5rem;
	height: 1.5rem;
	padding: 0;
	font-size: 0.75rem;
	color: var(--text-secondary);
	background: var(--surface-hover);
	border: none;
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s ease;
}

.btn-reset:hover {
	color: var(--error);
	background: var(--error-bg, rgba(239, 68, 68, 0.1));
}

.checkbox-label {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	font-size: 0.875rem;
	cursor: pointer;
}

.checkbox-label input {
	width: 1rem;
	height: 1rem;
	cursor: pointer;
}

/* Color Inputs */
.color-input-row {
	display: flex;
	gap: 0.5rem;
}

.color-input {
	width: 40px;
	height: 38px;
	padding: 2px;
	cursor: pointer;
	border: 1px solid var(--border);
	border-radius: var(--radius);
}

.color-input::-webkit-color-swatch-wrapper {
	padding: 2px;
}

.color-input::-webkit-color-swatch {
	border: none;
	border-radius: 4px;
}

.color-text {
	flex: 1;
	font-family: monospace;
	text-transform: uppercase;
}

/* Link Display */
.link-display {
	display: flex;
	gap: 0.5rem;
}

.link-input {
	flex: 1;
	font-size: 0.8125rem;
	color: var(--text-secondary);
	background: var(--surface);
}

/* Slug Input */
.slug-input-wrapper {
	display: flex;
	align-items: center;
	gap: 0;
	overflow: hidden;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius);
}

.slug-prefix {
	padding: 0.5rem 0.25rem 0.5rem 0.75rem;
	font-size: 0.8125rem;
	color: var(--text-secondary);
	background: var(--background);
	border-right: 1px solid var(--border);
}

.slug-input {
	flex: 1;
	border: none;
	border-radius: 0;
}

.slug-input:focus {
	box-shadow: none;
}

.slug-status {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 2.5rem;
	height: 2rem;
	font-size: 0.75rem;
}

.slug-status.checking {
	color: var(--text-secondary);
}

.slug-status.available {
	color: var(--success);
}

.slug-status.taken,
.slug-status.invalid {
	color: var(--error);
}

/* Status Selector */
.status-selector {
	display: flex;
	gap: 0.5rem;
}

.status-option {
	display: flex;
	flex: 1;
	gap: 0.5rem;
	align-items: center;
	justify-content: center;
	padding: 0.75rem 1rem;
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text-secondary);
	cursor: pointer;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	transition: all 0.15s ease;
}

.status-option:hover {
	color: var(--text);
	background: var(--surface-hover);
	border-color: var(--border-focus);
}

.status-option.active {
	color: var(--primary);
	background: rgba(99, 102, 241, 0.1);
	border-color: var(--primary);
}

.status-info {
	display: flex;
	gap: 0.5rem;
	align-items: flex-start;
	padding: 0.75rem;
	margin-top: 0.75rem;
	font-size: 0.8125rem;
	line-height: 1.4;
	border-radius: var(--radius);
}

.status-info.warning {
	color: #92400e;
	background: #fef3c7;
}

.status-info.info {
	color: #1e40af;
	background: #dbeafe;
}

/* Password Input */
.password-input-wrapper {
	display: flex;
	gap: 0.5rem;
}

.password-input-wrapper .input {
	flex: 1;
}

@media (max-width: 640px) {
	.form-row {
		grid-template-columns: 1fr;
	}

	.tab span {
		display: none;
	}
}
</style>
