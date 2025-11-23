from app.api.work_tracker_client import send_out_command

async def handle_out_command(user_name: str):
    """
    Twitchチャットから受け取った `/out` コマンドを送信

    Args:
        user_name: Twitchユーザー名

    Returns:
        OutCommandResponse

    Raises:
        RuntimeError: エラーが発生した場合（ユーザー未登録、有効なセッションなし等）
    """
    # エラーは上位に伝播させる
    return await send_out_command(user_name)
