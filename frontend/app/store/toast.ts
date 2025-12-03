import { defineStore } from "pinia";

export type ToastType = "success" | "error" | "warning" | "info";

export interface Toast {
	id: string;
	type: ToastType;
	title: string;
	message?: string;
	duration?: number;
}

export const useToastStore = defineStore("toast", () => {
	const toasts = ref<Toast[]>([]);

	const add = (toast: Omit<Toast, "id">) => {
		const id = crypto.randomUUID();
		const duration = toast.duration ?? 5000;

		toasts.value.push({ ...toast, id });

		if (duration > 0) {
			setTimeout(() => remove(id), duration);
		}
	};

	const remove = (id: string) => {
		const index = toasts.value.findIndex((t) => t.id === id);
		if (index > -1) {
			toasts.value.splice(index, 1);
		}
	};

	const success = (title: string, message?: string) => {
		add({ type: "success", title, message });
	};

	const error = (title: string, message?: string) => {
		add({ type: "error", title, message, duration: 8000 });
	};

	const warning = (title: string, message?: string) => {
		add({ type: "warning", title, message });
	};

	const info = (title: string, message?: string) => {
		add({ type: "info", title, message });
	};

	return {
		toasts,
		add,
		remove,
		success,
		error,
		warning,
		info,
	};
});
