import { defineConfig } from "vite";

export default defineConfig({
  server: {
    proxy: {
      "/upload": {
        target: "http://127.0.0.1:8888",
        changeOrigin: true,
      },
    },
  },
});
