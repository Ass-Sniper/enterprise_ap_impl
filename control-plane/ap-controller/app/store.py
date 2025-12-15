import os
import time
import re
from typing import Optional, Tuple, Dict, Any

import redis


REDIS_HOST = os.getenv("REDIS_HOST", "redis")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))
REDIS_DB = int(os.getenv("REDIS_DB", "0"))
REDIS_PASSWORD = os.getenv("REDIS_PASSWORD") or None

SESSION_PREFIX = os.getenv("SESSION_PREFIX", "session:")

# 连接 Redis（docker-compose 里 REDIS_HOST=redis 即可）
r = redis.Redis(
    host=REDIS_HOST,
    port=REDIS_PORT,
    db=REDIS_DB,
    password=REDIS_PASSWORD,
    decode_responses=True,
)


_MAC_RE = re.compile(r"^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$")


def normalize_mac(mac: str) -> str:
    mac = (mac or "").strip().lower()
    if not _MAC_RE.match(mac):
        raise ValueError(f"invalid mac: {mac}")
    return mac


def session_key(mac: str) -> str:
    return f"{SESSION_PREFIX}{mac}"


def set_session(mac: str, role: str, ttl: int) -> Dict[str, Any]:
    """
    Create/refresh session for mac with role and ttl seconds.
    Stores:
      - role
      - expires_at (unix ts)
    And sets Redis TTL via EXPIRE.
    """
    mac = normalize_mac(mac)
    ttl = int(ttl)
    if ttl <= 0:
        raise ValueError("ttl must be > 0")

    role = (role or "").strip() or "guest"
    now = int(time.time())
    expires_at = now + ttl

    key = session_key(mac)

    pipe = r.pipeline()
    pipe.hset(key, mapping={"role": role, "expires_at": str(expires_at)})
    pipe.expire(key, ttl)
    pipe.execute()

    return {"mac": mac, "role": role, "expires_at": expires_at, "ttl": ttl}


def get_session(mac: str) -> Tuple[Optional[Dict[str, Any]], int]:
    """
    Returns (session_dict_or_none, ttl_seconds).
    - ttl_seconds: Redis TTL (>=0), or -2 if missing, -1 if no expire (shouldn't happen).
    """
    mac = normalize_mac(mac)
    key = session_key(mac)

    pipe = r.pipeline()
    pipe.hgetall(key)
    pipe.ttl(key)
    data, ttl = pipe.execute()

    if not data:
        return None, -2

    # Normalize/validate fields
    role = data.get("role") or "guest"
    expires_at_raw = data.get("expires_at")
    try:
        expires_at = int(expires_at_raw) if expires_at_raw is not None else None
    except ValueError:
        expires_at = None

    session = {
        "mac": mac,
        "role": role,
        "expires_at": expires_at,
    }
    return session, int(ttl)


def del_session(mac: str) -> bool:
    mac = normalize_mac(mac)
    key = session_key(mac)
    return bool(r.delete(key))


def healthcheck() -> Dict[str, Any]:
    """
    Simple connectivity check.
    """
    try:
        pong = r.ping()
        return {
            "redis_host": REDIS_HOST,
            "redis_port": REDIS_PORT,
            "redis_db": REDIS_DB,
            "redis_ping": bool(pong),
        }
    except Exception as e:
        return {
            "redis_host": REDIS_HOST,
            "redis_port": REDIS_PORT,
            "redis_db": REDIS_DB,
            "redis_ping": False,
            "error": str(e),
        }
