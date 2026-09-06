<script setup lang="ts">
import { onMounted, ref } from "vue";
import { getMe, updateMe, type ProfileUser } from "../api/users";

const user = ref<ProfileUser | null>(null);
const name = ref("");
const loading = ref(false);
const saving = ref(false);
const message = ref("");

async function loadProfile() {
	loading.value = true;

	try {
		const result = await getMe();
		user.value = result;
		name.value = result.name ?? "";
	} catch (error) {
		message.value = "加载个人资料失败";
	} finally {
		loading.value = false;
	}
}

async function saveProfile() {
	saving.value = true;
	message.value = "";

	try {
		const result = await updateMe(name.value || null);
		user.value = result;
		message.value = "保存成功";
	} catch (error) {
		message.value = "保存失败";
	} finally {
		saving.value = false;
	}
}

onMounted(loadProfile);
</script>

<template>
	<section>
		<h1>个人资料</h1>

		<p v-if="loading">加载中...</p>

		<form v-else-if="user" @submit.prevent="saveProfile">
			<p>邮箱：{{ user.email }}</p>

			<label>
				昵称
				<input v-model.trim="name" maxlength="50" />
			</label>

			<button type="submit" :disabled="saving">
				{{ saving ? "保存中..." : "保存" }}
			</button>

			<p v-if="message">{{ message }}</p>
		</form>
	</section>
</template>
