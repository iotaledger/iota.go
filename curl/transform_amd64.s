// +build !gccgo,!appengine

#include "textflag.h"

// func transform(dst, src *[729]int8, rounds uint)
TEXT ·transform(SB),NOSPLIT,$0
    MOVQ dst+0(FP), AX
    MOVQ src+8(FP), BX
    MOVQ rounds+16(FP), SI     // r := rounds
    JZ DONE                    // if r == 0 goto DONE

    MOVQ $·Indices(SB), R8     // var Indices [730]int
    MOVQ $·TruthTable(SB), R9  // var TruthTable [11]int8

LOOP:
    XORQ DI, DI                // i := 0

ROUND:                         // three Curl-P rounds unrolled
    MOVQ 0(R8)(DI*8), R10
    MOVBQSX 0(BX)(R10*1), R10
    MOVQ 8(R8)(DI*8), R11
    MOVBQSX 0(BX)(R11*1), R11
    MOVQ 16(R8)(DI*8), R12
    MOVBQSX 0(BX)(R12*1), R12
    MOVQ 24(R8)(DI*8), R13
    MOVBQSX 0(BX)(R13*1), R13

    SALQ $2, R13
    ADDQ $5, R13
    ADDQ R12, R13
    SALQ $2, R12
    ADDQ $5, R12
    ADDQ R11, R12
    SALQ $2, R11
    ADDQ $5, R11
    ADDQ R10, R11

    MOVBQSX 0(R9)(R11*1), R11
    MOVB R11B, 0(AX)(DI*1)
    MOVBQSX 0(R9)(R12*1), R12
    MOVB R12B, 1(AX)(DI*1)
    MOVBQSX 0(R9)(R13*1), R13
    MOVB R13B, 2(AX)(DI*1)

    ADDQ $3, DI                // i+=3
    CMPQ DI, $727              // if i < 727 goto ROUND
    JL ROUND

    XCHGQ AX, BX               // swap src, dst

    DECQ SI                    // r--
    JG LOOP                    // if r > 0 goto LOOP

DONE:
    RET
