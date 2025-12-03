import { defineStore } from "pinia";

export const useFormsStore = defineStore("forms", () => {
	const { formsApi } = useApi();
	const toastStore = useToastStore();
	const { t } = useI18n();

	const forms = ref<Form[]>([]);
	const currentForm = ref<Form | null>(null);
	const isLoading = ref(false);
	const isLoadingForm = ref(false);
	const lastFetched = ref<number | null>(null);

	const CACHE_DURATION = 5 * 60 * 1000;

	const isCacheValid = computed(() => {
		if (!lastFetched.value) return false;
		return Date.now() - lastFetched.value < CACHE_DURATION;
	});

	const fetchForms = async (force = false) => {
		if (!force && isCacheValid.value && forms.value.length > 0) {
			return forms.value;
		}

		isLoading.value = true;
		try {
			const data = await formsApi.list();
			forms.value = data || [];
			lastFetched.value = Date.now();
			return forms.value;
		} catch (error) {
			toastStore.error(t("common.error"), t("store.forms.loadError"));
			throw error;
		} finally {
			isLoading.value = false;
		}
	};

	const fetchForm = async (id: string, force = false) => {
		if (!force && currentForm.value?.id === id) {
			return currentForm.value;
		}

		const cachedForm = forms.value.find((f) => f.id === id);
		if (!force && cachedForm && isCacheValid.value) {
			currentForm.value = cachedForm;
			return cachedForm;
		}

		isLoadingForm.value = true;
		try {
			const data = await formsApi.get(id);
			currentForm.value = data;

			const index = forms.value.findIndex((f) => f.id === id);
			if (index > -1) {
				forms.value[index] = data;
			}

			return data;
		} catch (error) {
			toastStore.error(t("common.error"), t("store.forms.loadFormError"));
			throw error;
		} finally {
			isLoadingForm.value = false;
		}
	};

	const createForm = async (formData: Partial<Form>) => {
		try {
			const newForm = await formsApi.create(formData);
			forms.value.unshift(newForm);
			toastStore.success(t("store.forms.created"), t("store.forms.createdMessage"));
			return newForm;
		} catch (error) {
			toastStore.error(t("common.error"), t("store.forms.createError"));
			throw error;
		}
	};

	const updateForm = async (id: string, formData: UpdateFormRequest) => {
		try {
			const updated = await formsApi.update(id, formData);
			const index = forms.value.findIndex((f) => f.id === id);
			if (index > -1) {
				forms.value[index] = updated;
			}

			if (currentForm.value?.id === id) {
				currentForm.value = updated;
			}

			return updated;
		} catch (error) {
			toastStore.error(t("common.error"), t("store.forms.saveError"));
			throw error;
		}
	};

	const checkSlugAvailability = async (slug: string, excludeId?: string) => {
		try {
			return await formsApi.checkSlugAvailability(slug, excludeId);
		} catch {
			return { available: false, slug };
		}
	};

	const deleteForm = async (id: string) => {
		try {
			await formsApi.delete(id);
			forms.value = forms.value.filter((f) => f.id !== id);

			if (currentForm.value?.id === id) {
				currentForm.value = null;
			}

			toastStore.success(t("store.forms.deleted"), t("store.forms.deletedMessage"));
		} catch (error) {
			toastStore.error(t("common.error"), t("store.forms.deleteError"));
			throw error;
		}
	};

	const duplicateForm = async (id: string) => {
		try {
			const duplicated = await formsApi.duplicate(id);
			forms.value.unshift(duplicated);
			toastStore.success(t("store.forms.duplicated"), t("store.forms.duplicatedMessage"));
			return duplicated;
		} catch (error) {
			toastStore.error(t("common.error"), t("store.forms.duplicateError"));
			throw error;
		}
	};

	const invalidateCache = () => {
		lastFetched.value = null;
	};

	const clearCurrentForm = () => {
		currentForm.value = null;
	};

	return {
		forms,
		currentForm,
		isLoading,
		isLoadingForm,
		isCacheValid,
		fetchForms,
		fetchForm,
		createForm,
		updateForm,
		deleteForm,
		duplicateForm,
		checkSlugAvailability,
		invalidateCache,
		clearCurrentForm,
	};
});
