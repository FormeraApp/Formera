<script lang="ts" setup>
const toastStore = useToastStore();

const iconMap: Record<string, string> = {
	success: "fa-solid fa-circle-check",
	error: "fa-solid fa-circle-xmark",
	warning: "fa-solid fa-triangle-exclamation",
	info: "fa-solid fa-circle-info",
};
</script>

<template>
	<Teleport to="body">
		<div class="toast-container" role="region" :aria-label="$t('toast.notifications')" aria-live="polite">
			<TransitionGroup name="toast">
				<div
					v-for="toast in toastStore.toasts"
					:key="toast.id"
					:class="['toast', `toast-${toast.type}`]"
					role="alert"
				>
					<UISysIcon :icon="iconMap[toast.type] ? iconMap[toast.type]! : 'fa-solid fa-circle-info'" class="toast-icon" />
					<div class="toast-content">
						<p class="toast-title">{{ toast.title }}</p>
						<p v-if="toast.message" class="toast-message">{{ toast.message }}</p>
					</div>
					<button
						class="toast-close"
						:aria-label="$t('toast.close')"
						@click="toastStore.remove(toast.id)"
					>
						<UISysIcon icon="fa-solid fa-xmark" />
					</button>
				</div>
			</TransitionGroup>
		</div>
	</Teleport>
</template>

<style scoped>
.toast-container {
	position: fixed;
	right: 1rem;
	bottom: 1rem;
	z-index: 9999;
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
	max-width: 400px;
}

.toast {
	display: flex;
	gap: 0.75rem;
	align-items: flex-start;
	padding: 1rem;
	background: var(--surface);
	border: 1px solid var(--border);
	border-radius: var(--radius-lg);
	box-shadow: var(--shadow-lg);
}

.toast-icon {
	flex-shrink: 0;
	margin-top: 0.125rem;
	font-size: 1.25rem;
}

.toast-success .toast-icon {
	color: var(--success);
}

.toast-error .toast-icon {
	color: var(--error);
}

.toast-warning .toast-icon {
	color: var(--warning);
}

.toast-info .toast-icon {
	color: var(--primary);
}

.toast-content {
	flex: 1;
	min-width: 0;
}

.toast-title {
	font-size: 0.9375rem;
	font-weight: 600;
	color: var(--text);
}

.toast-message {
	margin-top: 0.25rem;
	font-size: 0.875rem;
	color: var(--text-secondary);
}

.toast-close {
	flex-shrink: 0;
	padding: 0.25rem;
	color: var(--text-secondary);
	background: none;
	border: none;
	border-radius: var(--radius);
	transition: all 0.2s;
}

.toast-close:hover {
	color: var(--text);
	background: var(--background);
}

/* Animations */
.toast-enter-active,
.toast-leave-active {
	transition: all 0.3s ease;
}

.toast-enter-from {
	opacity: 0;
	transform: translateX(100%);
}

.toast-leave-to {
	opacity: 0;
	transform: translateX(100%);
}

.toast-move {
	transition: transform 0.3s ease;
}

@media (max-width: 480px) {
	.toast-container {
		right: 0.5rem;
		bottom: 0.5rem;
		left: 0.5rem;
		max-width: none;
	}
}
</style>
