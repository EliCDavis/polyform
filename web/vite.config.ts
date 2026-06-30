import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

const webRoot = path.resolve(__dirname);

export default defineConfig({
  root: webRoot,
  plugins: [react()],
  base: "./",
  resolve: {
    alias: {
      "@": path.resolve(webRoot, "./src"),
    },
  },
  build: {
    outDir: path.resolve(webRoot, "../generator/edit/html/js"),
    emptyOutDir: true,
    sourcemap: true,
    cssCodeSplit: false,
    rollupOptions: {
      input: path.resolve(webRoot, "src/main.tsx"),
      output: {
        entryFileNames: "index.js",
        inlineDynamicImports: true,
        assetFileNames: "assets/[name][extname]",
      },
    },
  },
  server: {
    headers: {
      "Cross-Origin-Opener-Policy": "same-origin",
      "Cross-Origin-Embedder-Policy": "require-corp",
    },
    proxy: {
      "/schema": { target: "http://localhost:8080", changeOrigin: true },
      "/graph": { target: "http://localhost:8080", changeOrigin: true },
      "/node": { target: "http://localhost:8080", changeOrigin: true },
      "/node-types": { target: "http://localhost:8080", changeOrigin: true },
      "/started": { target: "http://localhost:8080", changeOrigin: true },
      "/parameter": { target: "http://localhost:8080", changeOrigin: true },
      "/variable": { target: "http://localhost:8080", changeOrigin: true },
      "/profile": { target: "http://localhost:8080", changeOrigin: true },
      "/producer": { target: "http://localhost:8080", changeOrigin: true },
      "/manifest": { target: "http://localhost:8080", changeOrigin: true },
      "/zip": { target: "http://localhost:8080", changeOrigin: true },
      "/swagger": { target: "http://localhost:8080", changeOrigin: true },
      "/mermaid": { target: "http://localhost:8080", changeOrigin: true },
      "/new-graph": { target: "http://localhost:8080", changeOrigin: true },
      "/load-example": { target: "http://localhost:8080", changeOrigin: true },
      "/live": { target: "ws://localhost:8080", ws: true },
    },
  },
});
