package controller

import (
	"context"
	"crypto-wallet/pkg/db"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Name          string  `json:"name"`
	AccountNumber string  `json:"accountNumber"`
	Balance       float64 `json:"balance"`
}

// can add here validations on user details
func (u User) PostValidate() error {
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("name is required") // can use multierr if other fields were added
	}
	return nil
}

func (u *User) Create() error {
	// duplicate checking - not implemented in poc since only few details were setup
	uuID, err := uuid.NewRandom() // can improve to add more random values
	if err != nil {
		return err
	}
	u.AccountNumber = fmt.Sprintf("%012d%s", time.Now().UnixNano()%1e12, uuID.String()[:10])

	// save to redis
	userRedisKey := fmt.Sprintf("user:%s", u.AccountNumber)
	err = db.Redis().HSet(context.Background(), userRedisKey, map[string]any{
		"name":          u.Name,
		"accountNumber": u.AccountNumber,
		"balance":       "0",
	}).Err()
	if err != nil {
		return err
	}

	// save to postgres (for eventual consistency)
	go func() {
		query := `INSERT INTO users (name, account_number) VALUES ($1, $2) RETURNING id`
		var id int
		err = db.DB().QueryRow(query, u.Name, u.AccountNumber).Scan(&id)
		if err != nil {
			fmt.Println("err", err) // can put this to log or audit
		}
	}()
	return nil
}

func GetUser(accountNumber string) (user User, err error) {
	userRedisKey := fmt.Sprintf("user:%s", accountNumber)
	userMap, err := db.Redis().HGetAll(context.Background(), userRedisKey).Result()
	if err != nil {
		return user, err
	}

	balance, err := strconv.ParseFloat(userMap["balance"], 64)
	if err != nil {
		return user, err
	}

	user = User{
		Name:          userMap["name"],
		AccountNumber: userMap["accountNumber"],
		Balance:       balance,
	}
	return
}
