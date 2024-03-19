package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	Soapenv string   `xml:"soapenv,attr"`
	C2b     string   `xml:"c2b,attr"`
	Header  string   `xml:"Header"`
	Body    struct {
		Text                  string `xml:",chardata"`
		C2BPaymentQueryResult struct {
			Text          string `xml:",chardata"`
			ResultCode    string `xml:"ResultCode"`
			ResultDesc    string `xml:"ResultDesc"`
			TransID       string `xml:"TransID"`
			BillRefNumber string `xml:"BillRefNumber"`
			UtilityName   string `xml:"UtilityName"`
			CustomerName  string `xml:"CustomerName"`
			Amount        string `xml:"Amount"`
		} `xml:"C2BPaymentQueryResult"`
	} `xml:"Body"`
}

type Payload struct {
	Envelope              Envelope `xml:"Envelope"`
	C2BPaymentQueryResult string   `xml:"C2BPaymentQueryResult"`
}

func sendDataToServer2(xmlString, password string) {
	var envelope Envelope

	err := xml.NewDecoder(strings.NewReader(xmlString)).Decode(&envelope)
	if err != nil {
		fmt.Printf("error decode XML string: %s\n", err)
		return
	}

	payload := Payload{
		Envelope:              envelope,
		C2BPaymentQueryResult: password,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("error marshal payload: %s", err)
		return
	}

	url := "http://localhost:8080/receive"
	client := http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Unable to send request to Server 2.: ", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Server 2 failed to acknowledge.")
	}

}

func main() {
	xmlString := `
<soapenv:Envelope
xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
xmlns:c2b="http://cps.huawei.com/cpsinterface/c2bpayment">
<soapenv:Header/>
<soapenv:Body>
<c2b:C2BPaymentQueryResult>
<ResultCode>2</ResultCode>
<ResultDesc>Failed</ResultDesc>
<TransID>10111</TransID>
<BillRefNumber>12233</BillRefNumber>
<UtilityName>sddd</UtilityName>
<CustomerName>wee</CustomerName>
<Amount>30</Amount>
</c2b:C2BPaymentQueryResult>
</soapenv:Body>
</soapenv:Envelope>`

	password := "password123456"

	sendDataToServer2(xmlString, password)
}
