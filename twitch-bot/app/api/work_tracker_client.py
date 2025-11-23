import os
import logging
from typing import Optional

from .generated.workspace_backend_api_client import Client
from .generated.workspace_backend_api_client.api.default import join_command, out_command, more_command
from .generated.workspace_backend_api_client.models import (
    JoinCommandRequest,
    JoinCommandResponse,
    OutCommandRequest,
    OutCommandResponse,
    MoreCommandRequest,
    MoreCommandResponse,
    ErrorResponse,
)

logger = logging.getLogger(__name__)

WORK_TRACKER_URL = os.getenv("WORK_TRACKER_URL", "http://localhost:8000")


class AlreadyInSessionError(Exception):
    """User already has an active session"""
    pass


async def send_in_command(user_name: str, work_name: Optional[str] = None):
    """
    `/in` コマンドで作業セッションを作成する

    Args:
        user_name: Twitch/YouTube のユーザー名
        work_name: 作業名（任意）

    Returns:
        JoinCommandResponse オブジェクト

    Raises:
        AlreadyInSessionError: ユーザーが既にアクティブなセッションを持っている場合
        RuntimeError: その他のエラー
    """
    client = Client(base_url=WORK_TRACKER_URL)

    request = JoinCommandRequest(
        user_name=user_name,
        work_name=work_name if work_name else None,
    )

    logger.info(f"POST {WORK_TRACKER_URL}/api/commands/join request={request}")

    try:
        # asyncio_detailed を使ってステータスコードを確認
        detailed_response = await join_command.asyncio_detailed(client=client, body=request)

        if detailed_response.status_code == 200:
            response = detailed_response.parsed
            if isinstance(response, JoinCommandResponse):
                logger.info(f"Session created: session_id={response.session_id}, user_id={response.user_id}")
                return response
            else:
                logger.error("[IN失敗] Unexpected response type for 200 OK")
                raise RuntimeError("Unexpected response type for 200 OK")

        elif detailed_response.status_code == 409:
            error_response = detailed_response.parsed
            if isinstance(error_response, ErrorResponse):
                logger.warning(f"[IN失敗] Already in session: {error_response.error}")
                raise AlreadyInSessionError(error_response.error)
            else:
                logger.error("[IN失敗] Unexpected response type for 409 Conflict")
                raise RuntimeError("Unexpected response type for 409 Conflict")

        else:
            error_response = detailed_response.parsed
            error_msg = error_response.error if isinstance(error_response, ErrorResponse) else "Unknown error"
            logger.error(f"[IN失敗] Status {detailed_response.status_code}: {error_msg}")
            raise RuntimeError(f"Server returned {detailed_response.status_code}: {error_msg}")

    except AlreadyInSessionError:
        # 既知のエラーはそのまま再送出
        raise
    except Exception as e:
        logger.error(f"[IN失敗] {e}")
        raise


async def send_out_command(user_name: str):
    """
    `/out` コマンドで作業セッションを終了する

    Args:
        user_name: Twitch/YouTube のユーザー名

    Returns:
        OutCommandResponse オブジェクト

    Raises:
        RuntimeError: ユーザー未登録、有効なセッションなし、その他のエラー
    """
    client = Client(base_url=WORK_TRACKER_URL)

    request = OutCommandRequest(user_name=user_name)

    logger.info(f"POST {WORK_TRACKER_URL}/api/commands/out request={request}")

    try:
        # asyncio_detailed を使ってステータスコードを確認
        detailed_response = await out_command.asyncio_detailed(client=client, body=request)

        if detailed_response.status_code == 200:
            response = detailed_response.parsed
            if isinstance(response, OutCommandResponse):
                logger.info(f"Session ended: session_id={response.session_id}, user_id={response.user_id}")
                return response
            else:
                logger.error("[OUT失敗] Unexpected response type for 200 OK")
                raise RuntimeError("Unexpected response type for 200 OK")

        elif detailed_response.status_code == 404:
            error_response = detailed_response.parsed
            error_msg = error_response.error if isinstance(error_response, ErrorResponse) else "User not found or no active session"
            logger.warning(f"[OUT失敗] {error_msg}")
            raise RuntimeError(error_msg)

        else:
            error_response = detailed_response.parsed
            error_msg = error_response.error if isinstance(error_response, ErrorResponse) else "Unknown error"
            logger.error(f"[OUT失敗] Status {detailed_response.status_code}: {error_msg}")
            raise RuntimeError(f"Server returned {detailed_response.status_code}: {error_msg}")

    except Exception as e:
        logger.error(f"[OUT失敗] {e}")
        raise


async def send_more_command(user_name: str, minutes: int):
    """
    `/more` コマンドで作業セッションを延長する

    Args:
        user_name: Twitch/YouTube のユーザー名
        minutes: 延長時間（分）。1〜360の範囲

    Returns:
        MoreCommandResponse オブジェクト

    Raises:
        RuntimeError: ユーザー未登録、有効なセッションなし、無効な延長時間、その他のエラー
    """
    client = Client(base_url=WORK_TRACKER_URL)

    request = MoreCommandRequest(user_name=user_name, minutes=minutes)

    logger.info(f"POST {WORK_TRACKER_URL}/api/commands/more request={request}")

    try:
        # asyncio_detailed を使ってステータスコードを確認
        detailed_response = await more_command.asyncio_detailed(client=client, body=request)

        if detailed_response.status_code == 200:
            response = detailed_response.parsed
            if isinstance(response, MoreCommandResponse):
                logger.info(f"Session extended: session_id={response.session_id}, minutes={response.minutes}")
                return response
            else:
                logger.error("[MORE失敗] Unexpected response type for 200 OK")
                raise RuntimeError("Unexpected response type for 200 OK")

        elif detailed_response.status_code == 400:
            error_response = detailed_response.parsed
            error_msg = error_response.error if isinstance(error_response, ErrorResponse) else "Invalid extension minutes"
            logger.warning(f"[MORE失敗] {error_msg}")
            raise RuntimeError(error_msg)

        elif detailed_response.status_code == 404:
            error_response = detailed_response.parsed
            error_msg = error_response.error if isinstance(error_response, ErrorResponse) else "User not found or no active session"
            logger.warning(f"[MORE失敗] {error_msg}")
            raise RuntimeError(error_msg)

        else:
            error_response = detailed_response.parsed
            error_msg = error_response.error if isinstance(error_response, ErrorResponse) else "Unknown error"
            logger.error(f"[MORE失敗] Status {detailed_response.status_code}: {error_msg}")
            raise RuntimeError(f"Server returned {detailed_response.status_code}: {error_msg}")

    except Exception as e:
        logger.error(f"[MORE失敗] {e}")
        raise
