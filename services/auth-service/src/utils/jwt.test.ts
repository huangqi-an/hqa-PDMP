import assert from "node:assert/strict";
import test from "node:test";

process.env.JWT_ACCESS_SECRET = "test-access-secret";
process.env.JWT_REFRESH_SECRET = "test-refresh-secret";
process.env.JWT_ACCESS_EXPIRES_IN = "15m";
process.env.JWT_REFRESH_EXPIRES_IN = "7d";

const jwt = await import("./jwt");

test("generateAccessToken returns a valid access token", () => {
	const token = jwt.generateAccessToken({
		sub: "user-123",
		email: "test@example.com",
	});

	assert.equal(typeof token, "string");
	assert.ok(token.length > 0);

	const payload = jwt.verifyAccessToken(token);

	assert.ok(payload);
	assert.equal(payload.sub, "user-123");
	assert.equal(payload.email, "test@example.com");
	assert.equal(payload.type, "access");
});

test("generateRefreshToken returns a valid refresh token", () => {
	const token = jwt.generateRefreshToken({
		sub: "user-123",
		email: "test@example.com",
	});

	assert.equal(typeof token, "string");
	assert.ok(token.length > 0);

	const payload = jwt.verifyRefreshToken(token);

	assert.ok(payload);
	assert.equal(payload.sub, "user-123");
	assert.equal(payload.email, "test@example.com");
	assert.equal(payload.type, "refresh");
});

test("access token is not valid as refresh token", () => {
	const accessToken = jwt.generateAccessToken({
		sub: "user-123",
		email: "test@example.com",
	});

	const payload = jwt.verifyRefreshToken(accessToken);

	assert.equal(payload, null);
});

test("refresh token is not valid as access token", () => {
	const refreshToken = jwt.generateRefreshToken({
		sub: "user-123",
		email: "test@example.com",
	});

	const payload = jwt.verifyAccessToken(refreshToken);

	assert.equal(payload, null);
});

test("hashToken returns SHA-256 hex", () => {
	const hash = jwt.hashToken("refresh-token-value");

	assert.equal(typeof hash, "string");
	assert.equal(hash.length, 64);
});
