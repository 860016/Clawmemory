import json
import logging
from pathlib import Path
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
from app.config import settings

logger = logging.getLogger(__name__)


class MemoryFileHandler(FileSystemEventHandler):
    """Watches memory files and syncs changes to database"""

    def __init__(self, user_id: int, memory_service, vector_service):
        self.user_id = user_id
        self.memory_service = memory_service
        self.vector_service = vector_service

    def on_modified(self, event):
        if event.is_directory:
            return
        path = Path(event.src_path)
        if path.suffix in (".md", ".json", ".txt"):
            self._sync_file(path)

    def on_created(self, event):
        if event.is_directory:
            return
        path = Path(event.src_path)
        if path.suffix in (".md", ".json", ".txt"):
            self._sync_file(path)

    def _sync_file(self, path: Path):
        try:
            content = path.read_text(encoding="utf-8")
            if path.suffix == ".json":
                data = json.loads(content)
                if isinstance(data, list):
                    for item in data:
                        self._upsert_memory(item)
                elif isinstance(data, dict):
                    self._upsert_memory(data)
            else:
                self._upsert_memory({
                    "key": path.stem,
                    "value": content,
                    "layer": "knowledge",
                    "source": "file_sync",
                })
        except Exception as e:
            logger.error(f"Failed to sync file {path}: {e}")

    def _upsert_memory(self, data: dict):
        memory = self.memory_service.create(data)
        self.vector_service.add_memory(
            self.user_id, memory.id,
            f"{data.get('key', '')}: {data.get('value', '')}",
            metadata={"layer": data.get("layer", ""), "source": "file_sync"},
        )


class MemoryFileWatcher:
    def __init__(self):
        self.observer = Observer()
        self._watches = {}

    def start_watching(self, user_id: int, agent_id: int | None, watch_dir: Path, memory_service, vector_service):
        handler = MemoryFileHandler(user_id, memory_service, vector_service)
        watch = self.observer.schedule(handler, str(watch_dir), recursive=True)
        self._watches[user_id] = watch
        logger.info(f"Started watching: {watch_dir}")

    def stop_watching(self, user_id: int, agent_id: int | None = None):
        if user_id in self._watches:
            self.observer.unschedule(self._watches.pop(user_id))

    def start(self):
        self.observer.start()

    def stop(self):
        self.observer.stop()
        self.observer.join()


file_watcher = MemoryFileWatcher()
