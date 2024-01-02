package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	token      = flag.String("t", "", "Bot Token")
	dadRegex   = regexp.MustCompile(`(?i)\bI'?m\s+(\w+)`)
	pauseRegex = regexp.MustCompile(`(?i)\b(cigs|cigarette(s)?|milk)\b`)
	winRegex   = regexp.MustCompile(`(?i)(can'?t\s+win|keep\s+(losing))`)
	isPaused   bool
	pauseEnd   time.Time
)

func main() {
	flag.Parse()
	if *token == "" {
		fmt.Println("No Discord Bot Token provided. Please run the bot with -t <bot token>")
		return
	}

	discord, err := createDiscordSession(*token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	fmt.Println("Bot is now running. Press Ctrl + C to exit.")
	waitForInterrupt()
	discord.Close()
}

func createDiscordSession(token string) (*discordgo.Session, error) {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	discord.AddHandler(messageCreate)
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	if err := discord.Open(); err != nil {
		return nil, err
	}

	return discord, nil
}

func waitForInterrupt() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sig
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if shouldSkipMessage(s, m) {
		return
	}

	if isBotPaused() {
		return
	}

	handlePauseTrigger(s, m)
	handleWinLoseTrigger(s, m)
	handleJokeRequest(s, m)
	handleDadResponse(s, m)
}

func shouldSkipMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return m.Author.ID == s.State.User.ID
}

func isBotPaused() bool {
	if isPaused && time.Now().Before(pauseEnd) {
		return true
	}
	if isPaused && time.Now().After(pauseEnd) {
		isPaused = false
	}
	return false
}

func handlePauseTrigger(s *discordgo.Session, m *discordgo.MessageCreate) {
	matches := pauseRegex.FindStringSubmatch(m.Content)
	if len(matches) > 0 {
		pauseWord := matches[0]
		response := "Be back in 20, gonna go grab some " + pauseWord
		s.ChannelMessageSend(m.ChannelID, response)

		isPaused = true
		randomMinutes := rand.Intn(6)
		pauseEnd = time.Now().Add(time.Duration(15+randomMinutes) * time.Minute)
	}
}

func handleWinLoseTrigger(s *discordgo.Session, m *discordgo.MessageCreate) {
	if winRegex.MatchString(m.Content) {
		gifLink := "https://tenor.com/view/are-ya-winning-son-gif-18099517"
		s.ChannelMessageSend(m.ChannelID, gifLink)
	}
}

func handleJokeRequest(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.ToLower(m.Content) == "Hey dad tell me a joke" {
		joke, err := getDadJoke()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Oops, I couldn't fetch a joke right now.")
			return
		}
		s.ChannelMessageSend(m.ChannelID, joke)
	}
}

func handleDadResponse(s *discordgo.Session, m *discordgo.MessageCreate) {
	content := strings.ToLower(m.Content)
	matches := dadRegex.FindStringSubmatch(content)

	if len(matches) > 1 {
		if matches[1] == "dad" {
			// Special response for "I'm dad"
			s.ChannelMessageSend(m.ChannelID, "No, I'm dad!")
		} else {
			// Regular dad joke response
			response := "Hi " + matches[1] + ", I'm Dad!"
			s.ChannelMessageSend(m.ChannelID, response)
		}
	}
}

func getDadJoke() (string, error) {
	req, err := http.NewRequest("GET", "https://icanhazdadjoke.com/", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
