import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";

import KeysView from "./KeysView.vue";
import { listKeys } from "../api/keys";

vi.mock("../api/keys", () => ({
	listKeys: vi.fn(),
	createKey: vi.fn(),
	updateKey: vi.fn(),
	deleteKey: vi.fn(),
	revealKey: vi.fn(),
}));

describe("KeysView", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("renders empty state when no keys exist", async () => {
		vi.mocked(listKeys).mockResolvedValue({
			items: [],
			total: 0,
		});

		const wrapper = mount(KeysView);

		await flushPromises();

		expect(wrapper.text()).toContain("暂无数据");
		expect(wrapper.text()).toContain("共 0 条");
	});

	it("renders keys from api", async () => {
		vi.mocked(listKeys).mockResolvedValue({
			items: [
				{
					id: "key-1",
					provider: "openai",
					name: "default",
					maskedKey: "****abcd",
					keyHint: "abcd",
					notes: null,
					tags: ["test"],
					createdAt: "2026-09-07T00:00:00Z",
					updatedAt: "2026-09-07T00:00:00Z",
				},
			],
			total: 1,
		});

		const wrapper = mount(KeysView);

		await flushPromises();

		expect(wrapper.text()).toContain("default");
		expect(wrapper.text()).toContain("openai");
		expect(wrapper.text()).toContain("****abcd");
	});

	it("shows error message when list request fails", async () => {
		vi.mocked(listKeys).mockRejectedValue(new Error("request failed"));

		const wrapper = mount(KeysView);

		await flushPromises();

		expect(wrapper.text()).toContain("加载密钥列表失败");
	});
});
