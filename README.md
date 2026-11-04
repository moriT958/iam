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

| 名前           | 既定値 | 内容                                      |
| -------------- | ------ | ----------------------------------------- |
| `PORT`         | `8080` | app server の待ち受けポート               |
| `IAM_DATA_DIR` | `data` | SQLite (`iam.db`) を置くディレクトリ      |
| `IAM_API_URL`  | -      | `web:dev` が `/api` を転送する先 (開発用) |

## いいね機能

記事ページの「いいね」ボタンと、`/posts` の件数表示。件数は SQLite に保存する。

| メソッド | パス                | 内容                                     |
| -------- | ------------------- | ---------------------------------------- |
| `GET`    | `/api/likes`        | 全記事の件数 (`{"counts": {slug: n}}`)   |
| `GET`    | `/api/likes/{slug}` | 1 記事の件数 (`{"count": n}`)            |
| `POST`   | `/api/likes/{slug}` | いいねを 1 つ増やす                      |
| `DELETE` | `/api/likes/{slug}` | いいねを 1 つ減らす (0 未満にはならない) |

- 匿名なので重複は防げない。押した状態はブラウザの localStorage に持ち、サーバー側は IP ごとのレート制限 (10 回まで、以降 3 秒に 1 回) で抑える。
- 受け付ける slug は、ビルド時に Astro が出す `posts.json` に載っている記事だけ。
- 応答は `Cache-Control: no-store`。Cloudflare のキャッシュルールがオリジンの指定に従うため。

フロントエンドを開発するときは、API のために app server も起動しておく。
`astro dev` が `/api` を `http://localhost:8080` に転送する。

```sh
mise run web:build   # 初回のみ (Go が web/dist を埋め込むため)
mise run app:run     # ターミナル 1: API (http://localhost:8080)
mise run web:dev     # ターミナル 2: Astro (http://localhost:4321)
```
