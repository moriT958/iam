import type { AstroIntegration } from "astro";
import { execFileSync } from "node:child_process";
import { readdir, readFile, copyFile } from "node:fs/promises";
import { join, extname, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = join(dirname(fileURLToPath(import.meta.url)), "..");
const FULL_FONT = join(ROOT, "src/assets/fonts/plemol-jp-full.woff2");
const OUT_FONT = join(ROOT, "src/assets/fonts/plemol-jp.woff2");

const SCAN_EXTS = new Set([".md", ".astro", ".ts", ".tsx"]);
const SKIP_DIRS = new Set(["node_modules", ".astro", "dist", ".git"]);

async function collectUnicodes(dir: string): Promise<string> {
  const codepoints = new Set<number>();

  async function scan(current: string) {
    const entries = await readdir(current, { withFileTypes: true });
    for (const entry of entries) {
      if (SKIP_DIRS.has(entry.name)) continue;
      const fullPath = join(current, entry.name);
      if (entry.isDirectory()) {
        await scan(fullPath);
      } else if (SCAN_EXTS.has(extname(entry.name))) {
        const text = await readFile(fullPath, "utf-8");
        for (const ch of text) {
          const cp = ch.codePointAt(0);
          if (cp !== undefined) codepoints.add(cp);
        }
      }
    }
  }

  await scan(dir);

  return [...codepoints]
    .sort((a, b) => a - b)
    .map((cp) => `U+${cp.toString(16).toUpperCase().padStart(4, "0")}`)
    .join(",");
}

async function run(logger: {
  info(msg: string): void;
  warn(msg: string): void;
}) {
  logger.info("font-subset: scanning source files...");
  const unicodes = await collectUnicodes(join(ROOT, "src"));

  try {
    execFileSync(
      "pyftsubset",
      [
        FULL_FONT,
        `--output-file=${OUT_FONT}`,
        "--flavor=woff2",
        `--unicodes=${unicodes}`,
      ],
      { stdio: "pipe" },
    );
    logger.info("font-subset: subset generated.");
  } catch {
    logger.warn(
      "font-subset: pyftsubset not found. copying full font as fallback.",
    );
    await copyFile(FULL_FONT, OUT_FONT);
  }
}

export function fontSubset(): AstroIntegration {
  return {
    name: "font-subset",
    hooks: {
      "astro:build:start": async ({ logger }) => {
        await run(logger);
      },
      "astro:server:start": async ({ logger }) => {
        await run(logger);
      },
    },
  };
}
