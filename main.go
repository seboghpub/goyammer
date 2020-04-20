package main

import (
	"flag"
	"fmt"
	"github.com/mqu/go-notify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seboghpub/goyammer/internal"
	"io/ioutil"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

var buildVersion = "to be set by linker"
var buildGithash = "to be set by linker"

type app struct {
	users    *internal.Users
	messages *internal.Messages
	tmpdir   string
}

func main() {

	// initialze logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// initialize go-notify
	notify.Init("goyammer")

	// see: https://blog.rapid7.com/2016/08/04/build-a-simple-cli-tool-with-golang/

	// subcommands
	loginCommand := flag.NewFlagSet("login", flag.ExitOnError)
	pollCommand := flag.NewFlagSet("poll", flag.ExitOnError)

	// subcommand flag pointers
	loginClientId := loginCommand.String("client", "", "The client ID. (Required)")
	pollInterval := pollCommand.Uint("interval", 3, "The number of seconds to wait between request clientId. (Optional)")

	// verify that a subcommand has been provided
	if len(os.Args) < 2 {
		log.Fatal().Msg("expected 'login', 'poll', or 'version' subcommand")
	}

	// parse the flags for appropriate FlagSet
	switch os.Args[1] {
	case "login":
		_ = loginCommand.Parse(os.Args[2:])
	case "poll":
		_ = pollCommand.Parse(os.Args[2:])
	case "version":
		fmt.Printf("version: %s, git: %s", buildVersion, buildGithash)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	// doLogin
	if loginCommand.Parsed() {

		// required Flags
		if *loginClientId == "" {
			loginCommand.Usage()
			os.Exit(1)
		}

		// hand off to business logic
		internal.SetToken(*loginClientId)
	}

	if pollCommand.Parsed() {

		// get token from file
		token := internal.GetToken()

		// create a tmpdir dir where we store mug shot files
		tmpdir, errTmp := ioutil.TempDir("", "goyammer")
		if errTmp != nil {
			log.Fatal().Msg(fmt.Sprintf("couldn't create tmpdir directory: %v", errTmp))
		}
		client := internal.NewClient(token)
		users := internal.NewUsers(client, tmpdir)
		messages := internal.NewMessages(client)
		app := &app{users: users, messages: messages, tmpdir: tmpdir}
		app.setupCloseHandler()

		app.doPoll(*pollInterval)

	}
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func (app *app) setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Printf("\r")
		log.Info().Msg(fmt.Sprintf("SIGTERM received - cleaning up and shutting down"))
		errRm := os.RemoveAll(app.tmpdir)
		if errRm != nil {
			log.Fatal().Err(errRm).Msg(fmt.Sprintf("failed to remove temp dir %s", app.tmpdir))
		}
		os.Exit(0)
	}()
}

func (app *app) doPoll(interval uint) {

	log.Info().Msg(fmt.Sprint("goyamer started"))

	sleepTime := time.Duration(interval) * time.Second
	log.Info().Msg(fmt.Sprintf("* polling: every %s", sleepTime.String()))

	// get the current user
	var currentUser *internal.User
	for {
		user, errUser := app.users.GetUser(-1)
		currentUser = user
		if errUser == nil {
			break
		}
		log.Warn().Err(errUser).Msg("failed to get current user")
		time.Sleep(sleepTime)
	}
	log.Info().Msg(fmt.Sprintf("* user: %s", currentUser.FullName))
	log.Info().Msg(fmt.Sprint("* groups:"))
	for _, group := range *currentUser.Groups {
		log.Info().Msg(fmt.Sprintf("  - %s", group.FullName))
	}

	// poll messages
	for {
		for _, group := range *currentUser.Groups {
			gid := group.ID

			newMessages, errNM := app.messages.GetNewMessages(gid)
			if errNM != nil {
				log.Warn().Err(errNM).Msg(fmt.Sprintf("failed to get new messages for group %s", group.FullName))
			} else {
				if len(newMessages) > 0 {
					app.handleMessages(group.FullName, newMessages, currentUser)
				}
			}
			time.Sleep(sleepTime)
		}
	}
}

func (app *app) handleMessages(groupName string, messages []*internal.Message, currentUser *internal.User) {

	// regex matching newline newlines
	re := regexp.MustCompile(`\r?\n`)

	notified := false

	// go through all messages from newest to oldest
	for i := len(messages) - 1; i >= 0; i-- {

		message := messages[i]

		// get the sender
		senderId := message.SenderID
		user, errUser := app.users.GetUser(senderId)
		if errUser != nil {
			log.Warn().Err(errUser).Msg(fmt.Sprintf("failed to get user: %d", senderId))
			continue
		}

		// if there is plain text in the message and we have a full name
		if message.Body.Plain != "" && user.FullName != "" {

			// construct and format the logLine
			logLine := re.ReplaceAllString(message.Body.Plain, " ")
			logLine = fmt.Sprintf("%s - %s | %s",
				internal.ElipseMe(groupName, 6, true),
				internal.ElipseMe(user.FullName, 6, true),
				internal.ElipseMe(logLine, 50, false))

			// log
			log.Info().Msg(logLine)

			// only if no message from the batch has been notified and message was not send by current user
			if !notified && message.SenderID != currentUser.ID {

				// get mugshot file
				icon := "face-smile-big"
				file, errMug := app.users.GetMugFile(user)
				if errMug == nil {
					icon = file.Name()
				}

				summary := user.FullName
				body := message.Body.Plain
				if len(messages) > 1 {
					body = fmt.Sprintf("%s\n... and %d more", body, len(messages)-1)
				}

				internal.Notify(summary, body, icon)

				notified = true
			}
		}

	}
}
