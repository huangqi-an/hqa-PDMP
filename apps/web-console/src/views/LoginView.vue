<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";

const router = useRouter();
const authStore = useAuthStore();

const email = ref("");
const password = ref("");
const loading = ref(false);
const errorMessage = ref("");

async function submit() {
	loading.value = true;
	errorMessage.value = "";

	try {
		await authStore.login(email.value, password.value);

		const redirect =
			(router.currentRoute.value.query.redirect as string) || "/keys";
		await router.replace(redirect);
	} catch (error) {
		errorMessage.value = "登录失败，请检查邮箱和密码";
	} finally {
		loading.value = false;
	}
}
</script>

<template>
	<form @submit.prevent="submit">
		<h1>登录</h1>

		<label>
			邮箱
			<input v-model.trim="email" type="email" required />
		</label>

		<label>
			密码
			<input v-model="password" type="password" required minlength="8" />
		</label>

		<p v-if="errorMessage">{{ errorMessage }}</p>

		<button type="submit" :disabled="loading">
			{{ loading ? "登录中..." : "登录" }}
		</button>

		<RouterLink to="/register">还没有账号？去注册</RouterLink>
	</form>
</template>
