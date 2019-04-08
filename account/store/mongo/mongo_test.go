package mongo_test

import (
	"context"
	"github.com/iotaledger/iota.go/account/deposit"
	"github.com/iotaledger/iota.go/account/store"
	mongo_store "github.com/iotaledger/iota.go/account/store/mongo"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

const id = "d7e75aa9def2ef9c813313f0e0fb72b9"
const mongoDBURI = "mongodb://localhost:27017"
const dbName = "iota_accounts_test"

var _ = Describe("Mongo", func() {

	var zeroValBundleTrytes = []trinary.Trytes{
		"VDD999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999FNXCXZCECVIXHIBZKIFOPPWRKBC9BSN9B9QOTREKIJSBVSBLSETYPQOQSGTVWDDKHIWITNFUDWFFXUXTA999999999999999999999999999GHY9QLLL9XNY999999999999999ZI9FN9D99999999999999999999FPBVYDWIHZHJLBRQVZBTWEXIOBGTDYIUPQPBI9HIVVWGIDADHGFQOO9OPWYJVXJBIDHIHIPOCKHUQUCF9RDGYDVOGMSDQCBXLTLONMBVRLATCCLKPCCQRVFGPTVVRJPAITAKTFS9MLUKPPDCGJSPROOAYXKPO99999CEXUIXWFMUDXDNVGWIPCEDQD99LDAYNNYVUXLEZECXLBPLYAIKGWLCYEYPAXGKEW9REOOVFHEB9PA9999SHY9QLLL9XNY999999999999999ROUSUMMLE999999999MMMMMMMMMXRGQHAZWFAEVOAZ9VGEDQHRNZMC",
	}
	tx, err := transaction.AsTransactionObject(zeroValBundleTrytes[0])
	if err != nil {
		panic(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoDBURI))
	if err != nil {
		panic(err)
	}
	if err := client.Connect(ctx); err != nil {
		panic(err)
	}

	if err := client.Database(dbName).Drop(ctx); err != nil {
		panic(err)
	}

	st, err := mongo_store.NewMongoStore(mongoDBURI, &mongo_store.Config{DBName: dbName})
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

	var state *store.AccountState
	It("loads correctly an empty account", func() {
		var err error
		state, err = st.LoadAccount(id)
		Expect(err).ToNot(HaveOccurred())
		Expect(state.IsNew()).To(BeTrue())
	})

	Context("AddDepositAddress()", func() {
		It("adds the deposit address to the store", func() {
			ts := time.Now().AddDate(0, 0, 1)
			var expAm uint64 = 1337
			err := st.AddDepositAddress(id, 0, &store.StoredDepositAddress{
				SecurityLevel: consts.SecurityLevelMedium,
				Conditions: deposit.Conditions{
					TimeoutAt:      &ts,
					MultiUse:       true,
					ExpectedAmount: &expAm,
				},
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("GetDepositAddresses()", func() {
		It("returns all deposit addresses", func() {
			depositAddresses, err := st.GetDepositAddresses(id)
			Expect(err).ToNot(HaveOccurred())
			Expect(*depositAddresses[0].ExpectedAmount).To(Equal(uint64(1337)))
		})
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
			Expect(state.AccountState).To(Equal(stateToImport.AccountState))
		})
	})
})
