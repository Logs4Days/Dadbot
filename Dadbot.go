package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string

func main() {

	// Check cli input for bot token
	if token == "" {
		fmt.Println("No Discord Bot Token provided. Please run dadbot -t <bot token>")
		return
	}

	// Start discord session with provided bot token
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating discord session: ", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	discord.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	// Open discord websocket and begin listening for messages
	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}

	// Wait for CTRL-C or other interrupt
	fmt.Println("DadBot is now running. Press Ctrl + C to exit")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sig

	// If we see interrupt, cleanly close the discord websocket
	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages from the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Get all the "I'm" messages sent by users -
	// This can be much cleaner - condense this later - :)
	if strings.Contains(m.Content, "i'm ") {
		msg := m.Content
		msgSplit := strings.SplitAfter(msg, "i'm ")
		s.ChannelMessageSend(m.ChannelID, "Hi "+msgSplit[1]+", I'm Dad!")
		msg = ""
		return
	}
	if strings.Contains(m.Content, "I'm ") {
		msg := m.Content
		msgSplit := strings.SplitAfter(msg, "I'm ")
		s.ChannelMessageSend(m.ChannelID, "Hi "+msgSplit[1]+", I'm Dad!")
		msg = ""
		return
	}
	if strings.Contains(m.Content, "im ") {
		msg := m.Content
		msgSplit := strings.SplitAfter(msg, "im ")
		s.ChannelMessageSend(m.ChannelID, "Hi "+msgSplit[1]+", I'm Dad!")
		msg = ""
		return
	}
	if strings.Contains(m.Content, "i am ") {
		msg := m.Content
		msgSplit := strings.SplitAfter(msg, "i am ")
		s.ChannelMessageSend(m.ChannelID, "Hi "+msgSplit[1]+", I'm Dad!")
		msg = ""
		return
	}
	if strings.Contains(m.Content, "I am ") {
		msg := m.Content
		msgSplit := strings.SplitAfter(msg, "I am ")
		s.ChannelMessageSend(m.ChannelID, "Hi "+msgSplit[1]+", I'm Dad!")
		msg = ""
		return
	}
	if strings.Contains(m.Content, "IM ") {
		msg := m.Content
		msgSplit := strings.SplitAfter(msg, "IM ")
		s.ChannelMessageSend(m.ChannelID, "Hi "+msgSplit[1]+", I'm Dad!")
		msg = ""
		return
	}
	if strings.Contains(m.Content, "i m ") {
		msg := m.Content
		msgSplit := strings.SplitAfter(msg, "i m ")
		s.ChannelMessageSend(m.ChannelID, "Hi "+msgSplit[1]+", I'm Dad!")
		msg = ""
		return
	}
	if strings.Contains(m.Content, "Im ") {
		msg := m.Content
		msgSplit := strings.SplitAfter(msg, "Im ")
		s.ChannelMessageSend(m.ChannelID, "Hi "+msgSplit[1]+", I'm Dad!")
		msg = ""
		return
	}

}
