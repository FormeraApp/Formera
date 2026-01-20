<script lang="ts" setup>
const { t } = useI18n();

const props = defineProps<{
	open: boolean;
	webhook?: Webhook | null;
	initialSecret?: string;
	formFields?: FormField[];
}>();

// Filter only input fields (exclude layout elements like heading, paragraph, image, section)
const inputFields = computed(() => {
	if (!props.formFields) return [];
	const layoutTypes = ["heading", "paragraph", "image", "section", "divider"];
	return props.formFields.filter((f) => !layoutTypes.includes(f.type));
});

const emit = defineEmits<{
	"update:open": [value: boolean];
	save: [data: CreateWebhookRequest | UpdateWebhookRequest];
}>();

const isEditing = computed(() => !!props.webhook);

const form = ref<CreateWebhookRequest>({
	url: "",
	secret: "",
	events: ["submission.created"],
	headers: {},
	discord_config: undefined,
	enabled: true,
});

const showSecret = ref(false);
const newHeaderKey = ref("");
const newHeaderValue = ref("");
const newDiscordField = ref("");

// Available webhook events
const availableEvents: WebhookEvent[] = ["submission.created", "submission.deleted"];

// Check if URL is a Discord webhook
const isDiscordUrl = computed(() => {
	const url = form.value.url.toLowerCase();
	return url.includes("discord.com/api/webhooks/") || url.includes("discordapp.com/api/webhooks/");
});

// Initialize Discord config when Discord URL is detected
watch(isDiscordUrl, (isDiscord) => {
	if (isDiscord && !form.value.discord_config) {
		form.value.discord_config = {
			show_form_title: true,
			show_slug: true,
			fields: [],
			custom_title: "",
		};
	}
});

// Reset form when modal opens
watch(
	() => props.open,
	(isOpen) => {
		if (isOpen) {
			if (props.webhook) {
				form.value = {
					url: props.webhook.url,
					secret: "",
					events: [...props.webhook.events],
					headers: { ...props.webhook.headers },
					discord_config: props.webhook.discord_config ? { ...props.webhook.discord_config, fields: [...props.webhook.discord_config.fields] } : undefined,
					enabled: props.webhook.enabled,
				};
			} else {
				form.value = {
					url: "",
					secret: "",
					events: ["submission.created"],
					headers: {},
					discord_config: undefined,
					enabled: true,
				};
			}
			showSecret.value = false;
			newDiscordField.value = "";
		}
	}
);

const close = () => {
	emit("update:open", false);
};

const generateSecret = () => {
	const array = new Uint8Array(32);
	crypto.getRandomValues(array);
	form.value.secret = Array.from(array, (byte) => byte.toString(16).padStart(2, "0")).join("");
};

const toggleEvent = (event: WebhookEvent) => {
	const index = form.value.events.indexOf(event);
	if (index === -1) {
		form.value.events.push(event);
	} else if (form.value.events.length > 1) {
		form.value.events.splice(index, 1);
	}
};

const addHeader = () => {
	if (newHeaderKey.value && newHeaderValue.value) {
		if (!form.value.headers) {
			form.value.headers = {};
		}
		form.value.headers[newHeaderKey.value] = newHeaderValue.value;
		newHeaderKey.value = "";
		newHeaderValue.value = "";
	}
};

const removeHeader = (key: string) => {
	if (form.value.headers) {
		delete form.value.headers[key];
		form.value.headers = { ...form.value.headers };
	}
};

const addDiscordField = () => {
	if (newDiscordField.value && form.value.discord_config) {
		if (!form.value.discord_config.fields.includes(newDiscordField.value)) {
			form.value.discord_config.fields.push(newDiscordField.value);
		}
		newDiscordField.value = "";
	}
};

const removeDiscordField = (field: string) => {
	if (form.value.discord_config) {
		form.value.discord_config.fields = form.value.discord_config.fields.filter((f) => f !== field);
	}
};

