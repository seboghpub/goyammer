package main

import (
	"flag"
	"fmt"
	"github.com/seboghpub/goyammer/icon"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/mqu/go-notify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seboghpub/goyammer/internal"
	"github.com/shirou/gopsutil/process"
)

var buildVersion = "to be set by linker"
var buildGithash = "to be set by linker"

type app struct {
	users      *internal.Users
	messages   *internal.Messages
	tmpdir     string
	logo       string
	background bool
}

type Command int

const UsageMsg = `
usage: goyammer <command> [<args>]

commands:
  login      Login to Yammer and get an access token.
  poll       Poll for new messages and notify.  
  version    Display version infos.
  help       Display usage message.
`

const (
	POLL    Command = 0
	LOGIN   Command = 1
	VERSION Command = 2
	HELP    Command = 3
)

func (cmd Command) string() string {
	switch cmd {
	case POLL:
		return "poll"
	case LOGIN:
		return "login"
	case VERSION:
		return "version"
	case HELP:
		return "help"
	default:
		log.Fatal().Msgf("unknown command %d.\n", cmd)
	}
	return ""
}

func main() {

	background := isBackround()

	// initialze logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if !background {

		// a puristic console writer config
		writer := zerolog.ConsoleWriter{Out: os.Stderr}
		log.Logger = log.Output(writer)
	}

	// initialize go-notify
	notify.Init("goyammer")

	// see: https://blog.rapid7.com/2016/08/04/build-a-simple-cli-tool-with-golang/

	// subcommands
	loginCommand := flag.NewFlagSet("", flag.ExitOnError)
	pollCommand := flag.NewFlagSet("", flag.ExitOnError)

	// subcommand flag pointers
	loginClientId := loginCommand.String("client", "", "The client ID. (Required)")
	pollInterval := pollCommand.Uint("interval", 10, "The number of seconds to wait between request clientId. (Optional)")
	pollOutput := pollCommand.String("output", "", "Where to send output to (Optional)")
	pollForeground := pollCommand.Bool("foreground", false, "Run in foreground (Optional)")
	//pollDetached := pollCommand.Bool("detached", false, "internal flag")

	// parse the commandline
	var command = POLL
	var flagArgs []string
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case LOGIN.string():
			command = LOGIN
			flagArgs = os.Args[2:]
		case POLL.string():
			flagArgs = os.Args[2:]
		case VERSION.string():
			command = VERSION
		case HELP.string():
			command = HELP
		default:
			flagArgs = os.Args[1:]
		}
	}

	// depending on the command
	switch command {
	case VERSION:

		fmt.Printf("goyammer version %s git %s", buildVersion, buildGithash)

	case HELP:

		fmt.Print(UsageMsg)

	case LOGIN:

		// parse flags
		errFlags := loginCommand.Parse(flagArgs)
		if errFlags != nil {
			log.Fatal().Err(errFlags).Msgf("failed to parse command line for '%s' subcommand", LOGIN.string())
		}

		// ensure required flags
		if *loginClientId == "" {
			log.Fatal().Msg("missing '--client' parameter")
			os.Exit(1)
		}

		// hand off to business logic
		internal.SetToken(*loginClientId)

	case POLL:

		// parse flags
		errFlags := pollCommand.Parse(flagArgs)
		if errFlags != nil {
			log.Fatal().Err(errFlags).Msgf("failed to parse command line for '%s' subcommand", POLL.string())
		}

		// unless foreground is set
		if !*pollForeground {

			// restart in a detached mode

			// get cwd
			cwd, errCwd := os.Getwd()
			if errCwd != nil {
				log.Fatal().Err(errCwd).Msg("failed to get cwd")
			}

			// construct a file for connecting STDERR and STDOUT of the child, if pollOutput is given
			var file *os.File
			if *pollOutput != "" {

				f, err := os.OpenFile(*pollOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					log.Fatal().Err(err).Msgf("couldn't open %s", *pollOutput)
				}
				file = f
				defer func() {
					_ = f.Close()
				}()

			}

			// construct the command
			detachedFlags := append([]string{POLL.string(), "--foreground"}, flagArgs...)
			cmd := exec.Command(os.Args[0], detachedFlags...)
			cmd.Dir = cwd
			cmd.Stdout = file
			cmd.Stderr = file

			errStart := cmd.Start()
			if errStart != nil {
				log.Fatal().Err(errCwd).Msg("failed to restart")
			}

			pid := cmd.Process.Pid

			errRelease := cmd.Process.Release()
			if errRelease != nil {
				log.Fatal().Err(errRelease).Msg("failed to detach")
			}

			log.Info().Int("PID", pid).Str("logfile", *pollOutput).Msgf("DETACHED")

		} else {

			// start in the foreground

			// get token from file
			token := internal.GetToken()

			// create a tmpdir dir where we store mug shot files and the logo
			// note: we need a temp-dir as github.com/mqu/go-notify only supports file-based logos/mugshots.
			tmpdir, errTmp := ioutil.TempDir("", "goyammer-mugshots")
			if errTmp != nil {
				log.Fatal().Msg(fmt.Sprintf("couldn't create tmpdir directory: %v", errTmp))
			}

			// set the default application logo
			logo := "face-smile-big"
			logoFile, errLogo := internal.DumpImage(tmpdir, "logo", icon.Main)
			if errLogo != nil {
				log.Warn().Err(errLogo).Msg("failed to dump logo")
			} else {
				logo = logoFile.Name()
			}

			// collect application assets
			client := internal.NewClient(token)
			users := internal.NewUsers(client, tmpdir)
			messages := internal.NewMessages(client)
			app := &app{users: users, messages: messages, tmpdir: tmpdir, logo: logo, background: background}
			app.setupCloseHandler()

			systray.Run(func() {
				internal.Systray_init()
				app.doPoll(*pollInterval)
			}, func() {})

		}
	}
}

