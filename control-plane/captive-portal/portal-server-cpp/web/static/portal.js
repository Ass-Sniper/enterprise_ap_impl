document.addEventListener("DOMContentLoaded", () => {
  const form = document.getElementById("f");
  const out = document.getElementById("out");

  const setOut = (text, kind) => {
    out.textContent = text ?? "";
    out.classList.remove("success", "error");
    if (kind) out.classList.add(kind);
  };

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    setOut("Submitting...", null);

    const fd = new FormData(e.target);
    const body = new URLSearchParams(fd);

    try {
      const res = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body
      });

      const txt = await res.text();
      // 原始响应显示
      setOut(txt, res.ok ? "success" : "error");

      try {
        const j = JSON.parse(txt);
        if (j.ok && j.token) {
          localStorage.setItem("portal_token", j.token);
        }
      } catch {
        // 非 JSON 响应时忽略解析错误
      }
    } catch (err) {
      setOut(`Network error: ${err?.message || err}`, "error");
    }
  });
});
