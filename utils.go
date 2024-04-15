package main

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

const (
	ERC20_ABI = `[
		{
			"constant":true,
			"inputs":[
					{
						"name":"tokenOwner",
						"type":"address"
					}
			],
			"name":"balanceOf",
			"outputs":[
					{
						"name":"balance",
						"type":"uint256"
					}
			],
			"payable":false,
			"stateMutability":"view",
			"type":"function"
		}
	]`
	ETH_ABI = `[
		{
			"constant":true,
			"inputs":[
					{
						"name":"addr",
						"type":"address"
					}
			],
			"name":"getEthBalance",
			"outputs":[
					{
						"name":"balance",
						"type":"uint256"
					}
			],
			"payable":false,
			"stateMutability":"view",
			"type":"function"
		}
	]`
)

func readWalletsFromFile(filename string) ([]Wallet, error) {
	var wallets []Wallet

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	index := 1
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line != "" {
			wallets = append(wallets, Wallet{
				Index:   index,
				Address: line,
			})
			index++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при сканировании файла: %v", err)
	}

	return wallets, nil
}

func readChainsFromConfig(configName string) ([]Chain, error) {
    var c Config

    file, err := os.ReadFile(configName)
    if err != nil {
        return nil, err
    }

    err = yaml.Unmarshal(file, &c)
    if err != nil {
        return nil, err
    }

    return c.Chains, nil
}
