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
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	token      = flag.String("t", "", "Bot Token")
	dadRegex   = regexp.MustCompile(`(?i)\bI'?m\s+(.+)`)
	pauseRegex = regexp.MustCompile(`(?i)\b(cigs|cigarette(s)?|milk)\b`)
	catgirlRegex = regexp.MustCompile(`(?i)\b(budget|money|dollar?)\b`)
	thermostatRegex = regexp.MustCompile(`(?i)\btoo (hot|cold)\b`)
	// Pause state protected by mutex to prevent race conditions
	// when multiple Discord messages are processed concurrently
	pauseMu  sync.RWMutex
	isPaused bool
	pauseEnd time.Time

	// Rate limiting for joke API (one request per 5 seconds)
	jokeMu       sync.Mutex
	lastJokeTime time.Time
	jokeCooldown = 5 * time.Second

	goodnightMu       sync.Mutex
	lastGoodnightTime time.Time
	goodnightCooldown = 3 * time.Second
	goodnightMessages = []string{
		"Goodnight Snore-osaurus Rex.",
		"Goodnight? I’ll try… but I’ve been practicing for the greatnight.",
		"Goodnight! Sleep tight!",
		"Goodnight? Careful… last time I went to bed early, I woke up in tomorrow.",
		"Goodnight! Don’t let the bedbugs byte… they’re terrible at debugging.",
		"Don't forget to brush your teeth",
		
	}
	thermostatMessages = []string{
		"Don't touch that thermostat!",
		"Don't touch the thermostat! You don't pay the bills around here!",
	}
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
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
	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

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

	// Process triggers in priority order, stop after first match
	responseSent := handlePauseTrigger(s, m) ||
		handleThermostatRequest(s, m) ||
		handleDadResponse(s, m) ||
		handleWinLoseTrigger(s, m) ||
		handleJokeRequest(s, m) ||
		handleGoodnightRequest(s, m) ||
		handleCatGirlRequest(s, m)

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
	pauseMu.RLock()
	paused := isPaused
	end := pauseEnd
	pauseMu.RUnlock()

	if paused && time.Now().Before(end) {
		return true
	}
	if paused && time.Now().After(end) {
		pauseMu.Lock()
		isPaused = false
		pauseMu.Unlock()
	}
	return false
}

func handlePauseTrigger(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	matches := pauseRegex.FindStringSubmatch(m.Content)
	if len(matches) > 0 {
		pauseWord := matches[0]
		response := "Be back in 20, gonna go grab some " + pauseWord
		s.ChannelMessageSend(m.ChannelID, response)

		randomMinutes := rand.Intn(6)
		pauseDuration := time.Duration(15+randomMinutes) * time.Minute

		pauseMu.Lock()
		isPaused = true
		pauseEnd = time.Now().Add(pauseDuration)
		pauseMu.Unlock()

		slog.Info("Bot paused by trigger word",
			"event", "pause_triggered",
			"service", "dadbot",
			"pause_minutes", 15+randomMinutes)
		return true
	}
	return false
}

func handleWinLoseTrigger(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	msg := strings.ToLower(m.Content)
	if msg != "can't win" && msg != "cant win" && msg != "keep losing" {
		return false
	}

	gifLink := "https://tenor.com/view/are-ya-winning-son-gif-18099517"
	s.ChannelMessageSend(m.ChannelID, gifLink)

	slog.Info("Win/lose GIF sent",
		"event", "win_lose_response",
		"service", "dadbot",
		"trigger", msg)
	return true
}

func handleGoodnightRequest(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	msg := strings.ToLower(m.Content)
	if msg != "good night" && msg != "goodnight" {
		return false
	}

	goodnightMu.Lock()
	if time.Since(lastGoodnightTime) < goodnightCooldown {
		goodnightMu.Unlock()
		slog.Debug("Goodnight request rate limited",
			"event", "goodnight_rate_limited",
			"service", "dadbot")
		return false
	}
	lastGoodnightTime = time.Now()
	goodnightMu.Unlock()

	response := goodnightMessages[rand.Intn(len(goodnightMessages))]
	s.ChannelMessageSend(m.ChannelID, response)

	slog.Info("Bot said goodnight by trigger phrase",
		"event", "goodnight_triggered",
		"service", "dadbot",
		"trigger", msg)
	return true
}
func handleCatGirlRequest(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	matches := catgirlRegex.FindStringSubmatch(m.Content)
	if len(matches) > 0 {
		matchWord := matches[0]
		response := "Every dollar not spent on genetically engineering cat girls is a dollar wasted."
		s.ChannelMessageSend(m.ChannelID, response)

		slog.Info("Bot paused by trigger word",
			"event", "meow_triggered",
			"service", "dadbot",
			"trigger", matchWord)
		return true
	}
	return false
}
func handleThermostatRequest(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	matches := thermostatRegex.FindStringSubmatch(m.Content)
	if len(matches) > 0 {
		matchWord := matches[0]
		response := thermostatMessages[rand.Intn(len(thermostatMessages))]
		s.ChannelMessageSend(m.ChannelID, response)

		slog.Info("Bot paused by trigger word",
			"event", "thermostat_triggered",
			"service", "dadbot",
			"trigger", matchWord)
		return true
	}
	return false
}
func handleJokeRequest(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if strings.ToLower(m.Content) != "tell me a joke" {
		return false
	}

	jokeMu.Lock()
	if time.Since(lastJokeTime) < jokeCooldown {
		jokeMu.Unlock()
		slog.Debug("Joke request rate limited",
			"event", "joke_rate_limited",
			"service", "dadbot")
		return false
	}
	lastJokeTime = time.Now()
	jokeMu.Unlock()

	joke, err := getDadJoke()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Gosh dang joke AI always breakin. Tell Clutch to fix it.")
		slog.Error("Failed to fetch dad joke",
			"event", "joke_request_failed",
			"service", "dadbot",
			"error", err.Error())
		return true
	}
	s.ChannelMessageSend(m.ChannelID, joke)

	slog.Info("Dad joke sent",
		"event", "joke_request_fulfilled",
		"service", "dadbot")
	return true
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

	resp, err := httpClient.Do(req)
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
