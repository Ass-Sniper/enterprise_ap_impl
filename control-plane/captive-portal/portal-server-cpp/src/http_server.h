#pragma once
#include <memory>
#include <string>
#include "session_store.h"

namespace restbed { class Service; }

class HttpServer {
public:
    HttpServer(std::shared_ptr<SessionStore> store,
               std::string web_root,
               std::string hmac_key);

    void start(const std::string& host, uint16_t port);

private:
    void setupRoutes();

private:
    std::shared_ptr<SessionStore> store_;
    std::string web_root_;
    std::string hmac_key_;
    std::shared_ptr<restbed::Service> service_;
};
