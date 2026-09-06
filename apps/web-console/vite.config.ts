import vue from "@vitejs/plugin-vue";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
	plugins: [vue()],
	server: {
		port: 5173,
		proxy: {
			"/api/auth": "http://localhost:3001",
			"/api/users": "http://localhost:3001",
			"/api/keys": "http://localhost:8080",
		},
	},
});
