import assert from "node:assert/strict";
import test from "node:test";

import { loginSchema, refreshSchema, registerSchema } from "./auth.validator";

test("registerSchema accepts valid email and password", () => {
	const result = registerSchema.safeParse({
		email: "test@example.com",
		password: "12345678",
	});

	assert.equal(result.success, true);
});

test("registerSchema rejects invalid email", () => {
	const result = registerSchema.safeParse({
		email: "not-an-email",
		password: "12345678",
	});

	assert.equal(result.success, false);
});

test("registerSchema rejects password shorter than 8 characters", () => {
	const result = registerSchema.safeParse({
		email: "test@example.com",
		password: "1234567",
	});

	assert.equal(result.success, false);
});

test("loginSchema accepts valid email and password", () => {
	const result = loginSchema.safeParse({
		email: "test@example.com",
		password: "12345678",
	});

	assert.equal(result.success, true);
});

test("loginSchema rejects empty email", () => {
	const result = loginSchema.safeParse({
		email: "",
		password: "12345678",
	});

	assert.equal(result.success, false);
});

test("refreshSchema accepts non-empty refreshToken", () => {
	const result = refreshSchema.safeParse({
		refreshToken: "some-refresh-token",
	});

	assert.equal(result.success, true);
});

test("refreshSchema rejects empty refreshToken", () => {
	const result = refreshSchema.safeParse({
		refreshToken: "",
	});

	assert.equal(result.success, false);
});
