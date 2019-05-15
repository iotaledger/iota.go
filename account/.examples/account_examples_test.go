package account_examples_test

import (
	"fmt"
	"github.com/iotaledger/iota.go/account"
	"github.com/iotaledger/iota.go/account/deposit"
	"log"
	"time"
)

var acc account.Account

// o: string, The id of the account.
func ExampleID() {
	id := acc.ID()
	// VX9SJY9JLSDLLKBNAKXQYSYWAYBABIWOHXAAIDJAESFORTUKBICOLEPZLODPTKRZXDFUIBR99GETYLF9X
	fmt.Println(id)
}

// o: error, Returned when the account couldn't be started because of a misconfiguration or faulty plugin.
func ExampleStart() {
	if err := acc.Start(); err != nil {
		log.Fatal(err)
	}
}

// o: error, Return when the account couldn't shutdown.
func ExampleShutdown() {
	if err := acc.Start(); err != nil {
		log.Fatal(err)
	}
}

// i: recipients, The recipients to which to send funds or messages to.
// o: Bundle, The bundle which got sent off
// o: error, Returned when any error while gathering inputs, IRI API calls or storage related issues occur.
func ExampleSend() {
	// the Send() method expects Recipient(s), which are just
	// simple transfer objects found in the bundle package
	recipient := account.Recipient{
		Address: "SDOFSKOEFSDFKLG...",
		Value:   100,
	}

	bndl, err := acc.Send(recipient)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("sent off bundle %s\n", bndl[0].Hash)
}

// i: conds, The conditions for the newly allocated deposit address.
// o: CDA, The generated conditional deposit address.
// o: error, Returned when there's any issue persisting or creating the CDA.
func ExampleAllocateDepositAddress() {
	timeoutInOneHour := time.Now().Add(time.Duration(1) * time.Hour)
	conds := &deposit.Conditions{
		TimeoutAt: &timeoutInOneHour,
	}
	cda, err := acc.AllocateDepositAddress(conds)
	if err != nil {
		log.Fatal(err)
	}

	// EQSAUZXULTTYZCLNJNTXQTQHOMOFZERHTCGTXOLTVAHKSA9OGAZDEKECURBRIXIJWNPFCQIOVFVVXJVD9DGCJRJTHZ
	fmt.Println(cda.AsMagnetLink())
}

// o: uint64, The current available balance on the account.
// o: error, Returned when storage or IRI API call problems arise.
func ExampleAvailableBalance() {
	balance, err := acc.AvailableBalance()
	if err != nil {
		log.Fatal(err)
	}

	// 1337
	fmt.Println(balance)
}

// o: uint64, The total balance on the account.
// o: error, Returned when storage or IRI API call problems arise.
func ExampleTotalBalance() {
	balance, err := acc.TotalBalance()
	if err != nil {
		log.Fatal(err)
	}

	// 1337
	fmt.Println(balance)
}

// o: bool, Whether the account is new or not (has actual account data in the store).
// o: error, Returned when storage problems arise during the check.
func ExampleIsNew() {
	isNew, err := acc.IsNew()
	if err != nil {
		log.Fatal(err)
	}
	switch isNew {
	case true:
		fmt.Println("the account is new")
	case false:
		fmt.Println("the account is not new")
	}
}

// i: setts, The new settings to be applied to the account.
// o: error, Returned when any error occurs while applying the node settings.
func ExampleUpdateSettings() {
	newSetts := account.DefaultSettings()
	newSetts.Depth = 1
	newSetts.MWM = 9
	if err := acc.UpdateSettings(newSetts); err != nil {
		log.Fatal(err)
	}
	fmt.Println("updated account settings")
}

// o: uint64, Returns the sum of the transfer value.
func ExampleSum() {
	recipients := account.Recipients{
		{
			Address: "PWBJRKWJX...",
			Value:   200,
		},
		{
			Address: "ASXMROTUF...",
			Value:   400,
		},
	}

	// 600
	fmt.Println(recipients.Sum())
}

// o: Transfers,
func ExampleAsTransfers() {
	recipients := account.Recipients{
		{
			Address: "PWBJRKWJX...",
			Value:   200,
		},
		{
			Address: "ASXMROTUF...",
			Value:   400,
		},
	}

	transfers := recipients.AsTransfers()
	// PWBJRKWJX...
	fmt.Println(transfers[0].Address)
}

// i: setts, The settings to use for the account.
// o: Account, The account itself.
// o: error, Returned for misconfiguration or other types of errors.
func ExampleNewAccount() {
	newAccount, err := account.NewAccount(account.DefaultSettings())
	if err != nil {
		log.Fatal(err)
	}
	if err := newAccount.Start(); err != nil {
		log.Fatal(err)
	}
}

// o: Trytes, The seed in its Tryte representation.
// o: error, Returned when retrieving the seed fails.
func ExampleSeed() {
	seedProv := account.NewInMemorySeedProvider("WEOIOSDFX...")
	seed, err := seedProv.Seed()
	if err != nil {
		log.Fatal(err)
	}
	// WEOIOSDFX...
	fmt.Println(seed)
}

// i: seed, The seed to keep in memory.
// o: SeedProvider, The SeedProvider which will provide the seed.
func ExampleNewInMemorySeedProvider() {
	seedProv := account.NewInMemorySeedProvider("WEOIOSDFX...")
	seed, err := seedProv.Seed()
	if err != nil {
		log.Fatal(err)
	}
	// WEOIOSDFX...
	fmt.Println(seed)
}