func isBackround() bool {
	proc, errStat := process.NewProcess(int32(os.Getpid()))
	if errStat != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to stat current process: %s", errStat)
		os.Exit(1)
	}
	background, errBack := proc.Background()
	if errBack != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to get background info: %s", errBack)
		os.Exit(1)
	}
	return background
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func (app *app) setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		//fmt.Printf("\r")
		log.Info().Msg(fmt.Sprintf("SIGTERM received - cleaning up and shutting down"))
		errRm := os.RemoveAll(app.tmpdir)
		if errRm != nil {
			log.Fatal().Err(errRm).Msg(fmt.Sprintf("failed to remove temp dir %s", app.tmpdir))
		}
		os.Exit(0)
	}()
}

func (app *app) doPoll(interval uint) {

	log.Info().Msg(fmt.Sprint("goyammer started"))

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

	internal.Notify(
		"goyammer",
		fmt.Sprintf("Listening on %d groups for user %s.", len(*currentUser.Groups), currentUser.FullName),
		app.logo)

	// POLL messages
	for {
		for _, group := range *currentUser.Groups {
			internal.Systray_poll()
			gid := group.ID

			newMessages, errNM := app.messages.GetNewMessages(gid)
			if errNM != nil {
				log.Warn().Err(errNM).Msg(fmt.Sprintf("failed to get new messages for group %s", group.FullName))
			} else {
				if len(newMessages) > 0 {
					app.handleMessages(group.FullName, newMessages, currentUser)
				}
			}
			internal.Systray_reset()
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

			// replace newlines
			simpleMessage := re.ReplaceAllString(message.Body.Plain, " ")

			// log for background or foreground
			if app.background {

				// construct and format the logMsg
				logMsg := fmt.Sprintf("%s -- %s", simpleMessage, message.WebUrl)
				log.Info().Str("group", groupName).Str("user", user.FullName).Msg(logMsg)
			} else {

				// construct and format the logMsg
				logMsg := fmt.Sprintf("%s - %s | %s",
					internal.ElipseMe(groupName, 6, true),
					internal.ElipseMe(user.FullName, 6, true),
					internal.ElipseMe(simpleMessage, 50, false))
				log.Info().Msg(logMsg)
			}

			// only if no message from the batch has been notified and message was not send by current user
			if !notified && message.SenderID != currentUser.ID {

				// set icon (either mugshot or default logo)
				myIcon := app.logo
				file, errMug := app.users.GetMugFile(user)
				if errMug == nil {
					myIcon = file.Name()
				}

				summary := user.FullName
				body := fmt.Sprintf("%s\n\n%s", message.Body.Plain, message.WebUrl)
				if len(messages) > 1 {
					body = fmt.Sprintf("%s\n\n... and %d more", body, len(messages)-1)
				}

				internal.Notify(summary, body, myIcon)

				notified = true
			}
		}

	}
}
