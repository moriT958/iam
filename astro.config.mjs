// @ts-check
import { defineConfig } from "astro/config";
import { unified } from "@astrojs/markdown-remark";
import rehypeHighlight from "rehype-highlight";
import { visit } from "unist-util-visit";

function rehypeLazyImages() {
  /** @param {any} tree */
  return (tree) => {
    visit(tree, "element", (node) => {
      if (node.tagName === "img") {
        node.properties.loading ??= "lazy";
        node.properties.decoding ??= "async";
      }
    });
  };
}

// https://astro.build/config
export default defineConfig({
  site: "https://morit958.com",
  markdown: {
    syntaxHighlight: false,
    processor: unified({
      rehypePlugins: [rehypeHighlight, rehypeLazyImages],
    }),
  },
});
