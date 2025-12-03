<script lang="ts" setup>
const props = defineProps<{
	modelValue?: string;
	label?: string;
	accept?: string;
	maxSize?: number;
}>();

const emit = defineEmits(["update:modelValue"]);

const { t } = useI18n();
const { uploadApi } = useApi();

const acceptValue = computed(() => props.accept || "image/jpeg,image/png,image/gif,image/webp,image/svg+xml");
const maxSizeValue = computed(() => props.maxSize || 5);

const isUploading = ref(false);
const error = ref("");
const storedPath = ref(props.modelValue || "");
const previewUrl = computed(() => getFileUrl(storedPath.value));
const fileInput = ref<HTMLInputElement | null>(null);

watch(
	() => props.modelValue,
	(newVal) => {
		storedPath.value = newVal || "";
	}
);

const handleFileSelect = async (event: Event) => {
	const input = event.target as HTMLInputElement;
	const file = input.files?.[0];

	if (!file) return;

	error.value = "";

	if (file.size > maxSizeValue.value * 1024 * 1024) {
		error.value = t("imageUpload.sizeErrorMessage", { size: maxSizeValue.value });
		return;
	}

	const allowedTypes = acceptValue.value.split(",").map((t) => t.trim());
	if (!allowedTypes.includes(file.type)) {
		error.value = t("imageUpload.invalidFileType");
		return;
	}

	isUploading.value = true;

	try {
		const result = await uploadApi.uploadImage(file);
		storedPath.value = result.path;
		emit("update:modelValue", result.path);
	} catch (err) {
		error.value = err instanceof Error ? err.message : t("imageUpload.uploadErrorMessage");
	} finally {
		isUploading.value = false;
		if (input) input.value = "";
	}
};

const triggerFileSelect = () => {
	fileInput.value?.click();
};

const removeImage = () => {
	storedPath.value = "";
	emit("update:modelValue", "");
};
</script>

<template>
	<div class="image-upload">
		<label v-if="label" class="label">{{ label }}</label>

		<!-- Preview -->
		<div v-if="previewUrl" class="preview">
			<img :src="previewUrl" :alt="$t('imageUpload.preview')" />
			<div class="preview-actions">
				<button class="btn btn-sm btn-secondary" type="button" @click="triggerFileSelect">
					<UISysIcon icon="fa-solid fa-arrows-rotate" />
					{{ $t("imageUpload.change") }}
				</button>
				<button class="btn btn-sm btn-danger" type="button" @click="removeImage">
					<UISysIcon icon="fa-solid fa-trash" />
					{{ $t("imageUpload.remove") }}
				</button>
			</div>
		</div>

		<!-- Upload area -->
		<div v-else class="upload-area" :class="{ uploading: isUploading }" @click="triggerFileSelect">
			<div v-if="isUploading" class="uploading-state">
				<UISysIcon icon="fa-solid fa-spinner fa-spin" style="font-size: 24px" />
				<span>{{ $t("settings.design.uploading") }}</span>
			</div>
			<div v-else class="idle-state">
				<UISysIcon icon="fa-solid fa-cloud-arrow-up" style="font-size: 32px" />
				<span>{{ $t("imageUpload.dragDrop") }}</span>
				<span class="hint">{{ $t("imageUpload.formats") }} ({{ $t("imageUpload.maxSize", { size: maxSizeValue }) }})</span>
			</div>
		</div>

		<!-- Error message -->
		<div v-if="error" class="error">{{ error }}</div>

		<!-- Hidden file input -->
		<input
			ref="fileInput"
			type="file"
			:accept="acceptValue"
			style="display: none"
			@change="handleFileSelect"
		/>
	</div>
</template>

<style scoped>
.image-upload {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.label {
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text);
}

.preview {
	position: relative;
	overflow: hidden;
	border-radius: var(--radius);
	border: 1px solid var(--border);
}

.preview img {
	width: 100%;
	max-height: 200px;
	object-fit: cover;
}

.preview-actions {
	display: flex;
	gap: 0.5rem;
	justify-content: center;
	padding: 0.75rem;
	background: var(--surface);
	border-top: 1px solid var(--border);
}

.upload-area {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.5rem;
	min-height: 150px;
	padding: 2rem;
	color: var(--text-secondary);
	cursor: pointer;
	background: var(--background);
	border: 2px dashed var(--border);
	border-radius: var(--radius);
	transition: all 0.2s;
}

.upload-area:hover {
	border-color: var(--primary);
	background: var(--surface);
}

.upload-area.uploading {
	cursor: wait;
	border-color: var(--primary);
}

.idle-state,
.uploading-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.5rem;
}

.hint {
	font-size: 0.75rem;
	color: var(--text-secondary);
}

.error {
	font-size: 0.8125rem;
	color: var(--error);
}
</style>
