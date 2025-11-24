import os
import asyncio
import logging
import httpx
from typing import Optional

from dotenv import load_dotenv
from twitchAPI.twitch import Twitch
from twitchAPI.helper import first
from twitchAPI.eventsub.websocket import EventSubWebsocket
from twitchAPI.type import AuthScope

from app.commands.in_command import handle_in_command
from app.commands.out_command import handle_out_command
from app.commands.more_command import handle_more_command
from app.commands.change_command import handle_change_command
from app.commands.info_command import handle_info_command
from app.api.work_tracker_client import AlreadyInSessionError

load_dotenv()
logging.basicConfig(level=os.getenv("LOG_LEVEL", "INFO"))
log = logging.getLogger("bot")

CLIENT_ID         = os.environ["CLIENT_ID"]
CLIENT_SECRET     = os.environ["CLIENT_SECRET"]
USER_TOKEN        = os.environ["ACCESS_TOKEN"]     # 初期 access token
REFRESH_TOKEN_ENV = os.environ.get("REFRESH_TOKEN")
BROADCASTER_LOGIN = os.environ["CHANNELS"].lstrip("@").strip()  # 先頭@や空白の揺れを吸収

SEND_CHAT_URL = "https://api.twitch.tv/helix/chat/messages"
VALIDATE_URL  = "https://id.twitch.tv/oauth2/validate"
TOKEN_URL     = "https://id.twitch.tv/oauth2/token"

REQUIRED_SCOPES = [
    AuthScope.USER_READ_CHAT,   # 受信
    AuthScope.USER_WRITE_CHAT,  # 送信
    AuthScope.USER_BOT          # bot運用
]


class TokenManager:
    """access_token / refresh_token の管理と更新を一元化"""
    def __init__(self, client_id: str, client_secret: str, access_token: str, refresh_token: Optional[str]):
        self.client_id = client_id
        self.client_secret = client_secret
        self._access_token = access_token
        self._refresh_token = refresh_token
        self._lock = asyncio.Lock()

    @property
    def access_token(self) -> str:
        return self._access_token

    def apply_auth_headers(self, headers: dict) -> dict:
        # 呼び出し元の dict を壊さないようにコピー
        h = dict(headers) if headers else {}
        h["Authorization"] = f"Bearer {self._access_token}"
        h["Client-Id"] = self.client_id
        return h

    async def validate(self) -> dict:
        async with httpx.AsyncClient(timeout=10.0) as c:
            r = await c.get(VALIDATE_URL, headers={"Authorization": f"Bearer {self._access_token}"})
            if r.status_code == 200:
                return r.json()
            raise httpx.HTTPStatusError("invalid token", request=r.request, response=r)

    async def refresh(self) -> None:
        if not self._refresh_token:
            raise RuntimeError("refresh_token が未設定のため、更新できません。")

        async with self._lock:  # 多重更新防止
            # 二重チェック（同時401で複数タスクが待っていた場合に冪等化）
            try:
                await self.validate()
                return
            except Exception:
                pass

            async with httpx.AsyncClient(timeout=10.0) as c:
                data = {
                    "client_id": self.client_id,
                    "client_secret": self.client_secret,
                    "grant_type": "refresh_token",
                    "refresh_token": self._refresh_token,
                }
                r = await c.post(TOKEN_URL, data=data)
                r.raise_for_status()
                js = r.json()
                new_access = js["access_token"]
                new_refresh = js.get("refresh_token", self._refresh_token)

                # 更新
                self._access_token = new_access
                self._refresh_token = new_refresh
                # 環境変数へ反映したいならここで os.environ[...] を更新（任意）
                log.info("Access token refreshed successfully (scopes: %s)", js.get("scope"))

    async def ensure_fresh(self, min_expires: int = 180) -> None:
        """
        起動時や定期的に呼び出して、期限が近ければ先回りで refresh。
        min_expires: 残り秒数しきい値
        """
        try:
            v = await self.validate()
            if v.get("expires_in", 0) < min_expires:
                log.info("Token expires soon (%ss). Refreshing...", v.get("expires_in"))
                await self.refresh()
        except Exception:
            # 無効なら即更新を試みる
            log.info("Token invalid on startup. Refreshing...")
            await self.refresh()


async def send_chat(client: httpx.AsyncClient, token_mgr: TokenManager,
                    broadcaster_id: str, sender_id: str, message: str,
                    reply_to: Optional[str] = None):
    body = {"broadcaster_id": broadcaster_id, "sender_id": sender_id, "message": message}
    if reply_to:
        body["reply_parent_message_id"] = reply_to

    # 1回目試行
    r = await client.post(SEND_CHAT_URL, headers=token_mgr.apply_auth_headers(client.headers), json=body)
    if r.status_code == 401:
        # 自動リフレッシュ → 再試行
        log.warning("401 on send_chat. Trying token refresh...")
        await token_mgr.refresh()
        r = await client.post(SEND_CHAT_URL, headers=token_mgr.apply_auth_headers(client.headers), json=body)

    r.raise_for_status()
    return r.json()


