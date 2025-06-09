# OtelLab - OpenTelemetry & Jaeger 実験プロジェクト

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-20.10+-blue.svg)](https://docker.com)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-1.21+-orange.svg)](https://opentelemetry.io)
[![Jaeger](https://img.shields.io/badge/Jaeger-Latest-yellow.svg)](https://jaegertracing.io)

OtelLabは、OpenTelemetryとJaegerを使った分散トレーシングの実装を学習するためのマイクロサービスプロジェクトです。タスク管理システムを例に、実際の分散システムでのトレーシング実装パターンを体験できます。

## 🎯 学習目標

- **分散トレーシングの基礎**: Context Propagation、Span管理、サンプリング戦略
- **マイクロサービスパターン**: サービス間通信、障害の伝播と隔離
- **観測性エンジニアリング**: ゴールデンシグナル、SLI/SLO設定
- **Go言語の実践的活用**: コンテキスト管理、並行処理、エラーハンドリング

## 🏗️ アーキテクチャ

```
┌─────────────────────────────────────────────────────────────┐
│                        クライアント                          │
└───────────────────────────┬─────────────────────────────────┘
                            │ HTTP
                            ▼
┌───────────────────────────────────────────────────────────┐
│                   API Gateway (:8080)                      │
│  - ルーティング・認証・トレースの開始点                    │
└─────────────┬──────────────────────┬──────────────────────┘
              │ gRPC                 │ gRPC
              ▼                      ▼
┌─────────────────────────┐  ┌─────────────────────────┐
│   Task Service (:8081)  │  │   User Service (:8082)  │
│  - タスク管理           │  │  - ユーザー管理         │
│  - PostgreSQL連携       │  │  - インメモリストレージ │
└─────────────────────────┘  └─────────────────────────┘
```

### 使用技術

- **言語**: Go 1.21+
- **Webフレームワーク**: Gorilla Mux (API Gateway), gRPC (サービス間通信)
- **データベース**: PostgreSQL 15 (Task Service), インメモリ (User Service)
- **観測性**: OpenTelemetry + Jaeger
- **インフラ**: Docker Compose

## 🚀 クイックスタート

### 前提条件

- Go 1.21+
- Docker & Docker Compose
- Make (推奨)

### セットアップ

```bash
# 1. リポジトリクローン
git clone https://github.com/bonyuta0204/otel-lab.git
cd otel-lab

# 2. 依存関係インストール & サービス起動（ワンコマンド！）
make quick-start
```

これだけで全てのサービスが起動します！

### 確認

```bash
# ヘルスチェック
make health

# Jaeger UI を開く
make jaeger
# または直接: http://localhost:16686

# デモデータを作成
make demo
```

## 📚 API エンドポイント

### タスク管理

```bash
# タスク作成
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"新しいタスク","description":"説明","assignee_id":"user-001"}'

# タスク一覧取得
curl http://localhost:8080/api/tasks

# タスク詳細取得
curl http://localhost:8080/api/tasks/{task_id}

# タスク更新
curl -X PUT http://localhost:8080/api/tasks/{task_id} \
  -H "Content-Type: application/json" \
  -d '{"title":"更新されたタスク","status":"IN_PROGRESS"}'

# タスク削除
curl -X DELETE http://localhost:8080/api/tasks/{task_id}
```

### ユーザー管理

```bash
# ユーザー作成
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"山田太郎","email":"yamada@example.com"}'

# ユーザー一覧取得
curl http://localhost:8080/api/users

# ユーザー詳細取得
curl http://localhost:8080/api/users/{user_id}
```

## 🔍 トレーシングの確認

### Jaeger UIでのトレース分析

1. **Jaeger UI** にアクセス: http://localhost:16686
2. **Service** から `api-gateway` を選択
3. **Find Traces** をクリック

### 主要な観測ポイント

- **処理時間**: リクエスト全体とサービス間の処理時間
- **エラー追跡**: エラーの発生場所と伝播経路
- **依存関係**: サービス間の呼び出し関係
- **ボトルネック**: 遅い処理の特定

### サンプルトレース

タスク作成のトレースでは以下のような流れを確認できます：

```
api-gateway/CreateTask
├── task-service/CreateTask
│   ├── TaskRepository.CreateTask
│   │   └── PostgreSQL INSERT
│   └── user-service/GetUser (ユーザー検証)
│       └── cache.Get
└── Response
```

## 🛠️ 開発コマンド

```bash
# ヘルプ表示
make help

# 開発用コマンド
make build          # ビルド
make test           # テスト実行
make lint           # コード品質チェック
make format         # コードフォーマット

# Docker操作
make up             # サービス起動
make down           # サービス停止
make logs           # ログ表示
make restart        # サービス再起動

# protobuf生成
make proto          # gRPCコード生成

# データベース
make migrate        # マイグレーション実行
```

## 📈 学習のステップ

### Step 1: 基本的なトレースの確認

1. シンプルなGETリクエストを送信
2. Jaeger UIでトレースを確認
3. スパンの階層構造を理解

### Step 2: エラーケースの分析

```bash
# 存在しないタスクを取得（404エラー）
curl http://localhost:8080/api/tasks/nonexistent

# 不正なリクエストボディ（400エラー）
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"invalid": "data"}'
```

### Step 3: パフォーマンス分析

```bash
# 大量のタスクを作成
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/tasks \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"タスク$i\",\"assignee_id\":\"user-001\"}"
done

# リスト取得のパフォーマンスを確認
curl "http://localhost:8080/api/tasks?page_size=5&page_number=1"
```

### Step 4: 分散トランザクションの理解

タスク作成時のユーザー検証プロセスで、複数サービス間でのデータ整合性とエラーハンドリングを学習。

## 🔧 カスタマイズ

### 新しいサービスの追加

1. `proto/` ディレクトリにProtobufスキーマを追加
2. `make proto` でGoコードを生成
3. 新しいサービスディレクトリを作成
4. `docker-compose.yml` にサービスを追加

### トレース属性の追加

```go
span.SetAttributes(
    attribute.String("custom.attribute", "value"),
    attribute.Int("custom.count", 42),
)
```

### サンプリング戦略の変更

`tracing/tracer.go` の `WithSampler` を変更：

```go
// 50%サンプリング
sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.5))

// 条件付きサンプリング
sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample()))
```

## 🐛 トラブルシューティング

### サービスが起動しない

```bash
# ポートの確認
netstat -tulpn | grep -E ':(8080|8081|8082|5432|16686)'

# Docker コンテナの状態確認
docker-compose ps

# ログの確認
docker-compose logs -f [service-name]
```

### Jaegerでトレースが表示されない

1. **サービス名の確認**: 正しいサービス名でフィルタしているか
2. **時間範囲の確認**: 適切な時間範囲を選択しているか
3. **エンドポイント設定**: 環境変数 `JAEGER_ENDPOINT` が正しく設定されているか

### PostgreSQLに接続できない

```bash
# PostgreSQLコンテナの状態確認
docker-compose exec postgres pg_isready -U otellab

# 手動接続テスト
docker-compose exec postgres psql -U otellab -d taskdb
```

## 📖 参考資料

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Go gRPC Tutorial](https://grpc.io/docs/languages/go/)
- [分散トレーシング入門](https://www.oreilly.com/library/view/distributed-tracing-in/9781492056621/)

## 🤝 コントリビューション

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## 📄 ライセンス

このプロジェクトはMITライセンスの下で公開されています。詳細は [LICENSE](LICENSE) ファイルを参照してください。

---

**🎉 Happy Tracing!** 

分散システムの世界へようこそ！質問やフィードバックがあれば、Issueを作成してください。