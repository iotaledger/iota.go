package inmemory_test

import (
	"fmt"
	"github.com/iotaledger/iota.go/account/store"
	"github.com/iotaledger/iota.go/account/store/inmemory"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
	"time"
)

var emptyAddr = strings.Repeat("9", 81)

const id = "d7e75aa9def2ef9c813313f0e0fb72b9"

var _ = Describe("InMemory", func() {

	var zeroValBundleTrytes = []trinary.Trytes{
		"VDD999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999FNXCXZCECVIXHIBZKIFOPPWRKBC9BSN9B9QOTREKIJSBVSBLSETYPQOQSGTVWDDKHIWITNFUDWFFXUXTA999999999999999999999999999GHY9QLLL9XNY999999999999999ZI9FN9D99999999999999999999FPBVYDWIHZHJLBRQVZBTWEXIOBGTDYIUPQPBI9HIVVWGIDADHGFQOO9OPWYJVXJBIDHIHIPOCKHUQUCF9RDGYDVOGMSDQCBXLTLONMBVRLATCCLKPCCQRVFGPTVVRJPAITAKTFS9MLUKPPDCGJSPROOAYXKPO99999CEXUIXWFMUDXDNVGWIPCEDQD99LDAYNNYVUXLEZECXLBPLYAIKGWLCYEYPAXGKEW9REOOVFHEB9PA9999SHY9QLLL9XNY999999999999999ROUSUMMLE999999999MMMMMMMMMXRGQHAZWFAEVOAZ9VGEDQHRNZMC",
	}
	tx, err := transaction.AsTransactionObject(zeroValBundleTrytes[0])
	if err != nil {
		panic(err)
	}

	exportedDate := time.Now()
	const exportedAccountID = "12345"
	const exportedPendingTransferKey = "abcd"
	const exportedPendingTransfeTailHash = "9887656"

	stateToImport := store.ExportedAccountState{
		ID: exportedAccountID, Date: exportedDate, AccountState: store.AccountState{
			KeyIndex: 100,
			PendingTransfers: map[string]*store.PendingTransfer{
				exportedPendingTransferKey: {
					Tails:  trinary.Hashes{exportedPendingTransfeTailHash},
				},
			},
			DepositAddresses: map[uint64]*store.StoredDepositAddress{},
		},
	}

	st := inmemory.NewInMemoryStore()

	var state *store.AccountState
	It("loads correctly an empty account", func() {
		var err error
		state, err = st.LoadAccount(id)
		Expect(err).ToNot(HaveOccurred())
		Expect(state.IsNew()).To(BeTrue())
	})

	Context("AddPendingTransfer()", func() {
		It("adds the pending zero value transfer to the store", func() {
			err := st.AddPendingTransfer(id, tx.Hash, zeroValBundleTrytes)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("GetPendingTransfers()", func() {
		It("returns all pending transfers", func() {
			pendingTransfers, err := st.GetPendingTransfers(id)
			Expect(err).ToNot(HaveOccurred())
			bndl, err := store.PendingTransferToBundle(pendingTransfers[tx.Hash])
			Expect(err).ToNot(HaveOccurred())
			Expect(bndl[0].Address).To(Equal(tx.Address))
		})
	})

	Context("AddTailHash()", func() {
		It("adds the given tail hash", func() {
			newTail := strings.Repeat("A", 81)
			err := st.AddTailHash(id, tx.Hash, newTail)
			Expect(err).ToNot(HaveOccurred())
			state, err = st.LoadAccount(id)
			Expect(err).ToNot(HaveOccurred())
			Expect(state.PendingTransfers[tx.Hash].Tails[1]).To(Equal(newTail))
		})
	})

	Context("RemovePendingTransfer()", func() {
		It("removes the given transfer", func() {
			err := st.RemovePendingTransfer(id, tx.Hash)
			Expect(err).ToNot(HaveOccurred())
			state, err = st.LoadAccount(id)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(state.PendingTransfers)).To(Equal(0))
			pendingTransfers, err := st.GetPendingTransfers(id)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pendingTransfers)).To(Equal(0))
		})
	})

	Context("ImportAccount()", func() {
		It("imports the account", func() {
			err := st.ImportAccount(stateToImport)
			Expect(err).ToNot(HaveOccurred())
			state, err := st.LoadAccount(stateToImport.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(state.KeyIndex).To(Equal(stateToImport.KeyIndex))
			Expect(state.PendingTransfers[exportedPendingTransferKey].Tails).
				To(Equal(stateToImport.PendingTransfers[exportedPendingTransferKey].Tails))
		})
	})

	Context("ExportAccount()", func() {
		It("exports the account", func() {
			state, err := st.ExportAccount(stateToImport.ID)
			Expect(err).ToNot(HaveOccurred())
			fmt.Println(state.PendingTransfers[exportedPendingTransferKey].Bundle)
			Expect(state.AccountState).To(Equal(stateToImport.AccountState))
		})
	})
})
