import bcrypt from "bcryptjs";
import prisma from "../config/db";
import { type RegisterInput } from "../validator/auth.validator";
import { AppError } from "../utils/error";

const SALT_ROUNDS = 10;

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
