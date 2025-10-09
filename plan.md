  ---
  ライブラリ選定（2025年最新）

  1. OpenAPI関係: oapi-codegen v2

  - 選定理由:
    - Go界で最も人気があり、推奨されているOpenAPIジェネレーター
    - Chi、Echo、Gin、標準net/httpなど主要フレームワークに対応
    - サーバー・クライアント両方の型安全なコード生成
    - バリデーション自動生成
  - インストール: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

  2. マイグレーション: golang-migrate

  - 選定理由:
    - 2025年現在もGoコミュニティで最も広く使われている（v4.18.2）
    - シンプルで学習コストが低い
    - PostgreSQL、MySQL、SQLiteなど幅広いDB対応
    - CLI + Goライブラリ両対応
    - Atlasは強力だが、まずはシンプルなgolang-migrateで十分
  - 代替案: 将来的にAtlasを導入すれば、自動マイグレーション計画やエラー回復が強化される

  3. ORM: sqlc（推奨）

  - 選定理由:
    - 型安全性: SQLを書くと、型安全なGoコードが自動生成される
    - パフォーマンス: database/sqlと同等の速度（GORMより高速）
    - 明示性: SQLが隠蔽されず、複雑なクエリも制御しやすい
    - ゼロランタイムオーバーヘッド: コード生成型なので実行時のリフレクションなし
    - モック生成: テスト用のモックインターフェースも自動生成可能
  - トレードオフ: 動的クエリには不向き（固定SQLのみ）
  - 代替案:
    - sqlx: より柔軟だが型安全性は手動
    - GORM: プロトタイピングには便利だが、パフォーマンスと明示性で劣る

  ---
  /inコマンド実装のディレクトリ構成

  workspace-backend/
  ├── cmd/
  │   └── work-tracker/
  │       └── main.go                    # エントリポイント（DI初期化）
  │
  ├── domain/                            # ドメイン層（ビジネスロジック）
  │   ├── user.go                        # ✅ 既存
  │   ├── tier.go                        # ✅ 既存
  │   ├── session.go                     # Sessionエンティティ ← /in で作成
  │   │   # - NewSession(userID, workName, duration) (*Session, error)
  │   │   # - Extend(minutes int) error
  │   │   # - Complete() error
  │   │   # - IsActive() bool
  │   └── repository/                    # リポジトリインターフェース（依存性逆転）
  │       ├── user_repository.go         # type UserRepository interface { FindByName(...), Save(...) }
  │       └── session_repository.go      # type SessionRepository interface { Create(...), Update(...) }
  │
  ├── usecase/                           # ユースケース層（アプリケーションロジック）
  │   ├── command/
  │   │   ├── join_command.go            # /in コマンドのユースケース ← メイン実装箇所
  │   │   │   # type JoinCommandInput struct { UserName, WorkName string, Tier domain.Tier }
  │   │   │   # func (u *JoinCommandUsecase) Execute(ctx, input) (*JoinCommandOutput, error)
  │   │   │   #   1. UserRepositoryでユーザー取得or作成
  │   │   │   #   2. Sessionエンティティ生成（デフォルト60分）
  │   │   │   #   3. SessionRepository.Create()で永続化
  │   │   │   #   4. WebSocketへ通知（後述）
  │   │   └── join_command_test.go       # モックを使ったユニットテスト
  │   └── usecase.go                     # 共通インターフェース（必要なら）
  │
  ├── presentation/                      # プレゼンテーション層（外部I/F）
  │   ├── http/
  │   │   ├── router.go                  # ルーティング設定（Chi推奨）
  │   │   ├── middleware.go              # 認証・ログミドルウェア
  │   │   ├── handler/
  │   │   │   ├── command_handler.go     # /api/commands/join エンドポイント
  │   │   │   │   # - Twitch-botからPOSTリクエスト受信
  │   │   │   │   # - DTOをUsecaseInputに変換
  │   │   │   │   # - JoinCommandUsecase.Execute()呼び出し
  │   │   │   │   # - レスポンス返却
  │   │   │   └── health_handler.go      # ✅ 既存の /health
  │   │   └── dto/                       # OpenAPIから自動生成 ← oapi-codegen
  │   │       ├── command.gen.go         # type JoinCommandRequest struct { ... }
  │   │       └── server.gen.go          # Serverインターフェース（実装すべきAPI）
  │   └── ws/                            # WebSocket（後のPhaseで実装）
  │       └── hub.go
  │
  ├── infrastructure/                    # インフラ層（外部実装）
  │   ├── database/
  │   │   ├── postgres.go                # *sql.DB接続管理
  │   │   ├── sqlc.yaml                  # sqlc設定ファイル
  │   │   ├── query/                     # SQL定義（sqlcの入力）
  │   │   │   ├── user.sql               # -- name: FindUserByName :one SELECT ...
  │   │   │   └── session.sql            # -- name: CreateSession :one INSERT ...
  │   │   ├── sqlc/                      # sqlc生成コード（自動生成、Git管理）
  │   │   │   ├── db.go                  # DBTX interface, Queries struct
  │   │   │   ├── models.go              # User, Session struct
  │   │   │   ├── user.sql.go            # FindUserByName実装
  │   │   │   └── session.sql.go         # CreateSession実装
  │   │   └── repository/                # リポジトリ実装（domain/repository実装）
  │   │       ├── user_repository_impl.go      # sqlc.Queriesをラップ
  │   │       │   # type userRepositoryImpl struct { queries *sqlc.Queries }
  │   │       │   # func (r *userRepositoryImpl) FindByName(name) (*domain.User, error)
  │   │       │   #   1. r.queries.FindUserByName() → sqlc.User
  │   │       │   #   2. sqlc.User → domain.User に変換
  │   │       └── session_repository_impl.go   # 同様
  │   ├── config/
  │   │   └── config.go                  # 環境変数読込（DATABASE_URL等）
  │   └── logger/
  │       └── logger.go                  # 構造化ログ（slog等）
  │
  ├── shared/                            # 契約定義
  │   └── api/
  │       └── openapi.yaml               # OpenAPI 3.0定義 ← oapi-codegen入力
  │           # paths:
  │           #   /api/commands/join:
  │           #     post:
  │           #       requestBody: { user_name, work_name, tier }
  │           #       responses: { session_id, planned_end }
  │
  ├── migrations/                        # DBマイグレーション
  │   ├── 000001_create_users.up.sql    # CREATE TABLE users ...
  │   ├── 000001_create_users.down.sql  # DROP TABLE users;
  │   ├── 000002_create_sessions.up.sql
  │   └── 000002_create_sessions.down.sql
  │
  ├── scripts/
  │   ├── gen_sqlc.sh                    # sqlc generate 実行
  │   ├── gen_openapi.sh                 # oapi-codegen 実行
  │   └── migrate.sh                     # golang-migrate CLI実行
  │
  ├── tools.go                           # ツール依存管理（go mod tidy用）
  │   # //go:build tools
  │   # import (_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen")
  │
  ├── go.mod
  ├── go.sum
  ├── Makefile                           # make gen, make migrate-up, make test
  └── README.md

  ---
  テスト方針（DI + モック化）

  1. ドメイン層のテスト

