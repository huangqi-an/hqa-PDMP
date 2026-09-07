import assert from "node:assert/strict";
import test from "node:test";

import { updateProfileSchema } from "./user.validator";

test("updateProfileSchema accepts valid name", () => {
	const result = updateProfileSchema.safeParse({
		name: "Alice",
	});

	assert.equal(result.success, true);
});

test("updateProfileSchema accepts null name", () => {
	const result = updateProfileSchema.safeParse({
		name: null,
	});

	assert.equal(result.success, true);
});

test("updateProfileSchema rejects empty name", () => {
	const result = updateProfileSchema.safeParse({
		name: "",
	});

	assert.equal(result.success, false);
});

test("updateProfileSchema rejects name longer than 50 characters", () => {
	const result = updateProfileSchema.safeParse({
		name: "a".repeat(51),
	});

	assert.equal(result.success, false);
});
