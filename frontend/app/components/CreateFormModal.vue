<script setup lang="ts">
const props = defineProps<{
	show: boolean;
}>();

const emit = defineEmits<{
	close: [];
	blank: [];
	template: [];
}>();
</script>

<template>
	<Teleport to="body">
		<Transition name="modal">
			<div v-if="show" class="modal-overlay" @click.self="emit('close')">
				<div class="modal-container">
					<div class="modal-header">
						<h2>{{ $t("forms.createNew") }}</h2>
						<button class="close-button" @click="emit('close')">
							<i class="fa-solid fa-xmark"></i>
						</button>
					</div>

					<div class="modal-content">
						<div class="options-grid">
							<button class="option-card" @click="emit('blank')">
								<div class="option-icon">
									<i class="fa-solid fa-file-plus"></i>
								</div>
								<h3>{{ $t("templates.blankForm") }}</h3>
								<p>{{ $t("templates.blankFormDesc") }}</p>
							</button>

							<button class="option-card" @click="emit('template')">
								<div class="option-icon">
									<i class="fa-solid fa-grid-2"></i>
								</div>
								<h3>{{ $t("templates.fromTemplate") }}</h3>
								<p>{{ $t("templates.fromTemplateDesc") }}</p>
							</button>
						</div>
					</div>
				</div>
			</div>
		</Transition>
	</Teleport>
</template>

<style scoped>
.modal-overlay {
	position: fixed;
	inset: 0;
	background: rgba(0, 0, 0, 0.5);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
	padding: 1rem;
}

.modal-container {
	background: var(--surface);
	border-radius: 0.75rem;
	box-shadow: var(--shadow-lg);
	max-width: 42rem;
	width: 100%;
	max-height: 90vh;
	overflow: hidden;
	display: flex;
	flex-direction: column;
}

.modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 1.5rem;
	border-bottom: 1px solid var(--border);
}

.modal-header h2 {
	font-size: 1.25rem;
	font-weight: 600;
	color: var(--text);
	margin: 0;
}

.close-button {
	background: none;
	border: none;
	font-size: 1.5rem;
	color: var(--text-secondary);
	cursor: pointer;
	padding: 0.25rem;
	display: flex;
	align-items: center;
	justify-content: center;
	width: 2rem;
	height: 2rem;
	border-radius: 0.375rem;
	transition: background-color 0.15s;
}

.close-button:hover {
	background: var(--surface-hover);
	color: var(--text);
}

.modal-content {
	padding: 2rem;
	overflow-y: auto;
}

.options-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 1.5rem;
}

@media (max-width: 640px) {
	.options-grid {
		grid-template-columns: 1fr;
	}
}

.option-card {
	background: var(--surface);
	border: 2px solid var(--border);
	border-radius: 0.75rem;
	padding: 2rem;
	cursor: pointer;
	transition: all 0.2s;
	text-align: center;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 1rem;
}

.option-card:hover {
	border-color: var(--primary);
	background: var(--surface-hover);
	transform: translateY(-2px);
	box-shadow: var(--shadow-lg);
}

.option-icon {
	width: 4rem;
	height: 4rem;
	border-radius: 50%;
	background: color-mix(in srgb, var(--primary) 20%, transparent);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 2rem;
	color: var(--primary);
	transition: all 0.2s;
}

.option-card:hover .option-icon {
	background: var(--primary);
	color: white;
}

.option-card h3 {
	font-size: 1.125rem;
	font-weight: 600;
	color: var(--text);
	margin: 0;
}

.option-card p {
	font-size: 0.875rem;
	color: var(--text-secondary);
	margin: 0;
}

/* Modal transition */
.modal-enter-active,
.modal-leave-active {
	transition: opacity 0.2s;
}

.modal-enter-from,
.modal-leave-to {
	opacity: 0;
}

.modal-enter-active .modal-container,
.modal-leave-active .modal-container {
	transition: transform 0.2s;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
	transform: scale(0.95);
}
</style>
