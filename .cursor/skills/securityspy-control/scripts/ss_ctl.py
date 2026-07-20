#!/usr/bin/env python3
"""Control a SecuritySpy server via its HTTP web API (auth= query param)."""

from __future__ import annotations

import argparse
import base64
import os
import subprocess
import sys
import xml.etree.ElementTree as ET
from pathlib import Path
from typing import Any
from urllib.parse import urlencode, urljoin

import requests
import urllib3

DEFAULT_URL = "http://127.0.0.1:8000"


class SSClient:
    def __init__(
        self,
        base: str,
        user: str,
        password: str,
        *,
        insecure: bool = True,
        timeout: float = 30.0,
    ):
        self.base = base.rstrip("/") + "/"
        token = base64.urlsafe_b64encode(f"{user}:{password}".encode()).decode()
        self.auth = token
        self.timeout = timeout
        self.session = requests.Session()
        self.session.verify = not insecure
        if insecure:
            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

    def _params(self, extra: dict[str, Any] | None = None) -> dict[str, Any]:
        p: dict[str, Any] = {"auth": self.auth}
        if extra:
            p.update(extra)
        return p

    def get(self, api: str, params: dict[str, Any] | None = None, *, stream: bool = False) -> requests.Response:
        url = urljoin(self.base, api if api.startswith("++") else f"++{api}")
        r = self.session.get(url, params=self._params(params), timeout=self.timeout, stream=stream)
        r.raise_for_status()
        return r

    def get_text(self, api: str, params: dict[str, Any] | None = None) -> str:
        return self.get(api, params).text

    def get_ok(self, api: str, params: dict[str, Any] | None = None) -> None:
        body = self.get_text(api, params).strip()
        compact = "".join(body.split())
        if body.endswith("OK") or '"result":"OK"' in compact:
            return
        raise RuntimeError(f"{api} unexpected response: {body[:200]!r}")

    def system_info(self) -> ET.Element:
        text = self.get_text("++systemInfo", {"format": "xml"})
        return ET.fromstring(text)

    def cameras(self) -> list[dict[str, Any]]:
        root = self.system_info()
        cams: list[dict[str, Any]] = []
        # v6: camera-list/camera ; v5: cameralist/camera
        nodes = root.findall(".//camera-list/camera") or root.findall(".//cameralist/camera")
        for node in nodes:
            def t(tag: str, default: str = "") -> str:
                el = node.find(tag)
                return (el.text or default) if el is not None else default

            num = t("number") or t("cameraNum")
            name = t("name")
            # v6 modes / geometry
            mode_c = t("cc-mode") or t("mode-c")
            mode_m = t("mc-mode") or t("mode-m")
            mode_a = t("a-mode") or t("mode-a")
            w = t("video-width") or t("width")
            h = t("video-height") or t("height")
            connected = t("connected")
            cams.append(
                {
                    "number": int(num) if num.isdigit() else num,
                    "name": name,
                    "connected": connected,
                    "width": w,
                    "height": h,
                    "mode_c": mode_c,
                    "mode_m": mode_m,
                    "mode_a": mode_a,
                }
            )
        return cams

    def schedules(self) -> dict[int, str]:
        root = self.system_info()
        out: dict[int, str] = {}
        for node in root.findall(".//schedule-list/*") + root.findall(".//schedulelist/*"):
            n = node.find("name")
            i = node.find("id")
            if i is not None and i.text and i.text.isdigit():
                out[int(i.text)] = (n.text or "") if n is not None else ""
        return out

    def overrides(self) -> dict[int, str]:
        root = self.system_info()
        out: dict[int, str] = {}
        for node in root.findall(".//schedule-override-list/*") + root.findall(".//scheduleoverridelist/*"):
            n = node.find("name")
            i = node.find("id")
            if i is not None and i.text and i.text.isdigit():
                out[int(i.text)] = (n.text or "") if n is not None else ""
        return out

    def resolve_camera(self, ref: str) -> dict[str, Any]:
        cams = self.cameras()
        if ref.isdigit():
            num = int(ref)
            for c in cams:
                if c["number"] == num:
                    return c
            raise SystemExit(f"camera number not found: {ref}")
        low = ref.lower()
        exact = [c for c in cams if c["name"].lower() == low]
        if exact:
            return exact[0]
        partial = [c for c in cams if low in c["name"].lower()]
        if len(partial) == 1:
            return partial[0]
        if not partial:
            raise SystemExit(f"camera not found: {ref}")
        names = ", ".join(f'{c["number"]}:{c["name"]}' for c in partial)
        raise SystemExit(f"ambiguous camera {ref!r}: {names}")

    def resolve_named_id(self, mapping: dict[int, str], ref: str) -> int:
        if ref.isdigit():
            return int(ref)
        low = ref.lower()
        hits = [i for i, n in mapping.items() if n.lower() == low]
        if len(hits) == 1:
            return hits[0]
        hits = [i for i, n in mapping.items() if low in n.lower()]
        if len(hits) == 1:
            return hits[0]
        if not hits:
            raise SystemExit(f"name not found: {ref}")
        raise SystemExit(f"ambiguous name {ref!r}: " + ", ".join(f"{i}:{mapping[i]}" for i in hits))


