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


def read_version() -> str:
    return VERSION_FILE.read_text().strip()


def write_version(ver: str):
    VERSION_FILE.write_text(ver + "\n")
    print(f"[OK] VERSION -> {ver}")


def sync_pyproject(ver: str):
    path = ROOT / "backend" / "pyproject.toml"
    content = path.read_text()
    content = re.sub(r'version\s*=\s*"[^"]*"', f'version = "{ver}"', content, count=1)
    path.write_text(content)
    print(f"[OK] backend/pyproject.toml -> {ver}")


def sync_package_json(ver: str):
    path = ROOT / "frontend" / "package.json"
    content = path.read_text()
    content = re.sub(r'"version"\s*:\s*"[^"]*"', f'"version": "{ver}"', content, count=1)
    path.write_text(content)
    print(f"[OK] frontend/package.json -> {ver}")


def sync_cargo_toml(ver: str):
    path = ROOT / "backend" / "clawmemory_core" / "Cargo.toml"
    if not path.exists():
        print("[SKIP] backend/clawmemory_core/Cargo.toml not found (commercial module)")
        return
    content = path.read_text()
    content = re.sub(r'version\s*=\s*"[^"]*"', f'version = "{ver}"', content, count=1)
    path.write_text(content)
    print(f"[OK] backend/clawmemory_core/Cargo.toml -> {ver}")


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
    sync_cargo_toml(ver)

    print(f"\n[DONE] All components synced to v{ver}")
    print("\nAuto-read from VERSION (no sync needed):")
    print("  - backend/app/config.py (APP_VERSION)")
    print("  - backend/app/main.py")
    print("  - backend/app/services/stats_service.py")
    print("  - backend/app/services/license_service.py")


if __name__ == "__main__":
    main()
