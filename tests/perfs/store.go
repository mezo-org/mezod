package main

import (
	"encoding/json"
	"os"
)

type Accounts struct {
	Accounts map[string]string
}

func loadAccounts() Accounts {
	accounts := Accounts{Accounts: map[string]string{}}
	buf, err := os.ReadFile(accountsStore)
	if err != nil {
		return accounts
	}

	err = json.Unmarshal(buf, &accounts)
	if err != nil {
		panic(err)
	}

	return accounts
}

func saveAccounts(accounts Accounts) {
	buf, err := json.Marshal(accounts)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(accountsStore, buf, 0644)
	if err != nil {
		panic(err)
	}
}
