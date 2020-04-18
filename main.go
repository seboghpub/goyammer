package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"github.com/spf13/viper"
)

func makeHandler(dataChannel chan YammerOAuthAccessResponse) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// First, we need to get the value of the `code` query param
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		code := r.FormValue("code")

		// Next, lets for the HTTP request to call the github oauth enpoint
		// to get our access token
		reqURL := fmt.Sprintf("https://www.yammer.com/oauth2/access_token?client_id=%s&client_secret=%s&code=%s", viper.GetString("clientID"), viper.GetString("clientSecret"), code)
		req, err := http.NewRequest(http.MethodPost, reqURL, nil)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		// We set this header since we want the response
		// as JSON
		req.Header.Set("accept", "application/json")

		// Send out the HTTP request
		httpClient := http.Client{}
		res, err := httpClient.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		defer res.Body.Close()

		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not read body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		//fmt.Printf("body: %s\n", string(bodyBytes))

		// Parse the request body into the `OAuthAccessResponse` struct
		var t YammerOAuthAccessResponse
		if err := json.Unmarshal(bodyBytes, &t); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		dataChannel <- t
		close(dataChannel)

		// Finally, send a response to redirect the user to the "welcome" page
		// with the access token
		w.Header().Set("Location", "/welcome")
		w.WriteHeader(http.StatusFound)
	}
}

func makeWelcomeHandler(c chan bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "success")
		c <- true
	}
}

func main() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	dataChannel := make(chan YammerOAuthAccessResponse, 1)
	welcomeChannel := make(chan bool, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/redirect", makeHandler(dataChannel))
	mux.HandleFunc("/welcome", makeWelcomeHandler(welcomeChannel))

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server in a goroutine.
	go func() {

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stdout, "Server start failed: %v\n", err)
		}
	}()

	authUrl := fmt.Sprintf("https://www.yammer.com/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=http://localhost:8080/oauth/redirect", viper.GetString("clientID"))
	fmt.Fprintf(os.Stdout, "please authorize at: %s\n", authUrl)

	data := <-dataChannel
	<-welcomeChannel

	// Gracefully shutdown server (in a timeout context).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stdout, "Server shutdown failed.\n")
	}

	fmt.Printf("FullName: %s, token: %s\n", data.User.FullName, data.AccessToken.Token)

}

// YammerOAuthAccessResponse is the data structure for captiuring the access token
type YammerOAuthAccessResponse struct {
	AccessToken struct {
		UserID int    `json:"user_id"`
		Token  string `json:"token"`
	} `json:"access_token"`
	User struct {
		FullName string `json:"full_name"`
	} `json:"user"`
}
