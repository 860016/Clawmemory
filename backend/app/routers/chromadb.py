import logging
import subprocess
import sys
from fastapi import APIRouter

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/chromadb", tags=["chromadb"])


@router.get("/status")
def get_chromadb_status():
    """Check if chromadb is installed and available"""
    try:
        import chromadb
        return {"available": True, "version": getattr(chromadb, "__version__", "unknown")}
    except ImportError:
        return {"available": False}


@router.post("/install")
def install_chromadb():
    """Install chromadb package"""
    try:
        import chromadb
        return {"success": True, "message": "Already installed"}
    except ImportError:
        pass

    try:
        result = subprocess.run(
            [sys.executable, "-m", "pip", "install", "chromadb"],
            capture_output=True,
            text=True,
            timeout=300
        )
        if result.returncode == 0:
            logger.info("chromadb installed successfully")
            return {"success": True, "message": "Installation successful"}
        else:
            logger.error(f"chromadb install failed: {result.stderr}")
            return {"success": False, "message": result.stderr[:500]}
    except subprocess.TimeoutExpired:
        return {"success": False, "message": "Installation timed out"}
    except Exception as e:
        logger.error(f"chromadb install error: {e}")
        return {"success": False, "message": str(e)}