const toggleDiscordField = (fieldId: string) => {
	if (!form.value.discord_config) return;
	const index = form.value.discord_config.fields.indexOf(fieldId);
	if (index === -1) {
		form.value.discord_config.fields.push(fieldId);
	} else {
		form.value.discord_config.fields.splice(index, 1);
	}
};

const handleSubmit = () => {
	if (!form.value.url || form.value.events.length === 0) return;

	const data: CreateWebhookRequest | UpdateWebhookRequest = {
		url: form.value.url,
		events: form.value.events,
		headers: Object.keys(form.value.headers || {}).length > 0 ? form.value.headers : undefined,
		enabled: form.value.enabled,
	};

	if (form.value.secret) {
		data.secret = form.value.secret;
	}

	// Include Discord config if it's a Discord webhook
	if (isDiscordUrl.value && form.value.discord_config) {
		data.discord_config = form.value.discord_config;
	}

	emit("save", data);
};
</script>

<template>
	<Teleport to="body">
		<div v-if="open" class="modal-overlay" @click.self="close">
			<div class="modal">
				<div class="modal-header">
					<h2>{{ isEditing ? t("webhooks.editTitle") : t("webhooks.addTitle") }}</h2>
					<button class="modal-close" :aria-label="t('common.close')" @click="close">
						<UISysIcon icon="fa-solid fa-xmark" />
					</button>
				</div>

				<form class="modal-body" @submit.prevent="handleSubmit">
					<!-- URL -->
					<div class="form-group">
						<label class="label">{{ t("webhooks.form.url") }} *</label>
						<input
							v-model="form.url"
							type="url"
							class="input"
							:placeholder="t('webhooks.form.urlPlaceholder')"
							required
						/>
						<p class="form-hint">{{ t("webhooks.form.urlHint") }}</p>
					</div>

					<!-- Secret -->
					<div class="form-group">
						<label class="label">{{ t("webhooks.form.secret") }}</label>
						<div class="secret-input">
							<input
								v-model="form.secret"
								:type="showSecret ? 'text' : 'password'"
								class="input"
								:placeholder="isEditing ? t('webhooks.form.secretPlaceholderEdit') : t('webhooks.form.secretPlaceholder')"
							/>
							<button type="button" class="btn-icon" @click="showSecret = !showSecret">
								<UISysIcon :icon="showSecret ? 'fa-solid fa-eye-slash' : 'fa-solid fa-eye'" />
							</button>
							<button type="button" class="btn btn-outline btn-sm" @click="generateSecret">
								{{ t("webhooks.form.generate") }}
							</button>
						</div>
						<p class="form-hint">{{ t("webhooks.form.secretHint") }}</p>
						<div v-if="initialSecret && !isEditing" class="secret-display">
							<strong>{{ t("webhooks.form.generatedSecret") }}:</strong>
							<code>{{ initialSecret }}</code>
						</div>
					</div>

					<!-- Events -->
					<div class="form-group">
						<label class="label">{{ t("webhooks.form.events") }} *</label>
						<div class="events-grid">
							<label
								v-for="event in availableEvents"
								:key="event"
								class="event-checkbox"
							>
								<input
									type="checkbox"
									:checked="form.events.includes(event)"
									@change="toggleEvent(event)"
								/>
								<span class="checkmark"></span>
								<span class="event-label">{{ t(`webhooks.events.${event}`) }}</span>
							</label>
						</div>
					</div>

					<!-- Custom Headers -->
					<div class="form-group">
						<label class="label">{{ t("webhooks.form.headers") }}</label>
						<div class="headers-list">
							<div v-for="(value, key) in form.headers" :key="key" class="header-item">
								<span class="header-key">{{ key }}</span>
								<span class="header-value">{{ value }}</span>
								<button type="button" class="btn-icon btn-icon-sm" @click="removeHeader(String(key))">
									<UISysIcon icon="fa-solid fa-times" />
								</button>
							</div>
						</div>
						<div class="header-input">
							<input
								v-model="newHeaderKey"
								type="text"
								class="input input-sm"
								:placeholder="t('webhooks.form.headerKey')"
							/>
							<input
								v-model="newHeaderValue"
								type="text"
								class="input input-sm"
								:placeholder="t('webhooks.form.headerValue')"
							/>
							<button
								type="button"
								class="btn btn-outline btn-sm"
								:disabled="!newHeaderKey || !newHeaderValue"
								@click="addHeader"
							>
								<UISysIcon icon="fa-solid fa-plus" />
							</button>
						</div>
					</div>

					<!-- Discord Configuration (only shown for Discord webhooks) -->
					<div v-if="isDiscordUrl && form.discord_config" class="form-group discord-config">
						<label class="label">{{ t("webhooks.form.discordConfig") }}</label>

						<!-- Custom Title -->
						<div class="discord-option">
							<label class="label-sm">{{ t("webhooks.form.discordCustomTitle") }}</label>
							<input
								v-model="form.discord_config.custom_title"
								type="text"
								class="input input-sm"
								:placeholder="t('webhooks.form.discordCustomTitlePlaceholder')"
							/>
						</div>

						<!-- Show Form Title -->
						<label class="toggle-label toggle-label-sm">
							<input v-model="form.discord_config.show_form_title" type="checkbox" class="toggle-input" />
							<span class="toggle-switch toggle-switch-sm"></span>
							<span>{{ t("webhooks.form.discordShowFormTitle") }}</span>
						</label>

						<!-- Show Slug -->
						<label class="toggle-label toggle-label-sm">
							<input v-model="form.discord_config.show_slug" type="checkbox" class="toggle-input" />
							<span class="toggle-switch toggle-switch-sm"></span>
							<span>{{ t("webhooks.form.discordShowSlug") }}</span>
						</label>

						<!-- Field Filter -->
						<div class="discord-option">
							<label class="label-sm">{{ t("webhooks.form.discordFields") }}</label>
							<p class="form-hint-sm">{{ inputFields.length > 0 ? t("webhooks.form.discordFieldsHintSelect") : t("webhooks.form.discordFieldsHint") }}</p>

							<!-- Checkbox list when form fields are available -->
							<div v-if="inputFields.length > 0" class="discord-fields-checkboxes">
								<label
									v-for="field in inputFields"
									:key="field.id"
									class="field-checkbox"
								>
									<input
										type="checkbox"
										:checked="form.discord_config.fields.includes(field.id)"
										@change="toggleDiscordField(field.id)"
									/>
									<span class="checkmark-sm"></span>
									<span class="field-label">{{ field.label }}</span>
								</label>
							</div>

							<!-- Free text input for global webhooks (no form fields available) -->
							<template v-else>
								<div class="discord-fields-list">
									<span
										v-for="field in form.discord_config.fields"
										:key="field"
										class="discord-field-tag"
									>
										{{ field }}
										<button type="button" class="tag-remove" @click="removeDiscordField(field)">
											<UISysIcon icon="fa-solid fa-times" />
										</button>
									</span>
								</div>
								<div class="discord-field-input">
									<input
										v-model="newDiscordField"
										type="text"
										class="input input-sm"
										:placeholder="t('webhooks.form.discordFieldPlaceholder')"
										@keyup.enter.prevent="addDiscordField"
									/>
									<button
										type="button"
										class="btn btn-outline btn-sm"
										:disabled="!newDiscordField"
										@click="addDiscordField"
									>
										<UISysIcon icon="fa-solid fa-plus" />
									</button>
								</div>
							</template>
						</div>
					</div>

					<!-- Enabled -->
					<div class="form-group">
						<label class="toggle-label">
							<input v-model="form.enabled" type="checkbox" class="toggle-input" />
							<span class="toggle-switch"></span>
							<span>{{ t("webhooks.form.enabled") }}</span>
						</label>
					</div>
				</form>

				<div class="modal-footer">
					<button type="button" class="btn btn-outline" @click="close">
						{{ t("common.cancel") }}
					</button>
					<button
						type="submit"
						class="btn btn-primary"
						:disabled="!form.url || form.events.length === 0"
						@click="handleSubmit"
					>
						{{ isEditing ? t("common.save") : t("webhooks.form.create") }}
					</button>
				</div>
			</div>
		</div>
	</Teleport>
