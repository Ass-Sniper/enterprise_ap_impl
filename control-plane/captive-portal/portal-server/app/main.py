from fastapi import FastAPI, Form, Request, Query
from fastapi.templating import Jinja2Templates
from pathlib import Path
from fastapi.responses import JSONResponse
import requests
import os
from fastapi.staticfiles import StaticFiles

CONTROLLER_URL = os.getenv("CONTROLLER_URL", "http://ap-controller:8443")

app = FastAPI(title="Captive Portal")
templates = Jinja2Templates(directory=str(Path(__file__).resolve().parent / "templates"))
app.mount("/static", StaticFiles(directory=str(Path(__file__).resolve().parent / "static")), name="static")

@app.get("/")
def portal_page(request: Request):
    return templates.TemplateResponse("portal.html", {"request": request})

def _format_ttl(ttl: int | None) -> str:
    if ttl is None or ttl < 0:
        return "N/A"
    hours, remainder = divmod(ttl, 3600)
    minutes, seconds = divmod(remainder, 60)
    parts = []
    if hours > 0:
        parts.append(f"{hours}h")
    if minutes > 0:
        parts.append(f"{minutes}m")
    if seconds > 0 or not parts:
        parts.append(f"{seconds}s")
    return " ".join(parts)

@app.post("/login")
def login(request: Request, mac: str = Form(...)):
    try:
        r = requests.post(
            f"{CONTROLLER_URL}/portal/login",
            json={"mac": mac, "ip": "0.0.0.0"},
            headers={"Accept": "application/json"},
            timeout=5,
        )
        r.raise_for_status()
        try:
            data = r.json()
            return templates.TemplateResponse(
                "result.html",
                {
                    "request": request,
                    "mac": mac,
                    "result": data,
                    "ttl_human": _format_ttl(data.get("ttl")),
                },
            )
        except ValueError:
            return {"ok": False, "status": r.status_code, "body": r.text}
    except requests.RequestException as e:
        return {"ok": False, "error": str(e)}

@app.get("/status")
def status(mac: str = Query(..., min_length=1)):
    try:
        r = requests.get(f"{CONTROLLER_URL}/portal/status/{mac}", timeout=5)
        r.raise_for_status()
        return r.json()
    except requests.RequestException as e:
        return {"ok": False, "error": str(e)}

@app.post("/logout")
def logout(payload: dict):
    try:
        r = requests.post(f"{CONTROLLER_URL}/portal/logout", json=payload, timeout=5)
        r.raise_for_status()
        return r.json()
    except requests.RequestException as e:
        return {"ok": False, "error": str(e)}
