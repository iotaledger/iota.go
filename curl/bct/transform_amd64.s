// +build amd64,!gccgo,!appengine

#include "textflag.h"

#define SBOX(la, ha, lb, hb) \
    MOVQ lb, R14; \             // a := (lb ^ ha) & la
    XORQ ha, R14; \
    ANDQ la, R14; \
    MOVQ hb, R15; \             // b := (hb ^ la) | a
    XORQ la, R15; \
    ORQ R14, R15

// func transform(lto, hto, lfrom, hfrom *[729]uint)
TEXT ·transform(SB),NOSPLIT,$0
    MOVQ lto+0(FP), AX
    MOVQ hto+8(FP), BX
    MOVQ lfrom+16(FP), CX
    MOVQ hfrom+24(FP), DX
    MOVQ $81, SI                // r := curl.NumRounds

ROUND:
    MOVQ (CX), R10              // l0 := lfrom[0]
    MOVQ (DX), R11              // h0 := hfrom[0]
    MOVQ 2912(CX), R12          // l1 := lfrom[364]
    MOVQ 2912(DX), R13          // h1 := hfrom[364]

    SBOX(R10, R11, R12, R13)    // a, b := sBox(l0, h0, l1, h1)
    MOVQ R15, (BX)              // hto[0] = b
    NOTQ R14                    // lto[0] = ^a
    MOVQ R14, (AX)

    MOVQ $364, R8               // t := 364
    MOVQ $1, DI                 // i := 1

LOOP:
    MOVQ 2912(CX)(R8*8), R10    // l0 = lfrom[t+364]
    MOVQ 2912(DX)(R8*8), R11    // h0 = hfrom[t+364]

    SBOX(R12, R13, R10, R11)    // a, b = sBox(l1, h1, l0, h0)
    MOVQ R15, (BX)(DI*8)        // hto[i] = b
    NOTQ R14                    // lto[i] = ^a
    MOVQ R14, (AX)(DI*8)

    MOVQ -8(CX)(R8*8), R12       // l1 := lfrom[t-1]
    MOVQ -8(DX)(R8*8), R13       // h1 := hfrom[t-1]

    SBOX(R10, R11, R12, R13)    // a, b = sBox(l0, h0, l1, h1)
    MOVQ R15, 8(BX)(DI*8)       // hto[i+1] = b
    NOTQ R14                    // lto[i+1] = ^a
    MOVQ R14, 8(AX)(DI*8)

    MOVQ 2904(CX)(R8*8), R10    // l0 = lfrom[t+363]
    MOVQ 2904(DX)(R8*8), R11    // h0 = hfrom[t+363]

    SBOX(R12, R13, R10, R11)    // a, b = sBox(l1, h1, l0, h0)
    MOVQ R15, 16(BX)(DI*8)      // hto[i+2] = b
    NOTQ R14                    // lto[i+2] = ^a
    MOVQ R14, 16(AX)(DI*8)

    MOVQ -16(CX)(R8*8), R12     // l1 := lfrom[t-2]
    MOVQ -16(DX)(R8*8), R13     // h1 := hfrom[t-2]

    SBOX(R10, R11, R12, R13)    // a, b = sBox(l0, h0, l1, h1)
    MOVQ R15, 24(BX)(DI*8)      // hto[i+3] = b
    NOTQ R14                    // lto[i+3] = ^a
    MOVQ R14, 24(AX)(DI*8)

    SUBQ $2, R8                 // t -= 2
    ADDQ $4, DI                 // i += 4
    CMPQ DI, $726               // if i < 726 goto ROUND
    JL LOOP

    XCHGQ AX, CX                // lfrom, lto = lto, lfrom
    XCHGQ BX, DX                // hfrom, hto = hto, hfrom

    DECQ SI                     // r--
    JNZ ROUND                   // if r != 0 goto LOOP

    RET
