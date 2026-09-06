import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "../stores/auth";

const router = createRouter({
	history: createWebHistory(),
	routes: [
		{
			path: "/",
			component: () => import("../layouts/DefaultLayout.vue"),
			redirect: "/keys",
			children: [
				{
					path: "keys",
					name: "keys",
					component: () => import("../views/KeysView.vue"),
				},
				{
					path: "profile",
					name: "profile",
					component: () => import("../views/ProfileView.vue"),
				},
			],
		},
		{
			path: "/login",
			name: "login",
			component: () => import("../views/LoginView.vue"),
			meta: { public: true },
		},
		{
			path: "/register",
			name: "register",
			component: () => import("../views/RegisterView.vue"),
			meta: { public: true },
		},
	],
});

router.beforeEach((to) => {
	const authStore = useAuthStore();

	if (to.meta.public) {
		return true;
	}

	if (!authStore.isAuthenticated) {
		return {
			name: "login",
			query: { redirect: to.fullPath },
		};
	}

	return true;
});

export default router;
