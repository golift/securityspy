# RTSP clip spike results

Live test (`go test -tags=live ./internal/rtspclip/`) against SecuritySpy 6.20:

- **PASS**: `SaveMP4` wrote a playable fMP4 (`format_name=mov,mp4,…`, `codec_name=h264`).
- Example run: ~37 frames, ~2.5 MiB in ~1.7 s (stopped on `MaxBytes`), `ffprobe` duration ≈ 1.83 s.
- Package is the production remux path for `Camera.SaveVideo` / `StreamVideo`.

## Critical finding: RTSP auth

SecuritySpy **rejects** query-parameter `auth=` on RTSP/RTSPS (`DESCRIBE` → 401).

Working form:

```text
rtsps://USER:PASS@host:8001/stream?cameraNum=3&vcodec=h264&acodec=aac
```

HTTP APIs still use `auth=` query. RTSP URLs use **userinfo** (decode the library’s base64 auth blob back to `user:pass`).

## Notes

- Prefer `acodec=aac` for MP4 remux (Telegram-friendly stream-copy). Non-AAC audio tracks are ignored.
- Prefer TCP (`ProtocolTCP`) + `InsecureSkipVerify` when `VerifySSL` is false (LAN RTSPS).
- Wait for an IDR before writing samples; include SPS/PPS in the AVC descriptor.
- **AAC ASC must come from the SDP `config=`** (do not use mp4ff `SetAACDescriptor` — it hardcodes stereo and buzzes on SecuritySpy’s mono 64 kHz AAC).
- **Only declare an AAC track if AAC samples were actually received.** SecuritySpy often advertises AAC (especially with `height=`/`width=` resize) but sends no audio RTP on cameras without mics; an empty audio track freezes Telegram (5s of one frame).
- Do not mismatch `mp4a` ChannelCount vs ASC (e.g. 2ch entry + mono ASC): Telegram drops audio entirely. True dual-ear stereo needs a real stereo re-encode, not container lies.
