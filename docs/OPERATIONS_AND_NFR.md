# 運用・非機能要件ドキュメント

## 1. 非機能要件

### 1.1 可用性（Availability）

#### 目標稼働率
- **24時間365日稼働**: 年間稼働率 99.5% 以上を目指す
  - 許容ダウンタイム: 年間約43.8時間（月間約3.65時間）
  - 計画メンテナンス: 深夜帯（午前3-4時）に実施、5分以内

#### 無停止アップデート
- **デプロイ戦略**: ローリングアップデートまたはメンテナンスモード
- **ロールバック**: 簡単に前バージョンに戻せる仕組み
- **事前告知**: Discord等で事前にメンテナンス時間を告知

**関連Issue**: [#19 デプロイメント戦略（無停止アップデート）](https://github.com/yamada-ai/workspace-backend/issues/19)

---

### 1.2 信頼性（Reliability）

#### 自動復旧
- **PC再起動時の自動起動**: 電源障害やOS再起動時に人手を介さずにサービス復旧
- **コンテナ自動再起動**: `restart: always` ポリシーを全サービスに適用
- **ヘルスチェック**: 定期的なヘルスチェックで異常検知

**実装方針**:
- **Windows**: タスクスケジューラでシステム起動時に Docker Compose 起動
- **Linux**: systemd サービスで自動起動

**関連Issue**: [#18 自動復旧の仕組み（PC再起動時の自動起動）](https://github.com/yamada-ai/workspace-backend/issues/18)

#### 監視・アラート
- **外形監視**: UptimeRobot（無料枠: 50サイト、5分間隔）
- **監視対象**: `/health` エンドポイント
- **通知先**: Discord Webhook
- **リソース監視**: ディスク使用率、メモリ使用率

**関連Issue**: [#17 監視・アラートの導入（UptimeRobot + Discord）](https://github.com/yamada-ai/workspace-backend/issues/17)

---

### 1.3 保守性（Maintainability）

#### ログ管理
- **ログの永続化**: コンテナ再起動時にログが消失しないようボリュームマウント
- **ログローテーション**: 最大サイズ10MB、最大3ファイルまで保持
- **ログ分離**: エラーログとアクセスログの分離
- **保持期間**: 最低7日間

**関連Issue**: [#16 ログ管理の改善（永続化 + ローテーション）](https://github.com/yamada-ai/workspace-backend/issues/16)

#### バックアップ戦略
- **自動バックアップ**: 毎日深夜3時にPostgreSQLバックアップ
- **保持期間**: 最低7日分
- **外部保存**: Google Drive、外付けHDD等に定期的にコピー
- **復旧テスト**: 月1回、バックアップからの復元テストを実施

**バックアップ対象**:
- PostgreSQLデータベース（ユーザーデータ、作業時間、Raziiipo、ランキング）
- アプリケーション設定ファイル
- 環境変数ファイル（`.env.prod`）

**関連Issue**: [#14 PostgreSQLバックアップ戦略の実装](https://github.com/yamada-ai/workspace-backend/issues/14)

---

### 1.4 セキュリティ（Security）

#### シークレット管理
- **環境変数による管理**: `.env.prod` ファイルでシークレット情報を管理
- **強固なパスワード**: `openssl rand -base64 32` で生成
- **Git除外**: `.env.prod` は `.gitignore` に追加、Gitにコミットしない
- **アクセス制限**: 本番環境ファイルへのアクセスを最小限に制限

**管理対象**:
- データベースパスワード
- APIキー（Twitch Client ID/Secret）
- Webhook URL

**関連Issue**:
- [#15 本番環境設定の構築（compose.prod.yml + シークレット管理）](https://github.com/yamada-ai/workspace-backend/issues/15)
- [#6 環境変数管理整理](https://github.com/yamada-ai/workspace-backend/issues/6)

---

## 2. 本番環境構成

### 2.1 想定サーバースペック

#### ハードウェア
- **CPU**: Intel Core i7 7700 (Kaby Lake) 以上
- **メモリ**: 16GB以上
- **ストレージ**: SSD 256GB以上（ログ・バックアップ用に余裕を持たせる）
- **ネットワーク**: 安定した有線接続（24時間配信のため）

#### OS
- **Windows 10 Home 64bit** または **Linux**（Ubuntu 22.04 LTS 推奨）
- Docker Desktop（Windows）または Docker Engine（Linux）

---

### 2.2 Docker構成

#### compose.prod.yml
本番環境用の Docker Compose ファイルを `infrastructure/compose.prod.yml` として管理。

**主な設定**:
- `restart: always` - 自動再起動ポリシー
- 環境変数による設定（`.env.prod` から読み込み）
- ボリュームマウント（データ永続化）
- ログ設定（ローテーション）

**起動コマンド**:
```bash
docker compose -f infrastructure/compose.prod.yml --env-file .env.prod up -d
```

**関連Issue**: [#15 本番環境設定の構築](https://github.com/yamada-ai/workspace-backend/issues/15)

---

### 2.3 環境変数管理

#### .env.prod（本番環境）
```bash
# Database
POSTGRES_DB=workspace
POSTGRES_USER=workspace_user
POSTGRES_PASSWORD=<強固なランダムパスワード>

# Backend
DATABASE_URL=postgres://workspace_user:<password>@db:5432/workspace?sslmode=disable
PORT=8000
ENV=production

# Twitch Bot
TWITCH_CLIENT_ID=<your_client_id>
TWITCH_CLIENT_SECRET=<your_secret>
WORK_TRACKER_URL=http://localhost:8000
```

#### .env.example（各サービス）
- Backend: `.env.example`（このリポジトリ）
- Frontend: 別リポジトリ（[workspace-frontend](https://github.com/yamada-ai/workspace-frontend)）
- Twitch Bot: `twitch-bot/.env.example`（このリポジトリ）

新しい開発者が `.env.example` をコピーして `.env` を作成すれば即座に開発開始可能。

> **Note**: 現在フロントエンドは別リポジトリで管理されていますが、将来的にモノレポ化を検討中です。

**関連Issue**: [#6 環境変数管理整理](https://github.com/yamada-ai/workspace-backend/issues/6)

---

### 2.4 ネットワーク構成

#### ポート構成
| サービス | ポート | 用途 |
|---------|-------|------|
| work-tracker (Backend) | 8000 | HTTP API / WebSocket |
| PostgreSQL | 5432 | データベース（外部公開しない） |
| Frontend | 5173 (dev) | 開発環境のみ（本番はビルド後静的配信） |

#### 外部公開
- **開発環境**: localhost のみ
- **本番環境**: 必要に応じてリバースプロキシ（nginx）を設置
  - HTTPS化（Let's Encrypt）
  - 静的ファイル配信
  - レート制限

---

## 3. 運用方針

### 3.1 デプロイフロー

#### 通常デプロイ
1. **事前告知**: Discord等でメンテナンス時間を告知（24時間前）
2. **メンテナンスモード**: 深夜帯（午前3-4時）に実施
3. **デプロイ実行**:
   ```bash
   cd /path/to/workspace-backend
   git pull origin main
   docker compose -f infrastructure/compose.prod.yml build
   docker compose -f infrastructure/compose.prod.yml up -d
   ```
4. **ヘルスチェック**: `curl http://localhost:8000/health`
5. **動作確認**: 主要機能のスモークテスト
6. **完了通知**: Discord等で完了を報告

#### 緊急デプロイ（バグフィックス）
- メンテナンスモードを設定
- 最小限のダウンタイムで実施
- ロールバック準備（前バージョンのイメージを保持）

**関連Issue**: [#19 デプロイメント戦略](https://github.com/yamada-ai/workspace-backend/issues/19)

---

### 3.2 バックアップ運用

#### 自動バックアップ
- **頻度**: 毎日深夜3時
- **方式**: `pg_dump` による論理バックアップ
- **保存先**: `/path/to/backups/workspace_YYYYMMDD.sql`
- **保持期間**: 7日分（古いバックアップは自動削除）

#### バックアップスクリプト例
```bash
#!/bin/bash
# backup.sh
BACKUP_DIR=/path/to/backups
DATE=$(date +%Y%m%d)
docker exec workspace-prod-db pg_dump -U workspace_user workspace > $BACKUP_DIR/workspace_$DATE.sql
# 7日以上前のバックアップを削除
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
```

#### cron設定
```bash
0 3 * * * /path/to/backup.sh
```

#### 外部保存
- **Google Drive**: rclone で週次アップロード（15GB無料枠）
- **外付けHDD**: 月次でローカルバックアップをコピー

#### 復元手順
```bash
# 復元コマンド
docker exec -i workspace-prod-db psql -U workspace_user workspace < backup.sql
```

**関連Issue**: [#14 PostgreSQLバックアップ戦略](https://github.com/yamada-ai/workspace-backend/issues/14)

---

### 3.3 監視・アラート運用

#### 監視項目
| 項目 | ツール | 頻度 | アラート閾値 |
|-----|--------|------|-------------|
| サービス死活監視 | UptimeRobot | 5分 | 2回連続失敗 |
| ディスク使用率 | 監視スクリプト | 1時間 | 80%以上 |
| メモリ使用率 | 監視スクリプト | 1時間 | 90%以上 |
| エラーログ | 手動確認 | 日次 | - |

#### アラート通知先
- **Discord Webhook**: ダウン検知、リソース警告
- **メール**: UptimeRobot からの通知

#### 監視スクリプト例
```bash
#!/bin/bash
# monitor.sh
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    curl -X POST "$DISCORD_WEBHOOK" -d "{\"content\": \"⚠️ ディスク使用率: ${DISK_USAGE}%\"}"
fi
```

#### cron設定
```bash
0 * * * * /path/to/monitor.sh
```

**関連Issue**: [#17 監視・アラートの導入](https://github.com/yamada-ai/workspace-backend/issues/17)

---

### 3.4 ログ管理運用

#### ログファイル構成
```
logs/
├── work-tracker/
│   ├── app.log          # アプリケーションログ
│   └── error.log        # エラーログ
└── nginx/               # （将来的にリバースプロキシ導入時）
    ├── access.log
    └── error.log
```

#### ログローテーション設定
```yaml
# Docker Compose
services:
  work-tracker:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

#### logrotate設定（ホストOS）
```bash
# /etc/logrotate.d/workspace
/path/to/workspace-backend/logs/*/*.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
}
```

#### ログ確認コマンド
```bash
# アプリケーションログ確認
docker compose -f infrastructure/compose.prod.yml logs -f work-tracker

# エラーログのみ抽出
docker compose -f infrastructure/compose.prod.yml logs work-tracker | grep ERROR
```

**関連Issue**: [#16 ログ管理の改善](https://github.com/yamada-ai/workspace-backend/issues/16)

---

### 3.5 自動復旧運用

#### Windows環境
**タスクスケジューラ設定**:
1. タスクスケジューラを開く
2. 「基本タスクの作成」
3. トリガー: システム起動時
4. 操作: プログラムの起動
   - プログラム: `powershell.exe`
   - 引数: `-File C:\path\to\startup.ps1`

**startup.ps1**:
```powershell
# 30秒待機（Docker Desktopの起動を待つ）
Start-Sleep -Seconds 30
cd C:\path\to\workspace-backend
docker compose -f infrastructure\compose.prod.yml --env-file .env.prod up -d
```

#### Linux環境
**systemd サービス設定**:
```bash
# /etc/systemd/system/workspace.service
[Unit]
Description=Workspace Backend
After=docker.service

[Service]
Type=oneshot
WorkingDirectory=/path/to/workspace-backend
ExecStart=/usr/local/bin/docker-compose -f infrastructure/compose.prod.yml --env-file .env.prod up -d
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
```

**有効化**:
```bash
sudo systemctl enable workspace.service
sudo systemctl start workspace.service
```

#### 動作確認
1. PC再起動
2. `docker ps` でコンテナ確認
3. `curl http://localhost:8000/health` でヘルスチェック

**関連Issue**: [#18 自動復旧の仕組み](https://github.com/yamada-ai/workspace-backend/issues/18)

---

## 4. 開発・本番環境の分離

### 4.1 環境区分

| 環境 | 用途 | Docker Compose | 環境変数ファイル |
|-----|------|----------------|-----------------|
| 開発環境 | ローカル開発 | `compose.dev.yml` | `.env` |
| 本番環境 | 配信用サーバー | `compose.prod.yml` | `.env.prod` |

### 4.2 切り替え方法

#### 開発環境起動
```bash
docker compose -f infrastructure/compose.dev.yml up -d
```

#### 本番環境起動
```bash
docker compose -f infrastructure/compose.prod.yml --env-file .env.prod up -d
```

### 4.3 主な差分
| 項目 | 開発環境 | 本番環境 |
|-----|---------|---------|
| データベースパスワード | `postgres` | 強固なランダムパスワード |
| restart policy | `no` | `always` |
| ログレベル | DEBUG | INFO/WARN |
| ボリュームマウント | ホットリロード用 | データ永続化のみ |
| ポート公開 | localhost:8000 | localhost:8000（必要に応じてnginx経由） |

---

## 5. トラブルシューティング

### 5.1 よくある問題と対処法

#### コンテナが起動しない
```bash
# ログ確認
docker compose -f infrastructure/compose.prod.yml logs

# コンテナ状態確認
docker ps -a

# 強制再起動
docker compose -f infrastructure/compose.prod.yml down
docker compose -f infrastructure/compose.prod.yml up -d
```

#### データベース接続エラー
```bash
# PostgreSQL接続確認
docker exec -it workspace-prod-db psql -U workspace_user -d workspace

# 環境変数確認
docker compose -f infrastructure/compose.prod.yml config
```

#### ディスク容量不足
```bash
# 不要なDockerイメージ削除
docker system prune -a

# ログファイル削除
find logs/ -name "*.log" -mtime +7 -delete
```

#### バックアップからの復元

> ⚠️ **警告**: この手順は既存データを削除します。**最終手段**としてのみ使用してください。

**事前確認**:
1. バックアップファイルが存在し、正常であることを確認
2. 対象ボリューム名を確認: `docker volume ls | grep workspace`
3. 可能であれば、既存ボリュームのバックアップを取る

```bash
# 【事前確認】バックアップファイルの存在確認
ls -lh /path/to/backup/workspace_*.sql

# 【事前確認】ボリューム名の確認
docker volume ls | grep workspace-backend_db-data

# 1. データベースコンテナ停止
docker compose -f infrastructure/compose.prod.yml stop db

# 2. ⚠️ 危険: ボリューム削除（全データ削除）
# 対象ボリューム名が正しいか再確認してから実行
docker volume rm workspace-backend_db-data

# 3. コンテナ再起動（新規ボリューム作成）
docker compose -f infrastructure/compose.prod.yml up -d db

# 待機（PostgreSQL起動待ち）
sleep 10

# 4. バックアップから復元
docker exec -i workspace-prod-db psql -U workspace_user workspace < /path/to/backup/workspace_YYYYMMDD.sql

# 5. 復元確認
docker exec -it workspace-prod-db psql -U workspace_user workspace -c "SELECT COUNT(*) FROM users;"
```

**復元後の確認項目**:
- [ ] ユーザーデータが復元されているか
- [ ] 作業時間データが復元されているか
- [ ] Raziiipo残高が復元されているか
- [ ] アプリケーションが正常に起動するか

---

## 6. 今後の課題と優先度

### 高優先度（🔴）
- [#18 自動復旧の仕組み](https://github.com/yamada-ai/workspace-backend/issues/18)
- [#15 本番環境設定の構築](https://github.com/yamada-ai/workspace-backend/issues/15)
- [#14 PostgreSQLバックアップ戦略](https://github.com/yamada-ai/workspace-backend/issues/14)

### 中優先度（🟡）
- [#17 監視・アラートの導入](https://github.com/yamada-ai/workspace-backend/issues/17)
- [#16 ログ管理の改善](https://github.com/yamada-ai/workspace-backend/issues/16)
- [#6 環境変数管理整理](https://github.com/yamada-ai/workspace-backend/issues/6)

### 低優先度（🟢）
- [#19 デプロイメント戦略](https://github.com/yamada-ai/workspace-backend/issues/19)

---

## 7. チェックリスト

### 本番環境構築チェックリスト
- [ ] `compose.prod.yml` 作成
- [ ] `.env.prod` 作成（強固なパスワード設定）
- [ ] `.gitignore` に `.env.prod` 追加
- [ ] 自動起動設定（タスクスケジューラ or systemd）
- [ ] バックアップスクリプト設定
- [ ] cron設定（バックアップ・監視）
- [ ] UptimeRobot設定
- [ ] Discord Webhook設定
- [ ] ログローテーション設定
- [ ] `/health` エンドポイント実装確認
- [ ] バックアップからの復元テスト
- [ ] 再起動テスト（自動復旧確認）

### 定期メンテナンスチェックリスト（月次）
- [ ] ディスク容量確認
- [ ] バックアップファイル確認
- [ ] バックアップからの復元テスト
- [ ] エラーログ確認
- [ ] UptimeRobot稼働率確認
- [ ] セキュリティアップデート適用

---

## 8. 参考資料

### 公式ドキュメント
- [Docker Compose](https://docs.docker.com/compose/)
- [PostgreSQL Backup](https://www.postgresql.org/docs/current/backup.html)
- [UptimeRobot](https://uptimerobot.com/)

### 関連Issue
- [All Infrastructure Issues](https://github.com/yamada-ai/workspace-backend/issues?q=is%3Aissue+is%3Aopen+label%3Ainfrastructure)
- [Milestone: 基盤整備](https://github.com/yamada-ai/workspace-backend/milestone/1)

### 内部ドキュメント
- [外部仕様書](./EXTERNAL_SPECIFICATION.md)
- [README.md](../README.md)
