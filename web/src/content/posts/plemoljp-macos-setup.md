---
title: "PlemolJP を macOS で快適に使う設定"
date: "2026-04-03"
tags: ["font", "setup"]
description: "Console NF を kitty / Ghostty / VS Code で揃える。Nerd Font グリフが効くフォールバック順の指定方法。"
draft: true
---

## PlemolJP Console NF とは

PlemolJP は IBM Plex Mono と源ノ角ゴシックを合成した等幅フォント。Console NF は Nerd Fonts グリフを内包したバリアント。

## インストール

```bash
brew tap homebrew/cask-fonts
brew install font-plemol-jp-nf
```

## kitty の設定

```ini
# ~/.config/kitty/kitty.conf
font_family      PlemolJP Console NF
bold_font        auto
italic_font      auto
font_size        14.0
```

## Ghostty の設定

```toml
# ~/.config/ghostty/config
font-family = "PlemolJP Console NF"
font-size = 14
```

## VS Code の設定

```json
{
  "editor.fontFamily": "'PlemolJP Console NF', 'JetBrains Mono', monospace",
  "editor.fontLigatures": "'calt', 'liga'",
  "terminal.integrated.fontFamily": "'PlemolJP Console NF'"
}
```

## Nerd Font グリフのフォールバック

フォールバック順は `PlemolJP Console NF` を最初に置くことで Nerd Fonts の Private Use Area（U+E000–U+F8FF）が正しく表示される。
