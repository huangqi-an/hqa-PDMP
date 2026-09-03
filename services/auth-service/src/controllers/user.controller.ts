import type { NextFunction, Request, Response } from "express";
import { AppError } from "../utils/error";
import { getUserById } from "../services/user.service";
import { successResponse } from "../utils/response";

/**
 * @description: (controller)根据ID获取用户信息
 * @param {Request} _req
 * @param {Response} res
 * @param {NextFunction} next
 * @return {*}
 */
export async function getMe(_req: Request, res: Response, next: NextFunction) {
	try {
		const userId = res.locals.user?.id as string | undefined;
		if (!userId) {
			throw new AppError(401, 1005, "未认证");
		}
		const user = await getUserById(userId);
		return res.status(200).json(successResponse({ user }));
	} catch (error) {
		next(error);
	}
}
