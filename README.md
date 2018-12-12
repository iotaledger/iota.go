# IOTA Go client library

[![Build Status](https://travis-ci.org/iotaledger/iota.go.svg?branch=master)](https://travis-ci.org/iotaledger/iota.go)
[![GoDoc](https://godoc.org/github.com/iotaledger/iota.go?status.svg)](https://godoc.org/github.com/iotaledger/iota.go)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/iotaledger/iota.go/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/iotaledger/iota.go)](https://goreportcard.com/report/github.com/iotaledger/iota.go)

## Getting started

> The client library is in beta and is subject to change. Use of this library in production applications is not supported.

### Installation

It is suggested to use [vgo modules](https://github.com/golang/go/wiki/Modules) 
(since Go 1.11) in your project for dependency management:

In any directory outside of GOPATH:
```
$ go mod init <your-module-path>
```

`<your-module-path>` can be paths like github.com/me/awesome-project

```
$ go get github.com/iotaledger/iota.go/api
```
This downloads the latest version of iota.go and writes the used version into
the `go.mod` file (vgo is `go get` agnostic). **Make sure to include /api part in the url.**

### Connecting to the network

```go
package main

import (
    . "github.com/iotaledger/iota.go/api"
    "fmt"
)

var endpoint = "<node-url>"

func main() {
	// compose a new API instance
	api, err := ComposeAPI(HTTPClientSettings{URI: endpoint})
	must(err)
	
	nodeInfo, err := api.GetNodeInfo()
	must(err)
	
	fmt.Println("latest milestone index:", nodeInfo.LatestMilestoneIndex)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
```

### Creating & broadcasting transactions

Publish transfers by calling `PrepareTransfers()` and piping the prepared bundle to `SendTrytes()`.

```go
package main

import (
	"fmt"
	"github.com/iotaledger/iota.go/address"
	. "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/pow"
	"github.com/iotaledger/iota.go/trinary"
)

var endpoint = "<node-url>"

// must be 81 trytes long and truly random
var seed = trinary.Trytes("AAAA....")

// difficulty of the proof of work required to attach a transaction on the tangle
const mwm = 14

// how many milestones back to start the random walk from
const depth = 3

// can be 90 trytes long (with checksum)
const recipientAddress = "BBBB....."

func main() {

	// get the best available PoW implementation
	_, proofOfWorkFunc := pow.GetFastestProofOfWorkImpl()

	// create a new API instance
	api, err := ComposeAPI(HTTPClientSettings{
		URI: endpoint,
		// (!) if no PoWFunc is supplied, then the connected node is requested to do PoW for us
		// via the AttachToTangle() API call.
		LocalProofOfWorkFunc: proofOfWorkFunc,
	})
	must(err)

	// create a transfer to the given recipient address
	// optionally define a message and tag
	transfers := bundle.Transfers{
		{
			Address: recipientAddress,
			Value:   80,
		},
	}

	// create inputs for the transfer
	inputs := []Input{
		{
			Address:  "CCCCC....",
			Security: SecurityLevelMedium,
			KeyIndex: 0,
			Balance:  100,
		},
	}

	// create an address for the remainder.
	// in this case we will have 20 iotas as the remainder, since we spend 100 from our input
	// address and only send 80 to the recipient.
	remainderAddress, err := address.GenerateAddress(seed, 1, SecurityLevelMedium)
	must(err)

	// we don't need to set the security level or timestamp in the options because we supply
	// the input and remainder addresses.
	prepTransferOpts := PrepareTransfersOptions{Inputs: inputs, RemainderAddress: &remainderAddress}

	// prepare the transfer by creating a bundle with the given transfers and inputs.
	// the result are trytes ready for PoW.
	trytes, err := api.PrepareTransfers(seed, transfers, prepTransferOpts)
	must(err)

	// you can decrease your chance of sending to a spent address by checking the address before
	// broadcasting your bundle.
	spent, err := api.WereAddressesSpentFrom(transfers[0].Address)
	must(err)

	if spent[0] {
		fmt.Println("recipient address is spent from, aborting transfer")
		return
	}

	// at this point the bundle trytes are signed.
	// now we need to:
	// 1. select two tips
	// 2. do proof-of-work
	// 3. broadcast the bundle
	// 4. store the bundle
	// SendTrytes() conveniently does the steps above for us.
	bndl, err := api.SendTrytes(trytes, depth, mwm)
	must(err)

	fmt.Println("broadcasted bundle with tail tx hash: ", bundle.TailTransactionHash(bndl))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

```

## Native code and PoW
If the library is compiled with CGO enabled, certain functions such as Curl's transform() will
run native C code for increased speed. 

Certain PoW implementations are enabled if the correct flags are passed while compiling your program:
* `pow_avx` for AVX based PoW
* `pow_sse` for SSE based PoW
* `pow_c128` for C int128 based using PoW
* `pow_arm_c128` for ARM64 int128 C based PoW
* `pow_c` for C based PoW

PoW implementation in Go is always available.
Make sure to define `LocalProofOfWorkFunc` in your provider settings (i.e. `HTTPClientSettings`) if you want to use local PoW. 

## Contributing

We thank everyone for their contributions. In order for your pull requests to get accepted, 
they must fulfill following criterias:
* You must write tests for your additions with ginkgo.
* You must write example code describing the parameters and functionality of your additions. 
* The pull request must pass the CI config.

### Writing tests with ginkgo

First install ginkgo:
```
$ go get github.com/onsi/ginkgo/ginkgo
$ go get github.com/onsi/gomega/...
```

If you have written a new package, you can generate corresponding test suite files via:
```
$ cd <dir-of-your-package>
$ ginkgo bootstrap
```

Now generate a new testing file with:
```
$ ginkgo generate <package-name>
```

Executing the two commands above should give you two files:
```
<package-name>_suite_test.go
<package-name>_test.go
```

> You can use the existing tests as a reference on how to write ginkgo tests or
[read the documentation](https://onsi.github.io/ginkgo/).

Executing your tests:
```
$ go test -v
=== RUN   TestAddress
Running Suite: Address Suite
============================
Random Seed: 1542616006
Will run 11 of 11 specs

•••••••••••
Ran 11 of 11 Specs in 0.261 seconds
SUCCESS! -- 11 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestAddress (0.26s)
PASS
ok  	github.com/iotaledger/iota.go/address	0.264s
```

**Again, your tests must pass otherwise the pull request won't be accepted.**

### Writing documentation/example code
While godoc.org gives a good enough documentation of the package already, the IOTA Foundation's
documentation portal needs additional information, such as parameter description, examples and so on.

1. If non existent, add a `.examples` directory in your newly created package.
2. Create a new file with the following convention: `<package-name>_examples_test.go` inside
the `.examples` directory.
3. Write examples in the following schema:
```
// i req: s, The ASCII string to convert to Trytes.
// o: Trytes, The Trytes representation of the input ASCII string.
// o: error, Returned for non ASCII string inputs.
func ExampleASCIIToTrytes() {
	trytes, err := converter.ASCIIToTrytes("IOTA")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(trytes) // output: "SBYBCCKB"
}
```

| Symbol     | Description |
|:---------------|:--------|
| i req | Describes a parameter to the function. |
| i | Describes an optional parameter to the function. |
| o | Describes a return value of the function. |

Syntax:

* For parameters: `<symbol>: <parameter_name>, <description>.`  
* For return values: `<symbol>: <type>, <description>.`
* Example function: `Example<OriginFunctionName>`

## Reporting Issues

Please report any problems you encouter during development by [opening an issue](https://github.com/iotaledger/iota.go/issues/new).

## Join the discussion

Suggestions and discussion around specs, standardization and enhancements are highly encouraged.
You are invited to join the discussion on [IOTA Discord](https://discord.gg/DTbJufa).
