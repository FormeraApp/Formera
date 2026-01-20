<script lang="ts" setup>
const { t } = useI18n();
const { webhooksApi } = useApi();
const toastStore = useToastStore();

const props = defineProps<{
	formId?: string;
	isGlobal?: boolean;
	formFields?: FormField[];
}>();

// State
const webhooks = ref<Webhook[]>([]);
const loading = ref(true);
const showModal = ref(false);
const editingWebhook = ref<Webhook | null>(null);
const createdSecret = ref<string | null>(null);
const testingWebhookId = ref<string | null>(null);

// Load webhooks
const loadWebhooks = async () => {
	loading.value = true;
	try {
		if (props.isGlobal) {
			webhooks.value = await webhooksApi.listGlobal();
		} else if (props.formId) {
			webhooks.value = await webhooksApi.listForForm(props.formId);
		}
	} catch (error) {
		console.error("Failed to load webhooks:", error);
		toastStore.error(t("webhooks.loadError"));
	} finally {
		loading.value = false;
	}
};

// Open modal for creating
const handleAdd = () => {
	editingWebhook.value = null;
	createdSecret.value = null;
	showModal.value = true;
};

// Open modal for editing
const handleEdit = (webhook: Webhook) => {
	editingWebhook.value = webhook;
	createdSecret.value = null;
	showModal.value = true;
};

// Close modal
const handleCloseModal = () => {
	showModal.value = false;
	editingWebhook.value = null;
	createdSecret.value = null;
};

// Save webhook (create or update)
const handleSave = async (data: CreateWebhookRequest | UpdateWebhookRequest) => {
	try {
		if (editingWebhook.value) {
			// Update
			if (props.isGlobal) {
				await webhooksApi.updateGlobal(editingWebhook.value.id, data as UpdateWebhookRequest);
			} else if (props.formId) {
				await webhooksApi.updateForForm(props.formId, editingWebhook.value.id, data as UpdateWebhookRequest);
			}
			toastStore.success(t("webhooks.updated"));
		} else {
			// Create
			let result;
			if (props.isGlobal) {
				result = await webhooksApi.createGlobal(data as CreateWebhookRequest);
			} else if (props.formId) {
				result = await webhooksApi.createForForm(props.formId, data as CreateWebhookRequest);
			}
			if (result?.secret) {
				createdSecret.value = result.secret;
			}
			toastStore.success(t("webhooks.created"));
		}
		await loadWebhooks();
		if (!createdSecret.value) {
			handleCloseModal();
		}
	} catch (error) {
		console.error("Failed to save webhook:", error);
		toastStore.error(error instanceof Error ? error.message : t("webhooks.saveError"));
	}
};

// Delete webhook
const handleDelete = async (webhook: Webhook) => {
	if (!confirm(t("webhooks.confirmDelete"))) return;

	try {
		if (props.isGlobal) {
			await webhooksApi.deleteGlobal(webhook.id);
		} else if (props.formId) {
			await webhooksApi.deleteForForm(props.formId, webhook.id);
		}
		toastStore.success(t("webhooks.deleted"));
		await loadWebhooks();
	} catch (error) {
		console.error("Failed to delete webhook:", error);
		toastStore.error(error instanceof Error ? error.message : t("webhooks.deleteError"));
	}
};

// Test webhook
const handleTest = async (webhook: Webhook) => {
	testingWebhookId.value = webhook.id;
	try {
		let result;
		if (props.isGlobal) {
			result = await webhooksApi.testGlobal(webhook.id);
		} else if (props.formId) {
			result = await webhooksApi.testForForm(props.formId, webhook.id);
		}
		if (result?.success) {
			toastStore.success(t("webhooks.testSuccess"), `Status: ${result.status_code}`);
		} else {
			toastStore.error(t("webhooks.testFailed"), result?.error || `Status: ${result?.status_code}`);
		}
	} catch (error) {
		console.error("Failed to test webhook:", error);
		toastStore.error(t("webhooks.testError"));
	} finally {
		testingWebhookId.value = null;
	}
};

