package curl

func transformGeneric(dst, src *[StateSize]int8, rounds int) {
	for r := rounds; r > 0; r-- {
		// nine Curl-P rounds unrolled
		for i := 0; i < StateSize-8; i += 9 {
			s0 := src[Indices[i+0]]
			s1 := src[Indices[i+1]]
			s2 := src[Indices[i+2]]
			s3 := src[Indices[i+3]]
			s4 := src[Indices[i+4]]
			s5 := src[Indices[i+5]]
			s6 := src[Indices[i+6]]
			s7 := src[Indices[i+7]]
			s8 := src[Indices[i+8]]
			s9 := src[Indices[i+9]]

			dst[i+0] = TruthTable[s0+(s1<<2)+5]
			dst[i+1] = TruthTable[s1+(s2<<2)+5]
			dst[i+2] = TruthTable[s2+(s3<<2)+5]
			dst[i+3] = TruthTable[s3+(s4<<2)+5]
			dst[i+4] = TruthTable[s4+(s5<<2)+5]
			dst[i+5] = TruthTable[s5+(s6<<2)+5]
			dst[i+6] = TruthTable[s6+(s7<<2)+5]
			dst[i+7] = TruthTable[s7+(s8<<2)+5]
			dst[i+8] = TruthTable[s8+(s9<<2)+5]
		}
		// swap buffers
		src, dst = dst, src
	}
}
