import { defineConfig } from "@rsbuild/core";
import { pluginReact } from "@rsbuild/plugin-react";

export default defineConfig({
  plugins: [pluginReact()],
  source: {
    entry: {
      index: "./src/main.tsx"
    }
  },
  output: {
    distPath: {
      root: "frontend/static/app"
    },
    assetPrefix: "/sms/assets/app/",
    cleanDistPath: true
  },
  performance: {
    chunkSplit: {
      strategy: "all-in-one"
    }
  },
  html: {
    title: "SMS 历史记录"
  }
});
