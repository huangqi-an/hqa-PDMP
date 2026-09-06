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
		await authStore.register(email.value, password.value);
		await router.replace("/login");
	} catch (error) {
		errorMessage.value = "注册失败，请检查邮箱和密码";
	} finally {
		loading.value = false;
	}
}
</script>

<template>
	<form @submit.prevent="submit">
		<h1>注册</h1>

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
			{{ loading ? "注册中..." : "注册" }}
		</button>

		<RouterLink to="/login">已有账号？去登录</RouterLink>
	</form>
</template>
