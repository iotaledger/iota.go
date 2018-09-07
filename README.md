# giota

[![Build Status](https://travis-ci.org/iotaledger/giota.svg?branch=master)](https://travis-ci.org/iotaledger/giota)
[![GoDoc](https://godoc.org/github.com/iotaledger/giota?status.svg)](https://godoc.org/github.com/iotaledger/giota)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/iotaledger/giota/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/iotaledger/giota)](https://goreportcard.com/report/github.com/iotaledger/giota)

## Getting started

### Installation

It is suggested to use [vgo modules](https://github.com/golang/go/wiki/Modules) 
(since Go 1.11) in your project for dependency management:

In any directory outside of GOPATH:
```
$ go mod init <your-module-path>
```

`<your-module-path>` can be paths like github.com/me/awesome-project

```
$ go get github.com/iotaledger/giota
```
This downloads the latest version of giota and writes the used version into
the `go.mod` file (vgo is `go get` agnostic). Done.

### Connecting to the network

```go
package main

import (
    "github.com/iotaledger/giota"
    "fmt"
)

var endpoint = "<node-url>"

func main() {
	// create a new API instance, optionally provide your own http.Client
	api := giota.NewAPI(endpoint, nil)
	
	nodeInfo, err := api.GetNodeInfo()
	if err != nil {
	    panic(err)
	}
	
	fmt.Println("latest milestone index:", nodeInfo.LatestMilestoneIndex)
}
```

### Creating & broadcasting transactions

Publish transfers by calling `PrepareTransfers()` and piping the prepared bundle to `SendTrytes` command.

```go
package main

import (
    "github.com/iotaledger/giota"
    "github.com/iotaledger/giota/signing"
    "github.com/iotaledger/giota/bundle"
    "github.com/iotaledger/giota/trinary"
    "github.com/iotaledger/giota/pow"
    "fmt"
)


var endpoint = "<node-url>"
// must be 81 trytes long and truly random
var seed = trinary.Trytes("AAAA....") 
var securityLevel = signing.SecurityLevelMedium
// difficulty of the proof of work required to attach a transaction on the tangle
const mwm = 14
// how many milestones back to start the random walk from
const depth = 3
// can be 90 trytes long (with checksum)
const recipientAddrRaw = "BBBB....."

// use real error handling in your code instead of must()
func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// create a new API instance
	api := giota.NewAPI(endpoint, nil)
	
	// convert the recipient address to a signing.Address.
	// if the input string contains a checksum it is validated and an error
	// is returned if it is not valid.
	recipientAddr, err := signing.ToAddress(trinary.Trytes(recipientAddrRaw))
	must(err)
	
	transfers := bundle.Transfers{
		{
		    Address: recipientAddr,
		    Value: 1000, // deposit 1000 iota, 1 Ki
		    Tag: "", // optional tag
		    Message: "", // optional message in trytes
		},
	}
	
	// it is the library's user job to query and obtain valid inputs for the bundle.
	// this can be achieved by storing the seed's address states somewhere within the
	// application which uses the seed.
	
	// in this example we assume that the first address of our seed has
	// 5000 iotas, thereby enough funds for the transfer
	inputs := bundle.AddressInputs{
		{
		    Seed: seed,
		    Security: securityLevel,
		    Index: 0, // the index is optional as it isn't needed to construct the bundle
		},
	}
	
	// since in IOTA inputs must be spent completely, we need to send the remainder (4000 iotas)
	// to our next address. in this example this would simply be the address at the next
	// index which is 1 (we used address at index 0 for as input).
	remainderAddr, err := signing.NewAddress(seed, 1, securityLevel)
	must(err)
	
	// prepares the transfers by creating a bundle with the given output transaction (made from the transfer objects)
	// and input transactions from the given address inputs. in case not the entire input is spent to the
	// defined transfers, the remainder is sent to the given remainder address.
	// It also automatically checks whether the given input addresses have enough funds for the transfer.
	bundle, err := api.PrepareTransfers(seed, transfers, inputs, remainderAddr, securityLevel)
	must(err)
	
	// at this point it is good practice to check whether the destination address was already spent from
	spentStates, err := api.WereAddressesSpentFrom(recipientAddr)
	must(err)
	if spentStates[0] == true {
		fmt.Println("aborting, recipient address is already spent from")
		return
	}	
	
	// at this point the bundle contains input and output transactions and is signed.
	// now we need to first select two tips to approve and then do the proof of work.
	// we can do this in one call with SendTrytes() which does:
	// 1. select two tips (you can optionally provide a reference)
	// 2. create an attachToTangleRequest to the remote node or do PoW locally if powFunc is supplied
	// 3. broadcast the bundle to the network
	// 4. do a storeTransaction call to the connected node
	_, powFunc := pow.GetBestPoW()
	bundle, err = api.SendTrytes(3, bundle, 14, powFunc)
	must(err)
	
	fmt.Println("attached bundle with tail hash", bundle[0].Hash(), "to the tangle")
}
```

## PoW
If the library is compiled with CGO enabled, certain functions such as Curl's transform() will
run native C code for increased speed. Check the PoW files under the `pow` directory to see which
build flags must be enabled for which PoW function.

## Contributing

We thank everyone for their contributions. Here is quick guide to get started with giota:

### Clone and bootstrap

1. Fork the repo with <kbd>Fork</kbd> button at top right corner.
2. Clone your fork locally and `cd` in it.
3. Bootstrap your environment with:

```
go get ./...
```

This will install all needed dependencies.

### Run the tests

Make your changes on a single or across multiple packages and test the system in integration. Run from the _root directory_:

```
go test ./...
```

To run tests of specific package just `cd` to the package directory and run `go test` from there.

## Reporting Issues

Please report any problems you encouter during development by [opening an issue](https://github.com/iotaledger/giota/issues/new).

## Join the discussion

Suggestions and discussion around specs, standardization and enhancements are highly encouraged.
You are invited to join the discussion on [IOTA Discord](https://discord.gg/DTbJufa).