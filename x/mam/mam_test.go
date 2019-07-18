package mam_test

import (
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/converter"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/iotaledger/iota.go/x/mam"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	senderAPI *mam.MAM
	chID      string
	epID      string
	ch1ID     string
	ep1ID     string
)

const (
	testPreSharedKeyAStr   = "PSKIDAPSKIDAPSKIDAPSKIDAPSK"
	testPreSharedKeyANonce = "PSKANONCE"
	testPreSharedKeyBStr   = "PSKIDBPSKIDBPSKIDBPSKIDBPSK"
	testPreSharedKeyBNonce = "PSKBNONCE"
	testNTRUNonce          = "NTRUBNONCE"
	testMSSDepth           = 6
	apiSeed                = "APISEEDAPISEEDAPISEEDAPISEEDAPISEEDAPISEEDAPISEEDAPISEEDAPISEEDAPISEEDAPISEEDAPI9"
)

func testAPIWriteHeader(
	mamAPI *mam.MAM, pskA *mam.PSK, pskB *mam.PSK, ntruPubKey *mam.NTRUPK,
	msgPubKey mam.MsgPubKey, msgKeyload mam.MsgKeyload, bndl bundle.Bundle) (bundle.Bundle, trinary.Trits) {

	psks := []mam.PSK{}
	ntruPubKeys := []mam.NTRUPK{}

	if msgKeyload == mam.MsgKeyloadPSK {
		psks = append(psks, *pskA, *pskB)
	} else if msgKeyload == mam.MsgKeyloadNTRU {
		ntruPubKeys = append(ntruPubKeys, *ntruPubKey)
	}

	var err error
	var modBndl bundle.Bundle
	var msgIDTrits trinary.Trits

	if msgPubKey == mam.MsgPubKeyEndpointID {
		modBndl, msgIDTrits, err = mamAPI.BundleWriteHeaderOnEndpoint(bndl, chID, epID, psks, ntruPubKeys)
		Expect(err).ToNot(HaveOccurred())
	} else if msgPubKey == mam.MsgPubKeyChannelID1 {
		remainingSks := mamAPI.ChannelRemainingSecretKeys(chID)
		modBndl, msgIDTrits, err = mamAPI.BundleAnnounceChannel(bndl, chID, ch1ID, psks, ntruPubKeys)
		Expect(err).ToNot(HaveOccurred())
		Expect(mamAPI.ChannelRemainingSecretKeys(chID)).To(Equal(remainingSks - 1))
	} else if msgPubKey == mam.MsgPubKeyEndpointID1 {
		remainingSks := mamAPI.ChannelRemainingSecretKeys(chID)
		modBndl, msgIDTrits, err = mamAPI.BundleAnnounceEndpoint(bndl, chID, ep1ID, psks, ntruPubKeys)
		Expect(err).ToNot(HaveOccurred())
		Expect(mamAPI.ChannelRemainingSecretKeys(chID)).To(Equal(remainingSks - 1))
	} else {
		modBndl, msgIDTrits, err = mamAPI.BundleWriteHeaderOnChannel(bndl, chID, psks, ntruPubKeys)
		Expect(err).ToNot(HaveOccurred())
	}

	// TODO: look for freeing

	return modBndl, msgIDTrits
}

func testAPIWritePacket(
	mamAPI *mam.MAM, bndl bundle.Bundle, msgID trinary.Trits, checksum mam.MsgChecksum,
	payload string, isLastPacket bool, msgPubKey mam.MsgPubKey) bundle.Bundle {
	payloadTrytes, err := converter.ASCIIToTrytes(payload)
	Expect(err).ToNot(HaveOccurred())

	var remainingSKs int
	if checksum == mam.MsgChecksumSig {
		if msgPubKey == mam.MsgPubKeyChannelID {
			remainingSKs = mamAPI.ChannelRemainingSecretKeys(chID)
		} else if msgPubKey == mam.MsgPubKeyEndpointID {
			remainingSKs = mamAPI.EndpointRemainingSecretKeys(chID, epID)
		} else if msgPubKey == mam.MsgPubKeyChannelID1 {
			remainingSKs = mamAPI.ChannelRemainingSecretKeys(ch1ID)
		} else if msgPubKey == mam.MsgPubKeyEndpointID1 {
			remainingSKs = mamAPI.EndpointRemainingSecretKeys(chID, ep1ID)
		}
	}

	bndl, err = mamAPI.BundleWritePacket(msgID, payloadTrytes, checksum, isLastPacket, bndl)
	Expect(err).ToNot(HaveOccurred())

	if checksum == mam.MsgChecksumSig {
		if msgPubKey == mam.MsgPubKeyChannelID {
			Expect(mamAPI.ChannelRemainingSecretKeys(chID)).To(Equal(remainingSKs - 1))
		} else if msgPubKey == mam.MsgPubKeyEndpointID {
			Expect(mamAPI.EndpointRemainingSecretKeys(chID, epID)).To(Equal(remainingSKs - 1))
		} else if msgPubKey == mam.MsgPubKeyChannelID1 {
			Expect(mamAPI.ChannelRemainingSecretKeys(ch1ID)).To(Equal(remainingSKs - 1))
		} else if msgPubKey == mam.MsgPubKeyEndpointID1 {
			Expect(mamAPI.EndpointRemainingSecretKeys(chID, ep1ID)).To(Equal(remainingSKs - 1))
		}
	}

	return bndl
}

