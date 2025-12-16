import redis
from app.config import load_config

cfg = load_config()

REDIS_CFG = cfg["redis"]
SESSION_CFG = cfg["session"]
HEARTBEAT_CFG = cfg["heartbeat"]["ttl_policy"]
PREFIX = REDIS_CFG["prefix"]
DEFAULT_TTL = SESSION_CFG["default_ttl"]

r = redis.Redis(
    host=REDIS_CFG["host"],
    port=REDIS_CFG["port"],
    db=REDIS_CFG["db"],
    decode_responses=True,
)


def _key(mac: str) -> str:
    return f"{PREFIX}{mac.lower()}"


def redis_health():
    try:
        return r.ping()
    except Exception:
        return False


def create_session(mac: str, role: str, ttl: int | None = None):
    ttl = ttl or DEFAULT_TTL
    r.setex(_key(mac), ttl, role)
    return r.ttl(_key(mac))


def get_session_full(mac: str):
    """
    Authoritative session fetch.
    """
    key = _key(mac)
    role = r.get(key)
    if not role:
        return None

    ttl = r.ttl(key)
    if ttl <= 0:
        return None

    return {
        "mac": mac.lower(),
        "role": role,
        "ttl": ttl,
    }


def refresh_session(mac: str, source: str | None):
    role = r.get(_key(mac))
    if not role:
        return None, None

    src = source or "unknown"
    ttl = HEARTBEAT_CFG.get(src, HEARTBEAT_CFG.get("unknown", DEFAULT_TTL))

    r.expire(_key(mac), ttl)
    return role, r.ttl(_key(mac))


def delete_session(mac: str):
    return bool(r.delete(_key(mac)))