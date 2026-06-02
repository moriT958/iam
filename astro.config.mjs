// @ts-check
import { defineConfig } from "astro/config";
import { unified } from "@astrojs/markdown-remark";
import rehypeHighlight from "rehype-highlight";

// https://astro.build/config
export default defineConfig({
  markdown: {
    syntaxHighlight: false,
    processor: unified({
      rehypePlugins: [rehypeHighlight],
    }),
  },
});
