package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/oauth2.v3/models"
	"log"
	"net/http"

	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

var (
	username = "heimdall"
	password = "secret"
)

func main() {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	clientStore := store.NewClientStore()
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	http.HandleFunc("/protected", validateToken(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, I'm protected"))
	}, srv))

	http.HandleFunc("/credentials", func(w http.ResponseWriter, r *http.Request) {
		clientId := uuid.New().String()[:8]
		clientSecret := uuid.New().String()[:8]
		err := clientStore.Set(clientId, &models.Client{
			ID:     clientId,
			Secret: clientSecret,
			Domain: "http://localhost:9094",
		})
		if err != nil {
			fmt.Println(err.Error())
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"CLIENT_ID": clientId, "CLIENT_SECRET": clientSecret})
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		srv.HandleTokenRequest(w, r)
	})

	http.HandleFunc("/basic", func (w http.ResponseWriter, r *http.Request) {

		u, p, ok := r.BasicAuth()
		if !ok {
			fmt.Println("Error parsing basic auth")
			w.WriteHeader(401)
			return
		}
		if u != username {
			fmt.Printf("Username provided is correct: %s\n", u)
			w.WriteHeader(401)
			return
		}
		if p != password {
			fmt.Printf("Password provided is correct: %s\n", u)
			w.WriteHeader(401)
			return
		}
		fmt.Printf("Username: %s\n", u)
		fmt.Printf("Password: %s\n", p)
		w.WriteHeader(200)
		return
	})

	log.Fatal(http.ListenAndServe(":9096", logRequest(http.DefaultServeMux)))
}

func validateToken(f http.HandlerFunc, srv *server.Server) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := srv.ValidationBearerToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		f.ServeHTTP(w, r)
	})
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}


