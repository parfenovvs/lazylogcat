import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "../internal/web/static",
    emptyOutDir: true,
  },
  server: {
    proxy: {
      "/api": "http://localhost:8321",
      "/ws": { target: "http://localhost:8321", ws: true },
    },
  },
});
