import { defineStore } from "pinia";
export const useAuthStore = defineStore("auth", () => {
	const user = ref<User | null>(null);
	const token = ref<string | null>(null);
	const isLoading = ref(true);

	const { authApi } = useApi();

	const init = async () => {
		token.value = localStorage.getItem("token");
		if (token.value) {
			try {
				user.value = await authApi.me();
			} catch {
				localStorage.removeItem("token");
				token.value = null;
			}
		}
		isLoading.value = false;
	};

	const login = async (email: string, password: string) => {
		const response = await authApi.login(email, password);
		localStorage.setItem("token", response.token);
		token.value = response.token;
		user.value = response.user;
	};

	const register = async (email: string, password: string, name: string) => {
		const response = await authApi.register(email, password, name);
		localStorage.setItem("token", response.token);
		token.value = response.token;
		user.value = response.user;
	};

	const logout = () => {
		localStorage.removeItem("token");
		token.value = null;
		user.value = null;
	};

	return {
		user,
		token,
		isLoading,
		init,
		login,
		register,
		logout,
	};
});
