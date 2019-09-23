package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/simia-tech/env"

	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/mam/v1"
	"github.com/iotaledger/iota.go/pow"
)

var (
	endpointURL = env.String("ENDPOINT_URL", "https://nodes.thetangle.org:443")
	seed        = env.String("SEED", "")
	mwm         = env.Int("MWM", 9)
	mode        = env.String("MODE", "public", env.AllowedValues("public", "private", "restricted"))
	sideKey     = env.String("SIDE_KEY", "")
)

func main() {
	flag.Parse()
	messages := flag.Args()

	cm, err := mam.ParseChannelMode(mode.Get())
	if err != nil {
		log.Fatal(err)
	}

	_, powFunc := pow.GetFastestProofOfWorkImpl()

	api, err := api.ComposeAPI(api.HTTPClientSettings{
		URI:                  endpointURL.Get(),
		LocalProofOfWorkFunc: powFunc,
	})
	if err != nil {
		log.Fatal(err)
	}

	transmitter := mam.NewTransmitter(api, seed.Get(), uint64(mwm.Get()), consts.SecurityLevelMedium)
	if err := transmitter.SetMode(cm, sideKey.Get()); err != nil {
		log.Fatal(err)
	}

	for _, message := range messages {
		fmt.Printf("transmit message %q to %s channel...\n", message, cm)
		root, err := transmitter.Transmit(message)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("transmitted to root %q\n", root)
	}
}
