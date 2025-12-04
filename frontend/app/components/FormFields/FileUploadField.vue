<script lang="ts" setup>
// Pending file = selected but not yet uploaded (stored client-side)
interface PendingFile {
	id: string;
	file: File;
	filename: string;
	size: number;
	mimeType: string;
	previewUrl?: string;
}

// Uploaded file = already on server (from previous submission or after upload)
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

// Pending files (selected but not uploaded yet)
const pendingFiles = ref<PendingFile[]>([]);
// Already uploaded files (from modelValue - existing submissions)
const uploadedFiles = ref<UploadedFile[]>([]);
const validationError = ref<string | null>(null);

// Sync uploaded files with modelValue on mount (for existing data)
onMounted(() => {
	if (props.modelValue.length > 0) {
		uploadedFiles.value = props.modelValue.map((path, index) => ({
			id: `existing-${index}`,
			path,
			url: getFileUrl(path),
			filename: extractFilename(path),
			size: 0,
			mimeType: "",
		}));
	}
});

// Clean up preview URLs when component unmounts
onUnmounted(() => {
	pendingFiles.value.forEach((f) => {
		if (f.previewUrl) {
			URL.revokeObjectURL(f.previewUrl);
		}
	});
});

// Handle file selection - validate and store client-side only
const handleChange = (event: Event) => {
	const input = event.target as HTMLInputElement;
	const files = input.files;
	if (!files || files.length === 0) return;

	validationError.value = null;
	const filesToAdd = Array.from(files);

	// Validate file sizes
	for (const file of filesToAdd) {
		if (file.size > props.maxFileSize * 1024 * 1024) {
			validationError.value = t("fileUpload.fileTooLarge", { name: file.name, max: props.maxFileSize });
			input.value = "";
			return;
		}
	}

	// Validate file types if specified
	if (props.allowedTypes.length > 0) {
		for (const file of filesToAdd) {
			const isAllowed = props.allowedTypes.some((type) => {
				if (type.startsWith(".")) {
					return file.name.toLowerCase().endsWith(type.toLowerCase());
				}
				if (type.endsWith("/*")) {
					return file.type.startsWith(type.slice(0, -1));
				}
				return file.type === type;
			});
			if (!isAllowed) {
				validationError.value = t("fileUpload.invalidType", { name: file.name });
				input.value = "";
				return;
			}
		}
	}

	// Create pending file entries
	const newPending: PendingFile[] = filesToAdd.map((file) => ({
		id: `pending-${Date.now()}-${Math.random().toString(36).slice(2)}`,
		file,
		filename: file.name,
		size: file.size,
		mimeType: file.type,
		previewUrl: file.type.startsWith("image/") ? URL.createObjectURL(file) : undefined,
	}));

	// Add new files (or replace if not multiple)
	if (props.multiple) {
		pendingFiles.value = [...pendingFiles.value, ...newPending];
	} else {
		// Clean up old preview URLs
		pendingFiles.value.forEach((f) => {
			if (f.previewUrl) URL.revokeObjectURL(f.previewUrl);
		});
		pendingFiles.value = newPending;
		uploadedFiles.value = []; // Clear existing uploaded files in single mode
	}

	emit("change");
	input.value = "";
};

// Upload all pending files - called by parent form on submit
const uploadPendingFiles = async (): Promise<string[]> => {
	const uploadedPaths: string[] = [];

	for (const pending of pendingFiles.value) {
		const formData = new FormData();
		formData.append("file", pending.file);

		const response = await fetch(`${apiUrl}/api/public/upload`, {
			method: "POST",
			body: formData,
		});

		if (!response.ok) {
			const error = await response.json().catch(() => ({ error: t("fileUpload.uploadFailed") }));
			throw new Error(error.error || t("fileUpload.uploadFailed"));
		}

		const result = await response.json();
		uploadedPaths.push(result.path);
	}

	// Add existing uploaded file paths
	uploadedPaths.push(...uploadedFiles.value.map((f) => f.path));

	return uploadedPaths;
};

