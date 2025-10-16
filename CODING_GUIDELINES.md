# コーディングガイドライン

このプロジェクトのコーディング規約です。

## 基本方針

- **可読性と保守性を重視**
- **Go言語の慣習に従いつつ、日本語チームの特性を活かす**
- **自動生成コードは規約の対象外**

---

## 1. コメント

### 原則：日本語で記述

```go
// ✅ Good
// ユーザー情報を取得する
func FindUser(id int64) (*User, error) {
    // データベースから検索
    return db.Query(...)
}

// ❌ Bad
// Find user by ID
func FindUser(id int64) (*User, error) {
    // Query from database
    return db.Query(...)
}
```

### 例外：自動生成コード

- `sqlc` 生成コード（`infrastructure/database/sqlc/`）
- OpenAPI 生成コード（`presentation/http/dto/`）
- これらは英語コメントのままでOK（再生成時に上書きされるため）

---

## 2. 変数名

### usecase表記

- ❌ `useCase`（Cを大文字）を禁止
- ✅ `usecase` を使用

```go
// ✅ Good
joinUsecase := command.NewJoinCommandUseCase(...)

// ❌ Bad
joinUseCase := command.NewJoinCommandUseCase(...)
```

### repository表記

- ❌ `repo` への簡略化を禁止
- ✅ `repository` をそのまま使用

```go
// ✅ Good
userRepository := repository.NewUserRepository(queries)
sessionRepository := repository.NewSessionRepository(queries)

// ❌ Bad
userRepo := repository.NewUserRepository(queries)
sessionRepo := repository.NewSessionRepository(queries)
```

**理由**：
- 明示性の向上
- 検索性の向上（grepしやすい）
- 省略による曖昧さの排除

---

## 3. テストコード

### テスト関数名

**英語のままでOK**（Go言語の慣習に従う）

```go
// ✅ Good
func TestJoinCommand_NewUser(t *testing.T) { ... }
func TestUserRepository_FindByName(t *testing.T) { ... }
```

### テストケース名（t.Run内）

**原則：日本語で記述**

```go
// ✅ Good
t.Run("新規ユーザーの登録", func(t *testing.T) {
    // テストコード
})

t.Run("既存ユーザーが既にセッション中の場合", func(t *testing.T) {
    // テストコード
})

// ⚠️ Acceptable（英語でも許容するが、日本語推奨）
t.Run("NewUser_Registration", func(t *testing.T) {
    // テストコード
})
```

### テストケース名のガイドライン

**推奨パターン**：

1. **正常系**: 「〜の場合」「〜を実行」
   ```go
   t.Run("有効なユーザー名でユーザーを作成", func(t *testing.T) { ... })
   ```

2. **異常系**: 「〜の場合はエラー」「不正な〜」
   ```go
   t.Run("空のユーザー名の場合はエラー", func(t *testing.T) { ... })
   t.Run("不正なTier値でバリデーションエラー", func(t *testing.T) { ... })
   ```

3. **境界値**: 「〜の境界値」
   ```go
   t.Run("Tier値の境界値テスト", func(t *testing.T) { ... })
   ```

**注意**：
- 絶対的なルールではなく、チームで柔軟に運用
- 既存の英語テストケース名を無理に変更する必要はない
- 新規作成時は日本語を推奨

---

## 4. 適用範囲

### 対象

- ✅ `domain/` - ドメインロジック
- ✅ `usecase/` - ユースケース
- ✅ `infrastructure/database/repository/` - リポジトリ実装
- ✅ `infrastructure/config/` - 設定
- ✅ `presentation/http/handler/` - ハンドラー
- ✅ `cmd/` - エントリーポイント
- ✅ テストコード

### 対象外

- ❌ `infrastructure/database/sqlc/` - sqlc自動生成
- ❌ `presentation/http/dto/` - OpenAPI自動生成
- ❌ サードパーティライブラリのコード

---

## 5. 既存コードの扱い

- **新規コード**: このガイドラインに完全準拠
- **既存コード**: 修正時に可能な範囲で準拠させる
- **レガシーコード**: 無理に全修正する必要はなし

---

## 6. 例外の扱い

ガイドラインは「原則」であり「絶対」ではありません。

**例外が許容されるケース**：
- パフォーマンス上の理由
- 外部ライブラリとの互換性
- 業界標準的な用語（例: `ctx` for context）
- チーム内で合意が得られた場合

例外を適用する場合は、コードレビューで理由を明示してください。

---

## 7. レビュー時のチェックポイント

- [ ] コメントは日本語で書かれているか（自動生成を除く）
- [ ] `useCase`（Cが大文字）を使用していないか
- [ ] `repo` への省略を使用していないか
- [ ] テストケース名（`t.Run`内）は日本語で書かれているか（推奨）

---

## 改訂履歴

- 2025-10-12: 初版作成
