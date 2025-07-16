package main

import (
	"context"
	"log"
	"sync"
	"time"
)

type Subscription struct {
	ChannelID string
	GuildID   string
	Time      time.Time
	StopChan  chan struct{}
}

type SubscriptionManager struct {
	subscriptions map[string][]Subscription
	mu            sync.RWMutex
	wordService   *WordService
	sender        MessageSender
}

type MessageSender interface {
	SendMessage(channelID, message string) error
}

func NewSubscriptionManager(wordService *WordService, sender MessageSender) *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make(map[string][]Subscription),
		wordService:   wordService,
		sender:        sender,
	}
}

func (sm *SubscriptionManager) Subscribe(channelID, guildID string, scheduledTime time.Time) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stopChan := make(chan struct{})
	nextRun := sm.calculateNextRun(scheduledTime)

	subscription := Subscription{
		ChannelID: channelID,
		GuildID:   guildID,
		Time:      scheduledTime,
		StopChan:  stopChan,
	}

	sm.subscriptions[channelID] = append(sm.subscriptions[channelID], subscription)
	go sm.scheduleDaily(channelID, nextRun, stopChan)

	return nil
}

func (sm *SubscriptionManager) Unsubscribe(channelID string) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subs, exists := sm.subscriptions[channelID]
	if !exists {
		return 0
	}

	count := len(subs)
	for _, sub := range subs {
		if sub.StopChan != nil {
			close(sub.StopChan)
		}
	}

	delete(sm.subscriptions, channelID)
	return count
}

func (sm *SubscriptionManager) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, subs := range sm.subscriptions {
		for _, sub := range subs {
			if sub.StopChan != nil {
				close(sub.StopChan)
			}
		}
	}
}

func (sm *SubscriptionManager) calculateNextRun(scheduledTime time.Time) time.Time {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(),
		scheduledTime.Hour(), scheduledTime.Minute(), 0, 0, now.Location())

	if today.After(now) {
		return today
	}
	return today.Add(24 * time.Hour)
}

func (sm *SubscriptionManager) scheduleDaily(channelID string, nextRun time.Time, stopChan chan struct{}) {
	timer := time.NewTimer(time.Until(nextRun))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			sm.sendDailyWord(ctx, channelID)
			cancel()
			timer.Reset(24 * time.Hour)
		case <-stopChan:
			return
		}
	}
}

func (sm *SubscriptionManager) sendDailyWord(ctx context.Context, channelID string) {
	if sm.wordService == nil || sm.sender == nil {
		log.Printf("Word service or sender not configured")
		return
	}

	translation, err := sm.wordService.GetTodaysWordWithTranslation(ctx)
	if err != nil {
		log.Printf("Failed to get today's word for channel %s: %v", channelID, err)
		return
	}

	if err := sm.sender.SendMessage(channelID, translation); err != nil {
		log.Printf("Failed to send daily word to channel %s: %v", channelID, err)
	}
}
