#include "http_server.h"
#include "session_store.h"
#include <iostream>
#include <cstdlib>

int main(int argc, char** argv) {
    std::string host = "0.0.0.0";
    uint16_t port = 8080;

    if (const char* p = std::getenv("PORTAL_PORT")) port = (uint16_t)std::atoi(p);
    std::string web_root = std::getenv("WEB_ROOT") ? std::getenv("WEB_ROOT") : "./web";
    std::string hmac_key = std::getenv("PORTAL_HMAC_KEY") ? std::getenv("PORTAL_HMAC_KEY") : "devkey";

    // session TTL: 1 hour
    auto store = std::make_shared<SessionStore>(3600);

    HttpServer server(store, web_root, hmac_key);
    std::cerr << "Portal Server listening on " << host << ":" << port << "\n";
    server.start(host, port);
    return 0;
}
