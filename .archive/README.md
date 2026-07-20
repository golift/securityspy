# SecuritySpy Web Server Specification Archives

Offline copies of Ben Software's SecuritySpy web API documentation and live fixtures used to validate this SDK.

| File | Source | Notes |
|------|--------|-------|
| [web-server-spec-v5.html](web-server-spec-v5.html) | [Wayback Oct 2023](https://web.archive.org/web/20231002173428/https://bensoftware.com/securityspy/web-server-spec.html) | Explicitly **SecuritySpy v5** |
| [web-server-spec-v6.html](web-server-spec-v6.html) | [Ben Software current](https://www.bensoftware.com/securityspy/web-server-spec.html) | **SecuritySpy v6** |
| [systemInfo-v6.20.xml](systemInfo-v6.20.xml) | Live `++systemInfo` | Captured from SecuritySpy **6.20** |
| [settings-*-v6.20.xml](settings-general-v6.20.xml) | Live `++settings-*` | Captured from SecuritySpy **6.20** |

Live SDK validation (2026-07) against SecuritySpy **6.20** using query-parameter `auth=` (base64 `username:password`).
XML fixtures below are redacted (hostnames, IPs, paths, accounts, emails, passwords).

Note: v6 `++systemInfo` uses renamed XML tags (`camera-list`, `cc-mode`, `video-width`, `ptz-features`, …). This library dual-reads v5 and v6 schemas.
