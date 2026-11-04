---
title: "Neovim を lazy.nvim 構成へ移行した話"
date: "2026-05-19"
tags: ["nvim", "dotfiles"]
description: "packer.nvim から lazy.nvim へ移行するときに踏んだ罠と、起動時間を 410ms → 90ms まで削った経緯について書きます。"
draft: true
---

## 背景

長らく packer.nvim を使ってきたが、メンテナンスが止まったのをきっかけに lazy.nvim へ移行することにした。

## 移行手順

まず既存の `~/.config/nvim/lua/plugins/` 以下を lazy.nvim のフォーマットに変換する。packer の `use` を lazy の `{ "author/plugin", ... }` に書き換えるだけなので機械的にできる。

```lua
-- Before (packer)
use {
  "nvim-telescope/telescope.nvim",
  requires = { "nvim-lua/plenary.nvim" },
}

-- After (lazy)
{
  "nvim-telescope/telescope.nvim",
  dependencies = { "nvim-lua/plenary.nvim" },
}
```

## 起動時間の削減

`lazy=true` と `event`、`cmd` による遅延ロードを活用した。`--startuptime` で計測しながら各プラグインの読み込みタイミングを調整した結果、410ms → 90ms まで削れた。

```bash
nvim --startuptime /tmp/startup.log +qa && tail -1 /tmp/startup.log
```

## 踏んだ罠

- `config` 関数内で他プラグインに依存する設定を書くと、ロード順によってエラーになる
- `VeryLazy` イベントは UI 描画後に発火するため、カラースキームには使えない