func testAPIReadMsg(mam *mam.MAM, bndl bundle.Bundle) (string, bool) {
	msg, lastPacket, err := mam.BundleRead(bndl)
	Expect(err).ToNot(HaveOccurred())
	msgASCII, err := converter.TrytesToASCII(msg)
	Expect(err).ToNot(HaveOccurred())
	return msgASCII, lastPacket
}

func genericTests() {
	var err error
	var ntruSK *mam.NTRUSK
	var pskA, pskB *mam.PSK

	// generate recipient's spongos NTRU keys
	ntruSK = mam.NewNTRUSK(senderAPI, testNTRUNonce)

	err = senderAPI.AddNTRUSecretKey(ntruSK)
	Expect(err).ToNot(HaveOccurred())

	// generate pre shared keys
	pskA = mam.NewPSK(senderAPI, testPreSharedKeyAStr, testPreSharedKeyANonce)
	pskB = mam.NewPSK(senderAPI, testPreSharedKeyBStr, testPreSharedKeyBNonce)

	err = senderAPI.AddPreSharedKey(pskB)
	Expect(err).ToNot(HaveOccurred())

	/* chid=0, epid=1, chid1=2, epid1=3*/
	for pubKey := 0; pubKey < 4; pubKey++ {
		/* public=0, psk=1, ntru=2 */
		for keyload := 0; keyload < 3; keyload++ {
			/* none=0, mac=1, mssig=2 */
			for checksum := 0; checksum < 3; checksum++ {
				var msgID trinary.Trits
				bndl := bundle.Bundle{}

				// write packet
				bndl, msgID = testAPIWriteHeader(senderAPI, pskA, pskB, &ntruSK.PK, mam.MsgPubKey(pubKey), mam.MsgKeyload(keyload), bndl)
				bndl = testAPIWritePacket(senderAPI, bndl, msgID, mam.MsgChecksum(checksum), payload, true, mam.MsgPubKey(pubKey))

				// read packet
				msg, lastPacket := testAPIReadMsg(senderAPI, bndl)

				// assert
				Expect(lastPacket).To(BeTrue())
				Expect(msg).To(Equal(payload))
			}
		}
	}
}

func testMultiplePacketsRun(mamAPI *mam.MAM, numPackets int) {
	payloadIn := "RYFJ9ZCHZFYZSHCMBJPDHLQBODCMRMDH9CLGGVWJCZRJSNDWBTMWSBPYPFIIOLXEMKSJ9LJO" +
		"AFBFPJL9XMGGZXFCZHFDLOLODLMLNERWGUBXUJCMHJXWJPGX9R9HUPIHEIPNNMXTULSMHJGP" +
		"TDKYEFZ9FDKFOBBQS9QZ9LODBRQMVCKYIUXIPWMDNPSZEK9ZTDCJEFEQEAJAEABJQWPVCGSM" +
		"WEZLLSDYWLWQPRENHOTDQYP9QTIULYOE9OIVEWUXTMYUMPTMHOYFRJMWJUKM9QAJKVQW9ADY" +
		"TZFQTNDISJYNTSFIEESDZJJJUPBDJKNEYNMOIZOKFCARBCXVFTPQZEKVZBNZOSKMRHAJXG9U" +
		"ZBNEUXM9LARLTSRQXYACOVOCIFHUETWFXXOLSSQSKNFUANYIGVMXTOZYSYBIXTRTYOWQRVTF" +
		"VMMXSH9WVQKEYRALLTBJIBYJIMTV99PATCFBKXZPLIBPNQZYJLDUXRWKPRRJTPKWQAQFFQWS" +
		"VAHUOAAOWJLSYVYLI99RNUONKEEJYFIMEWBIKLGARGTABJDCHKQM9LFFMKQXHFSCGJXAYCLF" +
		"RFLWPKPPNHOWIMEFNRCGNDCHMEYYJWHPRJOYOFFPNISVUNMVYFW9ECUZBDOSUCFZPOREJMND" +
		"ZMZYWUBBFICWJ9IYHJDIDGLPERWCYXMFHXZGNLWXOCXBGEWZFKITEUMEVNUWLRUMHEZUJMRI" +
		"TTNKN9PBUR9MOZINMWWTRXVRRZHQVP9QDJPGBZALBVI9GXNZYQTOPKDJPXPLADTBUNRQFLTE" +
		"Q9XLMEPTJUWYIGNQMMLECGXAQOSFMDWFBFUYB9FEUMXSCRQVQMT9E9CEPRVQWQFVWT9UC9FH" +
		"NTCCRUHOOWXORIRHNNZUOQCSOGJCRUWCQHCLZMRNIWUESDEQWPHLLNEHXFDLRUEOTLQERNPT" +
		"OHNGGXIWJCKGKEGRFXYFLVOQVYQOVZ9QWGGBGZBLPVNQOBA9VYGKZE9MQYOHDKNE"

	var isLastPacket bool
	var payload trinary.Trytes
	bndl := bundle.Bundle{}

	// write and read header
	bndl, msgID, err := mamAPI.BundleWriteHeaderOnChannel(bndl, chID, nil, nil)
	Expect(err).ToNot(HaveOccurred())
	payload, _, err = mamAPI.BundleRead(bndl)
	Expect(err).ToNot(HaveOccurred())
	Expect(payload).To(BeEmpty())

	for i := 0; i < numPackets; i++ {
		remainingSks := mamAPI.ChannelRemainingSecretKeys(chID)

		isLastPacket = i == numPackets-1
		bndl = bundle.Bundle{}
		bndl, err = mamAPI.BundleWritePacket(msgID, payloadIn, mam.MsgChecksum(i%3), isLastPacket, bndl)
		Expect(err).ToNot(HaveOccurred())

		if mam.MsgChecksum(i) == mam.MsgChecksumSig {
			remainingSksAfter := mamAPI.ChannelRemainingSecretKeys(chID)
			Expect(remainingSksAfter).To(Equal(remainingSks - 1))
		}

		payload, isLastPacket, err = mamAPI.BundleRead(bndl)
		if i < numPackets-1 {
			Expect(isLastPacket).To(BeFalse())
		} else {
			Expect(isLastPacket).To(BeTrue())
		}

		Expect(payload).To(Equal(payloadIn))
	}
}

