package main

import (
	"fmt"
	"github.com/iotaledger/giota"
	"github.com/iotaledger/iota.go/bundle"
	. "github.com/iotaledger/iota.go/converter"
	"github.com/iotaledger/iota.go/pow"
	"github.com/iotaledger/iota.go/signing"
	"github.com/iotaledger/iota.go/units"
	"net/http"
)

const iriEndpoint = "https://trinity.iota-tangle.io:14265"
const seed = "QKFIKNNOLHEWDEATLDRTIQYTMJUBQQGIXFBJUQRIFYXVBIUSOGNIBCAKEDCWBKGVPQODZVQSWUVFGLJ9M"

var secLevel = signing.SecurityLevelMedium
var recipientAddress = "DLXGUQYGLC9HZXNVLEKPXJYVJUSNXJGOKYJLAXETSN9QLIPGKTMYNDUZYNHQFTWJJBIZRGDSJITXAKWCWVZWVRMLID"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// create an API instance
	api := giota.NewAPI(iriEndpoint, http.DefaultClient)

	// use NewAddressHash() to validate the address
	// if the address contains a checksum it will be thrown away
	targetAddr, err := signing.NewAddressHashFromTrytes(recipientAddress)
	must(err)

	// message for the recipient of the transaction
	msgTrytes, err := ASCIIToTrytes("this transaction was made via the Go IOTA library")
	must(err)

	// create a new transfer object representing our value transfer
	transfers := bundle.Transfers{
		{
			Address: targetAddr,
			Value:   int64(units.ConvertUnits(1, units.Ki, units.I)), // 1000 iotas
			Message: msgTrytes,
			Tag:     "GIOTALIBRARY", // gets automatically padded to 27 trytes
		},
	}

	// this function doesn't do any I/O but simply constructs objects
	// containing the information needed to derive the private key
	// and address at a given index with the provided seed and security level
	inputs := bundle.NewAddressInputs(seed, 0, 10, signing.SecurityLevelMedium)

	// compute the remainder address
	unusedAddr, _, err := api.GetUntilFirstUnusedAddress(seed, 2)

	// prepares the transfers by creating a bundle with the given output transaction (made from the transfer objects)
	// and input transactions from the given address inputs. in case not the entire input is spent to the
	// defined transfers, the remainder is sent to the given remainder address.
	// It also automatically checks whether the given input addresses have enough funds for the transfer.
	bndl, err := api.PrepareTransfers(seed, transfers, inputs, unusedAddr, 2)
	must(err)

	// at this point it is good practice to check whether the destination address was already spent from
	spentStates, err := api.WereAddressesSpentFrom(targetAddr)
	must(err)
	if spentStates[0] == true {
		fmt.Println("aborting, target address is already spent from")
		return
	}

	// at this point the bundle contains input and output transactions and is signed
	// now we need to first select two tips to approve and then do proof of work.
	// we can do this in one call with SendTrytes() which does:
	// 1. select two tips (you can optionally provide a reference)
	// 2. create an attachToTangleRequest to the remote node or do PoW locally if powFunc is supplied
	// 3. broadcast the bundle to the network
	// 4. do a storeTransaction call to the connected node
	_, powFunc := pow.GetBestPoW()
	_, err = api.SendTrytes(3, bndl, 14, powFunc)
	must(err)

}
