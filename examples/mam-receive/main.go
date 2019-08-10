package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/simia-tech/env"

	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/mam/v1"
)

var (
	endpointURL = env.String("ENDPOINT_URL", "https://nodes.thetangle.org:443")
	mode        = env.String("MODE", "public", env.AllowedValues("public", "private", "restricted"))
	sideKey     = env.String("SIDE_KEY", "")
)

func main() {
	follow := flag.Bool("f", false, "don't exit and check every second for new messages")
	flag.Parse()
	root := flag.Arg(0)

	cm, err := mam.ParseChannelMode(mode.Get())
	if err != nil {
		log.Fatal(err)
	}

	api, err := api.ComposeAPI(api.HTTPClientSettings{
		URI: endpointURL.Get(),
	})
	if err != nil {
		log.Fatal(err)
	}

	receiver := mam.NewReceiver(api)
	if err := receiver.SetMode(cm, sideKey.Get()); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("receive root %q from %s channel...\n", root, cm)
loop:
	nextRoot, messages, err := receiver.Receive(root)
	if err != nil {
		log.Fatal(err)
	}
	for _, message := range messages {
		fmt.Println(message)
	}
	if *follow {
		time.Sleep(time.Second)
		if len(messages) > 0 {
			root = nextRoot
		}
		goto loop
	}
	if len(messages) == 0 {
		fmt.Println("no messages found")
	}
}
