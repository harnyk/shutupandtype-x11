# shutupandtype-x11

<p align="center">
  <img src="icon.png" alt="shutupandtype-x11" width="128" />
</p>

Press **Ctrl+Shift+F12** once to start recording from your mic. Press it again to stop. The audio is transcribed via OpenAI Whisper and the result is copied to your clipboard.

## Installation

### From release (deb)

TODO

### From source

Build-time dependencies:

```sh
sudo apt install ffmpeg xclip \
  libayatana-appindicator3-dev libgtk-3-dev
```

Then install with Go:

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

- Press **Ctrl+Shift+F12** → recording starts (tray turns red)
- Press **Ctrl+Shift+F12** again → recording stops, transcription begins (tray turns amber)
- Transcribed text is copied to clipboard and a notification shows a preview (tray turns green)
- Tray returns to gray after 3 seconds