// Check if there are files ready (pending or uploaded)
const hasFiles = computed(() => pendingFiles.value.length > 0 || uploadedFiles.value.length > 0);

// Check if valid for submission
const isValid = computed(() => {
	if (props.required) {
		return hasFiles.value;
	}
	return true;
});

// Expose methods and state for parent form
defineExpose({
	uploadPendingFiles,
	hasFiles,
	isValid,
	hasPendingFiles: computed(() => pendingFiles.value.length > 0),
});

// Remove a pending file
const removePendingFile = (index: number) => {
	const file = pendingFiles.value[index];
	if (file?.previewUrl) {
		URL.revokeObjectURL(file.previewUrl);
	}
	pendingFiles.value.splice(index, 1);
	emit("change");
};

// Remove an already uploaded file
const removeUploadedFile = (index: number) => {
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
			:required="required && !hasFiles"
			:multiple="multiple"
			:accept="acceptTypes"
			@change="handleChange"
		/>
		<label :for="`file-${fieldId}`" class="file-label">
			<UISysIcon icon="fa-solid fa-upload" />
			<span>{{ multiple ? t("fileUpload.selectFiles") : t("fileUpload.selectFile") }}</span>
		</label>

		<!-- Validation Error -->
		<div v-if="validationError" class="upload-error">
			<UISysIcon icon="fa-solid fa-circle-exclamation" />
			{{ validationError }}
		</div>

		<!-- Pending Files List (not yet uploaded) -->
		<div v-if="pendingFiles.length > 0" class="file-list">
			<div v-for="(file, index) in pendingFiles" :key="file.id" class="file-item file-item-pending">
				<div class="file-preview">
					<img v-if="file.previewUrl" :src="file.previewUrl" :alt="file.filename" class="preview-image" />
					<UISysIcon v-else icon="fa-solid fa-file" class="file-icon" />
				</div>
				<div class="file-info">
					<span class="file-name">{{ file.filename }}</span>
					<span class="file-size">{{ formatFileSize(file.size) }}</span>
				</div>
				<span class="file-info-icon" :title="t('fileUpload.uploadOnSubmit')">
					<UISysIcon icon="fa-solid fa-circle-info" />
				</span>
				<button type="button" class="remove-btn" :title="t('fileUpload.remove')" @click="removePendingFile(index)">
					<UISysIcon icon="fa-solid fa-xmark" />
				</button>
			</div>
		</div>

		<!-- Already Uploaded Files List -->
		<div v-if="uploadedFiles.length > 0" class="file-list">
			<div v-for="(file, index) in uploadedFiles" :key="file.id" class="file-item">
				<div class="file-info">
					<UISysIcon icon="fa-solid fa-file" class="file-icon" />
					<span class="file-name">{{ file.filename }}</span>
					<span v-if="file.size > 0" class="file-size">{{ formatFileSize(file.size) }}</span>
				</div>
				<button type="button" class="remove-btn" :title="t('fileUpload.remove')" @click="removeUploadedFile(index)">
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

.file-label:hover {
	color: var(--primary);
	border-color: var(--primary);
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

.file-item-pending {
	border: 1px dashed var(--border);
	background: var(--surface);
}

.file-preview {
	flex-shrink: 0;
	width: 2.5rem;
	height: 2.5rem;
	display: flex;
	align-items: center;
	justify-content: center;
	background: var(--background);
	border-radius: var(--radius);
	overflow: hidden;
}

.preview-image {
	width: 100%;
	height: 100%;
	object-fit: cover;
}

.file-info-icon {
	display: flex;
	align-items: center;
	justify-content: center;
	color: var(--text-secondary);
	cursor: help;
	font-size: 0.875rem;
}

.file-hint {
	margin: 0;
	font-size: 0.75rem;
	color: var(--text-secondary);
}
</style>
