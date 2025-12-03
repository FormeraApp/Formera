interface AutoSaveOptions {
	delay?: number;
	onSave: () => Promise<void>;
	onError?: (error: unknown) => void;
}

export const useAutoSave = (options: AutoSaveOptions) => {
	const { t } = useI18n();
	const { delay = 2000, onSave, onError } = options;

	const isSaving = ref(false);
	const isDirty = ref(false);
	const lastSaved = ref<Date | null>(null);
	const error = ref<string | null>(null);

	let saveTimeout: ReturnType<typeof setTimeout> | null = null;

	const save = async () => {
		if (!isDirty.value || isSaving.value) return;

		if (saveTimeout) {
			clearTimeout(saveTimeout);
			saveTimeout = null;
		}

		isSaving.value = true;
		error.value = null;

		try {
			await onSave();
			isDirty.value = false;
			lastSaved.value = new Date();
		} catch (err) {
			error.value = t("autoSave.saveFailed");
			onError?.(err);
		} finally {
			isSaving.value = false;
		}
	};

	const scheduleSave = () => {
		isDirty.value = true;

		if (saveTimeout) {
			clearTimeout(saveTimeout);
		}

		saveTimeout = setTimeout(save, delay);
	};

	const markDirty = () => {
		isDirty.value = true;
		scheduleSave();
	};

	const saveNow = async () => {
		if (saveTimeout) {
			clearTimeout(saveTimeout);
			saveTimeout = null;
		}
		await save();
	};

	const cancel = () => {
		if (saveTimeout) {
			clearTimeout(saveTimeout);
			saveTimeout = null;
		}
	};

	// Keyboard shortcut: Ctrl+S / Cmd+S
	const handleKeydown = (event: KeyboardEvent) => {
		if ((event.ctrlKey || event.metaKey) && event.key === "s") {
			event.preventDefault();
			saveNow();
		}
	};

	onMounted(() => {
		window.addEventListener("keydown", handleKeydown);
	});

	onUnmounted(() => {
		window.removeEventListener("keydown", handleKeydown);
		cancel();
	});

	// Save before leaving page
	onBeforeUnmount(async () => {
		if (isDirty.value) {
			await saveNow();
		}
	});

	const lastSavedText = computed(() => {
		if (!lastSaved.value) return null;
		return t("autoSave.lastSaved", { time: lastSaved.value.toLocaleTimeString() });
	});

	const statusText = computed(() => {
		if (isSaving.value) return t("autoSave.saving");
		if (error.value) return error.value;
		if (isDirty.value) return t("autoSave.unsavedChanges");
		if (lastSaved.value) return lastSavedText.value;
		return null;
	});

	return {
		isSaving,
		isDirty,
		lastSaved,
		lastSavedText,
		statusText,
		error,
		markDirty,
		saveNow,
		cancel,
	};
};
