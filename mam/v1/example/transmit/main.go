package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/simia-tech/env"
	powsrvio "gitlab.com/powsrv.io/go/client"

	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/mam/v1"
	"github.com/iotaledger/iota.go/pow"
)

var (
	endpointURL  = env.String("ENDPOINT_URL", "https://nodes.thetangle.org:443")
	powSrvAPIKey = env.String("POWSRV_API_KEY", "")
	seed         = env.String("SEED", "", env.Required())
	mwm          = env.Int("MWM", 9)
)

func main() {
	flag.Parse()
	message := flag.Arg(0)

	powFunc := pow.ProofOfWorkFunc(nil)
	if apiKey := powSrvAPIKey.Get(); apiKey == "" {
		_, powFunc = pow.GetFastestProofOfWorkImpl()
	} else {
		powClient := &powsrvio.PowClient{
			ReadTimeOutMs: 5000,
			APIKey:        apiKey,
			Verbose:       true,
		}
		if err := powClient.Init(); err != nil {
			log.Fatal(err)
		}
		powFunc = powClient.PowFunc
	}

	// create a new API instance
	api, err := api.ComposeAPI(api.HTTPClientSettings{
		URI:                  endpointURL.Get(),
		LocalProofOfWorkFunc: powFunc,
	})
	if err != nil {
		log.Fatal(err)
	}

	transmitter := mam.NewTransmitter(api, seed.Get(), consts.SecurityLevelMedium)
	fmt.Printf("transmit message %q ...\n", message)
	address, err := transmitter.Transmit(message, uint64(mwm.Get()))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s: %s\n", address, message)
}
