package main

import (
	"math/big"
)

type balanceOutput struct {
	Balance *big.Int
}

type BalanceData struct {
	Index   int
	Address string
	Tokens  map[string]string
}

type Wallet struct {
	Index   int    `json:"index"`
	Address string `json:"address"`
}

type Token struct {
	Symbol   string `yaml:"Symbol"`
	Address  string `yaml:"Address"`
	Decimals int    `yaml:"Decimals"`
}

type Chain struct {
	Name   string  `yaml:"Name"`
	RPC    string  `yaml:"RPC"`
	Tokens []Token `yaml:"Tokens"`
}

type Config struct {
	Chains []Chain `yaml:"Chains"`
}