def cmd_info(ss: SSClient, _: argparse.Namespace) -> None:
    root = ss.system_info()
    server = root.find("server")
    if server is None:
        print(ET.tostring(root, encoding="unicode")[:500])
        return

    def t(tag: str) -> str:
        el = server.find(tag)
        return el.text or "" if el is not None else ""

    print(f"name={t('server-name') or t('name')} version={t('version')} cameras={t('camera-count')}")
    print(f"bonjour={t('bonjour-name')} wan={t('wan-address')}")


def cmd_cameras(ss: SSClient, _: argparse.Namespace) -> None:
    for c in ss.cameras():
        print(
            f"{c['number']:>3}  {c['name']:<16}  "
            f"{c['width']}x{c['height']}  connected={c['connected']}  "
            f"C:{c['mode_c']} M:{c['mode_m']} A:{c['mode_a']}"
        )


def cmd_schedules(ss: SSClient, _: argparse.Namespace) -> None:
    print("schedules:")
    for i, name in sorted(ss.schedules().items()):
        print(f"  {i:>3}  {name}")
    print("overrides:")
    for i, name in sorted(ss.overrides().items()):
        print(f"  {i:>3}  {name}")


def cmd_modes(ss: SSClient, args: argparse.Namespace) -> None:
    cam = ss.resolve_camera(args.camera)
    print(ss.get_text("++cameramodes", {"cameraNum": cam["number"]}).strip())


def cmd_set_schedule(ss: SSClient, args: argparse.Namespace) -> None:
    cam = ss.resolve_camera(args.camera)
    sid = ss.resolve_named_id(ss.schedules(), args.schedule)
    mode = args.mode.upper()
    ss.get_ok("++ssSetSchedule", {"cameraNum": cam["number"], "mode": mode, "id": sid})
    print(f"OK set-schedule camera={cam['number']}:{cam['name']} mode={mode} schedule={sid}")


def cmd_set_override(ss: SSClient, args: argparse.Namespace) -> None:
    cam = ss.resolve_camera(args.camera)
    oid = ss.resolve_named_id(ss.overrides(), args.override)
    mode = args.mode.upper()
    ss.get_ok("++ssSetOverride", {"cameraNum": cam["number"], "mode": mode, "id": oid})
    print(f"OK set-override camera={cam['number']}:{cam['name']} mode={mode} override={oid}")


def _arm(ss: SSClient, camera: str, api: str, arm: bool) -> None:
    cam = ss.resolve_camera(camera)
    ss.get_ok(api, {"cameraNum": cam["number"], "arm": "1" if arm else "0"})
    print(f"OK {api} camera={cam['number']}:{cam['name']} arm={arm}")


