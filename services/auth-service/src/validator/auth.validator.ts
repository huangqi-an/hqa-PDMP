import { z } from "zod";

export const registerSchema = z.object({
	email: z.email("邮箱格式错误"),
	password: z.string().min(8, "密码至少8位"),
});

export type RegisterInput = z.infer<typeof registerSchema>;

export const loginSchema = z.object({
	email: z.email("邮箱格式错误"),
	password: z.string().min(8, "密码至少8位"),
});

export type LoginInput = z.infer<typeof loginSchema>;
