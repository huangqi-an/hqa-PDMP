import z from "zod";

export const updateProfileSchema = z.object({
	name: z
		.string()
		.trim()
		.min(1, "昵称不能为空")
		.max(50, "昵称最多50个字符")
		.nullable(),
});
export type UpdateProfileInput = z.infer<typeof updateProfileSchema>;
