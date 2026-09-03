import { Router } from "express";
import { requireAuth } from "../middleware/auth.middleware";
import { getMe, updateMe } from "../controllers/user.controller";

const userRouter: Router = Router();
userRouter.get("/me", requireAuth, getMe);
userRouter.patch("/me", requireAuth, updateMe);

export default userRouter;
