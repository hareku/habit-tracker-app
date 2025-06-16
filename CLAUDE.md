# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 開発コマンド

### 環境セットアップ
```bash
# 初期設定（CSRF鍵生成、DynamoDB起動、テーブル作成）
make init

# 開発サーバー起動（localhost:3000）
make dev
```

### テスト・品質チェック
```bash
# 全テスト実行
go test -timeout 1m ./...

# リント実行（CI環境のみ：golangci-lint）
# 脆弱性チェック実行（CI環境のみ：govulncheck）
```

### ビルド・デプロイ
```bash
# ビルド
make build

# デプロイ
make deploy
```

## プロジェクトアーキテクチャ

### Clean Architecture設計
- `/cmd/lambda/` - アプリケーションエントリーポイント（AWS Lambda）
- `/internal/api/` - HTTPハンドラー・プレゼンテーション層
- `/internal/auth/` - Firebase認証ドメイン
- `/internal/repository/` - データアクセス層（Repository Pattern）
- `/internal/apperrors/` - アプリケーションエラー定義
- `/internal/applog/` - ログ設定

### 主要技術スタック
- **Webフレームワーク**: Chi router (`github.com/go-chi/chi/v5`)
- **認証**: Firebase Authentication（Session Cookie方式、14日間有効）
- **データベース**: AWS DynamoDB（複合主キー：PK + SK）
- **デプロイメント**: AWS SAM（Serverless Application Model）
- **テスト**: cupaloy（スナップショット）, go.uber.org/mock

### DynamoDB設計
- **ハビット**: `PK: USER#{userID}`, `SK: HABITS#{habitID}`
- **チェック**: `PK: USER#{userID}`, `SK: CHECKS#{habitID}#{date}`
- **LSI**: `CheckDateLSI`（チェック日付での検索用）

### 認証フロー
1. フロントエンド: Firebase ID Token取得
2. `/session-cookie`エンドポイント: Session Cookieに変換
3. 以降のリクエスト: Session Cookie検証

### エラーハンドリング
- `apperrors.ErrNotFound` → 404 Not Found
- `apperrors.ErrConflict` → 409 Conflict
- その他 → 500 Internal Server Error

### テスト戦略
- **スナップショットテスト**: HTMLレスポンステスト
- **モックテスト**: 依存関係の分離
- **リポジトリテスト**: `repositorytest`パッケージでのテストデータ生成
- **CI/CD**: GitHub Actionsで自動実行（テスト、リント、脆弱性チェック）

## 開発時の注意事項

### セキュリティ
- 全エンドポイントでCSRF保護有効
- Firebase認証による認可
- 本番環境ではセキュアクッキー設定

### 設定ファイル
- `template.yaml` - AWS SAM設定
- `dynamoconf/table.json` - DynamoDB テーブル定義
- `local-env.json` - ローカル環境変数

### 依存関係の注入
- `internal/api/dependency.go`でインターフェース定義
- テスト時はモック実装を使用
- 本番時は実際のDynamoDB実装を使用
