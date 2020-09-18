// +build amd64,!gccgo,!appengine

#include "textflag.h"

// func transform(dst, src *[729]int8, rounds uint)
TEXT ·transform(SB),NOSPLIT,$0
    MOVQ dst+0(FP), AX
    MOVQ src+8(FP), BX
    MOVQ rounds+16(FP), CX      // r = rounds
    TESTQ CX, CX                // if r == 0 goto DONE
    JZ DONE

    MOVQ $·Indices(SB), R8      // var Indices [730]int
    MOVQ $·TruthTable(SB), R9   // var TruthTable [11]int8

LOOP:
    MOVQ (R8), SI               // tmp = src[Indices[0]]
    MOVBLZX (BX)(SI*1), SI

    XORQ DX, DX                 // i := 0

ROUND:
    MOVQ SI, R10                // s0 = tmp
    MOVQ 8(R8)(DX*8), R11       // s1 = src[Indices[i+1]]
    MOVBLZX (BX)(R11*1), R11
    MOVQ 16(R8)(DX*8), R12      // s1 = src[Indices[i+2]]
    MOVBLZX (BX)(R12*1), R12
    MOVQ 24(R8)(DX*8), R13      // s2 = src[Indices[i+3]]
    MOVBLZX (BX)(R13*1), R13
    MOVQ R13, SI                // tmp = s3

    SHLB $2, R13B               // s3 = s3<<2 + 5 + s2
    ADDB $5, R13B
    ADDB R12B, R13B
    SHLB $2, R12B               // s2 = s2<<2 + 5 + s1
    ADDB $5, R12B
    ADDB R11B, R12B
    SHLB $2, R11B               // s1 = s1<<2 + 5 + s0
    ADDB $5, R11B
    ADDB R10B, R11B

    MOVBLZX (R9)(R11*1), R11    // dst[i] = TruthTable[s1]
    MOVB R11B, (AX)(DX*1)
    MOVBLZX (R9)(R12*1), R12    // dst[i+1] = TruthTable[s2]
    MOVB R12B, 1(AX)(DX*1)
    MOVBLZX (R9)(R13*1), R13    // dst[i+2] = TruthTable[s3]
    MOVB R13B, 2(AX)(DX*1)

    ADDQ $3, DX                 // i += 3
    CMPQ DX, $727               // if i < 727 goto ROUND
    JL ROUND

    XCHGQ AX, BX                // src, dst = dst, src

    DECQ CX                     // r--
    JNZ LOOP                    // if r != 0 goto LOOP

DONE:
    RET
