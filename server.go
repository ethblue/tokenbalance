package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type BalanceResponse struct {
	Name       string `json:"name,omitempty"`
	Wallet     string `json:"wallet,omitempty"`
	Symbol     string `json:"symbol,omitempty"`
	Balance    string `json:"balance"`
	EthBalance string `json:"eth_balance,omitempty"`
	Decimals   uint8  `json:"decimals,omitempty"`
	Block      uint64 `json:"block,omitempty"`
}

type T struct {
	Address string `json:"address"`
	Symbol  string `json:"symbol"`
	Decimal int    `json:"decimal"`
	Type    string `json:"type"`
}

type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func getInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	contract := vars["contract"]
	wallet := vars["wallet"]

	log.Println("Fetching Wallet:", wallet, "at Contract:", contract)

	name, balance, token, decimals, ethAmount, block, err := GetAccount(contract, wallet)

	if err != nil {
		m := ErrorResponse{
			Error:   true,
			Message: "could not find contract address",
		}
		msg, _ := json.Marshal(m)
		w.Write(msg)
		return
	}

	new := BalanceResponse{
		Name:       name,
		Symbol:     token,
		Decimals:   decimals,
		Wallet:     wallet,
		Balance:    balance,
		EthBalance: ethAmount,
		Block:      block,
	}

	j, err := json.Marshal(new)

	if err == nil {
		w.Write(j)
	}
}

func getAllTokensHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	wallet := vars["wallet"]
	var output []BalanceResponse

	raw, err := ioutil.ReadFile("./ethTokens.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var tokens []T
	json.Unmarshal(raw, &tokens)

	for _, t := range tokens {
		log.Println("Fetching Wallet:", wallet, "at Contract:", t.Address)
		name, balance, token, decimals, ethAmount, block, err := GetAccount(t.Address, wallet)

		if err != nil {
			m := ErrorResponse{
				Error:   true,
				Message: "could not find contract address",
			}
			msg, _ := json.Marshal(m)
			w.Write(msg)
			return
		}

		new := BalanceResponse{
			Name:       name,
			Symbol:     token,
			Decimals:   decimals,
			Wallet:     wallet,
			Balance:    balance,
			EthBalance: ethAmount,
			Block:      block,
		}

		output = append(output, new)
	}

	j, err := json.Marshal(output)

	if err == nil {
		w.Write(j)
	}
}

func getTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	vars := mux.Vars(r)
	contract := vars["contract"]
	wallet := vars["wallet"]

	log.Println("Fetching Wallet:", wallet, "at Contract:", contract)

	_, balance, _, _, _, _, err := GetAccount(contract, wallet)

	if err != nil {
		w.Write([]byte("0.0"))
		return
	} else {
		w.Write([]byte(balance))
	}
}

type TokenInfo struct {
	Address string `json:"address"`
	Symbol  string `json:"symbol"`
	Decimal int    `json:"decimal"`
	Type    string `json:"type"`
	Name    string `json:"name"`
}

// All eth tokens why not
func getETHTokensHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	contract := vars["contract"]

	log.Println("Fetching contract:", contract)

	name, symbol, tokenDecimals, err := GetContractInfo(contract)

	// name, balance, token, decimals, ethAmount, block, err := GetAccount(contract, wallet)

	if err != nil {
		m := ErrorResponse{
			Error:   true,
			Message: "could not find contract address",
		}
		msg, _ := json.Marshal(m)
		w.Write(msg)
		return
	}

	new := TokenInfo{
		Address: contract,
		Decimal: int(tokenDecimals),
		Symbol:  symbol,
		Name:    name,
	}

	j, err := json.Marshal(new)

	if err == nil {
		w.Write(j)
	}
}

type BlockInfo struct {
	Number uint64 `json:"number"`
	Data   uint64 `json:"data"`
}

func getBlockInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// vars := mux.Vars(r)
	// contract := vars["contract"]
	blockNumBigInt := new(big.Int)
	blockNumBigInt.SetInt64(4930986)
	//blockNumBigInt.
	blockN, dataInt, err := GetBlockInfo(blockNumBigInt)

	new := BlockInfo{
		Number: blockN,
		Data:   dataInt,
	}

	j, err := json.Marshal(new)

	if err == nil {
		w.Write(j)
	}
}

func StartServer() {
	log.Println("TokenBalance Server Running: http://" + UseIP + ":" + UsePort)
	http.Handle("/", Router())
	http.ListenAndServe(UseIP+":"+UsePort, nil)
}

func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/balance/{contract}/{wallet}", getTokenHandler).Methods("GET")
	r.HandleFunc("/token/{contract}/{wallet}", getInfoHandler).Methods("GET")
	r.HandleFunc("/tokens/{wallet}", getAllTokensHandler).Methods("GET")
	r.HandleFunc("/tokenInfo/{contract}", getETHTokensHandler).Methods("GET")
	r.HandleFunc("/getBlockInfo", getBlockInfo).Methods("GET")
	return r
}
