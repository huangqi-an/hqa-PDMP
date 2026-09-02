import jwt, { type JwtPayload, type SignOptions } from "jsonwebtoken";

const accessSecret = process.env.JWT_ACCESS_SECRET!;
const refreshSecret = process.env.JWT_REFRESH_SECRET!;

if (!accessSecret || !refreshSecret) {
	throw new Error("JWT secrets are not configured");
}

const accessExpiresIn =
	(process.env.JWT_ACCESS_EXPIRES_IN as SignOptions["expiresIn"]) ?? "15m";

const refreshExpiresIn =
	(process.env.JWT_REFRESH_EXPIRES_IN as SignOptions["expiresIn"]) ?? "7d";

export type TokenPayload = {
	sub: string;
	email: string;
	type: "access" | "refresh";
};

export function generateAccessToken(payload: Omit<TokenPayload, "type">) {
	return jwt.sign({ ...payload, type: "access" }, accessSecret, {
		expiresIn: accessExpiresIn,
	});
}

export function generateRefreshToken(payload: Omit<TokenPayload, "type">) {
	return jwt.sign({ ...payload, type: "refresh" }, refreshSecret, {
		expiresIn: refreshExpiresIn,
	});
}

export function verifyAccessToken(token: string): TokenPayload | null {
	try {
		const decoded = jwt.verify(token, accessSecret);

		if (typeof decoded === "string") {
			return null;
		}

		const payload = decoded as JwtPayload & Partial<TokenPayload>;

		if (payload.sub && payload.email && payload.type === "access") {
			return {
				sub: payload.sub,
				email: payload.email,
				type: "access",
			};
		}

		return null;
	} catch {
		return null;
	}
}

export function verifyRefreshToken(token: string): TokenPayload | null {
	try {
		const decoded = jwt.verify(token, refreshSecret);

		if (typeof decoded === "string") {
			return null;
		}

		const payload = decoded as JwtPayload & Partial<TokenPayload>;

		if (payload.sub && payload.email && payload.type === "refresh") {
			return {
				sub: payload.sub,
				email: payload.email,
				type: "refresh",
			};
		}

		return null;
	} catch {
		return null;
	}
}
