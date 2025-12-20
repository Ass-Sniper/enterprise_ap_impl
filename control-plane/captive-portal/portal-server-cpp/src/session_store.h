#pragma once
#include <string>
#include <unordered_map>
#include <mutex>
#include <chrono>

struct Session {
    std::string token;
    std::string ip;
    std::string mac;
    std::string username;
    std::chrono::steady_clock::time_point expire_at;
};

class SessionStore {
public:
    explicit SessionStore(int ttl_seconds);

    Session create(const std::string& ip, const std::string& mac, const std::string& username);
    bool validate(const std::string& token, const std::string& ip, const std::string& mac) const;
    void revoke(const std::string& token);

private:
    std::string genToken() const;
    void gcUnsafe(std::chrono::steady_clock::time_point now) const;

private:
    int ttl_;
    mutable std::mutex mu_;
    mutable std::unordered_map<std::string, Session> sessions_; // token -> session
};
