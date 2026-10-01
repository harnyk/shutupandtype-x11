# Smart mode (LLM post-processing + speech mode markers)

**Date:** 2026-10-01  
**Status:** draft (awaiting spec review)  
**Branch:** `feat/smart-mode`

## Goal

After speech-to-text, optionally run a **smart** pass that (1) edits prose like a light dictation assistant (punctuation, grammar, minimal rephrasing without changing meaning), and (2) restores **code and symbols** when the user dictates them in mixed speech. Users switch behavior **inside the utterance** using **spoken markers**; markers are configurable and may be slightly garbled by STT—the LLM must still interpret them and **must not** include marker phrases in the final pasted text.

## Decisions (locked)

| Topic | Choice |
|-------|--------|
| Default smart behavior | **Prose (B):** polish dictated text; do not rewrite intent |
| Mode markers | **D:** YAML-configured trigger phrases + LLM fuzzy matching for STT errors/synonyms |
| Smart activation | **C:** Off by default; tray menu toggle; persist `smart_mode` to config across restarts |
| Pipeline shape | **Approach 1:** Single LLM chat completion after STT (no two-step segment pipeline in v1) |
| LLM provider (v1) | OpenAI Chat Completions, same `openai_api_key` as STT |
| STT backends | Unchanged (`openai` \| `whisper`); smart step is text-only and independent of STT backend |
| LLM failure | Fall back to raw STT text; tray shows error tooltip (do not drop dictation) |
| Output format | Plain text only—no markdown fences, no “Here is the text:” preamble |

## Architecture

```text
Hotkey → Record → STT → [smart_mode off] → trim → clipboard → paste
                      → [smart_mode on]  → SmartTransform(transcript) → trim → clipboard → paste
```

### Components

| Unit | Responsibility |
|------|----------------|
| `Transcriber` (existing) | Audio → raw transcript string |
| `SmartTransformer` (new) | Raw transcript → final string per mode rules + markers |
| `main.run` | After successful STT, call transformer when `smart_mode` enabled |
| Tray menu | Checkbox “Smart mode”; writes `smart_mode` to config file on toggle |
| Config | `smart_mode`, `openai_model_smart`, `smart_markers`, validation when smart enabled |

Optional tray state: reuse `StateTranscribing` with tooltip **“Smart edit…”** after STT completes and before clipboard, or add `StateSmartEdit` (amber)—implementation may choose minimal diff (tooltip only).

### Mode semantics (for LLM)

| Mode | When active | LLM behavior |
|------|-------------|--------------|
| **prose** (default) | Start of utterance; after `prose` marker | Fix punctuation/capitalization; light grammar; preserve language (RU/EN); no factual additions |
| **code** | After `code` marker until another marker | Recover operators, brackets, quotes, `.`, `[]`, `??`, `{}`, etc.; **do not** refactor, optimize, or invent identifiers |
| **verbatim** | After `verbatim` marker | Minimal cleanup only (optional: drop filler “uh/эм”—**disabled in v1**: keep words as spoken aside from obvious STT glitches) |

Markers are **stripped** from output. Text before the first marker uses **prose**.

### Marker configuration

```yaml
smart_mode: false
openai_model_smart: gpt-4o-mini

smart_markers:
  code:
    - "режим код"
    - "mode code"
  verbatim:
    - "дословно"
    - "verbatim"
  prose:
    - "обычный текст"
    - "prose mode"
```

- Defaults above ship in code when YAML omits `smart_markers`.
- System prompt lists configured phrases and instructs: recognize minor STT corruption and natural synonyms; never echo marker phrases in output.

## LLM integration

### Request

- **Endpoint:** `POST https://api.openai.com/v1/chat/completions`
- **Model:** `openai_model_smart` (default `gpt-4o-mini`)
- **Messages:**
  - **System:** Mode definitions, marker list, strict output rules (plain text only, no markers, code mode rules, prose default).
  - **User:** Raw STT transcript (single block).

### Guardrails (system prompt essentials)

1. Output **only** the final text to paste—no explanation.
2. Do not wrap code in markdown code blocks.
3. In **code** segments, prefer faithful symbol recovery over “clean” natural language.
4. Do not add content the user did not dictate.
5. If uncertain in code mode, prefer literal STT tokens over guessing new code.
6. Markers may appear corrupted (e.g. “режим kot”); infer intent from configured phrases.

### Response handling

- Parse `choices[0].message.content`; trim whitespace.
- Empty response → treat as error, fallback to STT.
- HTTP/API errors → log, fallback to STT, `StateError` briefly.

## Config validation

| Condition | Rule |
|-----------|------|
| `smart_mode: false` | No LLM validation required at startup |
| User enables smart (tray) | If `openai_api_key` missing → disable toggle / show tray error; do not enable |
| `smart_mode: true` in file | Same as above at startup; if invalid, log warning and force smart off |

`openai_model_smart` default: `gpt-4o-mini`.

## Tray UX

- Menu item **Smart mode** with checkmark reflecting `viper.GetBool("smart_mode")`.
- On click: flip bool, `viper.Set`, **write config file** (`viper.WriteConfig()` or equivalent safe merge into `~/.config/shutupandtype/config.yaml`).
- If config file did not exist, create config dir + file on first toggle (document in README).

## Privacy & latency

- Smart mode sends **full transcript** to OpenAI on every completed recording while enabled.
- README must state this clearly next to smart mode docs.
- Expected added latency: one chat round-trip (typically sub-second to a few seconds on mini models).

## Error handling & observability

- Log LLM failures at `log.Printf` (same as transcribe).
- User-visible: tray tooltip + short error state; clipboard still gets **raw STT** on failure.
- Optional debug: log raw vs smart length only (never log API key).

## Testing

| Layer | Approach |
|-------|----------|
| Unit | `SmartTransformer` with injected HTTP client / mock OpenAI response: prose cleanup, code recovery fixture, marker stripping |
| Config | Marker defaults merge; persist `smart_mode` round-trip (test with temp config dir) |
| Manual | Mixed RU utterance with “режим код” + dictated JS snippet; toggle persistence across restart |

No live API calls in CI.

## Out of scope (v1)

- Non-OpenAI LLM providers (Anthropic, local Ollama)
- Two-pass LLM (segment JSON then transform)
- Context from focused app / open file / clipboard
- Separate hotkey for smart vs plain (tray-only toggle)
- Auto-download or local LLM

## Example

**STT raw (illustrative):**  
`Напиши пожалуйста режим код console dot log open brace resource colon stacks open bracket zero close bracket dot vpc dot id question question unknown close brace close paren`

**Expected paste (smart on, code segment):**  
`Напиши пожалуйста console.log({resource: stacks[0].Vpc.Id ?? '<unknown>'})`  
(with prose part polished per rules—exact polish left to model within guardrails)

## References

- Existing STT: `transcribe.go`, `transcribe_openai.go`, `transcribe_whisper.go`
- Orchestration: `main.go` (`transcribe` → clipboard)
