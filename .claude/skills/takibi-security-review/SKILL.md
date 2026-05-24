---
name: takibi-security-review
description: Security review checklist for the takibi Go web framework. Use this skill whenever reviewing a PR, auditing a change, or assessing security of new code in the takibi project. Trigger on: "security review", "セキュリティレビュー", "このPRは安全？", "脆弱性", "vulnerability", "is this safe", reviewing changes to cookie/, middlewares/, context.go, thttp/, router/, or any HTTP-handling code. Always apply this when the security-review skill or /security-review is invoked.
---

# takibi Security Review

takibi は OSS のウェブフレームワーク — ここのセキュリティバグは全ての downstream ユーザーに影響する。変更のブラスト半径を意識してレビューする。

実行環境が2つある: **native Go** (標準 net/http) と **WASM on Cloudflare Workers** (syumai/workers 経由)。OS レベルのエントロピー・ファイルシステム・ブロッキング syscall に依存するパターンは WASM で動作が変わるため必ずフラグを立てる。

---

## OWASP Top 10 クイックリファレンス

| OWASP 2021 | takibi チェック箇所 |
|---|---|
| A01 Broken Access Control | §1 アクセス制御 |
| A02 Cryptographic Failures | §3 Cookie署名・Secure属性 |
| A03 Injection | §5 入力・ヘッダー・テンプレートインジェクション |
| A04 Insecure Design | §6 新APIのデザインレビュー |
| A05 Security Misconfiguration | §2 CORS設定、§3 Cookieデフォルト |
| A06 Vulnerable & Outdated Components | §10 依存関係 |
| A07 Identification & Auth Failures | §3 Cookieセッション、§1 アクセス制御 |
| A08 Software & Data Integrity | §10 go.sum・サプライチェーン |
| A09 Security Logging & Monitoring | §8 エラー・ログ漏洩 |
| A10 SSRF | §7 Redirect・外部リクエスト |

---

## Review Checklist

diff に関係するセクションだけ通す。該当ファイルに触れていないセクションはスキップしてよい。

### 1. アクセス制御 — A01

- **ミドルウェアバイパス**: panic や early-return するミドルウェアが後続のセキュリティミドルウェアをスキップしないか確認。`next(c)` を呼ばずに `return nil` するケースは意図的か。
- **ルートレベル認証**: フレームワーク自体はビルトイン認証を持たない。新しいルートグループ API や `Use()` 相当の機能が追加される場合、auth middleware が適用されないデフォルト経路が生まれないか確認。
- **パラメータ改ざん**: `ctx.Param()` / `ctx.ParamBy()` はルーターが設定した値をそのまま返す。ルーター実装が外部入力でパラメータを上書きできる経路を持たないか確認。

### 2. CORS (`middlewares/cors.go`) — A05

- **Wildcard + credentials**: `AllowOrigins: ["*"]` と `AllowCredentials: true` の同時設定は仕様違反で危険。ブラウザは拒否するが misconfigured proxy は通す場合がある。両方セットされるコードパスをフラグ。
- **Origin 反射**: リクエストの `Origin` ヘッダーをホワイトリスト確認なしに `Access-Control-Allow-Origin` へ反射すると全オリジンを許可したのと同義になる。
- **Preflight バイパス**: OPTIONS がハンドラー実行前に処理されているか確認。Preflight がハンドラーに fallthrough すると意図しない副作用が起きる。

### 3. Cookie (`cookie/cookie.go`, `constants/cookie.go`) — A02 / A07

- **デフォルト値の上書き**: `DefaultCookieOptions` は `HttpOnly=true`, `Secure=true`, `SameSite=Strict` を設定している。カスタム `CookieOptions` を渡す際にこれらが省略/false になるとデフォルトが静かに失われる。新しい Cookie API は安全なデフォルトからの明示的なオプトアウトを要求する設計にする。
- **署名 Cookie**: `SetSignedCookie` は `gorilla/securecookie` (HMAC-SHA256) を使う。secret が空・短い・ハードコードされていないか確認。WASM では `securecookie` が `crypto/rand` を呼ぶ — syumai/workers v0.26+ であれば動作するが、それ以前のバージョンはフラグ。
- **Prefix セマンティクス**: `__Secure-` は `Secure=true` が必須、`__Host-` は `Secure=true` + `Path=/` + `Domain=""` が必須。`makeCookieSecure` と `makeCookieHost` がこれを正確に強制しているか確認。

### 4. ルーティング (`router/`) — A01

- **パストラバーサル**: パスパラメータ (`:id`, `*wildcard`) に `../` や URL エンコード版が含まれていても意図しないルートにマッチしないか確認。
- **メソッドオーバーライド**: `X-HTTP-Method-Override` ヘッダーのサポートが隠れていないか確認。GET リクエストが POST/DELETE ハンドラーに到達できる経路があれば HIGH。

### 5. インジェクション — A03

