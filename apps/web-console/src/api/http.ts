import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
const ACCESS_TOKEN_KEY = "accessToken";
const REFRESH_TOKEN_KEY = "refreshToken";

const http = axios.create({
	baseURL: "/api",
	timeout: 10_000,
});

http.interceptors.request.use((config) => {
	const token = localStorage.getItem(ACCESS_TOKEN_KEY);

	if (token) {
		config.headers.Authorization = `Bearer ${token}`;
	}

	return config;
});

let isRefreshing = false;
interface RefreshSession {
	accessToken: string;
	refreshToken: string;
}

let refreshPromise: Promise<RefreshSession | null> | null = null;

interface ApiResponse<T> {
	code: number;
	message: string;
	data: T;
}

http.interceptors.response.use(
	(response) => response,
	async (error: AxiosError) => {
		const original = error.config as
			| (InternalAxiosRequestConfig & { _retry?: boolean })
			| undefined;

		if (error.response?.status !== 401 || !original || original._retry) {
			return Promise.reject(error);
		}

		const url = original.url ?? "";

		if (url.includes("/auth/login") || url.includes("/auth/refresh")) {
			return Promise.reject(error);
		}

		original._retry = true;

		const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);
		if (!refreshToken) {
			clearTokens();
			window.location.assign("/login");
			return Promise.reject(error);
		}

		if (!isRefreshing) {
			isRefreshing = true;
			refreshPromise = axios
				.post<ApiResponse<RefreshSession>>("/api/auth/refresh", {
					refreshToken,
				})
				.then(({ data }) => data.data)
				.finally(() => {
					isRefreshing = false;
					refreshPromise = null;
				});
		}

		try {
			const data = await refreshPromise;
			const accessToken = data?.accessToken;
			const newRefreshToken = data?.refreshToken;

			if (!accessToken || !newRefreshToken) {
				throw new Error("refresh token failed");
			}

			localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
			localStorage.setItem(REFRESH_TOKEN_KEY, newRefreshToken);
			original.headers.Authorization = `Bearer ${accessToken}`;

			return http(original);
		} catch (refreshError) {
			clearTokens();
			window.location.assign("/login");
			return Promise.reject(refreshError);
		}
	},
);

export function clearTokens() {
	localStorage.removeItem(ACCESS_TOKEN_KEY);
	localStorage.removeItem(REFRESH_TOKEN_KEY);
}

export default http;
