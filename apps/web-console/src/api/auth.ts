import http from "./http";

export interface AuthUser {
	id: string;
	email: string;
}

export interface LoginResult {
	user: AuthUser;
	accessToken: string;
	refreshToken: string;
}

export interface RefreshResult {
	accessToken: string;
	refreshToken: string;
}

interface ApiResponse<T> {
	code: number;
	message: string;
	data: T;
}

export async function login(email: string, password: string) {
	const { data } = await http.post<ApiResponse<LoginResult>>("/auth/login", {
		email,
		password,
	});

	return data.data;
}

export async function register(email: string, password: string) {
	const { data } = await http.post<ApiResponse<AuthUser>>("/auth/register", {
		email,
		password,
	});

	return data.data;
}

export async function logout(refreshToken: string) {
	await http.post("/auth/logout", { refreshToken });
}
