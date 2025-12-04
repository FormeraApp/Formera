export const getFileUrl = (pathOrUrl: string | undefined | null): string => {
	if (!pathOrUrl) return "";

	if (pathOrUrl.startsWith("http://") || pathOrUrl.startsWith("https://") || pathOrUrl.startsWith("data:")) {
		return pathOrUrl;
	}

	const config = useRuntimeConfig();
	const apiUrl = config.public.apiUrl as string;

	const cleanPath = pathOrUrl.startsWith("/") ? pathOrUrl.slice(1) : pathOrUrl;
	return `${apiUrl}/uploads/${cleanPath}`;
};

export const useApi = () => {
	const config = useRuntimeConfig();
	const apiUrl = config.public.apiUrl as string;
	const apiBase = `${apiUrl}/api`;

	const getToken = () => {
		return localStorage.getItem("token");
	};

	const request = async <T>(endpoint: string, options: RequestInit = {}): Promise<T> => {
		const token = getToken();
		const headers: HeadersInit = {
			"Content-Type": "application/json",
			...(token ? { Authorization: `Bearer ${token}` } : {}),
			...options.headers,
		};

		const response = await fetch(`${apiBase}${endpoint}`, {
			...options,
			headers,
		});

		if (response.status === 401) {
			if (import.meta.client) {
				localStorage.removeItem("token");
				navigateTo("/login");
			}
			throw new Error("Unauthorized");
		}

		if (!response.ok) {
			const error = await response.json().catch(() => ({ error: "Request failed" }));
			throw new Error(error.error || "Request failed");
		}

		return response.json();
	};

	const authApi = {
		register: (email: string, password: string, name: string): Promise<AuthResponse> =>
			request("/auth/register", {
				method: "POST",
				body: JSON.stringify({ email, password, name }),
			}),
		login: (email: string, password: string): Promise<AuthResponse> =>
			request("/auth/login", {
				method: "POST",
				body: JSON.stringify({ email, password }),
			}),
		me: (): Promise<User> => request("/auth/me"),
	};

	const formsApi = {
		list: (params?: PaginationParams): Promise<PaginatedResponse<Form[]>> => {
			const searchParams = new URLSearchParams();
			if (params?.page) searchParams.append("page", params.page.toString());
			if (params?.pageSize) searchParams.append("page_size", params.pageSize.toString());
			const query = searchParams.toString();
			return request(`/forms${query ? `?${query}` : ""}`);
		},
		get: (id: string): Promise<Form> => request(`/forms/${id}`),
		getPublic: (id: string): Promise<Form> => request(`/public/forms/${id}`),
		create: (form: Partial<Form>): Promise<Form> =>
			request("/forms", {
				method: "POST",
				body: JSON.stringify(form),
			}),
		update: (id: string, form: UpdateFormRequest): Promise<Form> =>
			request(`/forms/${id}`, {
				method: "PUT",
				body: JSON.stringify(form),
			}),
		delete: (id: string): Promise<void> =>
			request(`/forms/${id}`, {
				method: "DELETE",
			}),
		duplicate: (id: string): Promise<Form> =>
			request(`/forms/${id}/duplicate`, {
				method: "POST",
			}),
		checkSlugAvailability: (slug: string, excludeId?: string): Promise<{ available: boolean; slug: string }> => {
			const params = new URLSearchParams({ slug });
			if (excludeId) params.append("exclude_id", excludeId);
			return request(`/forms/check-slug?${params.toString()}`);
		},
		verifyPassword: (id: string, password: string): Promise<{ valid: boolean; form?: Form }> =>
			request(`/public/forms/${id}/verify-password`, {
				method: "POST",
				body: JSON.stringify({ password }),
			}),
	};

	const submissionsApi = {
		submit: (formId: string, formData: Record<string, unknown>, metadata?: Record<string, string>): Promise<{ message: string; submission: Submission }> =>
			request(`/public/forms/${formId}/submit`, {
				method: "POST",
				body: JSON.stringify({ data: formData, metadata }),
			}),
		list: (formId: string, params?: PaginationParams): Promise<SubmissionsResponse> => {
			const searchParams = new URLSearchParams();
			if (params?.page) searchParams.append("page", params.page.toString());
			if (params?.pageSize) searchParams.append("page_size", params.pageSize.toString());
			const query = searchParams.toString();
			return request(`/forms/${formId}/submissions${query ? `?${query}` : ""}`);
		},
		get: (formId: string, submissionId: string): Promise<Submission> => request(`/forms/${formId}/submissions/${submissionId}`),
		delete: (formId: string, submissionId: string): Promise<void> =>
			request(`/forms/${formId}/submissions/${submissionId}`, {
				method: "DELETE",
			}),
		stats: (formId: string): Promise<FormStats> => request(`/forms/${formId}/stats`),
		exportCSV: (formId: string): string => {
			const token = getToken();
			return `${apiBase}/forms/${formId}/export/csv?token=${token}`;
		},
		exportJSON: (formId: string): string => {
			const token = getToken();
			return `${apiBase}/forms/${formId}/export/json?token=${token}`;
		},
	};

	const setupApi = {
		getStatus: (): Promise<SetupStatus> => request("/setup/status"),
		complete: (setupData: { email: string; password: string; name: string; app_name?: string; allow_registration: boolean }): Promise<AuthResponse> =>
			request("/setup/complete", {
				method: "POST",
				body: JSON.stringify(setupData),
			}),
	};

	const settingsApi = {
		get: (): Promise<Settings> => request("/settings"),
		update: (settings: Partial<Settings>): Promise<Settings> =>
			request("/settings", {
				method: "PUT",
				body: JSON.stringify(settings),
			}),
	};

	const usersApi = {
		list: (params?: PaginationParams): Promise<PaginatedResponse<User[]>> => {
			const searchParams = new URLSearchParams();
			if (params?.page) searchParams.append("page", params.page.toString());
			if (params?.pageSize) searchParams.append("page_size", params.pageSize.toString());
			const query = searchParams.toString();
			return request(`/users${query ? `?${query}` : ""}`);
		},
		get: (id: string): Promise<User> => request(`/users/${id}`),
		create: (user: { email: string; password: string; name: string; role?: UserRole }): Promise<User> =>
			request("/users", {
				method: "POST",
				body: JSON.stringify(user),
			}),
		update: (id: string, user: Partial<User & { password?: string }>): Promise<User> =>
			request(`/users/${id}`, {
				method: "PUT",
				body: JSON.stringify(user),
			}),
		delete: (id: string): Promise<void> =>
			request(`/users/${id}`, {
				method: "DELETE",
			}),
	};

	const uploadApi = {
		uploadImage: async (file: File): Promise<UploadResult> => {
			const token = getToken();
			if (!token) {
				throw new Error("Authentication required");
			}

			const formData = new FormData();
			formData.append("file", file);

			let response: Response;
			try {
				response = await fetch(`${apiBase}/uploads/image`, {
					method: "POST",
					headers: {
						Authorization: `Bearer ${token}`,
					},
					body: formData,
				});
			} catch {
				throw new Error("Network error - could not upload file");
			}

			if (!response.ok) {
				const error = await response.json().catch(() => ({ error: "Upload failed" }));
				throw new Error(error.error || `Upload failed (${response.status})`);
			}

			return response.json();
		},
		uploadFile: async (file: File): Promise<UploadResult> => {
			const token = getToken();
			const formData = new FormData();
			formData.append("file", file);

			let response: Response;
			try {
				response = await fetch(`${apiBase}/uploads/file`, {
					method: "POST",
					headers: {
						...(token ? { Authorization: `Bearer ${token}` } : {}),
					},
					body: formData,
				});
			} catch (networkError) {
				throw new Error("Network error - could not upload file");
			}

			if (!response.ok) {
				const error = await response.json().catch(() => ({ error: "Upload failed" }));
				throw new Error(error.error || `Upload failed (${response.status})`);
			}

			return response.json();
		},
		delete: (fileId: string): Promise<void> =>
			request(`/uploads/${fileId}`, {
				method: "DELETE",
			}),
	};

	return {
		authApi,
		formsApi,
		submissionsApi,
		setupApi,
		settingsApi,
		usersApi,
		uploadApi,
	};
};
