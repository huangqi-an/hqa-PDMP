<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import {
	createKey,
	deleteKey,
	listKeys,
	revealKey,
	updateKey,
	type APIKey,
	type UpdateAPIKeyInput,
} from "../api/keys";

const keys = ref<APIKey[]>([]);
const total = ref(0);
const loading = ref(false);
const errorMessage = ref("");
const showCreateForm = ref(false);
const creating = ref(false);
const revealedKeys = ref<Record<string, string>>({});
const editingKeyId = ref<string | null>(null);
const updating = ref(false);

const editForm = reactive({
	id: "",
	provider: "",
	name: "",
	notes: "",
	tags: "",
	key: "",
});

const createForm = reactive({
	provider: "openai",
	name: "",
	key: "",
	notes: "",
	tags: "",
});
const page = ref(1);
const pageSize = ref(20);
const query = ref("");
const providerFilter = ref("");

async function loadKeys() {
	loading.value = true;
	errorMessage.value = "";

	try {
		const result = await listKeys({
			page: page.value,
			pageSize: pageSize.value,
			provider: providerFilter.value || undefined,
			q: query.value || undefined,
		});

		keys.value = result.items;
		total.value = result.total;
	} catch (error) {
		errorMessage.value = "加载密钥列表失败";
	} finally {
		loading.value = false;
	}
}

async function handleCreate() {
	creating.value = true;
	errorMessage.value = "";

	try {
		const tags = createForm.tags
			.split(",")
			.map((tag) => tag.trim())
			.filter(Boolean);

		await createKey({
			provider: createForm.provider,
			name: createForm.name,
			key: createForm.key,
			notes: createForm.notes || null,
			tags,
		});

		createForm.provider = "openai";
		createForm.name = "";
		createForm.key = "";
		createForm.notes = "";
		createForm.tags = "";

		showCreateForm.value = false;
		await loadKeys();
	} catch (error) {
		errorMessage.value = "创建密钥失败";
	} finally {
		creating.value = false;
	}
}

async function handleReveal(key: APIKey) {
	try {
		const plaintext = await revealKey(key.id);
		revealedKeys.value[key.id] = plaintext;
	} catch (error) {
		revealedKeys.value[key.id] = "查看失败";
	}
}

async function handleDelete(key: APIKey) {
	const confirmed = window.confirm(`确定删除“${key.name}”吗？`);

	if (!confirmed) {
		return;
	}

	try {
		await deleteKey(key.id);
		await loadKeys();
	} catch (error) {
		errorMessage.value = "删除失败";
	}
}
function startEdit(key: APIKey) {
	editingKeyId.value = key.id;

	editForm.id = key.id;
	editForm.provider = key.provider;
	editForm.name = key.name;
	editForm.notes = key.notes ?? "";
	editForm.tags = key.tags.join(", ");
	editForm.key = "";
}

function cancelEdit() {
	editingKeyId.value = null;
	editForm.id = "";
	editForm.provider = "";
	editForm.name = "";
	editForm.notes = "";
	editForm.tags = "";
	editForm.key = "";
}

async function handleUpdate() {
	updating.value = true;
	errorMessage.value = "";

	try {
		const tags = editForm.tags
			.split(",")
			.map((tag) => tag.trim())
			.filter(Boolean);

		const input: UpdateAPIKeyInput = {
			provider: editForm.provider,
			name: editForm.name,
			notes: editForm.notes || null,
			tags,
		};

		if (editForm.key.trim()) {
			input.key = editForm.key.trim();
		}

		await updateKey(editForm.id, input);

		cancelEdit();
		await loadKeys();
	} catch (error) {
		errorMessage.value = "更新密钥失败";
	} finally {
		updating.value = false;
	}
}

onMounted(loadKeys);
</script>

<template>
	<section>
		<h1>API Key</h1>

		<button type="button" @click="showCreateForm = !showCreateForm">
			{{ showCreateForm ? "收起" : "新建密钥" }}
		</button>

		<form
			@submit.prevent="
				page = 1;
				loadKeys();
			"
		>
			<input v-model.trim="query" placeholder="搜索名称或 provider" />
			<input
				v-model.trim="providerFilter"
				placeholder="provider，例如 openai"
			/>
			<button type="submit">搜索</button>
		</form>
		<form v-if="showCreateForm" @submit.prevent="handleCreate">
			<label>
				Provider
				<input v-model.trim="createForm.provider" required />
			</label>

			<label>
				名称
				<input v-model.trim="createForm.name" required />
			</label>

			<label>
				API Key
				<input v-model.trim="createForm.key" required />
			</label>

			<label>
				备注
				<textarea v-model="createForm.notes" />
			</label>

			<label>
				标签，用逗号分隔
				<input v-model="createForm.tags" placeholder="prod,default" />
			</label>

			<button type="submit" :disabled="creating">
				{{ creating ? "创建中..." : "创建" }}
			</button>
		</form>

		<p v-if="loading">加载中...</p>
		<p v-if="errorMessage">{{ errorMessage }}</p>

		<table v-if="!loading">
			<thead>
				<tr>
					<th>名称</th>
					<th>Provider</th>
					<th>Masked Key</th>
					<th>标签</th>
					<th>操作</th>
				</tr>
			</thead>
			<tbody>
				<tr v-for="key in keys" :key="key.id">
					<td>{{ key.name }}</td>
					<td>{{ key.provider }}</td>
					<td>{{ revealedKeys[key.id] || key.maskedKey }}</td>
					<td>{{ key.tags.join(", ") }}</td>
					<td>
						<button type="button" @click="handleReveal(key)">查看明文</button>
						<button type="button" @click="startEdit(key)">编辑</button>
						<button type="button" @click="handleDelete(key)">删除</button>
					</td>
				</tr>
			</tbody>
		</table>

		<form v-if="editingKeyId" @submit.prevent="handleUpdate">
			<h2>编辑密钥</h2>

			<label>
				Provider
				<input v-model.trim="editForm.provider" required />
			</label>

			<label>
				名称
				<input v-model.trim="editForm.name" required />
			</label>

			<label>
				新 API Key，留空表示不修改
				<input v-model.trim="editForm.key" />
			</label>

			<label>
				备注
				<textarea v-model="editForm.notes" />
			</label>

			<label>
				标签，用逗号分隔
				<input v-model="editForm.tags" />
			</label>

			<button type="submit" :disabled="updating">
				{{ updating ? "保存中..." : "保存" }}
			</button>

			<button type="button" @click="cancelEdit">取消</button>
		</form>

		<p v-if="!loading && keys.length === 0">暂无数据</p>
		<p>共 {{ total }} 条</p>
		<div class="pagination">
			<button
				type="button"
				:disabled="page <= 1"
				@click="
					page -= 1;
					loadKeys();
				"
			>
				上一页
			</button>

			<span>第 {{ page }} 页 / 共 {{ total }} 条</span>

			<button
				type="button"
				:disabled="page * pageSize >= total"
				@click="
					page += 1;
					loadKeys();
				"
			>
				下一页
			</button>
		</div>
	</section>
</template>
