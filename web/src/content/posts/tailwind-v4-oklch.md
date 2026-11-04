---
title: "Tailwind v4 で oklch を使い倒す"
date: "2026-05-11"
tags: ["css", "frontend"]
description: "v4 から CSS-first の設定になり、color-mix と oklch でテーマシステムを書くと驚くほど短くなる。実例をいくつか。"
draft: true
---

## v4 の設定ファイルが CSS になった

Tailwind v4 は `tailwind.config.js` が廃止され、CSS ファイル内の `@theme` ブロックで設定する。

```css
@import "tailwindcss";

@theme {
  --color-primary: oklch(55% 0.22 280);
  --color-primary-soft: oklch(55% 0.22 280 / 15%);
}
```

## oklch のメリット

oklch は人間の知覚に合わせた色空間で、明度 `L`・彩度 `C`・色相 `H` の3軸を持つ。HSL と違い、同じ `L` の値なら異なる色相でも明度が揃って見える。

```css
/* 同じ明度・彩度で色相だけ変えたパレット */
--color-red: oklch(55% 0.22 25);
--color-green: oklch(55% 0.22 145);
--color-blue: oklch(55% 0.22 260);
--color-purple: oklch(55% 0.22 300);
```

## color-mix との組み合わせ

`color-mix` を使うとホバー時の色を CSS だけで計算できる。

```css
.btn:hover {
  background: color-mix(in oklch, var(--color-primary) 85%, black);
}
```
