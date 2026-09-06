import http from "./http";

export interface ProfileUser {
	id: string;
	email: string;
	name?: string | null;
	createdAt: string;
	updatedAt: string;
}

interface ApiResponse<T> {
	code: number;
	message: string;
	data: T;
}

export async function getMe() {
	const { data } =
		await http.get<ApiResponse<{ user: ProfileUser }>>("/users/me");

	return data.data.user;
}

export async function updateMe(name: string | null) {
	const { data } = await http.patch<ApiResponse<{ user: ProfileUser }>>(
		"/users/me",
		{ name },
	);

	return data.data.user;
}
