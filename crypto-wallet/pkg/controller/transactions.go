package controller

import (
	"context"
	"crypto-wallet/pkg/db"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Transaction struct {
	UserAccountNumber      string    `json:"userAccountNumber"`
	Type                   string    `json:"type"`
	Amount                 float64   `json:"amount"`
	RecipientAccountNumber string    `json:"recipientAccountNumber"`
	Timestamp              time.Time `json:"timestamp"`
}

// can use multierr if other fields were added
func (t Transaction) PostValidate() error {
	if strings.ToLower(t.Type) != "deposit" && strings.ToLower(t.Type) != "withdraw" && strings.ToLower(t.Type) != "transfer" && strings.ToLower(t.Type) != "check" {
		return errors.New("invalid transaction type")
	}
	if strings.ToLower(t.Type) == "transfer" {
		if strings.TrimSpace(t.RecipientAccountNumber) == "" {
			return errors.New("recipient account number is required")
		} else {
			if t.RecipientAccountNumber == t.UserAccountNumber {
				return errors.New("recipient account number is invalid")
			}
			_, err := getUserBalance(t.RecipientAccountNumber)
			if err != nil {
				return errors.New("recipient account number is invalid")
			}
		}
	}
	return nil
}

func (t Transaction) Execute() (balance float64, err error) {
	switch t.Type {
	case "deposit":
		balance, err = updateUserBalance(t.UserAccountNumber, t.Amount)
	case "withdraw":
		balance, err = getUserBalance(t.UserAccountNumber)
		if err != nil {
			return
		}
		if (balance - t.Amount) < 1 {
			return balance, errors.New("insufficient balance")
		}
		balance, err = updateUserBalance(t.UserAccountNumber, -1*t.Amount)
	case "transfer":
		balance, err = getUserBalance(t.UserAccountNumber)
		if err != nil {
			return
		}
		if (balance - t.Amount) < 1 {
			return balance, errors.New("insufficient balance")
		}
		// recipient
		_, err = updateUserBalance(t.RecipientAccountNumber, t.Amount)
		if err != nil {
			return
		}

		balance, err = updateUserBalance(t.UserAccountNumber, -1*t.Amount)
	default: // check
		balance, err = getUserBalance(t.UserAccountNumber)
	}
	return
}

func getUserBalance(userAccountNumber string) (float64, error) {
	userRedisKey := fmt.Sprintf("user:%s", userAccountNumber)
	userBalance, err := db.Redis().HGet(context.Background(), userRedisKey, "balance").Result()
	if err == redis.Nil {
		return 0, errors.New("user balance not found")
	}

	if err != nil {
		return -1, err
	}

	return strconv.ParseFloat(userBalance, 64)
}

func updateUserBalance(userAccountNumber string, depositAmount float64) (float64, error) {
	userRedisKey := fmt.Sprintf("user:%s", userAccountNumber)
	userUpdatedBalance, err := db.Redis().HIncrByFloat(context.Background(), userRedisKey, "balance", depositAmount).Result()
	if err != nil {
		return -1, err
	}

	return userUpdatedBalance, nil
}

func (t Transaction) SaveTransaction() error {
	// save to redis
	transactionJSON, err := json.Marshal(t)
	if err != nil {
		return err
	}

	userTransactionsRedisKey := fmt.Sprintf("user:%s:transactions", t.UserAccountNumber)
	_, err = db.Redis().RPush(context.Background(), userTransactionsRedisKey, transactionJSON).Result()
	if err != nil {
		return err
	}

	if t.Type == "transfer" {
		userTransactionsRedisKey := fmt.Sprintf("user:%s:transactions", t.RecipientAccountNumber)
		_, err = db.Redis().RPush(context.Background(), userTransactionsRedisKey, transactionJSON).Result()
		if err != nil {
			return err
		}
	}

	// save to postgres (for eventual consistency)
	go func() {
		query := `INSERT INTO transactions (user_account_number, transaction_type, transaction_amount, recipient_account_number) VALUES ($1, $2, $3, $4)`
		err := db.DB().QueryRow(query, t.UserAccountNumber, t.Type, t.Amount, t.RecipientAccountNumber).Err()
		if err != nil {
			fmt.Println("err", err) // can put this to log or audit
		}
	}()

	return nil
}

func GetTransactions(userAccountNumber string) (transactions []Transaction, err error) {
	userTransactionsRedisKey := fmt.Sprintf("user:%s:transactions", userAccountNumber)
	userTransactionsList, err := db.Redis().LRange(context.Background(), userTransactionsRedisKey, 0, -1).Result()
	for _, userTransaction := range userTransactionsList {
		var transaction Transaction
		err := json.Unmarshal([]byte(userTransaction), &transaction)
		if err != nil {
			return transactions, err
		}
		transactions = append(transactions, transaction)
	}
	// sort
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Timestamp.After(transactions[j].Timestamp)
	})
	return
}
