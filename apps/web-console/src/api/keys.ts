import http from "./http";

export interface APIKey {
	id: string;
	provider: string;
	name: string;
	maskedKey: string;
	keyHint: string;
	notes?: string | null;
	tags: string[];
	createdAt: string;
	updatedAt: string;
}

export interface APIKeyListResult {
	items: APIKey[];
	total: number;
}

interface ApiResponse<T> {
	code: number;
	message: string;
	data: T;
}

export interface CreateAPIKeyInput {
	provider: string;
	name: string;
	key: string;
	notes?: string | null;
	tags?: string[];
}

export interface UpdateAPIKeyInput {
	provider?: string;
	name?: string;
	key?: string;
	notes?: string | null;
	tags?: string[];
}

export async function listKeys(params?: {
	page?: number;
	pageSize?: number;
	provider?: string;
	q?: string;
}) {
	const { data } = await http.get<ApiResponse<APIKeyListResult>>("/keys", {
		params,
	});

	return data.data;
}

export async function createKey(input: CreateAPIKeyInput) {
	const { data } = await http.post<ApiResponse<APIKey>>("/keys", input);

	return data.data;
}

export async function getKey(id: string) {
	const { data } = await http.get<ApiResponse<APIKey>>(`/keys/${id}`);

	return data.data;
}

export async function updateKey(id: string, input: UpdateAPIKeyInput) {
	const { data } = await http.patch<ApiResponse<APIKey>>(`/keys/${id}`, input);

	return data.data;
}

export async function deleteKey(id: string) {
	await http.delete(`/keys/${id}`);
}

export async function revealKey(id: string) {
	const { data } = await http.post<ApiResponse<{ key: string }>>(
		`/keys/${id}/reveal`,
	);

	return data.data.key;
}
