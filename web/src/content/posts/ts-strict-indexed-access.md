---
title: "tsconfig: noUncheckedIndexedAccess を有効化した"
date: "2026-04-15"
tags: ["typescript"]
description: "後から有効化するときの差分の規模、Record<string, T> のリファクタ指針、配列アクセスの書き換えパターン。"
draft: true
---

## noUncheckedIndexedAccess とは

配列・オブジェクトのインデックスアクセスの戻り値を `T | undefined` に変える設定。

```ts
// before: string
// after:  string | undefined
const first = arr[0];
```

## 有効化したときの差分規模

中規模プロジェクト（TSファイル約200本）で有効化したところ、エラーが約400件出た。主な内訳:

- 配列の先頭・末尾アクセス: 約200件
- `Record<string, T>` のルックアップ: 約150件
- その他: 約50件

## 配列アクセスの書き換えパターン

```ts
// NG
const head = arr[0].name;

// OK: 非nullアサーション（確実に存在するとき）
const head = arr[0]!.name;

// OK: オプショナルチェーン（undefinedを伝播させるとき）
const head = arr[0]?.name;

// OK: 事前チェック
if (arr.length > 0) {
  const head = arr[0].name; // ここでは string
}
```

## Record<string, T> の指針

キーが不定なら `Map<string, T>` への移行も検討する。既存の `Record` を維持する場合は、アクセス後に `?? defaultValue` を付ける形が多い。
