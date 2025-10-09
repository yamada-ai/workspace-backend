- User

| カラム名 | 型 | 制約 | 備考 |
| --- | --- | --- | --- |
| id | SERIAL | PK | 主キー |
| name | TEXT | Not Null | ユーザ名 |
| Tier | integer | Not Null | Tier |
| created_at | TIMESTAMP | DEFAULT NOW() |  |
| updated_at | TIMESTAMP |  |  |


- Session

| カラム名 | 型 | 制約 |  |
| --- | --- | --- | --- |
| id | SERIAL | PK | 主キー |
| user_id | INTEGER | FK → user(id) | ユーザID |
| work_name | TEXT |  | 作業名(/in の場合null) |
| start_time | TIMESTAMP | Not Null | 開始時刻 |
| planned_end | TIMESTAMP | Not Null | 終了予定時間 |
| actual_end | TIMESTAMP |  | 実終了時間 |
| icon_id | INTEGER | FK → icon(id) | セッション中のアイコンID |
| created_at | TIMESTAMP | DEFAULT NOW() |  |
| updated_at | TIMESTAMP |  |  |
- Icon

| カラム名 | 型 | 制約 | 備考 |
| --- | --- | --- | --- |
| id | SERIAL | PK | 主キー |
| tier | SMALLINT | Not Null | どのティアか |
| created_at | TIMESTAMP | DEFAULT NOW() |  |
| updated_at | TIMESTAMP |  |  |
- IconMotion

| カラム名 | 型 | 制約 | 備考 |
| --- | --- | --- | --- |
| id | SERIAL | PK | 主キー |
| icon_id | INTEGER | Not Null | どのアイコンか |
| path | TEXT | Not Null | 画像URLまたはパス |
| motion_type | SMALLINT | Not Null | モーションのタイプ
1. 通常
2.  睡眠
3.  食事 |
| created_at | TIMESTAMP | DEFAULT NOW() |  |
| updated_at | TIMESTAMP |  |  |

- Comment

| カラム名 | 型 | 制約 | 備考 |
| --- | --- | --- | --- |
| id | SERIAL | PK | 主キー |
| user_id | INTEGER | Not Null |  |
| comment | TEXT | Not Null | コメント |
| created_at | TIMESTAMP | DEFAULT NOW() |  |
| updated_at | TIMESTAMP |  |  |

VirtualPoint

| カラム名 | 型 | 制約 | 備考 |
| --- | --- | --- | --- |
| id | SERIAL | PK | 主キー |
| user_id | INTEGER | Not Null Delete Cascade | ユーザID |
| point | INTEGER | Not Null | 仮想ポイント |
| created_at | TIMESTAMP | DEFAULT NOW() |  |
| updated_at | TIMESTAMP |  |  |

SlotLog

| カラム名 | 型 | 制約 | 備考 |
| --- | --- | --- | --- |
| id | SERIAL | PK | 主キー |
| user_id | INTEGER | Not Null | ユーザID |
| bet | INTEGER | Not Null | 仮想ポイント |
| multiplier_id | INTEGER  | NOT NULL | 倍率ID |
| multiplier_value | NUMERIC | NOT NULL | 実倍率(スナップショット) |
| reward | INTEGER  | NOT NULL | 実報酬値 |
| created_at | TIMESTAMP | DEFAULT NOW() |  |
| updated_at | TIMESTAMP |  |  |

SlotMultiplier

| カラム名 | 型 | 制約 | 備考 |
| --- | --- | --- | --- |
| id | SERIAL | PK | 主キー |
| name | TEXT | Not Null | 倍率表示 |
| value | NUMERIC  | Not Null | 実倍率 |
| created_at | TIMESTAMP | DEFAULT NOW() |  |
| updated_at | TIMESTAMP |  |  |
- Action

| カラム名 | 型 | 制約 | 備考 |
| --- | --- | --- | --- |
| id | SERIAL | PK | 主キー |
| name | TEXT | Not Null | アクション名 |
| cost | INTEGER | Not Null | 消費ポイント |
| icon_motion_id | INTEGER  |  | アニメーションのiconId |
| message | TEXT |  | 表示文言 |
| is_enable | BOOLEAN | DEFAULT TRUE |  |
| created_at | TIMESTAMP | DEFAULT NOW() |  |
| updated_at | TIMESTAMP |  |  |
- ActionLog

| カラム名 | 型 | 制約 | 備考 |
| --- | --- | --- | --- |
| id | SERIAL | PK | 主キー |
| user_id | TEXT | Not Null | ユーザID |
| action_id | INTEGER | Not Null |  |
| cost | INTEGER | Not Null | 消費ポイント |
| payload | JSONB |  | 追加情報( |
| created_at | TIMESTAMP | DEFAULT NOW() |  |