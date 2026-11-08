import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import path from "path";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  // @ds alias for the components directory
  resolve: {
    alias: {
      "@components": path.resolve(__dirname, "./src/components"),
      "@images": path.resolve(__dirname, "./src/assets/images"),
      "@constants": path.resolve(__dirname, "./src/constants"),
      "@context": path.resolve(__dirname, "./src/context"),
      "@assets": path.resolve(__dirname, "./src/assets"),
      "@views": path.resolve(__dirname, "./src/views"),
      "@utils": path.resolve(__dirname, "./src/utils"),
      "@ds": path.resolve(__dirname, "./src/ds"),
    },
  },
  server: {
    // bind to IPv4 too — on macOS vite can end up on IPv6 [::1] only, and
    // browsers that resolve localhost to 127.0.0.1 get connection refused
    host: "127.0.0.1",
    proxy: {
      // dev server forwards API calls to the Go backend (go run ./cmd/main/main.go)
      "/api": {
        target: "http://localhost:8016",
        changeOrigin: true,
      },
    },
  },
});
