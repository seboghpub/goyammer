package internal

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"path"
)

const tokeFile = ".goyammer-token"

// SetToken acquires an access token and store it to a file.
func SetToken(clientId string) {

	// tokenPath for the token
	home, _ := os.UserHomeDir()
	tokenPath := path.Join(home, tokeFile)

	// authenticate
	token := Authenticate(clientId)

	// save the token
	err := ioutil.WriteFile(tokenPath, []byte(token), 0600)
	if err != nil {
		log.Fatal().Err(err).Msg(fmt.Sprintf("failed to write token to %s", tokenPath))
	}
	log.Info().Msg(fmt.Sprintf("token written to %s", tokenPath))
}

// GetToken read an access token from a file and returns it.
func GetToken() string {

	// tokenPath for the token
	home, _ := os.UserHomeDir()
	tokenPath := path.Join(home, tokeFile)

	// stat the tokenPath
	_, errStat := os.Stat(tokenPath)

	if os.IsNotExist(errStat) {
		log.Fatal().Msg(fmt.Sprintf("%s missing, use 'login'", tokenPath))
	}
	if errStat != nil {
		log.Fatal().Err(errStat)
	}

	// read the file
	token, errRead := ioutil.ReadFile(tokenPath)
	if errRead != nil {
		log.Fatal().Err(errRead).Msg(fmt.Sprintf("failed to read token from %s", tokenPath))
	}

	// return the token
	return string(token)

}
