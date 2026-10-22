// expire_wrap_demo.go
// Bit-exact mirror of the C++ demo (expire_wrap_demo.cpp):
// Go integer arithmetic (uint32 modulo 2^32, int32 truncation) is identical to the
// *defined* parts of the C++ evaluation on MSVC/GCC/x86 (unsigned wrap is modulo in
// C++ too; unsigned->int narrowing wraps modulo 2^32 since C++20 / on all x86 compilers).
// Build&Run: go run ./expire_wrap_demo
package main

import "fmt"

type Entity struct {
	m_expireTime int32         // the suspect field: 32-bit signed (as in the C++ struct)
	m_sLogDesc   [16]byte
}

func main() {
	const nExitTime int32 = 10 // seconds
	const d uint32 = 10 * 1000 // nExitTime * 1000 (ms)

	// ---- A: 25.0 days uptime: tick = 2,160,000,000ms > INT_MAX(2,147,483,647ms = 24.855d) ----
	now := uint32(2160000000)
	raw := now + d                       // C++: unsigned add, modulo 2^32, no wrap yet
	var e Entity
	e.m_expireTime = int32(raw)          // C++: implicit unsigned->int narrowing -> negative
	guard := e.m_expireTime > 0
	fmt.Printf("[A] raw=%d stored=%d guard(expire>0)=%v => expiry branch %v\n",
		raw, e.m_expireTime, guard, map[bool]string{true: "runs", false: "never runs => item NEVER expires (leak)"}[guard])

	// ---- B: same store, but check is `now >= m_expireTime` (no guard): int->unsigned conversion ----
	asUnsigned := uint32(e.m_expireTime) // C++ usual arithmetic conversions: int -> unsigned
	fmt.Printf("[B] now=%d asUnsigned(expire)=%d -> expired=%v\n",
		now, asUnsigned, now >= asUnsigned)

	// ---- C: item created 1006ms before the 49.7-day tick wrap ----
	now2 := uint32(4294966290)
	expire2 := now2 + d                  // wraps to 8994
	fmt.Printf("[C] now2=%d expire2=%d -> expired at birth=%v\n",
		now2, expire2, now2 >= expire2)

	// ---- D: after full tick wrap, old item's expireTime still near 2^32 ----
	now3, expire3 := uint32(1000), uint32(4294000000)
	fmt.Printf("[D] now3=%d expire3=%d -> expired=%v (old item immortal => leak)\n",
		now3, expire3, now3 >= expire3)

	// ---- E: classic mixed-sign comparison trap (C++ silently converts int32 -> uint32) ----
	m1, u1 := int32(-1), uint32(1)
	fmt.Printf("[E] (-1 < 1u) = %v   // -1 converts to 4294967295\n", uint32(m1) < u1)

	// ---- F: 64-bit tick source stored into int32 (C++ converts int32 -> int64 for the compare) ----
	now64 := int64(2160000000)
	exp32 := int32(now64 + 10000)        // -2134957296
	fmt.Printf("[F] now64=%d exp32=%d -> (now64 >= exp32)=%v => delete-on-spawn\n",
		now64, exp32, now64 >= int64(exp32))

	// ---- G: 49.7-day wrap moment, SIGNED-domain check (getter returns int / local int copy) ----
	// Item created at day 40: tick was 3,456,000,000ms -> stored int32 = -838,967,296
	expOld := int32(3456000000 - 4294967296)   // -838967296
	nowAfterWrap := int32(2000)                // tick wrapped: now is a small positive int
	fmt.Printf("[G] signed check: now=%d expire=%d -> (now >= expire)=%v => MASS VANISH at wrap\n",
		nowAfterWrap, expOld, nowAfterWrap >= expOld)

	// ---- H: same item, UNSIGNED-domain check ----
	fmt.Printf("[H] unsigned check: now=%d expire=%d -> (now >= expire)=%v => immortal (leak)\n",
		uint32(2000), uint32(expOld), uint32(2000) >= uint32(expOld))

	// ---- I: item created BEFORE 24.855d, tick later reads as signed negative ----
	expOld2 := int32(1728010000)               // day-20 item: expire still positive
	nowNeg := int32(-2134967296)               // tick > 2^31 read as signed
	fmt.Printf("[I] signed check: now=%d expire=%d -> (now >= expire)=%v => pre-24.855d items become immortal\n",
		nowNeg, expOld2, nowNeg >= expOld2)
}
