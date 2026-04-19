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


def sync_pyproject(ver: str):
    path = ROOT / "backend" / "pyproject.toml"
    content = path.read_text(encoding=ENC)
    content = re.sub(r'version\s*=\s*"[^"]*"', f'version = "{ver}"', content, count=1)
    path.write_text(content, encoding=ENC)
    print(f"[OK] backend/pyproject.toml -> {ver}")


def sync_package_json(ver: str):
    path = ROOT / "frontend" / "package.json"
    content = path.read_text(encoding=ENC)
    content = re.sub(r'"version"\s*:\s*"[^"]*"', f'"version": "{ver}"', content, count=1)
    path.write_text(content, encoding=ENC)
    print(f"[OK] frontend/package.json -> {ver}")


def sync_package_lock(ver: str):
    path = ROOT / "frontend" / "package-lock.json"
    if not path.exists():
        print("[SKIP] frontend/package-lock.json not found")
        return
    content = path.read_text(encoding=ENC)
    # Update the top-level version
    content = re.sub(r'"version"\s*:\s*"\d+\.\d+\.\d+[^"]*"',
                     f'"version": "{ver}"', content, count=1)
    path.write_text(content, encoding=ENC)
    print(f"[OK] frontend/package-lock.json -> {ver}")


def sync_settings_view(ver: str):
    path = ROOT / "frontend" / "src" / "views" / "SettingsView.vue"
    if not path.exists():
        print("[SKIP] frontend/src/views/SettingsView.vue not found")
        return
    content = path.read_text(encoding=ENC)
    content = re.sub(r"appVersion\s*=\s*ref\(['\"][^'\"]*['\"]\)",
                     f"appVersion = ref('{ver}')", content)
    path.write_text(content, encoding=ENC)
    print(f"[OK] frontend/src/views/SettingsView.vue -> {ver}")


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

    print(f"\n[DONE] All components synced to v{ver}")
    print("\nAuto-read from VERSION (no sync needed):")
    print("  - backend/app/config.py (APP_VERSION)")
    print("  - backend/app/main.py")
    print("  - backend/app/services/stats_service.py")
    print("  - backend/app/services/license_service.py")


if __name__ == "__main__":
    main()
