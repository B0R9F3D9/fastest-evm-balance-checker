package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/forta-network/go-multicall"
	"github.com/AlecAivazis/survey/v2"
)

var (
	BalancesByChain map[string][]BalanceData
	mutex           sync.Mutex
)

func getBalance(chain Chain, token Token, wallets []Wallet, wg *sync.WaitGroup) {
	defer wg.Done()

	caller, err := multicall.Dial(context.Background(), chain.RPC)
	if err != nil {
		log.Printf("Ошибка создания caller для %s: %v\n", chain.RPC, err)
		return
	}

	var abi string
	var methodName string
	if (token.Symbol == "ETH" || token.Symbol == "MATIC" || token.Symbol == "BNB" || token.Symbol == "FTM") { // native
		abi = ETH_ABI
		methodName = "getEthBalance"
	} else {
		abi = ERC20_ABI
		methodName = "balanceOf"
	}

	contract, err := multicall.NewContract(abi, token.Address)
	if err != nil {
		log.Printf("Ошибка создания контракта для %s: %v\n", token.Symbol, err)
		return
	}

	var calls []*multicall.Call
	for _, wallet := range wallets {
		calls = append(
			calls,
			contract.NewCall(
				new(balanceOutput),
				methodName,
				common.HexToAddress(wallet.Address),
			).Name(wallet.Address),
		)
	}

	walletsResults, err := caller.Call(nil, calls...)
	if err != nil {
		log.Printf("Ошибк при вызове %s метода контракта %s: %v\n", chain.Name, token.Symbol, err)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for i, walletResult := range walletsResults {
		balance := walletResult.Outputs.(*balanceOutput).Balance
		balanceFloat := new(big.Float).SetInt(balance)
		exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token.Decimals)), nil)
		balanceFloat.Quo(balanceFloat, new(big.Float).SetInt(exp))

		if _, ok := BalancesByChain[chain.Name]; !ok {
			BalancesByChain[chain.Name] = make([]BalanceData, len(wallets))
		}

		BalancesByChain[chain.Name][i].Index = i
		BalancesByChain[chain.Name][i].Address = walletResult.CallName
		if BalancesByChain[chain.Name][i].Tokens == nil {
			BalancesByChain[chain.Name][i].Tokens = make(map[string]string)
		}
		BalancesByChain[chain.Name][i].Tokens[token.Symbol] = balanceFloat.Text('f', -1)
	}
}

func writeToCSV(chain Chain, balanceByChain []BalanceData) error {
	file, err := os.Create("results/" + chain.Name + ".csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"№", "Адрес"}
	for _, token := range chain.Tokens {
		headers = append(headers, token.Symbol)
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, balance := range balanceByChain {
		record := make([]string, len(headers))
		record[0] = fmt.Sprintf("%d", balance.Index)
		record[1] = balance.Address

		for j, tokenSymbol := range headers[2:] {
			record[j+2] = balance.Tokens[tokenSymbol]
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	fmt.Printf("Результаты %d %s успешно записаны в results/%s.csv\n", len(balanceByChain), chain.Name, chain.Name)
	return nil
}

func main() {
	chains, err := readChainsFromConfig("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка при чтении сетей из конфига: %v\n", err)
	}
	wallets, err := readWalletsFromFile("wallets.txt")
	if err != nil {
		log.Fatalf("Ошибка при чтении кошельков из файла: %v\n", err)
	}

	fmt.Printf("Найдено %d сетей и %d кошельков\n", len(chains), len(wallets))

	if _, err := os.Stat("results"); os.IsNotExist(err) {
        os.Mkdir("results", 0755)
    }

	var selectedChain string
	options := []string{"Все"}
	for _, chain := range chains {
		options = append(options, chain.Name)
	}

	prompt := &survey.Select{
		Message: "Выберите сеть:",
		Options: options,
	}
	err = survey.AskOne(prompt, &selectedChain)
	if err != nil {
		log.Fatalf("Ошибка при выборе сети: %v\n", err)
	}
	var chainsToProcess []Chain
	if selectedChain == "Все" {
		chainsToProcess = chains
	} else {
		for _, chain := range chains {
			if chain.Name == selectedChain {
				chainsToProcess = append(chainsToProcess, chain)
				break
			}
		}
	}

	BalancesByChain = make(map[string][]BalanceData)
	mutex = sync.Mutex{}

	startTime := time.Now()
	var wg sync.WaitGroup
	for _, chain := range chainsToProcess {
		for _, token := range chain.Tokens {
			wg.Add(1)
			go getBalance(chain, token, wallets, &wg)
		}
	}
	wg.Wait()

	for _, chain := range chainsToProcess {
		if err := writeToCSV(chain, BalancesByChain[chain.Name]); err != nil {
			log.Fatalf("Ошибка при записи в CSV файл: %v\n", err)
		}
	}

	fmt.Printf("Время выполнения: %s\n", time.Since(startTime))
}
