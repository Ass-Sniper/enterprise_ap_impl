#pragma once
#include <string>

namespace crypto {
    // 返回 hex string
    std::string hmacSha256Hex(const std::string& key, const std::string& data);
}