func testSerialization() {
	var deserializedAPI mam.MAM

	remainingSKs := senderAPI.ChannelRemainingSecretKeys(chID)

	serializedAPI := senderAPI.Serialize()
	Expect(deserializedAPI.Deserialize(serializedAPI)).ToNot(HaveOccurred())

	// should still have same remaining secret keys
	remainingSKsDeserializedAPI := deserializedAPI.ChannelRemainingSecretKeys(chID)
	Expect(remainingSKsDeserializedAPI).To(Equal(remainingSKs))

	testMultiplePacketsRun(&deserializedAPI, 2)
	Expect(deserializedAPI.Destroy()).ToNot(HaveOccurred())
}

func testAPISaveLoad() {
	var loadedAPI mam.MAM
	encrKeysTrytes := "NOPQRSTUVWXYZ9ABCDEFGHIJKLM" +
		"NOPQRSTUVWXYZ9ABCDEFGHIJKLM" +
		"NOPQRSTUVWXYZ9ABCDEFGHIJKLM"
	fileName := "mam-api.bin"
	remainingSks := senderAPI.ChannelRemainingSecretKeys(chID)

	Expect(senderAPI.Save(fileName, encrKeysTrytes)).ToNot(HaveOccurred())
	Expect(loadedAPI.Load(fileName, encrKeysTrytes)).ToNot(HaveOccurred())
	Expect(loadedAPI.ChannelRemainingSecretKeys(chID)).To(Equal(remainingSks))

	testMultiplePacketsRun(&loadedAPI, 2)
	Expect(loadedAPI.Destroy()).ToNot(HaveOccurred())
}

func testAPISaveLoadWithWrongKey() {
	var loadedAPI mam.MAM
	encrKeysTrytes := "NOPQRSTUVWXYZ9ABCDEFGHIJKLM" +
		"NOPQRSTUVWXYZ9ABCDEFGHIJKLM" +
		"NOPQRSTUVWXYZ9ABCDEFGHIJKLM"
	decrKeysTrytes := "MOPQRSTUVWXYZ9ABCDEFGHIJKLM" +
		"NOPQRSTUVWXYZ9ABCDEFGHIJKLM" +
		"NOPQRSTUVWXYZ9ABCDEFGHIJKLM"
	fileName := "mam-api.bin"

	Expect(senderAPI.Save(fileName, encrKeysTrytes)).ToNot(HaveOccurred())
	Expect(loadedAPI.Load(fileName, decrKeysTrytes)).To(HaveOccurred())
	Expect(loadedAPI.Destroy()).ToNot(HaveOccurred())
}

