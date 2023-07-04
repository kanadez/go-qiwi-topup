package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const qiwi_auth_url string = "https://edge.qiwi.com"
const qiwi_auth_token string = "cb1686d8f7c5cb34ca35f8ba73ad5ba7"

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		printHelp()
		os.Exit(0)
	}

	phone := args[0]
	sum, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatal("sum must be an integer number")

	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   20 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	processQiwiRequest(client, phone, sum)
}

func printHelp() {
	fmt.Println("Usage: program <phone> <sum>")
}

func processQiwiRequest(client *http.Client, phone string, sum int) (string, error) {
	operator := 99
	method := fmt.Sprintf("/sinap/api/v2/terms/%d/payments", operator)

	type Sum struct {
		Amount   int    `json:"amount"`
		Currency string `json:"currency"`
	}
	type Paymentmethod struct {
		Type      string `json:"type"`
		Accountid string `json:"accountId"`
	}
	type Fields struct {
		Account string `json:"account"`
	}
	type QiwiRequestPayload struct {
		ID            int64         `json:"id"`
		Sum           Sum           `json:"sum"`
		Paymentmethod Paymentmethod `json:"paymentMethod"`
		Fields        Fields        `json:"fields"`
	}

	var qiwi_payload = QiwiRequestPayload{
		ID: time.Now().Unix() * 1000,
		Sum: Sum{
			Amount:   sum,
			Currency: "643"},
		Paymentmethod: Paymentmethod{
			Type:      "Account",
			Accountid: "643"},
		Fields: Fields{
			Account: phone}}

	qiwi_payload_json, err := json.Marshal(qiwi_payload)

	if err != nil {
		log.Println(err)
	}

	qiwi_payload_json_as_string := string(qiwi_payload_json)

	req, err := http.NewRequest(http.MethodPost, qiwi_auth_url+method, bytes.NewBuffer(qiwi_payload_json))

	if err != nil {
		fmt.Println(err.Error())
		return "", errors.Wrap(err, "account NewRequest")
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+qiwi_auth_token)
	req.Header.Add("Host", "edge.qiwi.com")
	req.Header.Add("Content-Length", strconv.Itoa(len(qiwi_payload_json_as_string)))

	fmt.Printf("Sending POST request to %s trying to topup number %s with %d", qiwi_auth_url, phone, sum)

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
		return "", errors.Wrap(err, "account Do")
	}

	fmt.Printf("QIWI server response status: %s", resp.Status)

	defer resp.Body.Close() // defer очищает ресурс

	return "", nil
}
