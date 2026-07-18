# macOS support + local whisper.cpp

**Date:** 2026-07-18  
**Status:** approved (brainstorm)  
**Branch:** `feat/macos-version`

## Goal

Run shutupandtype on macOS with the same push-to-talk UX as Linux, and support a local Whisper backend via the external `whisper-cli` binary, while keeping the existing OpenAI Whisper API path.

## Decisions (locked)

| Topic | Choice |
|-------|--------|
| Transcription backends | Both: `openai` and `whisper` (whisper.cpp CLI), selectable in config on Linux and macOS |
| whisper.cpp integration | External process (`whisper-cli`), not CGO/libwhisper |
| macOS paste | Clipboard only in MVP; auto-paste later |
| Model management | User-provided `whisper_model` path now; auto-download later |
| Repo / binary name | Keep `shutupandtype-x11`; no rename |
| Structure | Thin `*_linux.go` / `*_darwin.go` files + shared `Transcriber` interface |
| Hotkey | Linux: Ctrl+Shift+F12. macOS: **Cmd+Shift+F12** |
| Suggested local model | `ggml-small-q5_1.bin` (~181MB, multilingual, fine for Russian on Apple Silicon) |

## Architecture

Build with Go filename suffixes (`_linux.go`, `_darwin.go`). Shared orchestration stays in `main.go`.

### Platform surface

| Concern | Linux | macOS (MVP) |
|---------|-------|-------------|
| Hotkey | X11 `GrabKey` (current `grab.go`) | In-process global hotkey (Cmd+Shift+F12); requires Accessibility |
| Recorder | `ffmpeg -f alsa -i default` | `ffmpeg -f avfoundation -i :0` (default mic) |
| Clipboard | `xclip` primary + clipboard | `pbcopy` |
| Paste | `xdotool` Shift+Insert | no-op |
| Tray | `getlantern/systray` | same |
| Single-instance lock | `flock` on lock file under `os.TempDir()` | same |

### Transcription (OS-independent)

```text
Transcriber.Transcribe(audioPath) (text, error)
  ├─ openai  — multipart POST to OpenAI (current behavior)
  └─ whisper — exec whisper-cli, parse stdout
```

Factory selects implementation from `backend` config at startup (or first use).

### Recording format

Format follows the transcription backend, not the OS:

- `backend: openai` → MP3 (current)
- `backend: whisper` → WAV 16 kHz mono PCM (what `whisper-cli` expects)

## Config

File: `~/.config/shutupandtype/config.yaml`

```yaml
backend: openai          # openai | whisper
openai_api_key: "sk-..."
openai_model_stt: whisper-1
timeout: 90s

# used when backend: whisper
whisper_bin: whisper-cli
whisper_model: ~/.config/shutupandtype/models/ggml-small-q5_1.bin
whisper_language: ru     # optional; empty = auto-detect
```

Validation:

- `backend: openai` → `openai_api_key` required
- `backend: whisper` → `whisper_bin` resolvable on `PATH` (or absolute path) and `whisper_model` file exists; fail fast with a clear error / tray tooltip

Defaults:

- `backend`: `openai` (backward compatible)
- `whisper_bin`: `whisper-cli`
- `whisper_language`: empty (whisper-cli auto-detect); README recommends `ru` for Russian dictation

## Data flow

```text
hotkey press
  → if idle: Recorder.Start() → tray Recording
  → if recording: Recorder.Stop() → path
       → tray Transcribing
       → Transcriber.Transcribe(path)
       → clipboard (+ Linux paste)
       → tray Done / Error
```

Timeout auto-stop behavior unchanged.

### whisper-cli invocation

```sh
whisper-cli -m <model> -f <wav> -l <lang> -nt -np
```

- `-nt` — no timestamps in stdout  
- `-np` — no progress bar  
- Omit `-l` when `whisper_language` is empty  
- Non-zero exit → error; include stderr in log  
- Success text = `strings.TrimSpace(stdout)`

## File layout (expected)

```text
main.go                 # orchestration (shared)
config.go               # viper + new keys
tray.go                 # shared
transcribe.go           # Transcriber interface + factory
transcribe_openai.go    # current HTTP client (moved)
transcribe_whisper.go   # whisper-cli exec
recorder_linux.go
recorder_darwin.go
hotkey_linux.go         # current grab.go content
hotkey_darwin.go
clipboard_linux.go
clipboard_darwin.go
paste_linux.go          # typeShiftInsert
paste_darwin.go         # no-op
```

Exact names may vary slightly; Linux behavior must remain identical for the openai path.

## Error handling

| Failure | Behavior |
|---------|----------|
| Missing ffmpeg / whisper-cli / model | Clear error at start or first record; tray Error |
| Recorder start/stop failure | tray Error + tooltip (current pattern) |
| Transcribe failure | tray Error + tooltip; cooldown reset to Idle if still idle |
| Accessibility denied (macOS hotkey) | Fatal or hard fail at register time with actionable message |
| Clipboard failure | tray Error (current pattern) |

Temp audio files: leave as today (no aggressive cleanup required in MVP).

## Testing

- Unit: whisper stdout trimming / empty output handling  
- Manual: Linux `openai` + `whisper`; macOS `whisper` + clipboard-only paste  
- Confirm Linux Ctrl+Shift+F12 and macOS Cmd+Shift+F12  

## Out of scope (MVP)

- Auto-download of models  
- Auto-paste on macOS  
- Mic device picker / config  
- Renaming module or binary away from `shutupandtype-x11`  
- CGO linking to libwhisper  
- Wayland support  

## Runtime dependencies

**Linux (unchanged + optional):** ffmpeg, xclip, xdotool, appindicator; `whisper-cli` if `backend: whisper`.

**macOS:** ffmpeg; `whisper-cli` if `backend: whisper` (e.g. Homebrew `whisper-cpp`); Accessibility permission for global hotkey; CGO for systray.

## Reference model path (already downloaded locally)

`~/.config/shutupandtype/models/ggml-small-q5_1.bin`
