---
name: go-security-spy-sdk
description: >-
  Develop against or call golift.io/securityspy/v2 (Go SDK for SecuritySpy).
  Use when writing Go code that uses this library, fixing SDK types/methods,
  or when the user asks how to do something via the Go package rather than the
  live HTTP control skill.
---

# SecuritySpy Go SDK (`golift.io/securityspy/v2`)

Use this skill when working **in Go** with this repository’s client library. For interactive control of a live SecuritySpy server (schedules, snapshots, downloads), prefer the sibling **`securityspy-control`** skill and its Python CLI.

Offline API HTML + redacted fixtures: [`.archive/`](../../../.archive/README.md). Method index: [reference.md](reference.md).

## Connect

```go
sspy, err := securityspy.New(&server.Config{
    Username:  "...",
    Password:  "...", // New() replaces with base64 auth blob
    URL:       "https://host:8001",
    VerifySSL: false,
})
```

- Auth is `auth=` query param (base64 `user:pass`), not HTTP Basic.
- `New` calls `Refresh()`. Lookup: `Cameras.ByNum` / `ByName`.
- Do not call `Refresh()` concurrently with other API use.

## Common Go patterns

**Schedule:** `cam.SetSchedule(CameraModeMotion, id)` — IDs from `Info.ServerSchedules`. Prefer schedules over `ToggleContinuous` (`ErrUnsupported` / 404).

**Snapshot:** `cam.SaveJPEG(&VidOps{Width: 1280, Quality: 80}, path)` — fails if path exists.

**Clip:** `cam.SaveVideo(ops, length, maxBytes, path)` — pure-Go RTSP remux (H.264 + AAC); no ffmpeg. Prefer `ACodec: "aac"` (default). `UseHTTP` is unsupported.

**Recorded files:** `sspy.Files.GetMCVideos` / `GetImages` / `GetAll` then `file.Save(path)`.

**Arm:** `ToggleMotion` / `ToggleActions`; live status via `Modes()`.

See [reference.md](reference.md) for the full surface.
