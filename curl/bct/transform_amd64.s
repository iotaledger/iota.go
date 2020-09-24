// +build amd64,!gccgo,!appengine

#include "textflag.h"

// func transform(lto, hto, lfrom, hfrom *[729]uint, rounds uint)
TEXT ·transform(SB),NOSPLIT,$0
    MOVQ lto+0(FP), AX
    MOVQ hto+8(FP), BX
    MOVQ lfrom+16(FP), CX
    MOVQ hfrom+24(FP), DX
    MOVQ rounds+32(FP), SI      // r = rounds
    TESTQ SI, SI                // if r == 0 goto DONE
    JZ DONE

LOOP:
    MOVQ $·Indices(SB), R8
    MOVQ (R8), R8
    MOVQ (CX)(R8*8), R14        // ltmp := lfrom[Indices[0]]
    MOVQ (DX)(R8*8), R15        // htmp := hfrom[Indices[0]]

    XORQ DI, DI                 // i := 0

ROUND:
    MOVQ $·Indices(SB), R12
    MOVQ 8(R12)(DI*8), R9       // t1 := Indices[i+1]
    MOVQ 16(R12)(DI*8), R10     // t2 := Indices[i+2]
    MOVQ 24(R12)(DI*8), R11     // t3 := Indices[i+3]

    MOVQ R14, R12               // l0 := ltmp
    MOVQ R15, R13               // h0 := htmp
    MOVQ (CX)(R9*8), R14        // l1 := lfrom[t1]
    MOVQ (DX)(R9*8), R15        // h1 := hfrom[t1]
    MOVQ (CX)(R10*8), R8
    MOVQ (DX)(R10*8), R9
    MOVQ (CX)(R11*8), R10
    MOVQ (DX)(R11*8), R11

    XORQ R14, R13               // a0 := (l1 ^ h0) & l0
    ANDQ R12, R13
    XORQ R15, R12               // b0 := (h1 ^ l0) | a0
    ORQ R13, R12
    MOVQ R12, (BX)(DI*8)        // hto[i] = b0
    NOTQ R13                    // lto[i] = ^a0
    MOVQ R13, (AX)(DI*8)

    XORQ R8, R15
    ANDQ R14, R15
    XORQ R9, R14
    ORQ R15, R14
    MOVQ R14, 8(BX)(DI*8)
    NOTQ R15
    MOVQ R15, 8(AX)(DI*8)

    XORQ R10, R9
    ANDQ R8, R9
    XORQ R11, R8
    ORQ R9, R8
    MOVQ R8, 16(BX)(DI*8)
    NOTQ R9
    MOVQ R9, 16(AX)(DI*8)

    MOVQ R10, R14               // ltmp = lfrom[t3]
    MOVQ R11, R15               // htmp = hfrom[t3]

    ADDQ $3, DI                 // i += 3
    CMPQ DI, $727               // if i < 727 goto ROUND
    JL ROUND

    XCHGQ AX, CX                // lfrom, lto = lto, lfrom
    XCHGQ BX, DX                // hfrom, hto = hto, hfrom

    DECQ SI                     // r--
    JNZ LOOP                    // if r != 0 goto LOOP

DONE:
    RET
