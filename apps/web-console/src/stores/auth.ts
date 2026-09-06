import { defineStore } from "pinia";
import {
	login as loginRequest,
	logout as logoutRequest,
	register as registerRequest,
	type AuthUser,
} from "../api/auth";

const ACCESS_TOKEN_KEY = "accessToken";
const REFRESH_TOKEN_KEY = "refreshToken";

export const useAuthStore = defineStore("auth", {
	state: () => ({
		user: null as AuthUser | null,
		accessToken: localStorage.getItem(ACCESS_TOKEN_KEY),
		refreshToken: localStorage.getItem(REFRESH_TOKEN_KEY),
	}),

	getters: {
		isAuthenticated: (state) => Boolean(state.accessToken),
	},

	actions: {
		async login(email: string, password: string) {
			const result = await loginRequest(email, password);

			this.user = result.user;
			this.accessToken = result.accessToken;
			this.refreshToken = result.refreshToken;

			localStorage.setItem(ACCESS_TOKEN_KEY, result.accessToken);
			localStorage.setItem(REFRESH_TOKEN_KEY, result.refreshToken);
		},

		async register(email: string, password: string) {
			await registerRequest(email, password);
		},

		async logout() {
			const refreshToken = this.refreshToken;

			if (refreshToken) {
				try {
					await logoutRequest(refreshToken);
				} catch {
					// 即使服务端登出失败，也继续清理本地状态
				}
			}

			this.user = null;
			this.accessToken = null;
			this.refreshToken = null;

			localStorage.removeItem(ACCESS_TOKEN_KEY);
			localStorage.removeItem(REFRESH_TOKEN_KEY);
		},
	},
});
