import express, { type Express } from "express";

const app: Express = express();

app.use(express.json());

app.get("/health", (req, res) => {
	return res.json({ status: "OK" });
});

export default app;
