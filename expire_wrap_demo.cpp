// expire_wrap_demo.cpp
// Demonstrate the C++ semantics of:  m_expireTime(int) = GetLogicCurrTickCount() + nExitTime*1000;
// Build: g++ -std=c++11 -O0 -Wall -Wextra -Wsign-compare -Wconversion expire_wrap_demo.cpp -o demo
#include <cstdio>
#include <cstdint>

typedef uint32_t DWORD;               // typical GetLogicCurrTickCount()/GetTickCount() return type

struct Entity {
    int  m_expireTime;                // the suspect field: 32-bit signed
    char m_sLogDesc[16];
};

int main() {
    const int   nExitTime = 10;       // seconds
    const DWORD d         = 10u * 1000u;

    // ---- A: 25.0 days uptime (tick=2,160,000,000ms > INT_MAX=2,147,483,647ms, i.e. > 24.855 days) ----
    DWORD now = 2160000000u;
    DWORD raw = now + d;                     // unsigned add: modulo 2^32, no wrap yet
    Entity e;
    e.m_expireTime = raw;                    // implicit unsigned->int narrowing (impl-defined pre-C++20, modulo since C++20)
    std::printf("[A] raw=%u stored=%d guard(expire>0)=%s => expiry branch %s\n",
        (unsigned)raw, e.m_expireTime, e.m_expireTime > 0 ? "true" : "FALSE",
        e.m_expireTime > 0 ? "runs" : "never runs => item NEVER expires (leak)");

    // ---- B: same store, but check is `now >= m_expireTime` (no guard) ----
    bool expiredB = (now >= (DWORD)e.m_expireTime);   // int->unsigned: modulo restores the bit pattern
    std::printf("[B] now=%u asUnsigned(expire)=%u -> expired=%s\n",
        (unsigned)now, (unsigned)e.m_expireTime, expiredB ? "true" : "false");

    // ---- C: item created 1006ms before the 49.7-day tick wrap ----
    DWORD now2 = 4294966290u, expire2 = now2 + d;     // expire2 wraps to 8994
    std::printf("[C] now2=%u expire2=%u -> expired at birth=%s\n",
        (unsigned)now2, (unsigned)expire2, (now2 >= expire2) ? "true" : "false");

    // ---- D: after full tick wrap, old item's expireTime still near 2^32 ----
    DWORD now3 = 1000u, expire3 = 4294000000u;
    std::printf("[D] now3=%u expire3=%u -> expired=%s (old item immortal => leak)\n",
        (unsigned)now3, (unsigned)expire3, (now3 >= expire3) ? "true" : "false");

    // ---- E: classic mixed-sign comparison trap ----
    std::printf("[E] (-1 < 1u) = %d   // -1 converts to 4294967295\n", (-1 < 1u));

    // ---- F: if the tick source is already 64-bit but stored into int32 ----
    int64_t now64 = 2160000000LL;
    int32_t exp32 = (int32_t)(now64 + 10000LL);       // -2134957296
    std::printf("[F] now64=%lld exp32=%d -> (now64 >= exp32)=%s => delete-on-spawn\n",
        (long long)now64, exp32, (now64 >= exp32) ? "true" : "false");
    return 0;
}
