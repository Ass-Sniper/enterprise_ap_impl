#pragma once
#include <string>
#include <unordered_map>

namespace util {
    std::string readFile(const std::string& path);
    std::string urlDecode(const std::string& s);
    std::unordered_map<std::string, std::string> parseFormUrlEncoded(const std::string& body);
    std::string randomHex(size_t bytes);
}
