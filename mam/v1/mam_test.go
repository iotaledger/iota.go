package mam_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/converter"
	. "github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/mam/v1"
	. "github.com/iotaledger/iota.go/merkle"
	"github.com/iotaledger/iota.go/trinary"
)

const (
	seed           = "TX9XRR9SRCOBMTYDTMKNEIJCSZIMEUPWCNLC9DPDZKKAEMEFVSTEVUFTRUZXEHLULEIYJIEOWIC9STAHW"
	sideKeyPublic  = ""
	sideKeyPrivate = "QOLOACG9BNUYLERQTZPPW9VKIOPDRTPMFZCYWGNVKIZJEYBWJDXASOXNDMZGBNYFVBCFBQBXSCCAFFRIO"
	message        = "{\"message\":\"Message from Alice\",\"timestamp\":\"2019-4-8 22:41:01\"}"
)

var _ = Describe("MAMCreate", func() {

	Context("MAMCreate and MAMParse", func() {

		It("Mam", func() {
			var index uint64 = 7
			var start uint64 = 0
			var count uint64 = 8
			var nextCount uint64 = 1
			nextStart := start + count

			for _, sideKey := range []string{sideKeyPublic, sideKeyPrivate} {
				for _, security := range []SecurityLevel{SecurityLevelLow, SecurityLevelMedium, SecurityLevelHigh} {
					treeSize := MerkleSize(count)
					messageTrytes, err := converter.ASCIIToTrytes(message)

					payloadLength := PayloadMinLength(uint64(len(messageTrytes)*3), treeSize*uint64(HashTrinarySize), index, security)

					merkleTree, err := MerkleCreate(count, seed, start, security, NewCurlP27())

					Expect(err).ToNot(HaveOccurred())

					nextRoot, err := MerkleCreate(nextCount, seed, nextStart, security, NewCurlP27())
					Expect(err).ToNot(HaveOccurred())

					payload, payloadLength, err := MAMCreate(payloadLength, messageTrytes, sideKey, merkleTree, treeSize*HashTrinarySize, count, index, nextRoot, start, seed, security)
					Expect(err).ToNot(HaveOccurred())

					parsedIndex, parsedNextRoot, parsedMessageTrytes, parsedSecurity, err := MAMParse(payload, payloadLength, sideKey, merkleTree)
					Expect(err).ToNot(HaveOccurred())

					parsedMessage, err := converter.TrytesToASCII(parsedMessageTrytes)
					Expect(index).To(Equal(parsedIndex))
					Expect(nextRoot).To(Equal(trinary.MustTrytesToTrits(parsedNextRoot)))
					Expect(message).To(Equal(parsedMessage))
					Expect(security).To(Equal(parsedSecurity))
				}
			}
		})

	})
})