</template>

<style scoped>
.modal-overlay {
	position: fixed;
	inset: 0;
	z-index: 1000;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 1rem;
	background: rgba(0, 0, 0, 0.5);
}

.modal {
	display: flex;
	flex-direction: column;
	width: 100%;
	max-width: 500px;
	max-height: 90vh;
	background: var(--surface);
	border-radius: var(--radius-lg);
	box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 1rem 1.25rem;
	border-bottom: 1px solid var(--border);
}

.modal-header h2 {
	margin: 0;
	font-size: 1.125rem;
	font-weight: 600;
}

.modal-close {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 2rem;
	height: 2rem;
	padding: 0;
	color: var(--text-secondary);
	background: transparent;
	border: none;
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s;
}

.modal-close:hover {
	color: var(--text);
	background: var(--surface-hover);
}

.modal-body {
	flex: 1;
	padding: 1.25rem;
	overflow-y: auto;
}

.modal-footer {
	display: flex;
	justify-content: flex-end;
	gap: 0.75rem;
	padding: 1rem 1.25rem;
	border-top: 1px solid var(--border);
}

.form-group {
	margin-bottom: 1.25rem;
}

.form-group:last-child {
	margin-bottom: 0;
}

.label {
	display: block;
	margin-bottom: 0.5rem;
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text);
}

