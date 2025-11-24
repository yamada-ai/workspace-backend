from app.api.work_tracker_client import get_user_info

async def handle_info_command(user_name: str):
    """
    Twitchチャットから受け取った `/info` コマンドを送信

    Args:
        user_name: Twitchユーザー名

    Returns:
        UserInfoResponse

    Raises:
        RuntimeError: エラーが発生した場合（ユーザー未登録、アクティブセッションなし等）
    """
    # エラーは上位に伝播させる
    return await get_user_info(user_name)
