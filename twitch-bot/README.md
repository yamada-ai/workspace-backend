# Twitch Bot

Twitch chat bot that integrates with the work-tracker backend API.

## Features

- Listen to Twitch chat messages
- Process `!in` command to start work sessions
- Auto-refresh OAuth tokens
- Reply to `!ping` with pong

## Setup

### Prerequisites

- Python 3.11+
- `openapi-python-client` for generating API client

### Install Dependencies

```bash
make install
# or
pip install -r requirements.txt
```

### Generate API Client

```bash
make gen-client
# or from root
cd .. && make gen-client-python
```

### Environment Variables

Create a `.env` file:

```bash
# Twitch API credentials
CLIENT_ID=your_twitch_client_id
CLIENT_SECRET=your_twitch_client_secret
ACCESS_TOKEN=your_initial_access_token
REFRESH_TOKEN=your_refresh_token
CHANNELS=your_channel_name

# Work tracker API
WORK_TRACKER_URL=http://localhost:8000

# Logging
LOG_LEVEL=INFO
```

## Running

```bash
make run
# or
python -m app.main
```

## Development

### Available Make Commands

```bash
make help           # Show available commands
make install        # Install dependencies
make gen-client     # Generate OpenAPI client
make run            # Run the bot locally
make test           # Run tests
make lint           # Run linters
```

## Architecture

```
twitch-bot/
├── app/
│   ├── api/
│   │   ├── generated/          # Auto-generated OpenAPI client
│   │   └── work_tracker_client.py  # Wrapper around generated client
│   ├── commands/
│   │   └── in_command.py       # Command handlers
│   └── main.py                 # Entry point
├── Dockerfile
├── requirements.txt
└── Makefile
```