// Format URL for display
const formatUrl = (url: string): string => {
	try {
		const urlObj = new URL(url);
		return urlObj.hostname + (urlObj.pathname !== "/" ? urlObj.pathname : "");
	} catch {
		return url;
	}
};

// Format events for display
const formatEvents = (events: WebhookEvent[]): string => {
	return events.map((e) => t(`webhooks.events.${e}`)).join(", ");
};

// Check if URL is a Discord webhook
const isDiscordWebhook = (url: string): boolean => {
	const lowerUrl = url.toLowerCase();
	return lowerUrl.includes("discord.com/api/webhooks/") || lowerUrl.includes("discordapp.com/api/webhooks/");
};

// Get field label by ID
const getFieldLabel = (fieldId: string): string => {
	if (!props.formFields) return fieldId;
	const field = props.formFields.find((f) => f.id === fieldId);
	return field?.label || fieldId;
};

// Format Discord fields for display
const formatDiscordFields = (webhook: Webhook): string => {
	if (!webhook.discord_config?.fields?.length) return t("webhooks.allFields");
	return webhook.discord_config.fields.map((f) => getFieldLabel(f)).join(", ");
};

// Load webhooks on mount
onMounted(() => {
	loadWebhooks();
});

// Reload when formId changes
watch(() => props.formId, () => {
	loadWebhooks();
});
</script>

<template>
	<div class="webhook-list">
		<div class="list-header">
			<h3>{{ t("webhooks.title") }}</h3>
			<button class="btn btn-primary btn-sm" @click="handleAdd">
				<UISysIcon icon="fa-solid fa-plus" />
				{{ t("webhooks.add") }}
			</button>
		</div>

		<div v-if="loading" class="loading-state">
			<UISysIcon icon="fa-solid fa-spinner fa-spin" />
			{{ t("common.loading") }}
		</div>

		<div v-else-if="webhooks.length === 0" class="empty-state">
			<UISysIcon icon="fa-solid fa-bolt" class="empty-icon" />
			<p>{{ t("webhooks.empty") }}</p>
			<button class="btn btn-outline btn-sm" @click="handleAdd">
				{{ t("webhooks.addFirst") }}
			</button>
		</div>

		<div v-else class="webhooks">
			<div v-for="webhook in webhooks" :key="webhook.id" class="webhook-item">
				<div class="webhook-info">
					<div class="webhook-url">
						<UISysIcon :icon="isDiscordWebhook(webhook.url) ? 'fa-brands fa-discord' : 'fa-solid fa-link'" />
						<span :title="webhook.url">{{ formatUrl(webhook.url) }}</span>
					</div>
					<div class="webhook-events">
						{{ formatEvents(webhook.events) }}
					</div>
					<!-- Discord Config Details -->
					<div v-if="isDiscordWebhook(webhook.url) && webhook.discord_config" class="webhook-discord-config">
						<span v-if="webhook.discord_config.custom_title" class="discord-detail">
							<UISysIcon icon="fa-solid fa-heading" />
							{{ webhook.discord_config.custom_title }}
						</span>
						<span class="discord-detail">
							<UISysIcon icon="fa-solid fa-list" />
							{{ formatDiscordFields(webhook) }}
						</span>
						<span v-if="!webhook.discord_config.show_form_title" class="discord-detail discord-detail-muted">
							{{ t("webhooks.hideFormTitle") }}
						</span>
						<span v-if="!webhook.discord_config.show_slug" class="discord-detail discord-detail-muted">
							{{ t("webhooks.hideSlug") }}
						</span>
					</div>
				</div>

				<div class="webhook-status">
					<span :class="['status-badge', webhook.enabled ? 'status-enabled' : 'status-disabled']">
						{{ webhook.enabled ? t("webhooks.enabled") : t("webhooks.disabled") }}
					</span>
				</div>

				<div class="webhook-actions">
					<button
						class="btn-icon"
						:class="{ 'btn-icon-loading': testingWebhookId === webhook.id }"
						:disabled="testingWebhookId === webhook.id"
						:title="t('webhooks.test')"
						@click="handleTest(webhook)"
					>
						<UISysIcon :icon="testingWebhookId === webhook.id ? 'fa-solid fa-spinner fa-spin' : 'fa-solid fa-play'" />
					</button>
					<button
						class="btn-icon"
						:title="t('webhooks.edit')"
						@click="handleEdit(webhook)"
					>
						<UISysIcon icon="fa-solid fa-pen" />
					</button>
					<button
						class="btn-icon btn-icon-danger"
						:title="t('webhooks.delete')"
						@click="handleDelete(webhook)"
					>
						<UISysIcon icon="fa-solid fa-trash" />
					</button>
				</div>
			</div>
		</div>

		<!-- Webhook Form Modal -->
		<WebhooksWebhookFormModal
			:open="showModal"
			:webhook="editingWebhook"
			:initial-secret="createdSecret || undefined"
			:form-fields="formFields"
			@update:open="showModal = $event"
			@save="handleSave"
		/>
	</div>
