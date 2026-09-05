import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
	history: createWebHistory(),
	routes: [
		{
			path: "/",
			redirect: "/keys",
		},
		{
			path: "/login",
			name: "login",
			component: () => import("../views/LoginView.vue"),
		},
		{
			path: "/register",
			name: "register",
			component: () => import("../views/RegisterView.vue"),
		},
		{
			path: "/keys",
			name: "keys",
			component: () => import("../views/KeysView.vue"),
		},
		{
			path: "/profile",
			name: "profile",
			component: () => import("../views/ProfileView.vue"),
		},
	],
});

export default router;
