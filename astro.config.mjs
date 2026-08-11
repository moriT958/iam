// @ts-check
import { defineConfig, fontProviders } from "astro/config";
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
  integrations: [],
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
  vite: {
    // Vite 8's default CSS minifier (Lightning CSS) drops the unprefixed
    // `backdrop-filter` declaration when no browser targets are configured,
    // keeping only `-webkit-backdrop-filter`. Pin back to esbuild so both
    // declarations survive minification.
    //
    // TODO: lightningcss 側の未修正バグと思われる。
    // parcel-bundler/lightningcss#695, 修正 PR #1259 は未マージ (2026-08-11 時点)
    // upstream で修正されたら cssMinify の固定を外す。
    build: {
      cssMinify: "esbuild",
    },
  },
});
