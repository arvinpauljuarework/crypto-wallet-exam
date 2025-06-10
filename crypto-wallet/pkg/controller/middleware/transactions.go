package middleware

import (
	"crypto-wallet/pkg/controller"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func Transactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// for poc, will just use query parameters
		urlParameters := r.URL.Query()
		accountnumber := urlParameters.Get("accountnumber")
		if strings.TrimSpace(accountnumber) == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		transactions, err := controller.GetTransactions(accountnumber)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactions)
	case http.MethodPost:
		var userTransaction controller.Transaction
		err := json.NewDecoder(r.Body).Decode(&userTransaction)
		if err != nil {
			http.Error(w, errors.Wrap(err, "invalid request").Error(), http.StatusBadRequest)
			return
		}
		userTransaction.Timestamp = time.Now()

		// validate
		err = userTransaction.PostValidate()
		if err != nil {
			http.Error(w, errors.Wrap(err, "invalid request").Error(), http.StatusBadRequest)
			return
		}

		// redis for realtime getting of balance, postgres for eventual consistency, reports, etc.
		balance, err := userTransaction.Execute()
		if err != nil {
			if balance == -1 {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}

		// save transaction
		err = userTransaction.SaveTransaction()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"accountNumber":          userTransaction.UserAccountNumber,
			"transactionType":        userTransaction.Type,
			"transactionAmount":      userTransaction.Amount,
			"balance":                balance,
			"recipientAccountNumber": userTransaction.RecipientAccountNumber,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