.input {
	width: 100%;
	padding: 0.625rem 0.875rem;
	font-size: 0.875rem;
	color: var(--text);
	background: var(--background);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	transition: border-color 0.15s, box-shadow 0.15s;
}

.input:focus {
	border-color: var(--primary);
	outline: none;
	box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.input-sm {
	padding: 0.5rem 0.75rem;
	font-size: 0.8125rem;
}

.form-hint {
	margin: 0.375rem 0 0;
	font-size: 0.75rem;
	color: var(--text-secondary);
}

.secret-input {
	display: flex;
	gap: 0.5rem;
}

.secret-input .input {
	flex: 1;
}

.secret-display {
	margin-top: 0.5rem;
	padding: 0.5rem;
	font-size: 0.8125rem;
	background: var(--background);
	border-radius: var(--radius-sm);
}

.secret-display code {
	display: block;
	margin-top: 0.25rem;
	word-break: break-all;
	font-family: monospace;
}

.events-grid {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.event-checkbox {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	cursor: pointer;
}

.event-checkbox input {
	display: none;
}

.checkmark {
	width: 1.125rem;
	height: 1.125rem;
	border: 2px solid var(--border);
	border-radius: var(--radius-sm);
	transition: all 0.15s;
}

.event-checkbox input:checked + .checkmark {
	background: var(--primary);
	border-color: var(--primary);
}

.event-checkbox input:checked + .checkmark::after {
	content: "";
	display: block;
	width: 0.375rem;
	height: 0.625rem;
	margin: 0.125rem auto;
	border: solid white;
	border-width: 0 2px 2px 0;
	transform: rotate(45deg);
}

.event-label {
	font-size: 0.875rem;
}

.headers-list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	margin-bottom: 0.5rem;
}

.header-item {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.375rem 0.5rem;
	font-size: 0.8125rem;
	background: var(--background);
	border-radius: var(--radius-sm);
}

.header-key {
	font-weight: 500;
}

.header-value {
	flex: 1;
	color: var(--text-secondary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.header-input {
	display: flex;
	gap: 0.5rem;
}

.header-input .input {
	flex: 1;
}

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
	width: 2.5rem;
	height: 1.5rem;
	background: var(--border);
	border-radius: 0.75rem;
	transition: background 0.2s;
}

.toggle-switch::after {
	content: "";
	position: absolute;
	top: 0.125rem;
	left: 0.125rem;
	width: 1.25rem;
	height: 1.25rem;
	background: white;
	border-radius: 50%;
	transition: transform 0.2s;
}

.toggle-input:checked + .toggle-switch {
	background: var(--primary);
}

.toggle-input:checked + .toggle-switch::after {
	transform: translateX(1rem);
}

.btn {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	gap: 0.5rem;
	padding: 0.625rem 1rem;
	font-size: 0.875rem;
	font-weight: 500;
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s;
}

.btn:disabled {
	opacity: 0.5;
	cursor: not-allowed;
}

.btn-sm {
	padding: 0.5rem 0.75rem;
	font-size: 0.8125rem;
}

.btn-primary {
	color: white;
	background: var(--primary);
	border: none;
}

.btn-primary:hover:not(:disabled) {
	background: var(--primary-dark);
}

.btn-outline {
	color: var(--text);
	background: transparent;
	border: 1px solid var(--border);
}

.btn-outline:hover:not(:disabled) {
	background: var(--surface-hover);
	border-color: var(--primary);
}

.btn-icon {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 2rem;
	height: 2rem;
	padding: 0;
	color: var(--text-secondary);
	background: transparent;
	border: none;
	border-radius: var(--radius-sm);
	cursor: pointer;
	transition: all 0.15s;
}

.btn-icon:hover {
	color: var(--text);
	background: var(--surface-hover);
}

.btn-icon-sm {
	width: 1.5rem;
	height: 1.5rem;
}

/* Discord Config Styles */
.discord-config {
	padding: 1rem;
	background: var(--background);
	border-radius: var(--radius);
}

.discord-option {
	margin-bottom: 0.75rem;
}

.label-sm {
	display: block;
	margin-bottom: 0.25rem;
	font-size: 0.8125rem;
	font-weight: 500;
	color: var(--text);
}

.form-hint-sm {
	margin: 0 0 0.5rem;
	font-size: 0.6875rem;
	color: var(--text-secondary);
}

.toggle-label-sm {
	margin-bottom: 0.5rem;
	font-size: 0.8125rem;
}

.toggle-switch-sm {
	width: 2rem;
	height: 1.25rem;
	border-radius: 0.625rem;
}

.toggle-switch-sm::after {
	width: 1rem;
	height: 1rem;
}

.toggle-input:checked + .toggle-switch-sm::after {
	transform: translateX(0.75rem);
}

.discord-fields-list {
	display: flex;
	flex-wrap: wrap;
	gap: 0.375rem;
	margin-bottom: 0.5rem;
}

.discord-field-tag {
	display: inline-flex;
	align-items: center;
	gap: 0.25rem;
	padding: 0.25rem 0.5rem;
	font-size: 0.75rem;
	color: var(--text);
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
}

.tag-remove {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 1rem;
	height: 1rem;
	padding: 0;
	color: var(--text-secondary);
	background: transparent;
	border: none;
	border-radius: 50%;
	cursor: pointer;
	transition: all 0.15s;
}

.tag-remove:hover {
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
}

.discord-field-input {
	display: flex;
	gap: 0.5rem;
}

.discord-field-input .input {
	flex: 1;
}

.discord-fields-checkboxes {
	display: flex;
	flex-direction: column;
	gap: 0.375rem;
	max-height: 150px;
	overflow-y: auto;
	padding: 0.5rem;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-sm);
}

.field-checkbox {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	cursor: pointer;
	font-size: 0.8125rem;
}

.field-checkbox input {
	display: none;
}

.checkmark-sm {
	width: 1rem;
	height: 1rem;
	border: 2px solid var(--border);
	border-radius: var(--radius-sm);
	transition: all 0.15s;
	flex-shrink: 0;
}

.field-checkbox input:checked + .checkmark-sm {
	background: var(--primary);
	border-color: var(--primary);
}

.field-checkbox input:checked + .checkmark-sm::after {
	content: "";
	display: block;
	width: 0.25rem;
	height: 0.5rem;
	margin: 0.0625rem auto;
	border: solid white;
	border-width: 0 2px 2px 0;
	transform: rotate(45deg);
}

.field-label {
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

@media (max-width: 640px) {
	.modal {
		max-height: 100vh;
		border-radius: 0;
	}

	.secret-input {
		flex-wrap: wrap;
	}

	.secret-input .input {
		width: 100%;
	}
}
</style>
