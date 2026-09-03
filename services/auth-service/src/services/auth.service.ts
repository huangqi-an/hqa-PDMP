import bcrypt from "bcryptjs";
import prisma from "../config/db";
import {
	type LoginInput,
	type RefreshInput,
	type RegisterInput,
} from "../validator/auth.validator";
import { AppError } from "../utils/error";
import {
	generateAccessToken,
	generateRefreshToken,
	verifyAccessToken,
	verifyRefreshToken,
} from "../utils/jwt";

const SALT_ROUNDS = 10;

/**
 * @description: (service)用户注册
 * @param {RegisterInput} data
 * @return {*}
 */
export async function registerUser(data: RegisterInput) {
	const { email, password } = data;
	//1. 检查是否已注册
	const existingUser = await prisma.user.findUnique({
		where: { email },
	});

	if (existingUser) {
		throw new AppError(409, 1001, "邮箱已注册");
	}

	const passwordHash = await bcrypt.hash(password, SALT_ROUNDS);
	const newUser = await prisma.user.create({
		data: {
			email,
			passwordHash,
		},
		select: {
			id: true,
			email: true,
		},
	});

	return newUser;
}

/**
 * @description: (service)用户登录
 * @param {LoginInput} data
 * @return {*}
 */
export async function loginUser(data: LoginInput) {
	const { email, password } = data;
	const user = await prisma.user.findUnique({
		where: { email },
		select: {
			id: true,
			email: true,
			passwordHash: true,
		},
	});
	if (!user) {
		throw new AppError(401, 1003, "邮箱或密码错误");
	}

	const passwordVaild = await bcrypt.compare(password, user.passwordHash);
	if (!passwordVaild) {
		throw new AppError(401, 1003, "邮箱或密码错误");
	}

	const payload = { sub: user.id, email: user.email };
	const accessToken = generateAccessToken(payload);
	const refreshToken = generateRefreshToken(payload);

	return {
		user: { id: user.id, email: user.email },
		accessToken,
		refreshToken,
	};
}

/**
 * @description:  (service)刷新accessToken
 * @param {RefreshInput} data
 * @return {*}
 */
export async function refreshAccessToken(data: RefreshInput) {
	const payload = verifyRefreshToken(data.refreshToken);

	if (!payload) {
		throw new AppError(401, 1004, "refreshToken 无效或过期");
	}
	const user = await prisma.user.findUnique({
		where: { id: payload.sub },
		select: {
			id: true,
			email: true,
		},
	});

	if (!user) {
		throw new AppError(401, 1004, "用户不存在");
	}

	const accessToken = generateAccessToken({
		sub: user.id,
		email: user.email,
	});

	return { accessToken };
}
