<script lang="ts" setup>
const props = defineProps<{
	open: boolean;
	title: string;
	message?: string;
	confirmText?: string;
	cancelText?: string;
	variant?: "danger" | "warning" | "info";
}>();

const emit = defineEmits(["confirm", "cancel", "update:open"]);

const { t } = useI18n();

const dialogRef = ref<HTMLDialogElement | null>(null);
const confirmButtonRef = ref<HTMLButtonElement | null>(null);

const confirmLabel = computed(() => props.confirmText || t("dialog.confirm"));
const cancelLabel = computed(() => props.cancelText || t("dialog.cancel"));
const variantValue = computed(() => props.variant || "danger");

const handleConfirm = () => {
	emit("confirm");
	emit("update:open", false);
};

const handleCancel = () => {
	emit("cancel");
	emit("update:open", false);
};

const handleBackdropClick = (event: MouseEvent) => {
	if (event.target === dialogRef.value) {
		handleCancel();
	}
};

const handleKeydown = (event: KeyboardEvent) => {
	if (event.key === "Escape") {
		handleCancel();
	}
};

watch(
	() => props.open,
	(isOpen) => {
		if (isOpen) {
			nextTick(() => {
				dialogRef.value?.showModal();
				confirmButtonRef.value?.focus();
			});
		}
	},
	{ immediate: true }
);
</script>

<template>
	<Teleport to="body">
		<dialog
			v-if="open"
			ref="dialogRef"
			class="dialog"
			aria-labelledby="dialog-title"
			aria-describedby="dialog-message"
			@click="handleBackdropClick"
			@keydown="handleKeydown"
		>
			<div class="dialog-content">
				<div class="dialog-icon" :class="`dialog-icon-${variantValue}`">
					<UISysIcon v-if="variantValue === 'danger'" icon="fa-solid fa-triangle-exclamation" />
					<UISysIcon v-else-if="variantValue === 'warning'" icon="fa-solid fa-circle-exclamation" />
					<UISysIcon v-else icon="fa-solid fa-circle-info" />
				</div>
				<h2 id="dialog-title" class="dialog-title">{{ title }}</h2>
				<p v-if="message" id="dialog-message" class="dialog-message">{{ message }}</p>
				<div class="dialog-actions">
					<button class="btn btn-secondary" @click="handleCancel">
						{{ cancelLabel }}
					</button>
					<button
						ref="confirmButtonRef"
						class="btn"
						:class="variantValue === 'danger' ? 'btn-danger' : 'btn-primary'"
						@click="handleConfirm"
					>
						{{ confirmLabel }}
					</button>
				</div>
			</div>
		</dialog>
	</Teleport>
</template>

<style scoped>
.dialog {
	position: fixed;
	inset: 0;
	z-index: 9998;
	display: flex;
	align-items: center;
	justify-content: center;
	width: 100%;
	max-width: none;
	height: 100%;
	max-height: none;
	padding: 1rem;
	background: transparent;
	border: none;
}

.dialog::backdrop {
	background: rgba(0, 0, 0, 0.5);
	backdrop-filter: blur(4px);
}

.dialog-content {
	width: 100%;
	max-width: 400px;
	padding: 1.5rem;
	text-align: center;
	background: var(--surface);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-lg);
	animation: dialogIn 0.2s ease-out;
}

@keyframes dialogIn {
	from {
		opacity: 0;
		transform: scale(0.95);
	}
	to {
		opacity: 1;
		transform: scale(1);
	}
}

.dialog-icon {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 3rem;
	height: 3rem;
	margin: 0 auto 1rem;
	font-size: 1.5rem;
	border-radius: 50%;
}

.dialog-icon-danger {
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
}

.dialog-icon-warning {
	color: var(--warning);
	background: rgba(245, 158, 11, 0.1);
}

.dialog-icon-info {
	color: var(--primary);
	background: rgba(99, 102, 241, 0.1);
}

.dialog-title {
	margin-bottom: 0.5rem;
	font-size: 1.125rem;
	font-weight: 600;
	color: var(--text);
}

.dialog-message {
	margin-bottom: 1.5rem;
	font-size: 0.9375rem;
	color: var(--text-secondary);
}

.dialog-actions {
	display: flex;
	gap: 0.75rem;
	justify-content: center;
}

.dialog-actions .btn {
	min-width: 100px;
}
</style>
