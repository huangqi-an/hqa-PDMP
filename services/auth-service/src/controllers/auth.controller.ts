import type { Request, Response, NextFunction } from "express";
import { loginUser, registerUser } from "../services/auth.service";
import { loginSchema, registerSchema } from "../validator/auth.validator";
import { successResponse, errorResponse } from "../utils/response";

export async function register(
	req: Request,
	res: Response,
	next: NextFunction,
) {
	try {
		const parsed = registerSchema.safeParse(req.body);

		if (!parsed.success) {
			return res
				.status(400)
				.json(
					errorResponse(1002, parsed.error.issues[0]?.message ?? "参数错误"),
				);
		}
		const user = await registerUser(parsed.data);
		return res.status(201).json(successResponse({ user }));
	} catch (error) {
		next(error);
	}
}

export async function login(req: Request, res: Response, next: NextFunction) {
	try {
		const parsed = loginSchema.safeParse(req.body);
		if (!parsed.success) {
			return res
				.status(400)
				.json(
					errorResponse(1002, parsed.error.issues[0]?.message ?? "参数错误"),
				);
		}

		const result = await loginUser(parsed.data);
		return res.status(200).json(successResponse(result));
	} catch (error) {
		next(error);
	}
}
