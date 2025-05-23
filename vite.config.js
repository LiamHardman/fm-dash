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
  // CSS preprocessing is handled by the quasar plugin's sassVariables option
  // No need for additionalData since we're using sassVariables in the quasar plugin config
  
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
