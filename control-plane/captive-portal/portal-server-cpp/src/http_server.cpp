#include "http_server.h"
#include "util.h"
#include "hmac_sha256.h"

#include <restbed>
#include <cstdlib>

using namespace std;
using namespace restbed;

static string headerOrEmpty(const shared_ptr<const Request>& req, const string& k) {
    return req->has_header(k) ? req->get_header(k) : "";
}

HttpServer::HttpServer(shared_ptr<SessionStore> store, string web_root, string hmac_key)
    : store_(std::move(store)), web_root_(std::move(web_root)), hmac_key_(std::move(hmac_key)),
      service_(make_shared<Service>()) {
    setupRoutes();
}

void HttpServer::start(const string& host, uint16_t port) {
    auto settings = make_shared<Settings>();
    settings->set_bind_address(host);
    settings->set_port(port);
    settings->set_default_header("Server", "portal-cpp11");
    service_->start(settings);
}

void HttpServer::setupRoutes() {
    // GET /healthz
    {
        auto r = make_shared<Resource>();
        r->set_path("/healthz");
        r->set_method_handler("GET", [](const shared_ptr<Session> session) {
            const string body = "ok\n";
            session->close(200, body, {{"Content-Type","text/plain"}});
        });
        service_->publish(r);
    }

    // GET /portal
    {
        auto r = make_shared<Resource>();
        r->set_path("/portal");
        r->set_method_handler("GET", [this](const shared_ptr<Session> session) {
            const string html = util::readFile(web_root_ + "/portal.html");
            session->close(200, html.empty() ? "<h1>portal.html missing</h1>" : html,
                           {{"Content-Type","text/html; charset=utf-8"}});
        });
        service_->publish(r);
    }

    // POST /api/login  (form urlencoded: username, password, ip, mac)
    {
        auto r = make_shared<Resource>();
        r->set_path("/api/login");
        r->set_method_handler("POST", [this](const shared_ptr<Session> session) {
            const auto req = session->get_request();
            const size_t len = req->get_header("Content-Length", 0);

            session->fetch(len, [this](const shared_ptr<Session> session, const Bytes& bodyBytes) {
                string body((char*)bodyBytes.data(), bodyBytes.size());
                auto form = util::parseFormUrlEncoded(body);

                string username = form["username"];
                string password = form["password"];
                string ip = form["ip"];
                string mac = form["mac"];

                // ✅ 这里先做最小可用：写死用户；你后面可以接 MySQL/FreeRADIUS 做校验
                if (!(username == "testuser" && password == "testpass")) {
                    session->close(401, R"({"ok":false,"err":"bad credentials"})",
                                   {{"Content-Type","application/json"}});
                    return;
                }

                Session s = store_->create(ip, mac, username);

                // token 也可以做成 HMAC( ip|mac|ts|rand )，这里直接随机
                string resp = string("{\"ok\":true,\"token\":\"") + s.token + "\"}";
                session->close(200, resp, {{"Content-Type","application/json"}});
            });
        });
        service_->publish(r);
    }

    // GET /api/check?token=...&ip=...&mac=...
    // 给 Nginx auth_request / 网关调用
    {
        auto r = make_shared<Resource>();
        r->set_path("/api/check");
        r->set_method_handler("GET", [this](const shared_ptr<Session> session) {
            const auto req = session->get_request();

            string token = req->get_query_parameter("token", "");
            string ip = req->get_query_parameter("ip", "");
            string mac = req->get_query_parameter("mac", "");

            if (token.empty()) {
                // 也允许从 Header/Cookie 拿
                string auth = headerOrEmpty(req, "Authorization");
                // 形如 "Bearer xxx"
                if (auth.rfind("Bearer ", 0) == 0) token = auth.substr(7);
            }

            bool ok = store_->validate(token, ip, mac);
            if (!ok) {
                session->close(401, "unauthorized\n", {{"Content-Type","text/plain"}});
                return;
            }

            // 你也可以回传网关需要的放行信息，例如 X-User / X-Policy
            session->close(200, "ok\n",
                           {{"Content-Type","text/plain"},
                            {"X-Portal-User", "ok"}});
        });
        service_->publish(r);
    }

    // POST /api/logout
    {
        auto r = make_shared<Resource>();
        r->set_path("/api/logout");
        r->set_method_handler("POST", [this](const shared_ptr<Session> session) {
            const auto req = session->get_request();
            string token = req->get_query_parameter("token", "");
            if (token.empty()) token = headerOrEmpty(req, "Authorization");
            if (token.rfind("Bearer ", 0) == 0) token = token.substr(7);

            if (!token.empty()) store_->revoke(token);
            session->close(200, R"({"ok":true})", {{"Content-Type","application/json"}});
        });
        service_->publish(r);
    }
}
