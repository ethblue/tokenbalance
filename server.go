package main

import (
	"bufio"
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

func getBlockInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	blockNumBigInt := new(big.Int)
	var allTokens []TokenInfo
	var i int64
	var minblock int64 = 4379066
	var maxblock int64 = 4380066
	for i = minblock; i <= maxblock; i++ {
		blockNumBigInt.SetInt64(i)
		contractAddresses, err := GetContractsAtBlock(blockNumBigInt)
		if i%50 == 0 {
			fmt.Printf("%d  %d \n", i, i/maxblock)
		}
		if err != nil {
			// Handle error for GetContractsAtBlock
			fmt.Printf("Error at block %d", i)
		} else {
			for _, contractAddress := range contractAddresses {
				name, symbol, tokenDecimals, err := GetContractInfo(contractAddress)
				if err != nil {
					// Do nothing
				} else {
					// Add the token info to the "database"
					newToken := TokenInfo{
						Address: contractAddress,
						Decimal: int(tokenDecimals),
						Symbol:  symbol,
						Name:    name,
					}
					allTokens = append(allTokens, newToken)
				}
			}
		}
	}

	j, err := json.Marshal(allTokens)

	if err == nil {
		w.Write(j)
	}
}

func StartERC20Dump() {
	blockNumBigInt := new(big.Int)
	// var allTokens []TokenInfo
	var i int64
	var minblock int64 = 4080066
	var maxblock int64 = 4380066
	blocksToTraverse := maxblock - minblock

	// For writing to cache
	os.MkdirAll("./output/", 0700)
	f, err := os.Create("./output/tokens" + string(minblock) + "-" + string(maxblock) + ".out")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	// fc, err := os.Create("./block.txt")
	defer f.Close()
	// defer fc.Close()

	w := bufio.NewWriter(f)
	// w2 := bufio.NewWriter(fc)

	w.WriteString("[")
	for i = minblock; i <= maxblock; i++ {
		blockNumBigInt.SetInt64(i)
		contractAddresses, err := GetContractsAtBlock(blockNumBigInt)
		if i%50 == 0 {
			fmt.Printf("%d:  %f%% \n", i, 100.0*float64(i-minblock)/float64(blocksToTraverse))
		}
		if err != nil {
			// Handle error for GetContractsAtBlock
			fmt.Printf("Error at block %d\n", i)
		} else {
			for _, contractAddress := range contractAddresses {
				name, symbol, tokenDecimals, err := GetContractInfo(contractAddress)
				if err != nil {
					// Do nothing
				} else {
					// Add the token info to the "database"
					newToken := TokenInfo{
						Address: contractAddress,
						Decimal: int(tokenDecimals),
						Symbol:  symbol,
						Name:    name,
					}
					j, err := json.Marshal(newToken)
					// allTokens = append(allTokens, newToken)
					if err == nil {
						w.Write(j)
						w.WriteString(",")
					}
				}
			}
		}
		// w2.WriteString("%d\n", i)
		w.Flush()
		// w2.Flush()
	}
	w.WriteString("]")
	w.Flush()
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
