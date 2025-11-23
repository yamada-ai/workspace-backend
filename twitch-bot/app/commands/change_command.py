from app.api.work_tracker_client import send_change_command

async def handle_change_command(user_name: str, content: str):
    """
    Twitchチャットから受け取った `/change` コマンドをパースして送信

    Examples:
        !change 資格勉強    → work_name="資格勉強"
        !change            → work_name="" (空の作業名)

    Args:
        user_name: Twitchユーザー名
        content: チャットメッセージ全体（例: "!change 資格勉強"）

    Returns:
        ChangeCommandResponse

    Raises:
        RuntimeError: その他のエラーが発生した場合
    """
    parts = content.strip().split(maxsplit=1)
    new_work_name = parts[1] if len(parts) >= 2 else ""

    # エラーは上位に伝播させる
    return await send_change_command(user_name, new_work_name)
