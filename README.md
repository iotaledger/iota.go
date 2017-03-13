[![Build Status](https://travis-ci.org/utamaro/giota.svg?branch=master)](https://travis-ci.org/utamaro/giota)
[![GoDoc](https://godoc.org/github.com/utamaro/giota?status.svg)](https://godoc.org/github.com/utamaro/giota)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/utamaro/giota/master/LICENSE)
[![Coverage Status](https://coveralls.io/repos/utamaro/giota/badge.svg?branch=master)](https://coveralls.io/r/utamaro/giota?branch=master)


gIOTA
=====

Client library for the IOTA reference implementation (IRI).

Refer to [godoc](https://godoc.org/github.com/utamaro/giota) for details.


Example
====

```go

//Trits
tritsFrom:=[]int8{1,-1,1,0,1,1,0,-1,0}
trits,err:=giota.ToTrits(tritsFrom)

//Trytes
trytes,err:=trits.ToTrytes()

//Hash
hash:=trits.Hash()

//API
api := giota.NewAPI("http://localhost:14265", nil)
ftr := &giota.FindTransactionsRequest{Bundles: []Trytes{"DEXRPL...SJRU"}}
resp, err := api.FindTransactions(ftr)

///Address
index:=0
security:=2
adr,err:=giota.NewAddress(trytes,index,seciruty) //without checksum.
adrWithChecksum := adr.WithChecksum() //adrWithChecksum is trytes type.

//transaction
tx,err:=giota.NewTransaction(trits)
if tx.HasValidNonce(){...}
trits2:=tx.Trits()

//create signature
key := giota.NewKey(seed.Trits(), 0, 2)
norm := bundleHash.Normalize()
sign := giota.Sign(norm[:27], key[:6561])

//validate signature
adr, err := giota.NewAddress(seed, 0, 2)
if giota.ValidateSig(adr, []Trits{sign}, bundleHash) {...}
```

Development Status: Alpha+
=========================

Tread lightly around here. This library is still very much
in flux and there are going to be breaking changes.


<hr>
Released under the [MIT License](LICENSE).
