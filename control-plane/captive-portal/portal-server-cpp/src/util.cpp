#include "util.h"
#include <fstream>
#include <sstream>
#include <iomanip>
#include <random>
#include <cctype>

namespace util {

std::string readFile(const std::string& path) {
    std::ifstream ifs(path.c_str(), std::ios::in | std::ios::binary);
    if (!ifs) return "";
    std::ostringstream oss;
    oss << ifs.rdbuf();
    return oss.str();
}

static int hexVal(char c) {
    if ('0' <= c && c <= '9') return c - '0';
    if ('a' <= c && c <= 'f') return 10 + (c - 'a');
    if ('A' <= c && c <= 'F') return 10 + (c - 'A');
    return -1;
}

std::string urlDecode(const std::string& s) {
    std::string out;
    out.reserve(s.size());
    for (size_t i = 0; i < s.size(); i++) {
        if (s[i] == '%' && i + 2 < s.size()) {
            int hi = hexVal(s[i+1]);
            int lo = hexVal(s[i+2]);
            if (hi >= 0 && lo >= 0) {
                out.push_back(static_cast<char>((hi << 4) | lo));
                i += 2;
            } else {
                out.push_back(s[i]);
            }
        } else if (s[i] == '+') {
            out.push_back(' ');
        } else {
            out.push_back(s[i]);
        }
    }
    return out;
}

std::unordered_map<std::string, std::string> parseFormUrlEncoded(const std::string& body) {
    std::unordered_map<std::string, std::string> kv;
    size_t start = 0;
    while (start < body.size()) {
        size_t amp = body.find('&', start);
        if (amp == std::string::npos) amp = body.size();
        size_t eq = body.find('=', start);
        if (eq != std::string::npos && eq < amp) {
            auto k = urlDecode(body.substr(start, eq - start));
            auto v = urlDecode(body.substr(eq + 1, amp - (eq + 1)));
            kv[k] = v;
        }
        start = amp + 1;
    }
    return kv;
}

std::string randomHex(size_t bytes) {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<int> dist(0, 255);

    std::ostringstream oss;
    for (size_t i = 0; i < bytes; i++) {
        int b = dist(gen);
        oss << std::hex << std::setw(2) << std::setfill('0') << (b & 0xff);
    }
    return oss.str();
}

}
