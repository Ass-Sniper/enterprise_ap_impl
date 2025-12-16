from fastapi import FastAPI
from pydantic import BaseModel
from app.store import (
    create_session,
    get_session_full,
    refresh_session,
    delete_session,
    redis_health,
)
from app.config import load_config
from app.audit import audit_log

cfg = load_config()

CONTROLLER_CFG = cfg["controller"]
SESSION_CFG = cfg["session"]
ROLE_CFG = cfg["roles"]

app = FastAPI(
    title=CONTROLLER_CFG["name"],
    version=CONTROLLER_CFG["version"],
)


class LoginReq(BaseModel):
    mac: str


class HeartbeatReq(BaseModel):
    mac: str
    source: str | None = None

    class Config:
        extra = "allow"


class LogoutReq(BaseModel):
    mac: str


# ---------- Helpers ----------

def build_session_resp(sess: dict) -> dict:
    """
    Build external session response with role-based network policy injected.
    """
    role = sess.get("role")
    ttl = sess.get("ttl")

    role_cfg = ROLE_CFG.get(role, {})
    network_cfg = role_cfg.get("network", {})

    return {
        "authorized": True,
        "role": role,
        "ttl": ttl,
        "network": {
            "vlan": network_cfg.get("vlan"),
            "policy": network_cfg.get("policy"),
            "ipset": network_cfg.get("ipset"),
        },
    }

# ---------- Endpoints ----------


@app.get("/")
def root():
    return {"status": "ok"}


@app.get("/healthz")
def healthz():
    return {
        "status": "ok",
        "redis": redis_health(),
    }


@app.post("/portal/login")
def portal_login(req: LoginReq):
    role = "guest"
    ttl = 3600

    create_session(req.mac, role, ttl)
    sess = get_session_full(req.mac)
    resp = build_session_resp(sess)

    audit_log(
        event="portal.login",
        mac=req.mac,
        authorized=True,
        role=resp["role"],
        ttl=resp["ttl"],
        network=resp["network"],
        result="ok",
    )

    return resp


@app.post("/portal/heartbeat")
def portal_heartbeat(req: HeartbeatReq):
    refreshed = refresh_session(req.mac, req.source)

    if not refreshed:
        audit_log(
            event="portal.heartbeat",
            mac=req.mac,
            authorized=False,
            source=req.source,
            result="not_found",
        )
        return {"authorized": False}

    sess = get_session_full(req.mac)
    if not sess:
        audit_log(
            event="portal.heartbeat",
            mac=req.mac,
            authorized=False,
            source=req.source,
            result="expired_after_refresh",
        )
        return {"authorized": False}

    resp = build_session_resp(sess)

    audit_log(
        event="portal.heartbeat",
        mac=req.mac,
        authorized=True,
        role=resp["role"],
        ttl=resp["ttl"],
        network=resp["network"],
        source=req.source,
        result="ok",
    )

    return resp


@app.post("/portal/logout")
def portal_logout(req: LogoutReq):
    sess = get_session_full(req.mac)
    existed = delete_session(req.mac)

    resp = build_session_resp(sess) if sess else None

    audit_log(
        event="portal.logout",
        mac=req.mac,
        authorized=False,
        role=resp["role"] if resp else None,
        network=resp["network"] if resp else None,
        result="ok" if existed else "not_found",
    )

    return {"authorized": False}

@app.get("/portal/status/{mac}")
def portal_status(mac: str):
    sess = get_session_full(mac)
    if not sess:
        return {
            "authorized": False,
            "role": None,
            "ttl": None,
            "network": None,
        }

    return build_session_resp(sess)


@app.post("/portal/batch_status")
def batch_status(req: dict):
    results = []

    for e in req.get("entries", []):
        mac = e.get("mac")
        sess = get_session_full(mac)

        if not sess:
            results.append({
                "mac": mac,
                "authorized": False,
            })
            continue

        resp = build_session_resp(sess)
        resp["mac"] = mac
        results.append(resp)

    return {"results": results}

