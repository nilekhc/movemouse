# movemouse
Programmatically move mouse pointer on the screen

## Usage

```
./movemouse -stop-in=1h
./movemouse -stop-in=8h -interval=30s -jitter=3
```

## Flags

- `-stop-in` — stop duration (default `8h`). eg `5s`, `1m`, `10h`.
- `-interval` — tick cadence between nudges (default `30s`).
- `-jitter` — nudge delta in pixels, uniform random in `[-jitter, +jitter]` (default `3`).

Behavior: each tick, nudge the cursor by a small random delta from its current position. If the cursor has moved since the last tick (i.e. you touched the mouse), yield and skip the nudge — it auto-resumes within one tick of you stopping. Ctrl-C (or SIGTERM) exits cleanly.
