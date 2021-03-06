package main

import (
	"bytes"
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
	"os/exec"
	"strings"
)

func (i *IRCCat) handleCommand(event *irc.Event) {
	msg := event.Message()
	channel := ""
	respond_to := event.Arguments[0]
	if respond_to[0] != '#' {
		respond_to = event.Nick
	} else {
		channel = respond_to
	}

	if event.Arguments[0][0] != '#' && !i.authorisedUser(event.Nick) {
		// Command not in a channel, or not from an authorised user
		log.Infof("Unauthorised command: %s (%s) %s", event.Nick, respond_to, msg)
		return
	}
	log.Infof("Authorised command: %s (%s) %s", event.Nick, respond_to, msg)

	parts := strings.SplitN(msg, " ", 1)

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(viper.GetString("commands.handler"), event.Nick, channel, respond_to, parts[0][1:])
	} else {
		cmd = exec.Command(viper.GetString("commands.handler"), event.Nick, channel, respond_to, parts[0][1:], parts[1])
	}
	i.runCommand(cmd, respond_to)
}

// Run a command with the output going to the nick/channel identified by respond_to
func (i *IRCCat) runCommand(cmd *exec.Cmd, respond_to string) {
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("Running command %s failed: %s", cmd.Args, err)
		i.irc.Privmsgf(respond_to, "Command failed: %s", err)
	}

	lines := strings.Split(out.String(), "\n")
	line_count := len(lines)
	if line_count > viper.GetInt("commands.max_response_lines") {
		line_count = viper.GetInt("commands.max_response_lines")
	}

	for _, line := range lines[0:line_count] {
		if line != "" {
			i.irc.Privmsg(respond_to, line)
		}
	}
}
