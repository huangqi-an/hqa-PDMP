import { Router } from "express";
import { requireAuth } from "../middleware/auth.middleware";
import { getMe } from "../controllers/user.controller";

const userRouter: Router = Router();
userRouter.get("/me", requireAuth, getMe);

export default userRouter;
