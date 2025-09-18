package main

import (
	"flag"
	"io"
	"log/slog"
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
	dadRegex   = regexp.MustCompile(`(?i)\bI'?m\s+(.+)`)
	pauseRegex = regexp.MustCompile(`(?i)\b(cigs|cigarette(s)?|milk)\b`)
	winRegex   = regexp.MustCompile(`(?i)(can'?t\s+win|keep\s+(losing))`)
	isPaused   bool
	pauseEnd   time.Time
)

func init() {
	// Configure structured logging for journald STONKS
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				// RFC3339 timestamp
				return slog.Attr{Key: "timestamp", Value: a.Value}
			}
			return a
		},
	}))
	slog.SetDefault(logger)
}

func main() {
	flag.Parse()

	// Use token from flag, or fall back to environment variable
	botToken := *token
	if botToken == "" {
		botToken = os.Getenv("DISCORD_BOT_TOKEN")
	}

	if botToken == "" {
		slog.Error("No Discord Bot Token provided. Please provide via -t flag or DISCORD_BOT_TOKEN environment variable")
		os.Exit(1)
	}

	discord, err := createDiscordSession(botToken)
	if err != nil {
		slog.Error("Error creating Discord session", "error", err)
		os.Exit(1)
	}

	slog.Info("Bot is now running. Press Ctrl + C to exit.", "service", "dadbot", "event", "startup")
	waitForInterrupt()
	discord.Close()
	slog.Info("Bot shutting down gracefully", "service", "dadbot", "event", "shutdown")
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
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sig
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if shouldSkipMessage(s, m) {
		return
	}

	if isBotPaused() {
		slog.Debug("Message skipped - bot is paused", "event", "message_skipped_paused", "service", "dadbot")
		return
	}

	// Track if any dad-response was sent
	responseSent := false
	responseSent = handlePauseTrigger(s, m) || responseSent
	responseSent = handleWinLoseTrigger(s, m) || responseSent
	responseSent = handleJokeRequest(s, m) || responseSent
	responseSent = handleDadResponse(s, m) || responseSent

	// Log non-dad-response messages
	if !responseSent {
		slog.Info("Message processed",
			"event", "message_processed",
			"service", "dadbot")
	}
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

func handlePauseTrigger(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	matches := pauseRegex.FindStringSubmatch(m.Content)
	if len(matches) > 0 {
		pauseWord := matches[0]
		response := "Be back in 20, gonna go grab some " + pauseWord
		s.ChannelMessageSend(m.ChannelID, response)

		isPaused = true
		randomMinutes := rand.Intn(6)
		pauseDuration := time.Duration(15+randomMinutes) * time.Minute
		pauseEnd = time.Now().Add(pauseDuration)

		slog.Info("Bot paused by trigger word",
			"event", "pause_triggered",
			"service", "dadbot",
			"pause_minutes", 15+randomMinutes)
		return true
	}
	return false
}

func handleWinLoseTrigger(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if winRegex.MatchString(m.Content) {
		gifLink := "https://tenor.com/view/are-ya-winning-son-gif-18099517"
		s.ChannelMessageSend(m.ChannelID, gifLink)

		slog.Info("Win/lose GIF sent",
			"event", "win_lose_response",
			"service", "dadbot",
			"trigger", "win_lose_pattern")
		return true
	}
	return false
}

func handleJokeRequest(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if strings.ToLower(m.Content) == "tell me a joke" {
		joke, err := getDadJoke()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Gosh dang joke AI always breakin. Tell Clutch to fix it.")
			slog.Error("Failed to fetch dad joke",
				"event", "joke_request_failed",
				"service", "dadbot",
				"error", err.Error())
			return true // Count an ERROR as reading message, maybe joke service is broke
		}
		s.ChannelMessageSend(m.ChannelID, joke)

		slog.Info("Dad joke sent",
			"event", "joke_request_fulfilled",
			"service", "dadbot")
		return true
	}
	return false
}

func handleDadResponse(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	matches := dadRegex.FindStringSubmatch(m.Content)

	if len(matches) > 1 {
		extracted := strings.TrimSpace(matches[1])
		var response string
		var responseType string

		if strings.ToLower(extracted) == "dad" {
			response = "No, I'm dad!"
			responseType = "dad_paradox"
		} else {
			response = "Hi " + extracted + ", I'm Dad!"
			responseType = "dad_joke"
		}

		s.ChannelMessageSend(m.ChannelID, response)

		slog.Info("Dad response sent",
			"event", "dad_response_sent",
			"service", "dadbot",
			"response_type", responseType)
		return true
	}
	return false
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
