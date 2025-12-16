const mac = document.querySelector('[data-mac]')?.dataset?.mac || '';

async function refreshStatus() {
  try {
    const res = await fetch(`/status?mac=${encodeURIComponent(mac)}`, { method: 'GET' });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();

    // 授权状态
    const authorized = typeof data?.authorized === 'boolean' ? data.authorized : false;
    const row = document.querySelector('.row');
    let badge = row?.querySelector('.badge');
    if (!badge && row) {
      badge = document.createElement('span');
      badge.classList.add('badge');
      row.appendChild(badge);
    }
    if (badge) {
      if (authorized) {
        badge.textContent = '已授权';
        badge.classList.remove('fail');
        badge.classList.add('ok');
      } else {
        badge.textContent = '未授权';
        badge.classList.remove('ok');
        badge.classList.add('fail');
      }
    }

    // 精确更新各字段
    const roleEl = document.getElementById('val-role');
    const ttlEl = document.getElementById('val-ttl');
    const vlanEl = document.getElementById('val-vlan');
    const policyEl = document.getElementById('val-policy');
    const ipsetEl = document.getElementById('val-ipset');

    roleEl && (roleEl.textContent = data?.role ?? '-');

    const ttl = data?.ttl;
    if (ttlEl) {
      if (typeof ttl === 'number' && Number.isFinite(ttl) && ttl >= 0) {
        ttlEl.textContent = `${formatTTL(ttl)}（${ttl} 秒）`;
      } else {
        ttlEl.textContent = 'N/A';
      }
    }

    const net = data?.network ?? null;
    vlanEl && (vlanEl.textContent = (net && net.vlan != null) ? net.vlan : '-');
    policyEl && (policyEl.textContent = (net && net.policy) ? net.policy : '-');
    ipsetEl && (ipsetEl.textContent = (net && net.ipset) ? net.ipset : '-');
  } catch (e) {
    alert(`刷新失败：${e}`);
  }
}

async function doLogout() {
  try {
    const res = await fetch(`/logout`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ mac })
    });
    const ok = res.ok;
    const payload = await res.json().catch(() => ({}));
    alert(ok ? '已注销' : `注销失败：${payload.error || res.status}`);
    if (ok) refreshStatus();
  } catch (e) {
    alert(`注销失败：${e}`);
  }
}

function formatTTL(ttl) {
  const h = Math.floor(ttl / 3600);
  const m = Math.floor((ttl % 3600) / 60);
  const s = ttl % 60;
  const parts = [];
  if (h) parts.push(`${h}h`);
  if (m) parts.push(`${m}m`);
  if (s || parts.length === 0) parts.push(`${s}s`);
  return parts.join(' ');
}

document.getElementById('btn-refresh')?.addEventListener('click', refreshStatus);
document.getElementById('btn-logout')?.addEventListener('click', doLogout);
