# shutupandtype-x11

Press **Scroll Lock** once to start recording from your mic. Press it again to stop. The audio is transcribed via OpenAI Whisper and the result is copied to your clipboard.

A tray icon shows the current state:

| Color | State |
|-------|-------|
| ⚫ Gray | Idle — waiting |
| 🔴 Red | Recording |
| 🟡 Amber | Transcribing |
| 🟢 Green | Done — text copied |
| 🟠 Orange | Error |

## Prerequisites

```sh
sudo apt install ffmpeg libnotify-bin xclip \
  libayatana-appindicator3-dev libgtk-3-dev
```

## Installation

```sh
go install github.com/harnyk/shutupandtype-x11@latest
```

## Configuration

Create `~/.config/shutupandtype/config.yaml`:

```yaml
openai_api_key: "sk-..."
openai_model_stt: "whisper-1"
timeout: "90s"
```

| Key | Default | Description |
|-----|---------|-------------|
| `openai_api_key` | — | OpenAI API key (required) |
| `openai_model_stt` | `whisper-1` | Whisper model for transcription |
| `timeout` | `90s` | Auto-stop recording after this duration |

The `--timeout` flag overrides the config value at runtime:

```sh
shutupandtype-x11 --timeout 2m
```

## Usage

```sh
shutupandtype-x11
```

- Press **Scroll Lock** → recording starts (tray turns red)
- Press **Scroll Lock** again → recording stops, transcription begins (tray turns amber)
- Transcribed text is copied to clipboard and a notification shows a preview (tray turns green)
- Tray returns to gray after 3 seconds
