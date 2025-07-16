# Daily Chinese Teacher Bot

A Discord bot that sends daily Chinese words with Japanese translations and pronunciation.

## Features

- **Daily Chinese Words**: Automatically fetches word-of-the-day from ChineseClass101
- **Japanese Translation**: Uses OpenAI GPT-4.1 for natural Japanese translations
- **Clean Messages**: Sends formatted text messages with OpenAI's response
- **Scheduling**: Set custom times for daily word delivery
- **Commands**: Interactive slash commands for subscription management

## Setup

### Prerequisites

- Go 1.24.5 or later
- Discord Bot Token
- OpenAI API Key

### Installation

1. Clone the repository:
```bash
git clone https://github.com/sglre6355/daily-chinese-teacher.git
cd daily-chinese-teacher
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export DISCORD_TOKEN="your_discord_bot_token"
export OPENAI_API_KEY="your_openai_api_key"
export OPENAI_TRANSLATION_PROMPT="Translate this English word/phrase to natural Japanese dictionary-style translation:"
```

4. Build and run:
```bash
go build .
./daily-chinese-teacher
```

## Commands

- `/subscribe <time>` - Subscribe channel to daily Chinese words (e.g., `/subscribe 09:00`)
- `/unsubscribe` - Unsubscribe channel from daily words
- `/word` - Get today's Chinese word immediately

## Environment Variables

- `DISCORD_TOKEN` - Your Discord bot token (required)
- `OPENAI_API_KEY` - Your OpenAI API key (required)
- `OPENAI_TRANSLATION_PROMPT` - Custom prompt for translation (required)

## Prompt Format

The bot appends only the Chinese word and pinyin to your prompt. For example:

```
[Your prompt here]

床 chuáng
```

See `prompt-example` file for a template.

## How it Works

1. Fetches daily Chinese word from ChineseClass101 API
2. Sends Chinese word + pinyin to OpenAI GPT-4.1 with your custom prompt
3. Sends OpenAI's raw response directly to Discord

## API Sources

- **ChineseClass101**: https://www.chineseclass101.com/api/word-day/YYYY-MM-DD
- **OpenAI**: GPT-4.1 for Japanese translations
