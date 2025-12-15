from fastapi import FastAPI, Form
import requests

CONTROLLER_URL = "http://ap-controller:8443"

app = FastAPI(title="Captive Portal")

@app.get("/")
def portal_page():
    return """
    <html>
      <body>
        <h3>Captive Portal</h3>
        <form method="post" action="/login">
          <input name="mac" placeholder="Client MAC" />
          <button type="submit">Login</button>
        </form>
      </body>
    </html>
    """

@app.post("/login")
def login(mac: str = Form(...)):
    r = requests.post(
        f"{CONTROLLER_URL}/portal/auth",
        json={"mac": mac, "ip": "0.0.0.0"}
    )
    return r.json()
