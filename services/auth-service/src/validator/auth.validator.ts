import { z } from "zod";

// 注册
export const registerSchema = z.object({
	email: z.email("邮箱格式错误"),
	password: z.string().min(8, "密码至少8位"),
});

export type RegisterInput = z.infer<typeof registerSchema>;

//登录
export const loginSchema = z.object({
	email: z.email("邮箱格式错误"),
	password: z.string().min(8, "密码至少8位"),
});

export type LoginInput = z.infer<typeof loginSchema>;

//刷新token
export const refreshSchema = z.object({
	refreshToken: z.string().min(1, "refreshToken不能为空"),
});

export type RefreshInput = z.infer<typeof refreshSchema>;