func testAPITrust() {
	var receiverAPI mam.MAM

	Expect(receiverAPI.Init(apiSeed)).ToNot(HaveOccurred())
	bndl := bundle.Bundle{}

	// we check that the read fails due to the channel not being trusted
	bndl, msgID, err := senderAPI.BundleWriteHeaderOnChannel(bndl, chID, nil, nil)
	Expect(err).ToNot(HaveOccurred())
	bndl, err = senderAPI.BundleWritePacket(msgID, "PAYLOAD", mam.MsgChecksumNone, true, bndl)
	Expect(err).ToNot(HaveOccurred())
	_, _, err = receiverAPI.BundleRead(bndl)
	Expect(err).To(Equal(mam.ErrMAMChannelNotTrusted))

	// we trust the channel and check that the read now succeeds
	Expect(receiverAPI.AddTrustedChannel(chID)).ToNot(HaveOccurred())
	_, _, err = receiverAPI.BundleRead(bndl)
	Expect(err).ToNot(HaveOccurred())

	// we check that the read fails due to the endpoint not being trusted
	bndl = bundle.Bundle{}
	bndl, msgID, err = senderAPI.BundleWriteHeaderOnEndpoint(bndl, chID, epID, nil, nil)
	Expect(err).ToNot(HaveOccurred())
	bndl, err = senderAPI.BundleWritePacket(msgID, "PAYLOAD", mam.MsgChecksumNone, true, bndl)
	Expect(err).ToNot(HaveOccurred())
	_, _, err = receiverAPI.BundleRead(bndl)
	Expect(err).To(Equal(mam.ErrMAMEndpointNotTrusted))

	// we trust the endpoint by announcing it and reading the same packet on the receiver side
	// which automatically updates our trusted endpoints as we trust the given channel
	bndl = bundle.Bundle{}
	bndl, msgID, err = senderAPI.BundleAnnounceEndpoint(bndl, chID, epID, nil, nil)
	Expect(err).ToNot(HaveOccurred())
	_, _, err = receiverAPI.BundleRead(bndl)
	Expect(err).ToNot(HaveOccurred())

	// reading the packet on the new endpoint should now succeed on the receiver side
	bndl = bundle.Bundle{}
	bndl, msgID, err = senderAPI.BundleWriteHeaderOnEndpoint(bndl, chID, epID, nil, nil)
	Expect(err).ToNot(HaveOccurred())
	_, _, err = receiverAPI.BundleRead(bndl)
	Expect(err).ToNot(HaveOccurred())

	// we check that the read fails due to the channel not being trusted
	bndl = bundle.Bundle{}
	bndl, msgID, err = senderAPI.BundleWriteHeaderOnChannel(bndl, ch1ID, nil, nil)
	Expect(err).ToNot(HaveOccurred())
	bndl, err = senderAPI.BundleWritePacket(msgID, "PAYLOAD", mam.MsgChecksumNone, true, bndl)
	Expect(err).ToNot(HaveOccurred())
	_, _, err = receiverAPI.BundleRead(bndl)
	Expect(err).To(Equal(mam.ErrMAMChannelNotTrusted))

	// we trust the new channel by announcing it through the trusted channel
	bndl = bundle.Bundle{}
	bndl, msgID, err = senderAPI.BundleAnnounceChannel(bndl, chID, ch1ID, nil, nil)
	Expect(err).ToNot(HaveOccurred())
	_, _, err = receiverAPI.BundleRead(bndl)
	Expect(err).ToNot(HaveOccurred())

	// reading the packet which was sent to the new channel should now success on the receiver side
	bndl = bundle.Bundle{}
	bndl, msgID, err = senderAPI.BundleWriteHeaderOnChannel(bndl, ch1ID, nil, nil)
	Expect(err).ToNot(HaveOccurred())
	bndl, err = senderAPI.BundleWritePacket(msgID, "PAYLOAD", mam.MsgChecksumNone, true, bndl)
	Expect(err).ToNot(HaveOccurred())
	_, _, err = receiverAPI.BundleRead(bndl)
	Expect(err).ToNot(HaveOccurred())

	Expect(receiverAPI.Destroy()).ToNot(HaveOccurred())
}

func testCreateChannels(depth uint) {
	var err error

	By("creating a new channel")
	chID, err = senderAPI.ChannelCreate(depth)
	Expect(err).ToNot(HaveOccurred())
	remainingChannSKs := senderAPI.ChannelRemainingSecretKeys(chID)
	Expect(remainingChannSKs).To(Equal(mam.MSSMaxSKN(depth) + 1))

	By("creating a new endpoint on that channel")
	epID, err = senderAPI.EndpointCreate(depth, chID)
	Expect(err).ToNot(HaveOccurred())
	remainingEndpSKs := senderAPI.EndpointRemainingSecretKeys(chID, epID)
	Expect(remainingEndpSKs).To(Equal(mam.MSSMaxSKN(depth) + 1))

	By("creating a second endpoint on the same channel")
	ep1ID, err = senderAPI.EndpointCreate(depth, chID)
	Expect(err).ToNot(HaveOccurred())
	remainingEndp1SKs := senderAPI.EndpointRemainingSecretKeys(chID, ep1ID)
	Expect(remainingEndp1SKs).To(Equal(mam.MSSMaxSKN(depth) + 1))

	By("creating a second channel")
	ch1ID, err = senderAPI.ChannelCreate(depth)
	Expect(err).ToNot(HaveOccurred())
	remainingChann1SKs := senderAPI.ChannelRemainingSecretKeys(ch1ID)
	Expect(remainingChann1SKs).To(Equal(mam.MSSMaxSKN(depth) + 1))
}

