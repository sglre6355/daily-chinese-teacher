package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

type Bot struct {
	session             *discordgo.Session
	wordService         *WordService
	subscriptionManager *SubscriptionManager
}

func NewChineseBot(config *Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("creating Discord session: %w", err)
	}

	openaiClient := openai.NewClient(config.OpenAIAPIKey)
	wordService := NewWordService(openaiClient, config.TranslationPrompt)

	bot := &Bot{
		session:     session,
		wordService: wordService,
	}

	bot.subscriptionManager = NewSubscriptionManager(wordService, bot)

	bot.setupHandlers()
	return bot, nil
}

func (b *Bot) SendMessage(channelID, message string) error {
	_, err := b.session.ChannelMessageSend(channelID, message)
	return err
}

func (b *Bot) setupHandlers() {
	b.session.AddHandler(b.onReady)
	b.session.AddHandler(b.onInteractionCreate)
	b.session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("opening Discord session: %w", err)
	}
	log.Println("Bot started successfully")
	return nil
}

func (b *Bot) Stop() {
	if b.subscriptionManager != nil {
		b.subscriptionManager.Stop()
	}
	if b.session != nil {
		if err := b.session.Close(); err != nil {
			log.Printf("Error closing Discord session: %v", err)
		}
	}
	log.Println("Bot stopped successfully")
}

func (b *Bot) RegisterCommands() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "subscribe",
			Description: "Subscribe this channel to receive daily Chinese words",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "time",
					Description: "Time to send daily word (format: HH:MM, e.g., 09:00)",
					Required:    true,
				},
			},
		},
		{
			Name:        "unsubscribe",
			Description: "Unsubscribe this channel from daily Chinese words",
		},
		{
			Name:        "word",
			Description: "Get today's Chinese word immediately",
		},
	}

	for _, cmd := range commands {
		if _, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", cmd); err != nil {
			return fmt.Errorf("creating command %s: %w", cmd.Name, err)
		}
	}

	return nil
}

func (b *Bot) onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	handlers := map[string]func(*discordgo.Session, *discordgo.InteractionCreate){
		"subscribe":   b.handleSubscribe,
		"unsubscribe": b.handleUnsubscribe,
		"word":        b.handleWord,
	}

	if handler, exists := handlers[i.ApplicationCommandData().Name]; exists {
		handler(s, i)
	}
}

func (b *Bot) handleSubscribe(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	timeStr := options[0].StringValue()

	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		b.respondWithError(s, i, "Invalid time format. Please use HH:MM format (e.g., 09:00)")
		return
	}

	if err := b.subscriptionManager.Subscribe(i.ChannelID, i.GuildID, parsedTime); err != nil {
		b.respondWithError(s, i, "Failed to subscribe channel to daily Chinese words")
		return
	}

	b.respondWithSuccess(s, i, fmt.Sprintf("Successfully subscribed this channel to receive daily Chinese words at %s", timeStr))
}

func (b *Bot) handleUnsubscribe(s *discordgo.Session, i *discordgo.InteractionCreate) {
	count := b.subscriptionManager.Unsubscribe(i.ChannelID)
	b.respondWithSuccess(s, i, fmt.Sprintf("Removed %d daily Chinese word subscription(s) from this channel", count))
}

func (b *Bot) handleWord(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		log.Printf("Error responding to interaction: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	translation, err := b.wordService.GetTodaysWordWithTranslation(ctx)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get today's word: %v", err)
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &errorMsg,
		}); err != nil {
			log.Printf("Error editing interaction response: %v", err)
		}
		return
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &translation,
	}); err != nil {
		log.Printf("Error editing interaction response: %v", err)
	}
}

func (b *Bot) respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		log.Printf("Error responding with error: %v", err)
	}
}

func (b *Bot) respondWithSuccess(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		log.Printf("Error responding with success: %v", err)
	}
}
