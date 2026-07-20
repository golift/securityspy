# SecuritySpy Go library reference

Import: `golift.io/securityspy/v2` and `golift.io/securityspy/v2/server`.

For live HTTP control without Go, use the sibling skill `securityspy-control`.

## Server

| Method | Notes |
|--------|-------|
| `New(cfg) (*Server, error)` | Connects + `Refresh` |
| `NewMust(cfg) *Server` | No connect; call `Refresh` before use |
| `Refresh()` / `RefreshContext(ctx)` | Reloads `Info`, `Cameras`, `Groups` |
| `GetScripts()` / `GetSounds()` | `[]string` |
| `SetSchedulePreset(id int)` | `++ssSetPreset`; global |
| `MultiplexURL(*MultiplexOps)` | `++multiplex` URL string |
| `Get*Settings` / `Set*Settings` | See Settings |

### `server.Config`

`Username`, `Password`, `URL`, `Timeout` (`server.Duration`), `VerifySSL`, optional `Client`, `JPEGRetries`.

After `New`, `Password` is the base64 auth token (`Config.Auth()`).

### After Refresh

- `Info *ServerInfo` — version, IPs, `ServerSchedules`, `ScheduleOverrides`, `SchedulePresets` (`map[int]string`)
- `Cameras *Cameras` — `All()`, `ByNum`, `ByName`
- `Groups []Group` — v6+
- `Files *Files`, `Events *Events`

## Camera control

| Method | Endpoint / notes |
|--------|------------------|
| `SetSchedule(mode, scheduleID)` | `++ssSetSchedule` |
| `SetScheduleOverride(mode, overrideID)` | `++ssSetOverride` |
| `ToggleContinuous(arm)` | Often `ErrUnsupported` |
| `ToggleMotion(arm)` | `++ssControlMotionCapture` |
| `ToggleActions(arm)` | `++ssControlActions` |
| `TriggerMotion()` | `++triggermd` |
| `Modes()` | `++cameramodes` → Continuous/Motion/Actions ARMED\|DISARMED |

`CameraArm` / `CameraDisarm` for toggle args.

Cached after Refresh: `ModeC`, `ModeM`, `ModeA`, schedule ID fields on `Camera`.

## Media

| Method | Notes |
|--------|-------|
| `GetJPEG(*VidOps)` | `image.Image`; retries |
| `SaveJPEG(*VidOps, path)` | No overwrite |
| `SaveVideo` | Pure-Go RTSP remux to file; length + maxsize; `UseHTTP` unsupported |
| `StreamVideo` | Progressive fMP4 pipe (init after IDR, periodic fragments); Close cancels |
| `StreamMJPG` / `StreamH264` / `StreamG711` | `io.ReadCloser`; Close when done |
| `PostG711(r)` | Talk-back audio |
| `HLSURL` / `HLSMediaPlaylistURL(q)` / `LiveURL` | URL strings |

`VidOps`: `Width`, `Height`, `FPS`, `Quality` (≤100), `UseHTTP` (not for Save/StreamVideo), `VCodec`, `ACodec` (prefer `aac` for remux).

`Server.Encoder` / `DefaultEncoder` are deprecated no-ops (kept for API compatibility).

## Files

| Method | Notes |
|--------|-------|
| `GetImages(nums, from, to)` | Captured stills |
| `GetMCVideos` / `GetCCVideos` / `GetAll` | Motion / continuous / all |
| `GetFile(name)` | Brittle; avoid |
| `(*File).Save(path)` | No overwrite |
| `(*File).Get(highBandwidth)` | `++getfilehb` / `++getfilelb` |

Date format for listing: `2006-01-02`.

## Settings

| Get | Set | Path |
|-----|-----|------|
| `GetGeneralSettings` | `SetGeneralSettings` | `++settings-general` |
| `GetDisplaySettings` | `SetDisplaySettings` | `++settings-display` |
| `GetStorageSettings` | `SetStorageSettings` | `++settings-storage` |
| `GetCompressionSettings` | `SetCompressionSettings` | `++settings-compression` |
| `GetEmailSettings` | `SetEmailSettings` | `++settings-email` |
| `GetWebSettings` | `SetWebSettings` | `++settings-web` |
| `GetCameraSettings(num)` | `SetCameraSettings` | `++settings-cameras` (Set needs `cameraNum` in form) |

Set methods take `url.Values` for partial updates. POST success: trailing `OK` or JSON `{"result":"OK"}`.

## Events (brief)

`Events.BindFunc` / `BindChan`, `Watch(retry, refreshOnConfigChange)`, `Stop`, `Custom`. Types include `EventTriggerMotion`, `EventFileWritten`, `EventClassify`, arm/disarm events, stream connect/disconnect.

## Source files

| Area | File |
|------|------|
| Construct / Refresh | `securityspy.go`, `securityspy_types.go` |
| Cameras / media / schedules on cam | `cameras.go`, `cameras_types.go`, `schedules.go` |
| Files | `files.go` |
| Settings | `settings.go`, `settings_types.go` |
| Transport | `server/server.go` |
| PTZ | `ptz.go` |
| Events | `events.go`, `events_types.go` |
