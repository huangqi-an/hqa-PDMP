import prisma from "../config/db";
import { AppError } from "../utils/error";
import type { UpdateProfileInput } from "../validator/user.validator";

/**
 * @description: 根据id获取用户
 * @param {string} id
 * @return {*}
 */
export async function getUserById(id: string) {
	const user = await prisma.user.findUnique({
		where: { id },
		select: {
			id: true,
			email: true,
			name: true,
			createdAt: true,
			updatedAt: true,
		},
	});
	if (!user) {
		throw new AppError(404, 1006, "用户不存在");
	}
	return user;
}

/**
 * @description: 根据id更新用户信息
 * @param {string} id
 * @param {UpdateProfileInput} data
 * @return {*}
 */
export async function updateUserProfile(id: string, data: UpdateProfileInput) {
	const exisitingUser = await prisma.user.findUnique({
		where: {
			id,
		},
		select: {
			id: true,
		},
	});
	if (!exisitingUser) {
		throw new AppError(404, 1006, "用户不存在");
	}
	const user = await prisma.user.update({
		where: { id },
		data: {
			name: data.name,
		},
		select: {
			id: true,
			email: true,
			name: true,
			createdAt: true,
			updatedAt: true,
		},
	});
	return user;
}