async def main():
    token_mgr = TokenManager(CLIENT_ID, CLIENT_SECRET, USER_TOKEN, REFRESH_TOKEN_ENV)

    # 1) Twitch クライアント
    twitch = await Twitch(CLIENT_ID, CLIENT_SECRET)
    # 起動時に token の鮮度を保証（期限切れならここで更新）
    await token_mgr.ensure_fresh()

    # twitchAPI 側にも最新トークンを渡す
    await twitch.set_user_authentication(token_mgr.access_token, REQUIRED_SCOPES, REFRESH_TOKEN_ENV)

    # 2) ID 解決
    me = await first(twitch.get_users())
    if me is None:
        raise RuntimeError("get_users() でbotユーザーを取得できませんでした。")
    broadcaster = await first(twitch.get_users(logins=[BROADCASTER_LOGIN]))
    if broadcaster is None:
        raise RuntimeError(f"チャンネル {BROADCASTER_LOGIN} が見つかりません。")

    bot_user_id = me.id
    broadcaster_id = broadcaster.id
    log.info("me=%s(%s) broadcaster=%s(%s)", me.display_name, bot_user_id, broadcaster.display_name, broadcaster_id)

    # 3) 返信用 HTTP クライアント（Client-Id だけ固定で持たせる）
    http = httpx.AsyncClient(headers={"Content-Type": "application/json", "Client-Id": CLIENT_ID}, timeout=10.0)

    # 4) EventSub WebSocket
    es = EventSubWebsocket(twitch)
    es.start()

    # 5) ハンドラ
    async def on_chat(ev):
        msg = getattr(ev.event.message, "text", None) or getattr(ev.event, "message", None)
        user_name = getattr(ev.event, "chatter_user_name", None) or getattr(ev.event, "user_name", None)
        message_id = getattr(ev.event, "message_id", None)
        log.info("[%s] %s: %s", broadcaster.display_name, user_name, msg)

        if not isinstance(msg, str):
            return

        if msg.strip().lower() == "!ping":
            try:
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id, "pong", reply_to=message_id)
            except httpx.HTTPStatusError as e:
                log.exception("send_chat failed: %s", e.response.text)

        if msg == "!in" or msg.startswith("!in "):
            try:
                result = await handle_in_command(user_name, msg)
                # 成功時のメッセージ
                work_display = result.work_name if result.work_name else "作業"
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} {work_display}を開始しました！",
                              reply_to=message_id)
            except AlreadyInSessionError as e:
                # 409 Conflict: 既にセッション中
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} 既に作業セッション中です。先に !out で終了してください。",
                              reply_to=message_id)
            except Exception as e:
                # その他のエラー
                log.exception("!in command failed")
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} コマンドの処理に失敗しました。",
                              reply_to=message_id)

        if msg.startswith("!out"):
            try:
                result = await handle_out_command(user_name)
                # 成功時のメッセージ
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} 作業を終了しました。お疲れ様でした！",
                              reply_to=message_id)
            except Exception as e:
                # エラー時のメッセージ（ユーザー未登録、有効なセッションなし等）
                log.exception("!out command failed")
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} コマンドの処理に失敗しました。有効なセッションがない可能性があります。",
                              reply_to=message_id)

        if msg.startswith("!more"):
            try:
                result = await handle_more_command(user_name, msg)
                # 成功時のメッセージ
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} 作業時間を{result.minutes}分延長しました！",
                              reply_to=message_id)
            except ValueError as e:
                # 無効な引数
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} {str(e)}",
                              reply_to=message_id)
            except Exception as e:
                # その他のエラー（ユーザー未登録、有効なセッションなし等）
                log.exception("!more command failed")
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} コマンドの処理に失敗しました。有効なセッションがない可能性があります。",
                              reply_to=message_id)

        if msg.startswith("!change"):
            try:
                result = await handle_change_command(user_name, msg)
                # 成功時のメッセージ
                work_display = f'"{result.work_name}"' if result.work_name else "作業"
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} {work_display}を開始しました",
                              reply_to=message_id)
            except Exception as e:
                # エラー時のメッセージ（ユーザー未登録、有効なセッションなし等）
                log.exception("!change command failed")
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} 入室していません",
                              reply_to=message_id)

        if msg.startswith("!info"):
            try:
                result = await handle_info_command(user_name)
                # 成功時のメッセージ
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name}さん→退出まで:{result.remaining_minutes}分/今日の累計作業時間:{result.today_total_minutes}分/累計作業時間:{result.lifetime_total_minutes}分",
                              reply_to=message_id)
            except Exception as e:
                # エラー時のメッセージ（ユーザー未登録、アクティブセッションなし等）
                log.exception("!info command failed")
                await send_chat(http, token_mgr, broadcaster_id, bot_user_id,
                              f"@{user_name} 入室していません",
                              reply_to=message_id)

    await es.listen_channel_chat_message(
        broadcaster_user_id=broadcaster_id,
        user_id=bot_user_id,
        callback=on_chat
    )
    log.info("Subscribed to channel.chat.message")

    try:
        while True:
            # 定期的に鮮度チェック（任意）
            await token_mgr.ensure_fresh(min_expires=300)
            await asyncio.sleep(60)
    finally:
        await es.stop()
        await http.aclose()
        await twitch.close()


if __name__ == "__main__":
    asyncio.run(main())
