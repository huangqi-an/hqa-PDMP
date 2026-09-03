import type { NextFunction, Request, Response } from "express";
import { AppError } from "../utils/error";
import { getUserById, updateUserProfile } from "../services/user.service";
import { errorResponse, successResponse } from "../utils/response";
import { updateProfileSchema } from "../validator/user.validator";

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

export async function updateMe(
	req: Request,
	res: Response,
	next: NextFunction,
) {
	try {
		const userId = res.locals.user?.id as string | undefined;
		if (!userId) {
			throw new AppError(401, 1005, "未认证");
		}
		const parsed = updateProfileSchema.safeParse(req.body);
		if (!parsed.success) {
			return res
				.status(400)
				.json(
					errorResponse(1002, parsed.error.issues[0]?.message ?? "参数错误"),
				);
		}
		const user = await updateUserProfile(userId, parsed.data);
		return res.status(200).json(successResponse({ user }));
	} catch (error) {
		next(error);
	}
}