```go
  // domain/session_test.go
  func TestSession_NewSession(t *testing.T) {
      session, err := domain.NewSession(1, "論文執筆", 60*time.Minute, time.Now)
      assert.NoError(t, err)
      assert.Equal(t, "論文執筆", session.WorkName)
      assert.True(t, session.IsActive())
  }

```
  2. ユースケース層のテスト（モックリポジトリ使用）

```go
  // usecase/command/join_command_test.go
  import (
      "testing"
      "github.com/stretchr/testify/mock"
  )

  // モック定義（mockeryやtestifyで自動生成可能）
  type MockUserRepository struct {
      mock.Mock
  }

  func (m *MockUserRepository) FindByName(ctx context.Context, name string) (*domain.User, error) {
      args := m.Called(ctx, name)
      if args.Get(0) == nil {
          return nil, args.Error(1)
      }
      return args.Get(0).(*domain.User), args.Error(1)
  }

  // テスト
  func TestJoinCommand_NewUser(t *testing.T) {
      // Arrange
      mockUserRepo := new(MockUserRepository)
      mockSessionRepo := new(MockSessionRepository)

      mockUserRepo.On("FindByName", mock.Anything, "yamada").Return(nil, domain.ErrUserNotFound)
      mockUserRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
      mockSessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

      usecase := NewJoinCommandUsecase(mockUserRepo, mockSessionRepo)

      // Act
      input := JoinCommandInput{UserName: "yamada", WorkName: "コーディング", Tier: domain.Tier1}
      output, err := usecase.Execute(context.Background(), input)

      // Assert
      assert.NoError(t, err)
      assert.NotNil(t, output.SessionID)
      mockUserRepo.AssertExpectations(t)
      mockSessionRepo.AssertExpectations(t)
  }
```

  3. プレゼンテーション層のテスト（HTTPハンドラー）

```go
  // presentation/http/handler/command_handler_test.go
  func TestCommandHandler_JoinCommand(t *testing.T) {
      // モックUsecaseを使用
      mockUsecase := new(MockJoinCommandUsecase)
      handler := NewCommandHandler(mockUsecase)

      // HTTPリクエスト作成
      body := `{"user_name": "yamada", "work_name": "勉強", "tier": 1}`
      req := httptest.NewRequest(http.MethodPost, "/api/commands/join", strings.NewReader(body))
      req.Header.Set("Content-Type", "application/json")
      rec := httptest.NewRecorder()

      // 実行
      handler.JoinCommand(rec, req)

      // 検証
      assert.Equal(t, http.StatusOK, rec.Code)
  }
  ```

  4. リポジトリ層のテスト（統合テスト）

