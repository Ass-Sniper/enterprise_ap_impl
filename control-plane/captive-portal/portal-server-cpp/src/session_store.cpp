#include "session_store.h"
#include "util.h"

SessionStore::SessionStore(int ttl_seconds) : ttl_(ttl_seconds) {}

Session SessionStore::create(const std::string& ip, const std::string& mac, const std::string& username) {
    const auto now = std::chrono::steady_clock::now();
    Session s;
    s.token = genToken();
    s.ip = ip;
    s.mac = mac;
    s.username = username;
    s.expire_at = now + std::chrono::seconds(ttl_);

    std::lock_guard<std::mutex> lk(mu_);
    gcUnsafe(now);
    sessions_[s.token] = s;
    return s;
}

bool SessionStore::validate(const std::string& token, const std::string& ip, const std::string& mac) const {
    const auto now = std::chrono::steady_clock::now();
    std::lock_guard<std::mutex> lk(mu_);
    gcUnsafe(now);

    auto it = sessions_.find(token);
    if (it == sessions_.end()) return false;

    const Session& s = it->second;
    if (s.expire_at < now) return false;
    if (!ip.empty() && s.ip != ip) return false;
    if (!mac.empty() && s.mac != mac) return false;
    return true;
}

void SessionStore::revoke(const std::string& token) {
    std::lock_guard<std::mutex> lk(mu_);
    sessions_.erase(token);
}

std::string SessionStore::genToken() const {
    // 32 bytes hex
    return util::randomHex(32);
}

void SessionStore::gcUnsafe(std::chrono::steady_clock::time_point now) const {
    for (auto it = sessions_.begin(); it != sessions_.end(); ) {
        if (it->second.expire_at < now) it = sessions_.erase(it);
        else ++it;
    }
}
