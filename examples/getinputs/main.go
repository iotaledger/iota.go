package main

import (
	"fmt"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/iotaledger/giota"

	"golang.org/x/crypto/ssh/terminal"
)

const Host = "http://node03.iotatoken.nl:15265"

func main() {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	fmt.Print("input your seed: ")
	api := giota.NewAPI(Host, &client)
	seed, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}

	seedT, err := giota.ToTrytes(string(seed))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\nhow many addresses should we check: ")
	var offset int
	n, err := fmt.Scanf("%d\n", &offset)
	if err != nil || n < 1 {
		log.Fatal("expecting integer offset")
	}

	fmt.Print("which security level should the addresses be generated at (0, 1, 2; default is 2): ")
	var slevel int
	n, err = fmt.Scanf("%d\n", &slevel)
	if err != nil || n < 1 || slevel < 0 || slevel > 2 {
		slevel = 2
	}

	println("Getting balances")
	// GetInputs(API, seed, start index, end index, threshold, security level)
	inputs, err := giota.GetInputs(api, seedT, 0, offset, 0, slevel)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(inputs)
}