var _ = Describe("MAM", func() {

	senderAPI = &mam.MAM{}

	Context("API", func() {
		It("works", func() {
			Expect(senderAPI.Init(apiSeed)).ToNot(HaveOccurred())
		})
	})

	Context("Channels", func() {
		It("works", func() {
			testCreateChannels(testMSSDepth)
		})
	})

	Context("Generic", func() {
		It("works", func() {
			genericTests()
		})
	})

	Context("Serialization", func() {
		It("works", func() {
			testSerialization()
		})
	})

	Context("Save and load", func() {
		It("works", func() {
			testAPISaveLoad()
		})
	})

	Context("Save and load with wrong keys", func() {
		It("returns errors", func() {
			testAPISaveLoadWithWrongKey()
		})
	})

	Context("Trust", func() {
		It("works", func() {
			testAPITrust()
		})
	})

	Context("Destroy", func() {
		It("works", func() {
			Expect(senderAPI.Destroy()).ToNot(HaveOccurred())
		})
	})

	Context("Multiple Packets", func() {
		It("works", func() {
			Expect(senderAPI.Init(apiSeed)).ToNot(HaveOccurred())
			testCreateChannels(4)
			testMultiplePacketsRun(senderAPI, 36)
			Expect(senderAPI.Destroy()).ToNot(HaveOccurred())
		})
	})
})

