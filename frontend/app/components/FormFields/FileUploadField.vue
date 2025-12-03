<script lang="ts" setup>
interface UploadedFile {
	id: string;
	path: string;
	url: string;
	filename: string;
	size: number;
	mimeType: string;
}

const props = withDefaults(
	defineProps<{
		modelValue?: string[];
		required?: boolean;
		multiple?: boolean;
		allowedTypes?: string[];
		maxFileSize?: number;
		fieldId: string;
	}>(),
	{
		modelValue: () => [],
		required: false,
		multiple: false,
		allowedTypes: () => [],
		maxFileSize: 10,
	}
);

const emit = defineEmits<{
	"update:modelValue": [value: string[]];
	change: [];
}>();

const { t } = useI18n();
const config = useRuntimeConfig();
const apiUrl = config.public.apiUrl as string;

const uploadedFiles = ref<UploadedFile[]>([]);
const isUploading = ref(false);
const uploadError = ref<string | null>(null);

// Sync uploaded files with modelValue on mount
onMounted(() => {
	if (props.modelValue.length > 0) {
		uploadedFiles.value = props.modelValue.map((path, index) => ({
			id: `existing-${index}`,
			path,
			url: getFileUrl(path),
			filename: path.split("/").pop() || "file",
			size: 0,
			mimeType: "",
		}));
	}
});

const handleChange = async (event: Event) => {
	const input = event.target as HTMLInputElement;
	const files = input.files;
	if (!files || files.length === 0) return;

	uploadError.value = null;
	isUploading.value = true;

	const filesToUpload = Array.from(files);

	// Validate file sizes
	for (const file of filesToUpload) {
		if (file.size > props.maxFileSize * 1024 * 1024) {
			uploadError.value = t("fileUpload.fileTooLarge", { name: file.name, max: props.maxFileSize });
			isUploading.value = false;
			input.value = "";
			return;
		}
	}

	try {
		const newFiles: UploadedFile[] = [];

		for (const file of filesToUpload) {
			const formData = new FormData();
			formData.append("file", file);

			const response = await fetch(`${apiUrl}/public/upload`, {
				method: "POST",
				body: formData,
			});

			if (!response.ok) {
				const error = await response.json().catch(() => ({ error: t("fileUpload.uploadFailed") }));
				throw new Error(error.error || t("fileUpload.uploadFailed"));
			}

			const result = await response.json();

			newFiles.push({
				id: result.id,
				path: result.path,
				url: getFileUrl(result.path),
				filename: result.filename || file.name,
				size: result.size || file.size,
				mimeType: result.mimeType || file.type,
			});
		}

		// Add new files to existing ones (or replace if not multiple)
		if (props.multiple) {
			uploadedFiles.value = [...uploadedFiles.value, ...newFiles];
		} else {
			uploadedFiles.value = newFiles;
		}

		// Emit the paths as the model value
		const paths = uploadedFiles.value.map((f) => f.path);
		emit("update:modelValue", paths);
		emit("change");
	} catch (err) {
		uploadError.value = err instanceof Error ? err.message : t("fileUpload.uploadFailed");
	} finally {
		isUploading.value = false;
		input.value = "";
	}
};

const removeFile = (index: number) => {
	uploadedFiles.value.splice(index, 1);
	const paths = uploadedFiles.value.map((f) => f.path);
	emit("update:modelValue", paths);
	emit("change");
};

const formatFileSize = (bytes: number): string => {
	if (bytes === 0) return "";
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
	return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};

const acceptTypes = computed(() => {
	return props.allowedTypes.length > 0 ? props.allowedTypes.join(",") : undefined;
});
</script>

<template>
	<div class="file-upload">
		<input
			:id="`file-${fieldId}`"
			type="file"
			class="file-input"
			:required="required && uploadedFiles.length === 0"
			:multiple="multiple"
			:accept="acceptTypes"
			:disabled="isUploading"
			@change="handleChange"
		/>
		<label :for="`file-${fieldId}`" class="file-label" :class="{ disabled: isUploading }">
			<template v-if="isUploading">
				<UISysIcon icon="fa-solid fa-spinner fa-spin" />
				<span>{{ t("fileUpload.uploading") }}</span>
			</template>
			<template v-else>
				<UISysIcon icon="fa-solid fa-upload" />
				<span>{{ multiple ? t("fileUpload.selectFiles") : t("fileUpload.selectFile") }}</span>
			</template>
		</label>

		<!-- Upload Error -->
		<div v-if="uploadError" class="upload-error">
			<UISysIcon icon="fa-solid fa-circle-exclamation" />
			{{ uploadError }}
		</div>

		<!-- Uploaded Files List -->
		<div v-if="uploadedFiles.length > 0" class="file-list">
			<div v-for="(file, index) in uploadedFiles" :key="file.id" class="file-item">
				<div class="file-info">
					<UISysIcon icon="fa-solid fa-file" class="file-icon" />
					<span class="file-name">{{ file.filename }}</span>
					<span v-if="file.size > 0" class="file-size">{{ formatFileSize(file.size) }}</span>
				</div>
				<button type="button" class="remove-btn" :title="t('fileUpload.remove')" @click="removeFile(index)">
					<UISysIcon icon="fa-solid fa-xmark" />
				</button>
			</div>
		</div>

		<p v-if="maxFileSize || allowedTypes.length" class="file-hint">
			<template v-if="maxFileSize">{{ t("fileUpload.maxSize", { size: maxFileSize }) }}</template>
			<template v-if="maxFileSize && allowedTypes.length"> · </template>
			<template v-if="allowedTypes.length">{{ allowedTypes.join(", ") }}</template>
		</p>
	</div>
</template>

<style scoped>
.file-upload {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.file-input {
	position: absolute;
	width: 1px;
	height: 1px;
	padding: 0;
	margin: -1px;
	overflow: hidden;
	clip: rect(0, 0, 0, 0);
	white-space: nowrap;
	border: 0;
}

.file-label {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	justify-content: center;
	padding: 1rem;
	color: var(--text-secondary);
	cursor: pointer;
	background: var(--background);
	border: 2px dashed var(--border);
	border-radius: var(--radius);
	transition: all 0.2s;
}

.file-label:hover:not(.disabled) {
	color: var(--primary);
	border-color: var(--primary);
}

.file-label.disabled {
	cursor: not-allowed;
	opacity: 0.7;
}

.upload-error {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	padding: 0.5rem 0.75rem;
	font-size: 0.8125rem;
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
	border-radius: var(--radius);
}

.file-list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.file-item {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	justify-content: space-between;
	padding: 0.5rem 0.75rem;
	background: var(--background);
	border-radius: var(--radius);
}

.file-info {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	flex: 1;
	min-width: 0;
}

.file-icon {
	flex-shrink: 0;
	color: var(--text-secondary);
}

.file-name {
	font-size: 0.875rem;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.file-size {
	flex-shrink: 0;
	font-size: 0.75rem;
	color: var(--text-secondary);
}

.remove-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 1.5rem;
	height: 1.5rem;
	padding: 0;
	color: var(--text-secondary);
	cursor: pointer;
	background: transparent;
	border: none;
	border-radius: var(--radius);
	transition: all 0.2s;
}

.remove-btn:hover {
	color: var(--error);
	background: rgba(239, 68, 68, 0.1);
}

.file-hint {
	margin: 0;
	font-size: 0.75rem;
	color: var(--text-secondary);
}
</style>