</template>

<style scoped>
.webhook-list {
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius);
}

.list-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 1rem;
	border-bottom: 1px solid var(--border);
}

.list-header h3 {
	margin: 0;
	font-size: 1rem;
	font-weight: 600;
}

.loading-state,
.empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.75rem;
	padding: 2rem;
	color: var(--text-secondary);
	text-align: center;
}

.empty-icon {
	font-size: 2rem;
	opacity: 0.5;
}

.webhooks {
	display: flex;
	flex-direction: column;
}

.webhook-item {
	display: flex;
	align-items: center;
	gap: 1rem;
	padding: 1rem;
	border-bottom: 1px solid var(--border);
}

.webhook-item:last-child {
	border-bottom: none;
}

.webhook-info {
	flex: 1;
	min-width: 0;
}

.webhook-url {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	font-weight: 500;
	color: var(--text);
}

.webhook-url span {
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.webhook-url :deep(i),
.webhook-url :deep(svg) {
	color: var(--text-secondary);
	font-size: 0.75rem;
}

.webhook-events {
	margin-top: 0.25rem;
	font-size: 0.8125rem;
	color: var(--text-secondary);
}

.webhook-discord-config {
	display: flex;
	flex-wrap: wrap;
	gap: 0.5rem;
	margin-top: 0.375rem;
}

.discord-detail {
	display: inline-flex;
	align-items: center;
	gap: 0.25rem;
	padding: 0.125rem 0.375rem;
	font-size: 0.6875rem;
	color: var(--text-secondary);
	background: var(--background);
	border-radius: var(--radius-sm);
}

.discord-detail :deep(i),
.discord-detail :deep(svg) {
	font-size: 0.625rem;
	color: #5865f2;
}

.discord-detail-muted {
	opacity: 0.7;
	text-decoration: line-through;
}

.webhook-status {
	flex-shrink: 0;
}

.status-badge {
	display: inline-flex;
	align-items: center;
	padding: 0.25rem 0.5rem;
	font-size: 0.75rem;
	font-weight: 500;
	border-radius: var(--radius-sm);
}

.status-enabled {
	color: var(--success);
	background: rgba(34, 197, 94, 0.1);
}

.status-disabled {
	color: var(--text-secondary);
	background: var(--background);
}

.webhook-actions {
	display: flex;
	align-items: center;
	gap: 0.25rem;
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

.btn-icon:disabled {
	cursor: not-allowed;
	opacity: 0.7;
}

.btn-icon-danger:hover {
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
}

.btn {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem 1rem;
	font-size: 0.875rem;
	font-weight: 500;
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.15s;
}

.btn-sm {
	padding: 0.375rem 0.75rem;
	font-size: 0.8125rem;
}

.btn-primary {
	color: white;
	background: var(--primary);
	border: none;
}

.btn-primary:hover {
	background: var(--primary-dark);
}

.btn-outline {
	color: var(--text);
	background: transparent;
	border: 1px solid var(--border);
}

.btn-outline:hover {
	background: var(--surface-hover);
	border-color: var(--primary);
}

@media (max-width: 640px) {
	.webhook-item {
		flex-wrap: wrap;
	}

	.webhook-info {
		width: 100%;
	}

	.webhook-status,
	.webhook-actions {
		margin-top: 0.5rem;
	}
}
</style>
