iam [Morita](https://morit958.com).

Astro で静的サイトをビルドし、Go のサーバーに埋め込んだシングルバイナリとして配信する。

```
.
├─ cmd/iam/       エントリポイント
├─ internal/      Go のアプリケーションコード (静的配信、ミドルウェア)
├─ web/           Astro のプロジェクト (記事は web/src/content/posts)
│  └─ embed.go    web/dist を go:embed する Go パッケージ
├─ mise.toml      ランタイムのバージョンとタスク
└─ Dockerfile
```

## 開発

ランタイム (node / pnpm / go) とタスクは [mise](https://mise.jdx.dev) で管理している。

```sh
mise install            # ランタイムを入れる
mise tasks              # タスク一覧
```

| タスク                  | 内容                                                                  |
| ----------------------- | --------------------------------------------------------------------- |
| `mise run start`        | web をビルドしてから app を起動する (1 発起動。http://localhost:8080) |
| `mise run web:dev`      | Astro の dev server だけを起動する (http://localhost:4321)            |
| `mise run app:run`      | Go の app server だけを起動する (事前に `web:build` が必要)           |
| `mise run web:build`    | `web/dist` を作る                                                     |
| `mise run build`        | `bin/iam` (シングルバイナリ) を作る                                   |
| `mise run ci`           | CI と同じチェックをすべて実行する                                     |
| `mise run fmt`          | web と Go をフォーマットする                                          |
| `mise run docker:build` | コンテナイメージをビルドする                                          |

Go は `web/dist` を埋め込むので、`go build` / `go vet` / `go test` の前に `mise run web:build` が必要。
タスクは必要な `web:build` を依存として自動で実行する。

環境変数:

| 名前   | 既定値 | 内容                        |
| ------ | ------ | --------------------------- |
| `PORT` | `8080` | app server の待ち受けポート |
