package main

import (
	"fmt"
	"os"
)

type Config struct {
	DiscordToken      string
	OpenAIAPIKey      string
	TranslationPrompt string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		DiscordToken:      os.Getenv("DISCORD_TOKEN"),
		OpenAIAPIKey:      os.Getenv("OPENAI_API_KEY"),
		TranslationPrompt: os.Getenv("OPENAI_TRANSLATION_PROMPT"),
	}

	if config.DiscordToken == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN environment variable is required")
	}

	if config.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	if config.TranslationPrompt == "" {
		config.TranslationPrompt = `
			Translate the following Chinese word/saying to Japanese. I'll also give you its pinyin.
			Give me a sentence that sounds like a dictionary explanation. Give me the result in the following format:

			読み方：${ピンインのカタカナ読み}（${ピンイン}）
			意味：${日本語での意味}（${単語を英語で、直訳がない/複数該当する英訳がある場合は2語を上限として与えてもよい}）

			For ピンインのカタカナ読み, be very careful not to give wrong answer.
			For 日本語での意味, write in a very, very specific style of dictionary explanations.

			The following is an exemplary response:

			読み方：エックスグアン（X guāng）
			意味：エックス線。またはレントゲン線。医療や工業で物体内部を透過して観察するために使われる放射線。（X-ray）
		`
	}

	return config, nil
}
