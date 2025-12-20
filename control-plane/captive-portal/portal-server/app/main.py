from fastapi import FastAPI, Request, HTTPException
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
import httpx
import os
from typing import Dict, Any

app = FastAPI()
templates = Jinja2Templates(directory="app/templates")

# Controller endpoint
CONTROLLER_BASE = os.getenv("CONTROLLER_BASE", "http://ap-controller:8443")

# -------------------------------------------------------------------
# Helpers
# -------------------------------------------------------------------

def _format_ttl(ttl: Any) -> str:
    if not ttl:
        return "unknown"
    try:
        ttl = int(ttl)
    except Exception:
        return "unknown"

    if ttl < 60:
        return f"{ttl}s"
    if ttl < 3600:
        return f"{ttl // 60}m"
    return f"{ttl // 3600}h"


def get_trusted_context(request: Request) -> Dict[str, Any]:
    """
    Build trusted context from headers injected by gateway / nginx.
    Portal Server does NOT perform any security verification.
    """
    h = request.headers

    required = ["X-Client-MAC", "X-Client-IP"]
    missing = [k for k in required if not h.get(k)]
    if missing:
        raise HTTPException(
            status_code=400,
            detail=f"missing required headers: {', '.join(missing)}",
        )

    return {
        "client": {
            "mac": h.get("X-Client-MAC"),
            "ip": h.get("X-Client-IP"),
            "os": h.get("X-Client-OS"),
        },
        "wireless": {
            "ssid": h.get("X-Client-SSID"),
            "radio_id": h.get("X-Client-Radio-ID"),
        },
        "access": {
            "ap_id": h.get("X-Portal-AP-ID"),
            "vlan_id": h.get("X-Portal-VLAN-ID"),
        },
        "meta": {
            "source": "portal-server",
        },
    }


async def controller_post(path: str, payload: Dict[str, Any]) -> Dict[str, Any]:
    url = f"{CONTROLLER_BASE}{path}"
    async with httpx.AsyncClient(timeout=5.0) as client:
        resp = await client.post(url, json=payload)
        if resp.status_code != 200:
            raise HTTPException(
                status_code=resp.status_code,
                detail=f"controller error: {resp.text}",
            )
        return resp.json()


# -------------------------------------------------------------------
# Routes
# -------------------------------------------------------------------

@app.get("/", response_class=HTMLResponse)
async def portal_index(request: Request):
    return templates.TemplateResponse(
        "portal.html",
        {"request": request},
    )


@app.post("/login", response_class=HTMLResponse)
async def portal_login(request: Request):
    try:
        ctx = get_trusted_context(request)
    except HTTPException as e:
        return templates.TemplateResponse(
            "result.html",
            {
                "request": request,
                "authorized": False,
                "error": e.detail,
                "ttl_human": "n/a",
            },
        )

    # Forward trusted context to controller
    data = await controller_post("/portal/login", ctx)

    authorized = data.get("authorized", False)
    session = data.get("session", {})
    ttl = session.get("ttl")

    return templates.TemplateResponse(
        "result.html",
        {
            "request": request,
            "authorized": authorized,
            "session": session,
            "ttl_human": _format_ttl(ttl),
        },
    )


@app.get("/hotspot-detect.html", response_class=HTMLResponse)
async def hotspot_detect():
    # iOS / macOS captive portal detection
    return HTMLResponse("Success")


@app.get("/generate_204")
async def generate_204():
    # Android / HarmonyOS captive portal detection
    return HTMLResponse(status_code=204)
