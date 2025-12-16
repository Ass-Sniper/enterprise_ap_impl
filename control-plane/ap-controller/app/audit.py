import json
import time
import hmac
import hashlib
from typing import Dict, Any

from app.config import load_config

_cfg = load_config()
_audit_cfg = _cfg.get("audit", {})

AUDIT_ENABLED = _audit_cfg.get("enabled", True)
AUDIT_SECRET = _audit_cfg.get("secret")

if not AUDIT_SECRET:
    raise RuntimeError("audit.secret is missing in controller.yaml")


def sign_audit(payload: Dict[str, Any]) -> str:
    raw = json.dumps(payload, sort_keys=True, separators=(",", ":"))
    return hmac.new(
        AUDIT_SECRET.encode(),
        raw.encode(),
        hashlib.sha256
    ).hexdigest()


def audit_log(event: str, **context):
    if not AUDIT_ENABLED:
        return

    record = {
        "ts": int(time.time()),
        "event": event,
        **context,
    }

    record["sig"] = sign_audit(record)

    # stdout → docker logs → log pipeline
    print(json.dumps(record, separators=(",", ":")))
