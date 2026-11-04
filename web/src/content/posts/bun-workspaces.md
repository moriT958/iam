---
title: "Bun の Workspace で Monorepo を組む"
date: "2026-05-02"
tags: ["bun", "monorepo"]
description: "Bun 1.2 の workspaces は pnpm を置き換えられるか? 実プロジェクトで2週間運用したログ。"
draft: true
---

## 構成

`package.json` のルートに `workspaces` を定義するだけで使える。

```json
{
  "name": "my-monorepo",
  "workspaces": ["packages/*", "apps/*"]
}
```

## pnpm との比較

| 機能                    | pnpm       | Bun        |
| ----------------------- | ---------- | ---------- |
| インストール速度        | 速い       | さらに速い |
| `workspace:` プロトコル | ○          | ○          |
| `--filter`              | ○          | ○          |
| `pnpx` 相当             | `pnpm dlx` | `bunx`     |

## 2週間運用した感想

- インストールが体感で pnpm の2〜3倍速い
- `bun run --filter` でパッケージを絞った実行が快適
- Node.js との互換性は概ね問題なし。一部 `node:` プレフィックス必須の API がある
- CI キャッシュは `~/.bun/install/cache` を保存すると効く
