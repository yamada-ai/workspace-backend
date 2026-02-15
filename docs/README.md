# ドキュメント

workspace-backendプロジェクトの各種ドキュメントです。

## 📚 ドキュメント一覧

### [外部仕様書](./EXTERNAL_SPECIFICATION.md)
24時間オンラインコワーキングスペースの機能仕様書。

**内容**:
- システム概要
- ユーザー階層（Tier分類）
- コマンド仕様（/in, /out, /more, /sleep, /dance等）
- 仮想ポイントシステム（Raziiipo）
- ランキングシステム
- 画面表示仕様
- アニメーション仕様

**対象読者**: 開発者、企画者、UI/UXデザイナー

---

### [運用・非機能要件](./OPERATIONS_AND_NFR.md)
本番環境の構築・運用方針、非機能要件（可用性、信頼性、保守性、セキュリティ）をまとめたドキュメント。

**内容**:
- 非機能要件（可用性、信頼性、保守性、セキュリティ）
- 本番環境構成（Docker、環境変数、ネットワーク）
- 運用方針（デプロイフロー、バックアップ、監視、ログ管理、自動復旧）
- トラブルシューティング
- 今後の課題と優先度
- チェックリスト

**対象読者**: インフラ担当者、運用担当者、開発リーダー

---

## 🔗 関連リンク

### GitHub
- [Issues（課題管理）](https://github.com/yamada-ai/workspace-backend/issues)
- [Milestone: 基盤整備](https://github.com/yamada-ai/workspace-backend/milestone/1)

### 参考資料
- [オンライン自習室を作った（Zenn）](https://zenn.dev/soraride/articles/a546dbfc4bb6ee)
- [UnityでTwitchチャットと連携（Qiita）](https://qiita.com/mojomojopon/items/c74b3027e1302d489f77)

---

## 📝 更新履歴

| 日付 | ドキュメント | 変更内容 |
|-----|-------------|---------|
| 2025-02-15 | 全体 | 初版作成 |
