"""
Version sync script - reads version from root VERSION file, syncs to all components.
Usage: python sync_version.py [new_version]
  - No args: read VERSION file, sync to component configs
  - With arg: update VERSION file first, then sync
"""
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
VERSION_FILE = ROOT / "VERSION"
ENC = "utf-8"


def read_version() -> str:
    return VERSION_FILE.read_text(encoding=ENC).strip()


def write_version(ver: str):
    VERSION_FILE.write_text(ver + "\n", encoding=ENC)
    print(f"[OK] VERSION -> {ver}")


def _sync_file(path: Path, pattern: str, replacement: str, label: str):
    """Generic file sync: replace pattern with replacement, report result."""
    if not path.exists():
        print(f"[SKIP] {label} not found")
        return
    content = path.read_text(encoding=ENC)
    new_content = re.sub(pattern, replacement, content, count=1)
    if new_content == content:
        print(f"[SKIP] {label} — version already up to date or pattern not found")
        return
    path.write_text(new_content, encoding=ENC)
    print(f"[OK] {label}")


def sync_pyproject(ver: str):
    _sync_file(
        ROOT / "backend" / "pyproject.toml",
        r'version\s*=\s*"[^"]*"',
        f'version = "{ver}"',
        f"backend/pyproject.toml -> {ver}",
    )


def sync_package_json(ver: str):
    _sync_file(
        ROOT / "frontend" / "package.json",
        r'"version"\s*:\s*"[^"]*"',
        f'"version": "{ver}"',
        f"frontend/package.json -> {ver}",
    )


def sync_package_lock(ver: str):
    path = ROOT / "frontend" / "package-lock.json"
    if not path.exists():
        print("[SKIP] frontend/package-lock.json not found")
        return
    content = path.read_text(encoding=ENC)
    new_content = content

    # 1. Replace top-level "version": "X.Y.Z" (right after "name": "clawmemory")
    new_content = re.sub(
        r'("name":\s*"clawmemory",\s*\n\s*"version":\s*)"\d+\.\d+\.\d+[^"]*"',
        rf'\1"{ver}"',
        new_content, count=1
    )

    # 2. Replace packages."".version — the root package entry
    #    Pattern: "": { "name": "clawmemory", ... "version": "X.Y.Z"
    new_content = re.sub(
        r'("":\s*\{\s*\n\s*"name":\s*"clawmemory",\s*\n\s*"version":\s*)"\d+\.\d+\.\d+[^"]*"',
        rf'\1"{ver}"',
        new_content, count=1
    )

    if new_content != content:
        path.write_text(new_content, encoding=ENC)
        print(f"[OK] frontend/package-lock.json -> {ver}")
    else:
        print(f"[SKIP] frontend/package-lock.json — already up to date")


def sync_settings_view(ver: str):
    _sync_file(
        ROOT / "frontend" / "src" / "views" / "SettingsView.vue",
        r"appVersion\s*=\s*ref\(['\"][^'\"]*['\"]\)",
        f"appVersion = ref('{ver}')",
        f"frontend/src/views/SettingsView.vue -> {ver}",
    )


def sync_install_sh(ver: str):
    _sync_file(
        ROOT / "install.sh",
        r"v\d+\.\d+\.\d+",
        f"v{ver}",
        f"install.sh -> v{ver}",
    )


def sync_install_ps1(ver: str):
    _sync_file(
        ROOT / "install.ps1",
        r"v\d+\.\d+\.\d+",
        f"v{ver}",
        f"install.ps1 -> v{ver}",
    )


def main():
    if len(sys.argv) > 1:
        new_ver = sys.argv[1].strip()
        if not re.match(r'^\d+\.\d+\.\d+', new_ver):
            print(f"[ERR] Invalid version format: {new_ver}")
            print("   Expected: X.Y.Z or X.Y.Z-suffix")
            sys.exit(1)
        write_version(new_ver)
    else:
        print("[INFO] Reading version from VERSION file...")

    ver = read_version()
    print(f"\n[SYNC] Syncing version {ver} to all components...\n")

    sync_pyproject(ver)
    sync_package_json(ver)
    sync_package_lock(ver)
    sync_settings_view(ver)
    sync_install_sh(ver)
    sync_install_ps1(ver)

    print(f"\n[DONE] All components synced to v{ver}")
    print("\nAuto-read from VERSION (no sync needed):")
    print("  - backend/app/config.py (APP_VERSION)")
    print("  - backend/app/main.py")
    print("  - backend/app/services/stats_service.py")
    print("  - backend/app/services/license_service.py")


if __name__ == "__main__":
    main()