```go
  // infrastructure/database/repository/user_repository_impl_test.go
  func TestUserRepository_Integration(t *testing.T) {
      if testing.Short() {
          t.Skip("skipping integration test")
      }

      // テスト用DBコンテナ起動（testcontainers-go使用）
      ctx := context.Background()
      pgContainer, err := postgres.Run(ctx, "postgres:16", ...)
      require.NoError(t, err)
      defer pgContainer.Terminate(ctx)

      // マイグレーション実行
      db := setupTestDB(t, pgContainer.ConnectionString())
      repo := NewUserRepositoryImpl(sqlc.New(db))

      // テスト実行
      user, err := repo.Save(ctx, &domain.User{Name: "test", Tier: domain.Tier1})
      require.NoError(t, err)

      found, err := repo.FindByName(ctx, "test")
      require.NoError(t, err)
      assert.Equal(t, user.ID, found.ID)
  }
```

  ---
  DI（依存性注入）設計

  cmd/work-tracker/main.go

```go
  package main

  import (
      "database/sql"
      "log"
      "net/http"

      "github.com/go-chi/chi/v5"
      _ "github.com/lib/pq"

      "github.com/yamada-ai/workspace-backend/infrastructure/config"
      "github.com/yamada-ai/workspace-backend/infrastructure/database"
      infraRepo "github.com/yamada-ai/workspace-backend/infrastructure/database/repository"
      "github.com/yamada-ai/workspace-backend/infrastructure/database/sqlc"
      "github.com/yamada-ai/workspace-backend/presentation/http/handler"
      "github.com/yamada-ai/workspace-backend/usecase/command"
  )

  func main() {
      // 設定読込
      cfg := config.Load()

      // DB接続
      db, err := sql.Open("postgres", cfg.DatabaseURL)
      if err != nil {
          log.Fatalf("failed to connect database: %v", err)
      }
      defer db.Close()

      // === 依存性注入（下層→上層） ===

      // 1. sqlc Queries生成（最下層）
      queries := sqlc.New(db)

      // 2. リポジトリ実装（インフラ層）
      userRepo := infraRepo.NewUserRepositoryImpl(queries)
      sessionRepo := infraRepo.NewSessionRepositoryImpl(queries)

      // 3. ユースケース（ユースケース層）
      joinUsecase := command.NewJoinCommandUsecase(userRepo, sessionRepo)

      // 4. ハンドラー（プレゼンテーション層）
      commandHandler := handler.NewCommandHandler(joinUsecase)

      // 5. ルーター設定
      r := chi.NewRouter()
      r.Post("/api/commands/join", commandHandler.JoinCommand)
      r.Get("/health", handler.HealthCheck)

      // サーバー起動
      log.Printf("Server listening on :8000")
      log.Fatal(http.ListenAndServe(":8000", r))
  }
```

  ---
  実装順序（/inコマンドに集中）

  Phase 1: 環境セットアップ

  1. ライブラリインストール（oapi-codegen, golang-migrate, sqlc）
  2. shared/api/openapi.yaml 作成（/api/commands/joinエンドポイン定義）
  3. migrations/ 作成（users, sessionsテーブル）
  4. infrastructure/database/query/ にSQL定義

  Phase 2: ドメイン層

  5. domain/session.go 実装（Sessionエンティティ）
  6. domain/repository/user_repository.go インターフェース定義
  7. domain/repository/session_repository.go インターフェース定義

  Phase 3: インフラ層

  8. sqlc.yaml 設定 + make gen-sqlc でコード生成
  9. infrastructure/database/repository/*_impl.go 実装
  10. 統合テスト作成（testcontainers-go使用）

  Phase 4: ユースケース層

  11. usecase/command/join_command.go 実装
  12. モックリポジトリでユニットテスト

  Phase 5: プレゼンテーション層

  13. make gen-openapi でDTO生成
  14. presentation/http/handler/command_handler.go 実装
  15. HTTPハンドラーテスト

  Phase 6: 統合

  16. cmd/work-tracker/main.go でDI配線
  17. E2Eテスト（curl等でAPI叩く）
  18. Docker Compose作成（PostgreSQL + Backend）

  ---
  追加の推奨ライブラリ

  | 用途       | ライブラリ             | 理由                         |
  |----------|-------------------|----------------------------|
  | HTTPルーター | go-chi/chi/v5     | 軽量、標準http互換、oapi-codegen対応 |
  | モック生成    | vektra/mockery    | インターフェースから自動生成             |
  | アサーション   | stretchr/testify  | assert、mockパッケージが便利        |
  | テストDB    | testcontainers-go | Dockerで隔離されたテストDB          |
  | ログ       | log/slog（標準）      | 構造化ログ、Go 1.21以降標準          |
  | バリデーション  | oapi-codegen組込    | OpenAPIスキーマから自動生成          |

  ---
  この構成で進めますか？まず Phase 1 から始めましょうか？