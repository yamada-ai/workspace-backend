import os
import logging
from typing import Optional

from .generated.workspace_backend_api_client import Client
from .generated.workspace_backend_api_client.api.default import join_command
from .generated.workspace_backend_api_client.models import JoinCommandRequest

logger = logging.getLogger(__name__)

WORK_TRACKER_URL = os.getenv("WORK_TRACKER_URL", "http://localhost:8000")


async def send_in_command(user_name: str, work_name: Optional[str] = None):
    """
    `/in` コマンドで作業セッションを作成する

    Args:
        user_name: Twitch/YouTube のユーザー名
        work_name: 作業名（任意）

    Returns:
        JoinCommandResponse オブジェクト
    """
    client = Client(base_url=WORK_TRACKER_URL)

    request = JoinCommandRequest(
        user_name=user_name,
        work_name=work_name if work_name else None,
    )

    logger.info(f"POST {WORK_TRACKER_URL}/api/commands/join request={request}")

    try:
        response = await join_command.asyncio(client=client, body=request)

        if response is None:
            logger.error("[IN失敗] No response received")
            raise RuntimeError("No response from server")
        print(response)
        logger.info(f"Session created: session_id={response.session_id}, user_id={response.user_id}")
        return response

    except Exception as e:
        logger.error(f"[IN失敗] {e}")
        raise