def cmd_snapshot(ss: SSClient, args: argparse.Namespace) -> None:
    cam = ss.resolve_camera(args.camera)
    out = Path(args.output)
    if out.exists() and not args.force:
        raise SystemExit(f"refusing to overwrite {out} (pass --force)")
    params: dict[str, Any] = {"cameraNum": cam["number"]}
    if args.width:
        params["width"] = args.width
    if args.height:
        params["height"] = args.height
    if args.quality:
        params["quality"] = args.quality
    r = ss.get("++image", params)
    out.write_bytes(r.content)
    print(f"saved {out} ({len(r.content)} bytes) camera={cam['number']}:{cam['name']}")


def cmd_clip(ss: SSClient, args: argparse.Namespace) -> None:
    cam = ss.resolve_camera(args.camera)
    out = Path(args.output)
    if out.exists() and not args.force:
        raise SystemExit(f"refusing to overwrite {out} (pass --force)")
    params = {"cameraNum": str(cam["number"]), "auth": ss.auth}
    stream_url = urljoin(ss.base, "++stream") + "?" + urlencode(params)
    cmd = [
        args.ffmpeg,
        "-y" if args.force else "-n",
        "-i",
        stream_url,
        "-t",
        str(args.seconds),
        "-c",
        "copy",
        str(out),
    ]
    print(f"recording {args.seconds}s from ++stream camera={cam['number']}:{cam['name']} -> {out}")
    subprocess.run(cmd, check=True)
    print(f"saved {out} camera={cam['number']}:{cam['name']}")


def cmd_list_files(ss: SSClient, args: argparse.Namespace) -> None:
    cam = ss.resolve_camera(args.camera) if args.camera else None
    params: dict[str, Any] = {
        "format": "xml",
        "ageText": str(args.days),
        "results": str(args.limit),
    }
    if cam is not None:
        params["cameraNum"] = cam["number"]
    kind = args.type
    if kind in ("all", "continuous", "cc"):
        params["ccFilesCheck"] = "1"
    if kind in ("all", "motion", "mc"):
        params["mcFilesCheck"] = "1"
    if kind in ("all", "image", "images"):
        params["imageFilesCheck"] = "1"
    if kind == "motion":
        params.setdefault("mcFilesCheck", "1")
    text = ss.get_text("++download", params)
    root = ET.fromstring(text)
    # Files appear as item/entry with title + link href depending on version
    items = root.findall(".//item") or root.findall(".//entry") or root.findall(".//file")
    if not items:
        found = 0
        for el in root.iter():
            href = el.attrib.get("href") or (el.findtext("link") or "")
            title = el.findtext("title") or el.findtext("name") or el.tag
            if href and ("getfile" in href or href.startswith("++")):
                print(f"{title}\t{href}")
                found += 1
        if found == 0:
            print(text[:1000])
        return
    for item in items:
        title = item.findtext("title") or item.findtext("name") or ""
        link = item.find("link")
        href = ""
        if link is not None:
            href = link.attrib.get("href") or (link.text or "")
        if not href:
            href = item.findtext("href") or item.attrib.get("href") or ""
        print(f"{title}\t{href}")


