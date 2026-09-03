import express, { type Express } from "express";
import { errorHandler } from "./middleware/error.middleware";
import registerRouter from "./routes/auth.routes";
import userRouter from "./routes/user.routes";

const app: Express = express();

app.use(express.json());

app.use("/api/auth", registerRouter);
app.use("/api/users", userRouter);

app.use(errorHandler);

export default app;
