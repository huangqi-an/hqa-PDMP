import type { NextFunction, Request, Response } from "express";
import { AppError } from "../utils/error";
import { verifyAccessToken } from "../utils/jwt";

/**
 * @description: JWT鉴权中间件
 * @param {Request} req
 * @param {Response} res
 * @param {NextFunction} next
 * @return {*}
 */
export function requireAuth(req: Request, res: Response, next: NextFunction) {
	const authHeader = req.headers.authorization;
	if (!authHeader?.startsWith("Bearer ")) {
		return next(new AppError(401, 1005, "未提供认证信息"));
	}
	const token = authHeader.slice("Bearer ".length);
	const payload = verifyAccessToken(token);
	console.log("🚀 ~ requireAuth ~ payload:", payload);
	if (!payload) {
		return next(new AppError(401, 1005, "未认证或token已过期"));
	}
	res.locals.user = {
		id: payload.sub,
		email: payload.email,
	};
	next();
}
