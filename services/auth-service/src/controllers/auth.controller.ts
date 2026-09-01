import type { Request, Response, NextFunction } from "express";
import { registerUser } from "../services/auth.service";
import { registerSchema } from "../validator/auth.validator";
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
