export class AppError extends Error {
	constructor(
		public statusCode: number,
		public code: number,
		message: string,
	) {
		super(message);
	}
}
