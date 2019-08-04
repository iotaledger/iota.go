package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/simia-tech/env"

	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/mam/v1"
)

var (
	endpointURL  = env.String("ENDPOINT_URL", "https://nodes.thetangle.org:443")
	powSrvAPIKey = env.String("POWSRV_API_KEY", "")
)

// PN9CYZTVHPKVRVBOOIRZDZSHQLURKQYSQTTDOGAKZZ9SCGIWTTOBPRWVPZRHZJHIWOKLZE9SQJWGPGFTX

func main() {
	flag.Parse()
	root := flag.Arg(0)

	// create a new API instance
	api, err := api.ComposeAPI(api.HTTPClientSettings{
		URI: endpointURL.Get(),
	})
	if err != nil {
		log.Fatal(err)
	}

	receiver := mam.NewReceiver(api, root)
	fmt.Printf("receive root %q ...\n", root)
	messages, err := receiver.Receive()
	if err != nil {
		log.Fatal(err)
	}
	for _, message := range messages {
		fmt.Println(message)
	}
}
