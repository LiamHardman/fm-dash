import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { quasar, transformAssetUrls } from "@quasar/vite-plugin";
import path from "path";

export default defineConfig({
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  plugins: [
    vue({
      template: { transformAssetUrls },
    }),
    quasar({
      // This tells the Quasar plugin to use your custom SASS variables file.
      // The plugin handles merging these with Quasar's defaults and making them
      // globally available for Quasar components and potentially other SCSS processing.
      sassVariables: "@/quasar-variables.scss",
    }),
  ],
  css: {
    preprocessorOptions: {
      scss: {
        // Prepend your custom Quasar variables to all Vue component style blocks.
        // This ensures your theme overrides ($primary, etc.) are available in components.
        // Quasar's default variables (like $grey-4) should be made globally available
        // by the quasar() plugin itself.
        // We will import app.scss directly in main.js instead of here.
        additionalData: `@import "@/quasar-variables.scss";\n`,
      },
    },
  },
  server: {
    port: 3000,
    proxy: {
      "/upload": {
        target: "http://localhost:8091",
        changeOrigin: true,
      },
    },
  },
});
