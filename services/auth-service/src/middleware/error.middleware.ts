import type { Request, Response, NextFunction } from "express";
import { errorResponse } from "../utils/response";
import { ZodError } from "zod";
import { AppError } from "../utils/error";

export function errorHandler(
	err: unknown,
	_req: Request,
	res: Response,
	_next: NextFunction,
) {
	if (err instanceof ZodError) {
		res
			.status(400)
			.json(errorResponse(1002, err.issues[0]?.message ?? "参数错误"));
		return;
	}
	if (err instanceof AppError) {
		res.status(err.statusCode).json(errorResponse(err.code, err.message));
		return;
	}
	console.error(err);
	res.status(500).json(errorResponse(9999, "服务器内部错误"));
}
