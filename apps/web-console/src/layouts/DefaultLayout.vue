<script setup lang="ts">
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";

const router = useRouter();
const authStore = useAuthStore();

async function handleLogout() {
	await authStore.logout();
	await router.replace("/login");
}
</script>

<template>
	<div class="layout">
		<aside>
			<RouterLink to="/keys">API Key</RouterLink>
			<RouterLink to="/profile">个人资料</RouterLink>
			<button type="button" @click="handleLogout">退出登录</button>
		</aside>

		<main>
			<RouterView />
		</main>
	</div>
</template>
