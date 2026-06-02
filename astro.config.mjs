// @ts-check
import { defineConfig, fontProviders } from "astro/config";
import { fontSubset } from "./integrations/font-subset.ts";
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
  integrations: [fontSubset()],
  fonts: [
    {
      provider: fontProviders.local(),
      name: "PlemolJP",
      cssVariable: "--font-plemol",
      options: {
        variants: [
          {
            src: ["./src/assets/fonts/plemol-jp.woff2"],
            weight: "400",
            style: "normal",
          },
        ],
      },
    },
  ],
  site: "https://morit958.com",
  markdown: {
    syntaxHighlight: false,
    processor: unified({
      rehypePlugins: [rehypeHighlight, rehypeLazyImages],
    }),
  },
});
