import { defineConfig } from "astro/config";
import clerk from "@clerk/astro";

import cloudflare from "@astrojs/cloudflare";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  integrations: [clerk()],
  adapter: cloudflare({
    platformProxy: {
      enabled: true,
    },
    imageService: "cloudflare",
  }),
  output: "server",
  vite: {
    plugins: [tailwindcss()],
  },
});

