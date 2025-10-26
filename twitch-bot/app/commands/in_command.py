from app.api.work_tracker_client import send_in_command

async def handle_in_command(user_name: str, content: str):
    """
    Twitchチャットから受け取った `/in` コマンドをパースして送信

    Examples:
        !in                → 作業名なし
        !in 資料作成       → work_name="資料作成"
        !in 論文執筆       → work_name="論文執筆"

    Args:
        user_name: Twitchユーザー名
        content: チャットメッセージ全体（例: "!in 資料作成"）

    Returns:
        JoinCommandResponse

    Raises:
        AlreadyInSessionError: ユーザーが既にアクティブなセッションを持っている場合
        RuntimeError: その他のエラーが発生した場合
    """
    parts = content.strip().split(maxsplit=1)
    work_name = parts[1] if len(parts) >= 2 else None

    # エラーは上位に伝播させる
    return await send_in_command(user_name, work_name)