- **ヘッダーインジェクション / レスポンス分割**: `ctx.Response().Header().Set(key, value)` にユーザー入力由来の改行文字 (`\r\n`) が渡されると HTTP レスポンス分割になる。新しいヘッダー操作ヘルパーは値のバリデーションを確認。
- **テンプレートインジェクション (XSS)**: `ctx.Render()` は `templ` コンポーネントを使う。`templ` は Go の型システムで XSS を防ぐが、`templ.Raw()` / unsafe 系の使用は要注意。新しい Render 系 API が生 HTML を直接 `Write` する経路を持たないか確認。
- **Content-Type スニッフィング**: `ctx.Bytes()` は `Content-Type: application/octet-stream` を返す。ブラウザが MIME スニッフィングで HTML と解釈しないよう、`X-Content-Type-Options: nosniff` ヘッダーの設定を推奨。
- **クエリパラメータの信頼境界**: `QueryBy` / `Query` は raw・非バリデートな文字列を返す。SQL・シェル・テンプレートのコンテキストで使われるリスクがある。新しいパラメータ解析ヘルパーを追加するなら信頼境界をドキュメントに明記。

### 6. 新API のデザインレビュー — A04

新しいフレームワーク機能を追加する場合は以下を確認する:

- デフォルトが安全か (secure by default)。ユーザーが何もしなければ安全な動作になるか。
- 危険なオプションは明示的に opt-in が必要か。
- 公開 API のシグネチャが将来の安全強化を妨げないか (例: 後から入力バリデーションを追加できるか)。

### 7. Redirect / SSRF (`context.go`) — A10

- `ctx.Redirect(url)` はバリデーションなしで `http.Redirect` に渡す。`url` がクエリパラメータ・パスパラメータ・リクエストボディ由来の場合はオープンリダイレクト。`ctx.Redirect(ctx.Req().QueryBy(...))` のようなパターンをフラグ。
- フレームワーク内部または examples でアウトバウンド HTTP リクエストを行う場合、ユーザー制御の URL を使う経路は SSRF のリスク。

### 8. エラーハンドリングとログ — A09

- `OnError` コールバックは raw `error` を受け取る。スタックトレース・内部メッセージ・設定値がレスポンスに漏れないか確認。example・docs のコードが `ctx.Text(err.Error())` を使っていないか確認。
- ログに認証トークン・Cookie 値・PII が出力されないか (特に `Bindings` の内容をログに流すパターン)。

### 9. リクエスト解析 (`thttp/request.go`, `context.go`) — A03

- **ボディサイズ制限なし**: `json.NewDecoder(r.request.Body).Decode(dest)` に `io.LimitReader` ガードがない。native モードでは DoS ベクター。WASM/Workers ランタイムは独自制限を持つが、フレームワークがそれに依存するのはリスク。
- **Content-Type チェックの位置**: `ctx.Json()` (レスポンス) が*リクエスト*の Content-Type を確認している — 意図的な設計か、JSON ボディを持たないリクエストからの JSON レスポンスを妨げるバグか確認。

### 10. 依存関係 (`go.mod`, `go.sum`) — A06 / A08

- 新しい直接依存: 既知の CVE を `govulncheck ./...` で確認。ライセンス互換性・メンテナンス状況も確認。
- バージョンダウングレード: 明示的にフラグ — パッチ済みの脆弱性が再導入される可能性がある。
- `go.sum` の整合性: `go.sum` が更新されているか、削除されていないか確認。サプライチェーン攻撃の起点になりうる。

### 11. WASM 固有の懸念

diff が crypto・乱数・I/O に触れる場合に確認:

| パターン | WASM/Workers でのリスク |
|---|---|
| `crypto/rand` | Workers runtime ≥ v0.26 で動作、それ以前は失敗 |
| `os.Getenv` | 空文字列を返す — env vars によるシークレット注入が静かに失敗する |
| `time.Sleep` / blocking I/O | WASM イベントループのデッドロックの原因になりうる |
| `net.Listen` / raw TCP | 利用不可; syumai/workers ブリッジ経由の `net/http` のみ |

---

## Report Format

```
## Security Review: <PR タイトル / 変更内容>

### Findings

| Severity | OWASP | Area | Finding | Recommendation |
|---|---|---|---|---|
| HIGH | A05 | CORS | AllowOrigins=* + AllowCredentials=true | 同時設定を禁止 |
| MEDIUM | A07 | Cookie | カスタム opts で HttpOnly が落ちる | 明示的 opt-out を要求 |
| LOW | A10 | Redirect | URL 未検証 | 信頼境界をドキュメント化 |
| INFO | - | WASM | crypto/rand の依存バージョン | syumai/workers ≥ v0.26 を確認 |

### 問題なし
- [ ] パストラバーサル / ルーティング
- [ ] エラーメッセージ漏洩
- [ ] 署名 Cookie の暗号実装

### Summary
<1–2 文でリスクの全体像とマージの可否>
```

Severity: **HIGH** (悪用可能・データ漏洩), **MEDIUM** (安全でないデフォルト), **LOW** (多層防御), **INFO** (情報提供のみ)

findings がゼロの場合も「問題なし」と明示する — 無言はレビュー未完了と区別できない。
