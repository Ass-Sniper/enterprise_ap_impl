#include "hmac_sha256.h"
#include <openssl/hmac.h>
#include <sstream>
#include <iomanip>

namespace crypto {

std::string hmacSha256Hex(const std::string& key, const std::string& data) {
    unsigned char out[EVP_MAX_MD_SIZE];
    unsigned int outlen = 0;

    HMAC(EVP_sha256(),
         reinterpret_cast<const unsigned char*>(key.data()), (int)key.size(),
         reinterpret_cast<const unsigned char*>(data.data()), data.size(),
         out, &outlen);

    std::ostringstream oss;
    for (unsigned int i = 0; i < outlen; i++) {
        oss << std::hex << std::setw(2) << std::setfill('0') << (int)out[i];
    }
    return oss.str();
}

}
