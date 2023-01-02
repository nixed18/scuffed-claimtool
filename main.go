package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)


const (
	PULL_WAIT = time.Second*10
	MEM_URL = "https://bitcoinexplorer.org/mempool-transactions"
	CHECK_URL = "https://blockchain.info/multiaddr?active="
)

var (
	last_pull time.Time
	limit = 100
)

type ReqForm struct {
	Limit int
	Offset int
}

type CheckResult struct {
	Addresses []CheckResult_Addr `json:"addresses"`
}
type CheckResult_Addr struct {
	Address string `json:"address"`
	NTX int `json:"n_tx"`
}

func request(form ReqForm) {
	defer func() {
		last_pull = time.Now()
	}()

	fmt.Println("--- INITIATING PULL ---")

	var highest_tx *Tx

	url := MEM_URL
	if form.Limit != 0 {
		url += "?limit="+fmt.Sprint(form.Limit)
		if form.Offset != 0 {
			url += "&offset="+fmt.Sprint(form.Offset)
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	resp.Body.Close()

	txes, lowest_tx_fee := scrape(string(body))

	if lowest_tx_fee == -1 {
		// Error
		return
	}

	if len(txes) != 0 {
		url = CHECK_URL

		// Build the check url
		for _, tx := range txes {
			for _, out := range tx.Outputs{
				url += out.Address+"|"
			}
		}
	
		url = url[:len(url)-1] + "&n=1"

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			log.Fatal(err)
		}
	
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			log.Fatal(err)
		}
	
		resp.Body.Close()

		cres := CheckResult{}
		err = json.Unmarshal(body, &cres)
		if err != nil {
			log.Fatal(err)
		}

		highest_tx = func() *Tx {
			for _, tx := range txes {
				// Ordered from highest fee to lowest
				for _, out := range tx.Outputs {
					// Find the corresponding info pull
					var chk_addr *CheckResult_Addr
					for _, chk := range cres.Addresses {
						if out.Address == chk.Address {
							chk_addr = &chk
							break
						}
					}
					if chk_addr == nil {
						fmt.Println("ERROR: MISSING CHK_ADDR FOR ", out.Address)
						continue
					} 
	
					if chk_addr.NTX == 1 {
						// Found the highest
						return tx
					}
				}
			}
			return nil
		}()
	}

	if highest_tx == nil {
		fmt.Println("COULD NOT FIND UNSEEN SEGWIT TX IN MEMPOOL; SEARCHED", limit, "TXES, LOWEST sats/vB SEEN=", lowest_tx_fee)
		return
	}

	fmt.Println("HIGHEST TX DATA:")
	fmt.Println(`	TXID: `+highest_tx.TXID)
	fmt.Println(`	FEE: `+fmt.Sprint(highest_tx.Fee)+` sats/vB`)
	for _, out := range highest_tx.Outputs {
		fmt.Println(`	OUTPUT: ADDR=`+out.Address)
	}
	fmt.Println("")
}

func main() {

	flag.IntVar(&limit, "limit", 100, "how many mempool txes to pull")
	flag.Parse()


	for {
		if time.Since(last_pull) > PULL_WAIT {
			request(ReqForm{
				Limit: limit,
				Offset: 0,
			})
		}
	}
}