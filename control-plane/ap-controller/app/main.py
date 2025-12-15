import os
from typing import Optional, Literal

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from app.store import set_session, get_session, del_session, healthcheck, normalize_mac

APP_TITLE = os.getenv("APP_TITLE", "AP Controller")
DEFAULT_TTL = int(os.getenv("SESSION_TTL_DEFAULT", "3600"))
MAX_TTL = int(os.getenv("SESSION_TTL_MAX", "86400"))  # 1 day cap for safety

app = FastAPI(title=APP_TITLE)


class LoginReq(BaseModel):
    mac: str = Field(..., description="Client MAC address, aa:bb:cc:dd:ee:ff")
    # 可扩展：用户名/密码/券码/短信等
    # username: Optional[str] = None
    # password: Optional[str] = None


class HeartbeatReq(BaseModel):
    mac: str


class LogoutReq(BaseModel):
    mac: str


class SessionResp(BaseModel):
    authorized: bool
    role: Optional[str] = None
    ttl: Optional[int] = None


def clamp_ttl(ttl: int) -> int:
    if ttl <= 0:
        return DEFAULT_TTL
    return min(ttl, MAX_TTL)


@app.get("/")
def root():
    return {"status": "ap-controller ok", "title": APP_TITLE}


@app.get("/healthz")
def healthz():
    return {"status": "ok", **healthcheck()}


@app.post("/portal/login", response_model=SessionResp)
def portal_login(req: LoginReq):
    """
    登录：创建 session，并给 Redis 设置 TTL。
    这里先做 demo：全部允许，role=guest。
    你后续可以接入：RADIUS/LDAP/券码/短信/企业账号等。
    """
    try:
        mac = normalize_mac(req.mac)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    role: Literal["guest", "staff", "admin"] = "guest"
    ttl = clamp_ttl(DEFAULT_TTL)

    set_session(mac, role, ttl)
    return {"authorized": True, "role": role, "ttl": ttl}


@app.post("/portal/heartbeat", response_model=SessionResp)
def portal_heartbeat(req: HeartbeatReq):
    """
    心跳：若已有 session，则刷新 TTL（续期）。
    """
    try:
        mac = normalize_mac(req.mac)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    sess, ttl = get_session(mac)
    if not sess:
        return {"authorized": False}

    role = sess.get("role") or "guest"
    ttl_new = clamp_ttl(DEFAULT_TTL)
    set_session(mac, role, ttl_new)
    return {"authorized": True, "role": role, "ttl": ttl_new}


@app.post("/portal/logout", response_model=SessionResp)
def portal_logout(req: LogoutReq):
    """
    下线：主动删除 session（例如用户点击退出）。
    """
    try:
        mac = normalize_mac(req.mac)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    del_session(mac)
    return {"authorized": False}


@app.get("/portal/status/{mac}", response_model=SessionResp)
def portal_status(mac: str):
    """
    给 ImmortalWRT 同步脚本用：查询是否授权 + 剩余 TTL。
    """
    try:
        mac = normalize_mac(mac)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    sess, ttl = get_session(mac)
    if not sess or ttl <= 0:
        return {"authorized": False}

    return {"authorized": True, "role": sess.get("role"), "ttl": ttl}