def cmd_get_file(ss: SSClient, args: argparse.Namespace) -> None:
    out = Path(args.output)
    if out.exists() and not args.force:
        raise SystemExit(f"refusing to overwrite {out} (pass --force)")
    href = args.href.lstrip("/")
    if not href.startswith("++"):
        href = "++" + href
    r = ss.get(href, stream=True)
    size = 0
    with out.open("wb") as fh:
        for chunk in r.iter_content(chunk_size=1 << 16):
            if chunk:
                fh.write(chunk)
                size += len(chunk)
    print(f"saved {out} ({size} bytes)")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="SecuritySpy HTTP API control")
    p.add_argument("--url", default=os.environ.get("SECURITYSPY_URL", DEFAULT_URL))
    p.add_argument("--user", default=os.environ.get("SECURITYSPY_USER"))
    p.add_argument("--password", default=os.environ.get("SECURITYSPY_PASS"))
    p.add_argument(
        "--insecure",
        action=argparse.BooleanOptionalAction,
        default=False,
        help="skip TLS verify (needed for self-signed HTTPS; default: verify)",
    )
    sub = p.add_subparsers(dest="cmd", required=True)

    sub.add_parser("info", help="server summary")
    sub.add_parser("cameras", help="list cameras")
    sub.add_parser("schedules", help="list schedules and overrides")

    m = sub.add_parser("modes", help="live arm modes for a camera")
    m.add_argument("camera")

    s = sub.add_parser("set-schedule", help="assign schedule to camera mode")
    s.add_argument("camera")
    s.add_argument("--mode", required=True, help="C, M, A, or X")
    s.add_argument("--schedule", required=True, help="schedule id or name")

    o = sub.add_parser("set-override", help="assign schedule override")
    o.add_argument("camera")
    o.add_argument("--mode", required=True)
    o.add_argument("--override", required=True)

    for name, help_ in (
        ("arm-motion", "arm motion capture"),
        ("disarm-motion", "disarm motion capture"),
        ("arm-actions", "arm actions"),
        ("disarm-actions", "disarm actions"),
        ("trigger-motion", "trigger motion detection"),
    ):
        sp = sub.add_parser(name, help=help_)
        sp.add_argument("camera")

    snap = sub.add_parser("snapshot", help="download a still JPEG")
    snap.add_argument("camera")
    snap.add_argument("-o", "--output", required=True)
    snap.add_argument("--width", type=int)
    snap.add_argument("--height", type=int)
    snap.add_argument("--quality", type=int)
    snap.add_argument("--force", action="store_true")

    clip = sub.add_parser("clip", help="record a short clip via ffmpeg ++stream")
    clip.add_argument("camera")
    clip.add_argument("-o", "--output", required=True)
    clip.add_argument("--seconds", type=int, default=10)
    clip.add_argument("--ffmpeg", default="ffmpeg")
    clip.add_argument("--force", action="store_true")

    lf = sub.add_parser("list-files", help="list recorded files")
    lf.add_argument("camera", nargs="?", help="camera name/number (optional)")
    lf.add_argument("--days", type=int, default=1)
    lf.add_argument("--type", choices=["all", "motion", "continuous", "image"], default="motion")
    lf.add_argument("--limit", type=int, default=50)

    gf = sub.add_parser("get-file", help="download a file by ++getfile* href")
    gf.add_argument("--href", required=True)
    gf.add_argument("-o", "--output", required=True)
    gf.add_argument("--force", action="store_true")

    return p


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    if not args.user or not args.password:
        print("error: --user/--password or SECURITYSPY_USER/SECURITYSPY_PASS required", file=sys.stderr)
        return 2

    ss = SSClient(args.url, args.user, args.password, insecure=args.insecure)

    handlers = {
        "info": cmd_info,
        "cameras": cmd_cameras,
        "schedules": cmd_schedules,
        "modes": cmd_modes,
        "set-schedule": cmd_set_schedule,
        "set-override": cmd_set_override,
        "snapshot": cmd_snapshot,
        "clip": cmd_clip,
        "list-files": cmd_list_files,
        "get-file": cmd_get_file,
    }

    if args.cmd == "arm-motion":
        _arm(ss, args.camera, "++ssControlMotionCapture", True)
    elif args.cmd == "disarm-motion":
        _arm(ss, args.camera, "++ssControlMotionCapture", False)
    elif args.cmd == "arm-actions":
        _arm(ss, args.camera, "++ssControlActions", True)
    elif args.cmd == "disarm-actions":
        _arm(ss, args.camera, "++ssControlActions", False)
    elif args.cmd == "trigger-motion":
        cam = ss.resolve_camera(args.camera)
        ss.get_ok("++triggermd", {"cameraNum": cam["number"]})
        print(f"OK triggermd camera={cam['number']}:{cam['name']}")
    else:
        handlers[args.cmd](ss, args)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except requests.HTTPError as e:
        print(f"HTTP error: {e} {getattr(e.response, 'text', '')[:300]}", file=sys.stderr)
        raise SystemExit(1) from e
    except subprocess.CalledProcessError as e:
        print(f"command failed: {e}", file=sys.stderr)
        raise SystemExit(e.returncode) from e
