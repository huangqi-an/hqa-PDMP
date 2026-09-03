import type { Request, Response, NextFunction } from "express";
import {
	loginUser,
	logoutUser,
	refreshAccessToken,
	registerUser,
} from "../services/auth.service";
import {
	loginSchema,
	refreshSchema,
	registerSchema,
} from "../validator/auth.validator";
import { successResponse, errorResponse } from "../utils/response";

/**
 * @description: 用户注册
 * @return {*}
 */
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

/**
 * @description: 用户登录
 * @param {Request} req
 * @param {Response} res
 * @param {NextFunction} next
 * @return {*}
 */
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

/**
 * @description: 刷新Token
 * @param {Request} req
 * @param {Response} res
 * @param {NextFunction} next
 * @return {*}
 */
export async function refresh(req: Request, res: Response, next: NextFunction) {
	try {
		const parsed = refreshSchema.safeParse(req.body);
		if (!parsed.success) {
			return res
				.status(400)
				.json(
					errorResponse(1002, parsed.error.issues[0]?.message ?? "参数错误"),
				);
		}
		const result = await refreshAccessToken(parsed.data);
		return res.status(200).json(successResponse(result));
	} catch (error) {
		next(error);
	}
}

/**
 * @description: 用户登出
 * @param {Request} req
 * @param {Response} res
 * @param {NextFunction} next
 * @return {*}
 */
export async function logout(req: Request, res: Response, next: NextFunction) {
	try {
		const parsed = refreshSchema.safeParse(req.body);
		if (!parsed.success) {
			return res
				.status(400)
				.json(
					errorResponse(1002, parsed.error.issues[0]?.message ?? "参数错误"),
				);
		}
		await logoutUser(parsed.data);

		return res.status(200).json(successResponse(null, "ok"));
	} catch (error) {}
}
