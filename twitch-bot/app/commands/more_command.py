from app.api.work_tracker_client import send_more_command

async def handle_more_command(user_name: str, content: str):
    """
    Twitchチャットから受け取った `/more` コマンドをパースして送信

    Examples:
        !more 30       → 30分延長
        !more 60       → 60分延長
        !more          → エラー（延長時間が必要）

    Args:
        user_name: Twitchユーザー名
        content: チャットメッセージ全体（例: "!more 30"）

    Returns:
        MoreCommandResponse

    Raises:
        ValueError: 延長時間が指定されていない、または無効な値の場合
        RuntimeError: その他のエラーが発生した場合
    """
    parts = content.strip().split()

    if len(parts) < 2:
        raise ValueError("延長時間（分）を指定してください。例: !more 30")

    try:
        minutes = int(parts[1])
    except ValueError:
        raise ValueError("延長時間は数値で指定してください。例: !more 30")

    if minutes < 1 or minutes > 360:
        raise ValueError("延長時間は1〜360分の範囲で指定してください。")

    # エラーは上位に伝播させる
    return await send_more_command(user_name, minutes)
