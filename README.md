# shutupandtype-x11

<p align="center">
  <img src="icon.png" alt="shutupandtype-x11" width="128" />
</p>

Press a hotkey to start recording from your mic. Press it again to stop. Audio is transcribed (OpenAI Whisper API or local `whisper-cli`) and copied to the clipboard.

| OS | Hotkey | Paste |
|----|--------|-------|
| Linux (X11) | **Ctrl+Shift+F12** | auto via Shift+Insert |
| macOS | **Ctrl+Shift+F12** | auto via Cmd+V |

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

### macOS app (no Terminal)

Build a menu-bar `.app` into `~/Applications`:

```sh
task app:open
```

Or: `task app` then open **ShutUpAndType** from Spotlight / Launchpad.

Grant **Accessibility** and **Microphone** to **ShutUpAndType** (not Terminal) in System Settings → Privacy & Security. Login Items can also start it at login.


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
whisper_language: ru     # empty = auto-detect; set ru/en only if you always speak that language
```

| Key | Default | Description |
|-----|---------|-------------|
| `backend` | `openai` | `openai` or `whisper` (local whisper.cpp CLI) |
| `openai_api_key` | — | Required when `backend: openai` |
| `openai_model_stt` | `whisper-1` | OpenAI STT model |
| `timeout` | `90s` | Auto-stop recording after this duration |
| `whisper_bin` | `whisper-cli` | Path or name of whisper.cpp CLI |
| `whisper_model` | — | Path to ggml model (required for whisper) |
| `whisper_language` | empty | Passed as `-l` to whisper-cli; empty means `auto`. Do not omit — whisper-cli’s own default is `en`, which turns non-English into “(speaking in foreign language)” |
| `smart_mode` | `false` | Tray-toggleable; LLM post-process after STT |
| `smart_llm_base_url` | `https://api.openai.com/v1` | OpenAI-compatible API root (includes `/v1`) |
| `smart_llm_api_key` | empty | Bearer token for smart mode; falls back to `openai_api_key` |
| `smart_llm_model` | `gpt-4o-mini` | Chat completions model id for your endpoint |
| `smart_markers` | built-in defaults | Spoken phrases for code / verbatim / prose modes (see spec) |

### Smart mode

When **Smart mode** is enabled (tray menu checkbox; persisted in config), each transcript is sent to an **OpenAI-compatible** `POST …/chat/completions` endpoint for punctuation, light editing, and code/symbol recovery. You can use OpenAI or any compatible server (local LM Studio, Groq, etc.) via `smart_llm_base_url`.

**Privacy:** while smart mode is on, the **full transcript** is sent to whatever host you configure—not only OpenAI.

On macOS and Linux (with `notify-send`), **errors** and successfully **typed text** also appear as short system notifications; other steps stay in the log and menu bar tooltip (tooltip shows on hover only).

Example (local compatible server):

```yaml
smart_mode: true
smart_llm_base_url: http://127.0.0.1:1234/v1
smart_llm_model: your-model-id
# smart_llm_api_key: "..."   # optional if openai_api_key is set
```

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
- Text is copied to the clipboard and auto-pasted into the focused app; tray turns green
- Tray returns to gray after a few seconds
