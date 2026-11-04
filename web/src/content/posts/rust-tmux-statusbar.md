---
title: "Rust で自作 tmux ステータスバーを書く"
date: "2026-04-24"
tags: ["rust", "tmux"]
description: "シェルスクリプトの限界を感じたので Rust に書き換え。tokio + watch チャネルで省電力な常駐型に。"
draft: true
---

## なぜ Rust か

tmux の `status-interval` でシェルスクリプトを定期実行する構成は、CPU・メモリを毎秒消費する。常駐プロセスにして tmux にソケット越しに通知する構成に変えたかった。

## アーキテクチャ

```
[daemon]  watch チャネルで状態変化を検知
   ↓ Unix socket
[tmux]   status-left / status-right に表示
```

tokio の `watch::channel` を使い、バッテリー・ネットワーク・時刻などを別タスクで監視する。変化があったときだけ tmux にコマンドを送る。

```rust
let (tx, mut rx) = tokio::sync::watch::channel(StatusState::default());

tokio::spawn(async move {
    loop {
        let state = collect_state().await;
        tx.send(state).ok();
        tokio::time::sleep(Duration::from_secs(5)).await;
    }
});
```

## 結果

- CPU 使用率: 常時 ~0% (以前は毎秒スパイク)
- 起動からステータス表示まで: 数ミリ秒
