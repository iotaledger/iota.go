package mam_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/iotaledger/iota.go/consts"
	mam "github.com/iotaledger/iota.go/mam/v1"
	"github.com/iotaledger/iota.go/trinary"
)

var _ = Describe("State", func() {

	Context("NewState", func() {

		It("Should return a new state", func() {
			state := mam.NewState(mam.Settings{ProviderURL: ""}, "", consts.SecurityLevelLow)

			Expect(state.Channel().Mode).To(Equal(mam.ChannelModePublic))
			Expect(state.Channel().SideKey).To(Equal(trinary.Trytes("")))
			Expect(state.Channel().SecurityLevel).To(Equal(consts.SecurityLevelLow))
			Expect(state.Channel().Start).To(Equal(0))
			Expect(state.Channel().Count).To(Equal(1))
			Expect(state.Channel().NextCount).To(Equal(1))
			Expect(state.Channel().Index).To(Equal(0))
		})

	})

	Context("Subcribe", func() {

		It("Should add a subscription to the state", func() {
			state := mam.NewState(mam.Settings{ProviderURL: ""}, "", consts.SecurityLevelLow)

			state.Subscribe("", mam.ChannelModePublic, "")

			Expect(state.SubscriptionCount()).To(Equal(1))
		})

	})

	Context("SetMode", func() {

		It("Should set the mode to private", func() {
			state := mam.NewState(mam.Settings{ProviderURL: ""}, "", consts.SecurityLevelLow)

			err := state.SetMode(mam.ChannelModePrivate, "")

			Expect(err).NotTo(HaveOccurred())
			Expect(state.Channel().Mode).To(Equal(mam.ChannelModePrivate))
		})

		It("Should set the mode to restricted", func() {
			state := mam.NewState(mam.Settings{ProviderURL: ""}, "", consts.SecurityLevelLow)

			err := state.SetMode(mam.ChannelModeRestricted, "ABC")

			Expect(err).NotTo(HaveOccurred())
			Expect(state.Channel().Mode).To(Equal(mam.ChannelModeRestricted))
			Expect(state.Channel().SideKey).To(Equal(trinary.Trytes("ABC")))
		})

	})

})
