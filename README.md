# shutupandtype-x11

<p align="center">
  <img src="icon.png" alt="shutupandtype-x11" width="128" />
</p>

Press a hotkey to start recording from your mic. Press it again to stop. Audio is transcribed (OpenAI Whisper API or local `whisper-cli`) and copied to the clipboard.

| OS | Hotkey | Paste |
|----|--------|-------|
| Linux (X11) | **Ctrl+Shift+F12** | auto via Shift+Insert |
| macOS | **Cmd+Shift+F12** | clipboard only (paste manually) |

## Installation

### From source

**Linux build dependencies:**

```sh
sudo apt install ffmpeg xclip xdotool \
  libayatana-appindicator3-dev libgtk-3-dev
```

**macOS:**

```sh
brew install ffmpeg
# optional, for local STT:
brew install whisper-cpp
```

Then:

```sh
CGO_ENABLED=1 go install github.com/harnyk/shutupandtype-x11@latest
```

On macOS, grant **Accessibility** (and Microphone) to the terminal or app that runs the binary, or global hotkey registration will fail.

## Configuration

Create `~/.config/shutupandtype/config.yaml`:

```yaml
backend: openai          # openai | whisper
openai_api_key: "sk-..."
openai_model_stt: "whisper-1"
timeout: "90s"

# when backend: whisper
whisper_bin: whisper-cli
whisper_model: ~/.config/shutupandtype/models/ggml-small-q5_1.bin
whisper_language: ru     # empty = auto-detect; ru recommended for Russian
```

| Key | Default | Description |
|-----|---------|-------------|
| `backend` | `openai` | `openai` or `whisper` (local whisper.cpp CLI) |
| `openai_api_key` | — | Required when `backend: openai` |
| `openai_model_stt` | `whisper-1` | OpenAI STT model |
| `timeout` | `90s` | Auto-stop recording after this duration |
| `whisper_bin` | `whisper-cli` | Path or name of whisper.cpp CLI |
| `whisper_model` | — | Path to ggml model (required for whisper) |
| `whisper_language` | empty | Language code for whisper-cli; empty = auto |

The `--timeout` flag overrides the config value at runtime:

```sh
shutupandtype-x11 --timeout 2m
```

### Local whisper model

Download a multilingual ggml model (example: small q5_1) into `~/.config/shutupandtype/models/` and set `whisper_model`. Russian works with multilingual models (avoid `*.en` variants).

## Usage

```sh
shutupandtype-x11
```

- Press the hotkey → recording starts (tray turns red)
- Press again → recording stops, transcription begins (tray turns amber)
- Text is copied to the clipboard (and auto-pasted on Linux); tray turns green
- Tray returns to gray after a few seconds
