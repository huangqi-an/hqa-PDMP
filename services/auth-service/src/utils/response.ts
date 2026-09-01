export interface ApiResponse<T = any> {
	code: number;
	message: string;
	data?: T;
}

export function successResponse<T>(data: T, message = "ok"): ApiResponse<T> {
	return { code: 0, message, data };
}

export function errorResponse(
	code: number,
	message: string,
): ApiResponse<null> {
	return { code, message, data: null };
}