const payload = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec vel tortor " +
	"semper, lacinia augue eu, interdum ligula. Duis nec nulla arcu. Sed " +
	"commodo sem urna, vitae tristique mauris tempor at. Nullam vel faucibus " +
	"enim. In vel maximus urna, nec pretium ipsum. Nam vitae augue consequat, " +
	"lacinia lacus eget, maximus ante. Quisque egestas nibh ut sem molestie, " +
	"at ornare odio congue. In turpis libero, euismod eget euismod eget, " +
	"interdum pretium orci. Nulla facilisi. In faucibus, purus non " +
	"pellentesque ultrices, nisi urna finibus leo, id convallis lacus ante " +
	"quis nisi. Suspendisse vehicula est ipsum, at feugiat est congue vitae. " +
	"Duis id tincidunt purus. Ut ornare leo vel venenatis suscipit. Mauris " +
	"pulvinar sapien sit amet velit varius, quis condimentum orci lobortis. " +
	"Praesent eget est porta, rutrum eros eu, porta eros. Donec et felis a " +
	"metus imperdiet imperdiet. Cras a fringilla mi. Aliquam auctor dapibus " +
	"dolor, quis euismod augue condimentum ut. Curabitur a nisl id purus " +
	"efficitur tincidunt dapibus non arcu. Integer efficitur faucibus massa. " +
	"Ut vitae ultrices ex. In hendrerit tincidunt lacus, nec mollis urna " +
	"iaculis vitae. Aliquam volutpat sodales augue, sed posuere eros lobortis " +
	"vel. Cras rutrum elit nec gravida volutpat. Aliquam erat volutpat. Cras " +
	"dignissim tellus eget magna dapibus, in finibus ipsum facilisis. " +
	"Curabitur eleifend elit nibh, eu egestas tortor consequat ac. Nam " +
	"facilisis dolor sit amet lacus elementum, vitae porta libero iaculis. " +
	"Duis nisl diam, dignissim eu luctus pellentesque, mattis ac magna. Sed " +
	"malesuada dignissim orci vitae cursus. Pellentesque viverra massa vitae " +
	"neque tincidunt ornare. Phasellus iaculis lobortis sapien eu interdum. " +
	"Aenean facilisis a enim a malesuada. Maecenas ut arcu bibendum, interdum " +
	"odio ac, hendrerit est. Quisque dui diam, venenatis non purus eget, " +
	"volutpat facilisis nisi. Nullam mi urna, imperdiet eu eros tincidunt, " +
	"congue dapibus velit. Donec non eleifend nunc, sed sodales nisl. Aliquam " +
	"tristique pulvinar ante, nec tincidunt ex. Fusce eget porta arcu. Cras " +
	"quis ornare magna, ac ullamcorper dolor. Etiam hendrerit posuere " +
	"hendrerit. Donec sit amet fermentum quam, non euismod dui. Sed suscipit " +
	"posuere est ac tempor. Vivamus id aliquam urna. Integer cursus, tellus " +
	"nec malesuada varius, lectus justo maximus est, ac molestie dolor nulla " +
	"ut mauris. Suspendisse molestie ornare consequat. Nullam convallis ex eu " +
	"tortor pretium pulvinar. Cras in gravida augue, a porttitor libero. Morbi " +
	"orci sapien, egestas at consectetur non, egestas ut orci. Praesent at " +
	"nisl sollicitudin, elementum nunc sit amet, placerat urna. Aliquam ut " +
	"felis auctor, scelerisque mi ut, posuere arcu. Donec venenatis ac ex eget " +
	"suscipit. Quisque ut nibh vitae lacus interdum pretium ut ac tellus. Sed " +
	"cursus dui vel fringilla porta. Suspendisse in rhoncus lectus. Ut " +
	"interdum risus at consequat porta. Donec commodo accumsan magna quis " +
	"ornare. Pellentesque habitant morbi tristique senectus et netus et " +
	"malesuada fames ac turpis egestas. Donec sed mollis purus, eget viverra " +
	"ipsum. Mauris vulputate rutrum leo a lobortis. Mauris vitae semper nibh. " +
	"Maecenas vestibulum justo libero. Duis ornare nunc nec dolor auctor " +
	"congue. Phasellus a quam ultricies, pharetra purus ac, pretium ipsum. " +
	"Pellentesque dignissim nulla at dolor ullamcorper, at aliquam nunc " +
	"cursus. Proin in maximus ligula, sed imperdiet nunc. Proin eu ante neque. " +
	"Ut consectetur ultricies ipsum, et laoreet sapien iaculis rutrum. Sed " +
	"ultrices magna metus, in volutpat felis convallis lobortis. Cras " +
	"elementum mauris ut nibh tincidunt, sed venenatis ante lacinia. Nunc " +
	"ullamcorper ante sit amet congue vulputate. Vivamus magna odio, molestie " +
	"ut risus ac, lobortis vehicula mauris. Proin bibendum id sapien id " +
	"dictum. Duis eu porta elit. Nullam auctor, diam eu sagittis pretium, " +
	"nulla ligula pharetra ligula, in blandit tortor mi in nisl. Donec " +
	"accumsan dolor orci, a lacinia lectus sollicitudin a. Sed odio eros, " +
	"mattis eget lorem fringilla, egestas tempus turpis. Orci varius natoque " +
	"penatibus et magnis dis parturient montes, nascetur ridiculus mus. In hac " +
	"habitasse platea dictumst. Nulla facilisi. Fusce vel dolor accumsan, " +
	"sagittis velit vitae, pretium erat. Etiam tempor erat quis neque " +
	"vulputate, vel convallis velit porttitor. Etiam risus urna, consectetur " +
	"et justo ut, accumsan consectetur est. Proin eu eros et felis ultrices " +
	"suscipit eget et enim. Mauris sem neque, blandit sed mollis eget, feugiat " +
	"non mauris. Etiam pellentesque est nulla, ac ornare elit vulputate et. " +
	"Donec quis cursus lacus. Donec scelerisque tincidunt massa at rutrum. " +
	"Quisque commodo erat non cursus euismod. Pellentesque vitae luctus sem, " +
	"ac volutpat quam. Aliquam vitae consectetur nibh. Nunc eu pharetra velit. " +
	"Praesent at tincidunt felis. Aenean non ultrices neque. Duis finibus urna " +
	"ut est rhoncus pharetra. Curabitur commodo eget sem in posuere. Praesent " +
	"at blandit purus. In pretium urna vel purus commodo, ac sollicitudin odio " +
	"pretium. Praesent sit amet magna id mi tincidunt ultrices et vel lectus. " +
	"Aenean nec finibus est. Etiam eu purus ac ligula auctor malesuada a eget " +
	"purus. Curabitur convallis, ligula ut tristique rutrum, urna purus " +
	"pulvinar risus, a suscipit ante est ac eros. In ut dignissim mi. Aenean " +
	"vel posuere tellus. Phasellus eu sem in nunc malesuada aliquet. In at " +
	"ullamcorper ipsum. Aliquam erat volutpat. Pellentesque sagittis ligula " +
	"non venenatis mollis. Curabitur non cursus nulla. Phasellus vestibulum " +
	"venenatis sapien vel imperdiet. Integer id ultricies sem, id fringilla " +
	"dui. Aenean id massa id tortor finibus euismod id at diam. Integer " +
	"aliquam est metus, ac aliquam lacus auctor ut. Quisque vel convallis mi. " +
	"Sed ultrices nisi nec ipsum tempus blandit. Aenean at cursus justo, nec " +
	"feugiat erat. Etiam mollis in metus ac dignissim. Vestibulum sed massa " +
	"arcu. Nam condimentum id orci molestie ornare. Quisque semper non orci a " +
	"pharetra. Praesent id interdum massa. Morbi non porta libero. Nulla " +
	"dignissim purus enim, et elementum velit feugiat eu. Fusce dapibus est " +
	"eget dolor dignissim, vitae placerat risus condimentum. Praesent " +
	"tincidunt efficitur massa, porttitor mollis velit rhoncus id. Nullam " +
	"dignissim, nisi sed placerat faucibus, massa mauris ullamcorper ipsum, " +
	"nec tristique velit tellus quis tortor. Vestibulum ante ipsum primis in " +
	"faucibus orci luctus et ultrices posuere cubilia Curae; Donec a urna dui. " +
	"Sed et augue eget dolor efficitur volutpat sed a turpis. Integer cursus, " +
	"arcu vitae aliquam fermentum, lectus metus elementum elit, vitae suscipit " +
	"felis sapien tristique erat. Suspendisse et dui magna. Aenean pharetra " +
	"augue augue. Nam eu eros sed tortor hendrerit venenatis at nec sem. Etiam " +
	"in risus ac dolor lobortis ornare. Orci varius natoque penatibus et " +
	"magnis dis parturient montes, nascetur ridiculus mus. Mauris pellentesque " +
	"nibh ut imperdiet eleifend. Nam rhoncus rhoncus porta. Ut sed " +
	"pellentesque quam. Pellentesque lobortis convallis tincidunt. Praesent " +
	"non venenatis sapien. Nulla facilisi. Phasellus egestas tristique " +
	"sagittis. Fusce volutpat tristique rhoncus. Duis semper ipsum sed " +
	"volutpat lobortis. Suspendisse ut eros nunc. Pellentesque eu felis et " +
	"tortor cursus dictum. Aliquam viverra augue a fermentum lobortis. " +
	"Phasellus viverra placerat velit sed aliquet. Duis id rutrum sem. Vivamus " +
	"feugiat placerat ipsum at vehicula. Ut sit amet metus bibendum libero " +
	"semper cursus. Integer in quam eu leo interdum eleifend. Cras tristique " +
	"eros vel fermentum vestibulum. In hac habitasse platea dictumst. Donec id " +
	"libero ut elit tempor vestibulum. Nullam vel porttitor dolor. In pretium " +
	"bibendum posuere. Donec congue scelerisque hendrerit. Ut sed tincidunt " +
	"enim. Ut eleifend lectus sem. Vestibulum accumsan euismod purus ac " +
	"lacinia. Morbi ornare iaculis nulla, eu ultrices ante aliquam sit amet. " +
	"Phasellus tincidunt vehicula viverra. Mauris ut massa nulla. Maecenas sit " +
	"amet velit in justo blandit mattis eu quis massa. Sed vel tincidunt " +
	"metus. Curabitur euismod erat ac ante consectetur interdum. Aliquam erat " +
	"volutpat. Duis sem arcu, hendrerit in quam vitae, molestie vehicula " +
	"lorem. Morbi sed sagittis dui, vel gravida diam. In nulla ante, pharetra " +
	"vel leo sit amet, maximus commodo metus. Integer posuere nisi nisi, eu " +
	"dapibus nisl viverra ac. Etiam euismod sapien quis justo laoreet " +
	"lobortis. Donec sagittis est vel tincidunt blandit. Sed ac tortor eu eros " +
	"gravida ultricies. Nam id neque sed elit pharetra rutrum. Suspendisse " +
	"maximus varius urna quis eleifend. Donec dictum finibus interdum. " +
	"Maecenas lectus risus, fringilla vel ipsum eget, porttitor mattis magna. " +
	"Fusce massa nulla, bibendum ut dictum nec, congue vitae lectus. Sed ac " +
	"pharetra ante. Integer sed nisi sem. Sed mattis pretium porta. Donec in " +
	"efficitur urna. In placerat, justo at ultrices tempor, tortor turpis " +
	"bibendum massa, ut maximus augue odio a felis. Curabitur suscipit erat eu " +
	"lectus porttitor sodales. Nam facilisis placerat commodo. Integer vel " +
	"odio ut dui faucibus sollicitudin. Suspendisse nec ultrices turpis. " +
	"Vestibulum nisl magna, mollis nec porta at, facilisis nec erat. Interdum " +
	"et malesuada fames ac ante ipsum primis in faucibus. Sed id dui in mi " +
	"dictum ullamcorper. Curabitur purus est, pretium ut urna et, posuere " +
	"laoreet dolor. Mauris blandit felis non massa cursus, et luctus turpis " +
	"imperdiet. Vestibulum ante ipsum primis in faucibus orci luctus et " +
	"ultrices posuere cubilia Curae; Praesent imperdiet, leo id vehicula " +
	"viverra, mi diam pellentesque dui, sed faucibus augue nisi id orci. " +
	"Suspendisse in sagittis metus, vitae vestibulum odio. Morbi tempus ex " +
	"massa, consequat aliquam turpis tristique vitae. Quisque accumsan " +
	"interdum sagittis. Integer eget malesuada nulla, id blandit mauris. " +
	"Integer elementum quam et leo pretium sodales. Fusce tempus venenatis " +
	"fermentum. Nulla lobortis eros molestie risus cursus, ac suscipit lorem " +
	"fermentum. Etiam a augue ac quam accumsan ultrices eu id leo. Mauris " +
	"laoreet nulla non magna vestibulum, et ullamcorper eros suscipit. Etiam " +
	"hendrerit purus et ex porta tempus. Vestibulum ac lorem quis arcu semper " +
	"feugiat. Aenean ac orci ut urna rhoncus pretium. Duis condimentum non " +
	"mauris sit amet semper. Nunc vehicula elit scelerisque lectus vulputate, " +
	"eget imperdiet lectus tincidunt. Mauris feugiat dolor imperdiet rhoncus " +
	"vehicula. Etiam sit amet suscipit augue. Morbi elit sapien, convallis ut " +
	"tristique in, pharetra vitae ligula. Praesent volutpat, enim facilisis " +
	"auctor tincidunt, est erat fringilla nunc, nec consectetur lectus augue a " +
	"dolor. Donec interdum, eros eget lacinia ultricies, risus ligula posuere " +
	"tellus, non maximus sem nibh quis lacus. Donec vehicula porttitor dictum. " +
	"Proin fringilla, enim quis auctor tempor, mi leo rhoncus ipsum, id " +
	"dapibus est libero vel libero. Etiam nisl erat, cursus eu eros vel, " +
	"vestibulum scelerisque risus. Proin in arcu pellentesque, pretium nisl " +
	"sit amet, rhoncus lectus. Etiam id tortor ac erat malesuada convallis " +
	"eget a enim. Etiam varius dolor nec feugiat porta. Etiam rhoncus suscipit " +
	"ligula, sit amet gravida justo faucibus nec. In mollis finibus purus. " +
	"Integer sit amet massa a nisl malesuada viverra nec et sem. Sed placerat " +
	"neque at commodo porttitor. Sed eget nunc sit amet magna tincidunt " +
	"mattis. Pellentesque habitant morbi tristique senectus et netus et " +
	"malesuada fames ac turpis egestas. In mollis ante auctor, sagittis felis " +
	"rhoncus, sodales est. Nulla non lacus tempor, consectetur enim at, semper " +
	"magna. Donec felis odio, finibus ac sollicitudin ut, pretium non felis. " +
	"Nunc in venenatis leo. Curabitur fringilla neque lacus, sit amet " +
	"consectetur risus luctus sed. Donec convallis vel enim dignissim iaculis. " +
	"Nulla tristique non ipsum vel faucibus. Nulla suscipit tortor nec felis " +
	"malesuada porta. Nam elit risus, auctor eu enim sit amet, pharetra " +
	"efficitur erat. Vestibulum vel quam arcu. Sed vehicula consequat auctor. " +
	"Quisque blandit massa sit amet ipsum fringilla, et congue ipsum pulvinar. " +
	"Aenean vel mauris justo. Nam non molestie mauris. Quisque ut arcu sit " +
	"amet nisl congue pretium. Suspendisse facilisis augue ut sapien vehicula " +
	"iaculis. Proin pretium ornare erat, vitae suscipit risus bibendum non. " +
	"Integer viverra quam dui, vel mattis orci hendrerit eu. Integer aliquam " +
	"egestas mattis. Phasellus hendrerit id metus sit amet cursus. In hac " +
	"habitasse platea dictumst. Mauris consectetur, felis convallis ultrices " +
	"pellentesque, eros magna viverra est, a interdum metus dui vitae ante. " +
	"Vivamus eget rhoncus dui. Suspendisse luctus, nunc et sagittis viverra, " +
	"orci orci feugiat lectus, ut suscipit tellus risus id odio. Duis porta " +
	"nibh ac fermentum interdum. Praesent sed felis eleifend, porttitor eros " +
	"id, ornare nisl. Ut sit amet pharetra odio. Pellentesque ac mi viverra, " +
	"congue lacus vel, eleifend sapien. Etiam gravida ipsum ut erat " +
	"vestibulum, nec viverra velit gravida. Donec non eleifend turpis, a " +
	"tempor leo. Duis sollicitudin ligula vel pharetra finibus. Vestibulum " +
	"justo lectus, tempor eget porttitor et, gravida ac felis. Praesent dictum " +
	"sem sed ornare interdum. Mauris vel tellus blandit, mattis tellus at, " +
	"ultricies neque. Nullam sit amet lectus augue. Nunc sed ultricies ipsum. " +
	"Nullam risus ligula, rutrum ullamcorper magna nec, tempus viverra lorem. " +
	"Donec laoreet felis sed erat elementum mattis. Integer accumsan turpis " +
	"nec egestas consequat. Quisque molestie augue vitae magna rhoncus mollis. " +
	"Morbi at justo in ipsum imperdiet consectetur non et urna. Nam dapibus ex " +
	"quis ornare maximus. Etiam sit amet libero molestie."
