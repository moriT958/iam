// @ts-check
import { defineConfig } from "astro/config";
import rehypeHighlight from "rehype-highlight";

// https://astro.build/config
export default defineConfig({
  markdown: {
    syntaxHighlight: false,
    rehypePlugins: [rehypeHighlight],
  },
});
