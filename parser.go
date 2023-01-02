package main

import (
	"errors"
	"fmt"
	"strings"
	"strconv"
)

type Tx struct {
	TXID string
	Fee int64
	Outputs []Output
}

type Output struct {
	Address string
	Amt string
}

func pull_to(target, data string) (string, error) {
	i := 0
	tlen := len(target)
	all_len := len(data) - tlen +1
	for i < all_len {
		if data[i:i+tlen] == target {
			out := data[:i]
			return out, nil
		}
		i += 1
	}
	return data, errors.New("pull_to data missing target "+target)
}

func strip_to(target, data string) (string, error) {
	my_string := data
	i := 0
	tlen := len(target)
	length := len(data) - tlen +1
	
	for i < length {
		if my_string[i:i+tlen] == target {
			// Target found, return all after this
			out := my_string[i+tlen:]
			return out, nil
		}
		i += 1
	}
	return data, errors.New("pull_to data missing target "+target)
}

func scrape(data string) ([]*Tx, int64) {
	fmt.Sprint("a")
	var err error
	errs := []error{}
	var lowest_tx_fee int64

	txes := []*Tx{}

	// Remove early chunk
	for {
		var txid string
		var fee int64
		this_tx := &Tx{Outputs: []Output{}}

		// Find the TX block
		data, err = strip_to(`<div class="border bg-content p-3">`, data)
		errs = append(errs, err)
		if err != nil {break} // No more transactions

		// Find the TXID
		data, err = strip_to(`<span class="badge bg-primary fw-normal me-2"`, data)
		errs = append(errs, err)
		if err != nil {fmt.Println("ERRORS:", errs); return []*Tx{}, -1}

		data, err = strip_to(`./tx/`, data)
		errs = append(errs, err)
		if err != nil {fmt.Println("ERRORS:", errs); return []*Tx{}, -1}

		// Pull the TXID
		txid, err = pull_to(`"`, data)
		errs = append(errs, err)
		if err != nil {fmt.Println("ERRORS:", errs); return []*Tx{}, -1}
		this_tx.TXID = txid

		// Find the fee
		data, err = strip_to(`<span class="badge bg-light text-dark border me-2">`, data)
		errs = append(errs, err)
		if err != nil {fmt.Println("ERRORS:", errs); return []*Tx{}, -1}

		// Pull the fee
		sfee, err := pull_to(`<`, data)
		errs = append(errs, err)
		if err != nil {fmt.Println("ERRORS:", errs); return []*Tx{}, -1}

		sfee = strings.Replace(sfee, ",", "", -1)

		fee, err = strconv.ParseInt(sfee, 10, 0)
		errs = append(errs, err)
		if err != nil {fmt.Println("ERRORS:", errs); return []*Tx{}, -1}
		this_tx.Fee = fee

		// Find the output list start
		data, err = strip_to(`<div class="col-lg-6 border-lg-`, data)
		errs = append(errs, err)
		if err != nil {fmt.Println("ERRORS:", errs); return []*Tx{}, -1}

		// Find all the outputs
		txid_line := `<div data-txid="`+txid+`">`

		has_segwit := false

		for {
			var addr string
			var amt string

			//fmt.Println(txid_line)
			// Strip to the txid_line
			data, err = strip_to(txid_line, data)
			if err != nil {break} // No more outputs in this tx

			// Find the output addr
			data, err = strip_to(`./address/`, data)
			errs = append(errs, err)
			if err != nil {fmt.Println("ERRORS:", errs); return []*Tx{}, -1}

			// Pull the addr
			addr, err = pull_to(`"`, data)
			errs = append(errs, err)
			if err != nil {fmt.Println("ERRORS:", errs); return []*Tx{}, -1}
		
			/*// Find the amt (small) Not working atm
			data, err = strip_to(`<span class="text-small text-darken ms-1">`, data)
			errs = append(errs, err)
			if err != nil {log.Fatal(errs)}

			// Pull the amt (small)
			amt, err = pull_to(`<`, data)
			errs = append(errs, err)
			if err != nil {log.Fatal(errs)}
			*/

			if len(addr) == 62 {
				has_segwit = true
				this_tx.Outputs = append(this_tx.Outputs, Output{Address: addr, Amt: amt})
			}
		}

		// Check if transaction is segwit. If so, add to the list.
		if has_segwit {
			txes = append(txes, this_tx)
		}

		lowest_tx_fee = fee
	}

	return txes, lowest_tx_fee

}