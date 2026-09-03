import prisma from "../config/db";
import { AppError } from "../utils/error";

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
