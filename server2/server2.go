package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"

	"os"

	"golang.org/x/crypto/bcrypt"
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

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashPassword := string(hash)
	return hashPassword, err
}

func matchPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	match := err == nil
	return match, err
}

func receiveHandler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		msg := fmt.Sprintf("error reading request body: %s", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var data Payload
	if err = json.Unmarshal(body, &data); err != nil {
		msg := fmt.Sprintf("error unmarshal json: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	password := data.C2BPaymentQueryResult
	xml, err := xml.MarshalIndent(data.Envelope, "", " ")
	if err != nil {
		msg := fmt.Sprintf("error MarshalIndent Envelope: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	hashPassword, err := hashPassword(password)
	if err != nil {
		msg := fmt.Sprintf("error hash password: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	match, err := matchPassword(password, hashPassword)
	if err != nil {
		msg := fmt.Sprintf("error matching password: %s", err)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	if match {
		if err := os.WriteFile("success.xml", xml, 0644); err != nil {
			msg := fmt.Sprintf("error writing xml in success.xml: %s", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		fmt.Println("XML data written to success.xml")
	}
	fmt.Println("xml:", string(xml))

	if !match {
		if err := os.WriteFile("failed.xml", xml, 0644); err != nil {
			msg := fmt.Sprintf("error writing xml in failex.xml: %s", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		fmt.Println("XML data written to failed.xml")
	}
}

func main() {
	http.HandleFunc("/receive", receiveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
