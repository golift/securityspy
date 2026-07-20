---
name: securityspy-control
description: >-
  Control a live SecuritySpy server over its HTTP API with the Python CLI
  (schedules, arm/disarm, snapshots, list/download recordings). Use when the
  user mentions SecuritySpy cameras, secspy, setting schedules, arming motion,
  downloading a snapshot or video, or gives a SecuritySpy URL/username/password.
---

# SecuritySpy live control

Talk to SecuritySpy’s **web API** directly (not the Go SDK). Run the CLI under this skill’s `scripts/` directory.

## Auth

Pass credentials every time (or via env). Never commit passwords.

| Flag / env | Meaning |
| ------------ | --------- |
| `--url` / `SECURITYSPY_URL` | Base URL, default `http://127.0.0.1:8000` |
| `--user` / `SECURITYSPY_USER` | Username (required) |
| `--password` / `SECURITYSPY_PASS` | Password (required) |
| `--insecure` | Skip TLS verify (opt-in; use for self-signed HTTPS) |

Auth is the `auth=` query parameter: base64url(`user:pass`). TLS is verified by default; add `--insecure` for typical LAN self-signed certs.

## Preferred commands

Resolve `SCRIPT` to this skill’s script path (repo: `.cursor/skills/securityspy-control/scripts/ss_ctl.py`).

```bash
# Inspect (add --insecure for https:// with a self-signed cert)
python3 "$SCRIPT" --url https://HOST:8001 --insecure --user USER --password PASS info
python3 "$SCRIPT" --user USER --password PASS cameras
python3 "$SCRIPT" --user USER --password PASS schedules
python3 "$SCRIPT" --user USER --password PASS modes Door

# Set schedule (mode: C=continuous M=motion A=actions; schedule by id or name)
python3 "$SCRIPT" --user USER --password PASS set-schedule Door --mode M --schedule "Armed 24/7"
python3 "$SCRIPT" --user USER --password PASS set-override Door --mode M --override "Armed For 1 Hour"

# Arm / disarm (motion & actions; continuous soft-toggle often 404 — use set-schedule)
python3 "$SCRIPT" --user USER --password PASS arm-motion Door
python3 "$SCRIPT" --user USER --password PASS disarm-motion Door
python3 "$SCRIPT" --user USER --password PASS arm-actions Door
python3 "$SCRIPT" --user USER --password PASS trigger-motion Door

# Live snapshot / short clip (clip needs ffmpeg)
python3 "$SCRIPT" --user USER --password PASS snapshot Door -o /tmp/door.jpg
python3 "$SCRIPT" --user USER --password PASS snapshot 3 --width 1280 --quality 80 -o /tmp/cam3.jpg
python3 "$SCRIPT" --user USER --password PASS clip Door -o /tmp/door.mp4 --seconds 10

# Recorded files
python3 "$SCRIPT" --user USER --password PASS list-files Door --days 1 --type motion
python3 "$SCRIPT" --user USER --password PASS get-file --href '++getfilehb/3/2026-07-19/...' -o /tmp/clip.m4v
```

Camera args accept **number** or **name** (case-insensitive).

Requires: `pip3 install --user requests` (stdlib `xml.etree` for XML). `ffmpeg` only for `clip`.

## Agent workflow

1. Get URL/user/password from the user (or env already set). Do not invent credentials.
2. Run `cameras` / `schedules` if the target name is ambiguous.
3. Prefer `set-schedule` for continuous arming; do not rely on continuous toggle.
4. Confirm before `set-schedule` with camera `-1` / all-cameras or schedule presets if added later.
5. Report command output (OK / saved path / error) to the user.

## Gotchas

- `++ssControlContinuous` is often missing → use an armed/disarmed **schedule** on mode `C`.
- `list-files` returns download `href` values; pass those to `get-file`.
- Snapshot/clip overwrite: script refuses to clobber an existing `-o` path unless `--force`.
