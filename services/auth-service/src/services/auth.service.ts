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
	hashToken,
	verifyRefreshToken,
} from "../utils/jwt";

const SALT_ROUNDS = 10;

/**
 * @description: 用户注册
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
 * @description: 用户登录
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

	await prisma.refreshToken.create({
		data: {
			tokenHash: hashToken(refreshToken),
			userId: user.id,
			expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
		},
	});

	return {
		user: { id: user.id, email: user.email },
		accessToken,
		refreshToken,
	};
}

/**
 * @description: 刷新accessToken
 * @param {RefreshInput} data
 * @return {*}
 */
export async function refreshAccessToken(data: RefreshInput) {
	const payload = verifyRefreshToken(data.refreshToken);

	if (!payload) {
		throw new AppError(401, 1004, "refreshToken 无效或过期");
	}

	const tokenHash = hashToken(data.refreshToken);
	const stored = await prisma.refreshToken.findUnique({
		where: { tokenHash },
	});

	if (!stored || stored.revokedAt || stored.expiresAt <= new Date()) {
		throw new AppError(401, 1004, "refreshToken 已失效");
	}
	await prisma.refreshToken.update({
		where: {
			id: stored.id,
		},
		data: {
			revokedAt: new Date(),
		},
	});

	const newRefreshToken = generateRefreshToken({
		sub: payload.sub,
		email: payload.email,
	});

	await prisma.refreshToken.create({
		data: {
			tokenHash: hashToken(newRefreshToken),
			userId: payload.sub,
			expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
		},
	});

	const accessToken = generateAccessToken({
		sub: payload.sub,
		email: payload.email,
	});

	return {
		accessToken,
		refreshToken: newRefreshToken,
	};
}

export async function logoutUser(data: RefreshInput) {
	const tokenHash = hashToken(data.refreshToken);
	await prisma.refreshToken.updateMany({
		where: {
			tokenHash,
			revokedAt: null,
		},
		data: {
			revokedAt: new Date(),
		},
	});
}
