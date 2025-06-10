package middleware

import (
	"crypto-wallet/pkg/controller"
	"encoding/json"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func Users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// for poc, will just use query parameters
		urlParameters := r.URL.Query()
		accountnumber := urlParameters.Get("accountnumber")
		if strings.TrimSpace(accountnumber) == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		user, err := controller.GetUser(accountnumber)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	case http.MethodPost:
		var user controller.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, errors.Wrap(err, "invalid request").Error(), http.StatusBadRequest)
			return
		}

		// validate
		err = user.PostValidate()
		if err != nil {
			http.Error(w, errors.Wrap(err, "invalid request").Error(), http.StatusBadRequest)
			return
		}

		// create
		err = user.Create()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
