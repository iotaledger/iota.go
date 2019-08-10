package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math/big"

	"github.com/simia-tech/env"
	powsrvio "gitlab.com/powsrv.io/go/client"
	"golang.org/x/crypto/argon2"

	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/mam/v1"
	"github.com/iotaledger/iota.go/pow"
	"github.com/iotaledger/iota.go/trinary"
)

var (
	endpointURL  = env.String("ENDPOINT_URL", "https://nodes.thetangle.org:443")
	powSrvAPIKey = env.String("POWSRV_API_KEY", "")
	seed         = env.String("SEED", "")
	seedPassword = env.String("SEED_PASSWORD", "")
	mwm          = env.Int("MWM", 9)
)

func main() {
	flag.Parse()
	messages := flag.Args()

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

	transmitter := mam.NewTransmitter(api, seedFromEnv(), consts.SecurityLevelMedium)
	for _, message := range messages {
		fmt.Printf("transmit message %q ...\n", message)
		address, err := transmitter.Transmit(message, uint64(mwm.Get()))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", address, message)
	}
}

func seedFromEnv() trinary.Trytes {
	if v := seed.Get(); v != "" {
		return v
	}
	if v := seedPassword.Get(); v != "" {
		password := bytes.TrimSpace([]byte(v))
		seedBytes := argon2.IDKey(password, []byte(""), 50, 64*1024, 4, 48)
		log.Printf("b = %d", len(seedBytes))
		return toTrytes(seedBytes)
	}
	seedBytes := make([]byte, 48)
	rand.Read(seedBytes)
	seed := toTrytes(seedBytes)
	fmt.Printf("generated random seed %s\n", seed)
	return seed
}

func toTrytes(b []byte) trinary.Trytes {
	b = []byte(new(big.Int).SetBytes(b).Text(27))
	index := 0
	for range b {
		switch {
		case b[index] == 48:
			b[index] = 57
			break
		case b[index] > 48 && b[index] < 58:
			b[index] += 16
			break
		case b[index] > 96 && b[index] < 123:
			b[index] -= 23
			break
		default:
			b[index] += 9
		}
		index++
	}
	return trinary.Trytes(b)
}
