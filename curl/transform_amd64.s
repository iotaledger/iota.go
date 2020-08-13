// +build !gccgo,!appengine

#include "textflag.h"

// func transform(dst, src *[729]int8, rounds int)
TEXT ·transform(SB),NOSPLIT,$0
    MOVQ dst+0(FP), R10
    MOVQ src+8(FP), R11
    MOVQ rounds+16(FP), DI     // r := rounds
    JZ DONE                    // if r == 0 goto DONE

    MOVQ $·Indices(SB), R8     // var Indices [730]int
    MOVQ $·TruthTable(SB), R9  // var TruthTable [11]int8

LOOP:
    XORQ SI, SI                // i := 0

ROUND:                         // three Curl-P rounds unrolled
    MOVQ 0(R8)(SI*8), R12
    MOVBQSX 0(R11)(R12*1), R12
    MOVQ 8(R8)(SI*8), R13
    MOVBQSX 0(R11)(R13*1), R13
    MOVQ 16(R8)(SI*8), R14
    MOVBQSX 0(R11)(R14*1), R14
    MOVQ 24(R8)(SI*8), R15
    MOVBQSX 0(R11)(R15*1), R15

    SALQ $2, R15
    ADDQ $5, R15
    ADDQ R14, R15
    SALQ $2, R14
    ADDQ $5, R14
    ADDQ R13, R14
    SALQ $2, R13
    ADDQ $5, R13
    ADDQ R12, R13

    MOVBQSX 0(R9)(R13*1), R13
    MOVB R13B, 0(R10)(SI*1)
    MOVBQSX 0(R9)(R14*1), R14
    MOVB R14B, 1(R10)(SI*1)
    MOVBQSX 0(R9)(R15*1), R15
    MOVB R15B, 2(R10)(SI*1)

    ADDQ $3, SI                // i+=3
    CMPQ SI, $727              // if i < 727 goto ROUND
    JL ROUND

    XCHGQ R10, R11             // swap src, dst

    DECQ DI                    // r--
    JG LOOP                    // if r > 0 goto LOOP

DONE:
    RET
