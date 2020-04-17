package internal

import (
	"context"
	"fmt"
	"github.com/phayes/freeport"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"time"
)

const YammerOAuthUrl = "https://www.yammer.com/dialog/oauth"

// "Controller HTML" to orchestrate the passing of an access token to the server.
//
// This HTML executes Javascript that extracts the access token from the URI fragment (set by the authentication server)
// and then sends the extracted access token via URL parameter of an Ajax-Fetch request to the server (SYN). As a
// response for the fetch request, the server acknowledges the token (SYN-ACK). The client, finally fetches a second
// endpoint and thereby signals the server that it may shot down now (ACK).
const fragmentExtractHtml = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <script>
        if(window.location.hash) {
            const fragment=window.location.hash.substr(1);
            const token=fragment.split('=')[1]
            fetch('/token?token='+token)
                .then((response) => {
                    response.text().then(function(text) {
                        document.getElementById("result").innerHTML = text;
                        fetch('/done')
                    })
                })
        } else {
            document.getElementById("result").innerHTML = "something went wrong";
        }
    </script>
</head>
<body>
    <span id="result"></span>
</body>
</html>
`

// The redirect handler simply returns the "Controller HTML".
func redirectHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := io.WriteString(w, fragmentExtractHtml)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to send controller HTML")
	}
}

// The handler to receive the access token via URL parameter (SYN) and to acknowledge the token (SYN-ACK)
func makeTokenHandler(tokenChannel chan string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()["token"]
		if len(values) < 1 {
			log.Fatal().Msg("missing token in client request")
		}
		// send the token to the parent
		tokenChannel <- values[0]
		close(tokenChannel)
		_, _ = fmt.Fprintf(w, "success")
	}
}

// The handler to receive the final call indicating that the server may shot down now (ACK).
func makeDoneHandler(c chan bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		c <- true
	}
}

// Authenticate authenticate a user via OAUTH2 Implicit Flow and returns the access token.
func Authenticate(clientId string) string {

	// create channels for communicating with the handlers
	tokenChannel := make(chan string, 1)
	resultChannel := make(chan bool, 1)

	// create handler
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/redirect", redirectHandler)
	mux.HandleFunc("/token", makeTokenHandler(tokenChannel))
	mux.HandleFunc("/done", makeDoneHandler(resultChannel))

	// get a free port
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get a free TCP port")
	}

	// configure server
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", +port),
		Handler: mux,
	}

	// start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// tell the user which URL to open in his browser
	redirectURI := fmt.Sprintf("http://localhost:%d/oauth/redirect", port)
	authUrl := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=token", YammerOAuthUrl, clientId, redirectURI)
	fmt.Printf("please authorize at: %s\n", authUrl)

	// wait for the token (SYN)
	token := <-tokenChannel

	// wait for the signal to shutdown the server (ACK)
	<-resultChannel

	// shutdown server (in a timeout context).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to shutdown server")
	}

	return token
}
