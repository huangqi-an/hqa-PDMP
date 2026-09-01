import express, { type Express } from "express";
import { errorHandler } from "./middleware/error.middleware";
import registerRouter from "./routes/auth.routes";

const app: Express = express();

app.use(express.json());

app.use("/api/auth", registerRouter);

app.use(errorHandler);

export default app;
