FROM python:3.12-slim

# Install system deps for C extension compilation
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libssl-dev curl && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install Python deps
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Try to download pre-built C core wheel
RUN PYVER=$(python -c "import sys; print(f'cp{sys.version_info.major}{sys.version_info.minor}')") && \
    ARCH=$(python -c "import platform; print(platform.machine().lower())") && \
    if [ "$ARCH" = "x86_64" ] || [ "$ARCH" = "amd64" ]; then \
        PLATFORM="manylinux*_x86_64"; \
    elif [ "$ARCH" = "aarch64" ]; then \
        PLATFORM="manylinux*_aarch64"; \
    else \
        PLATFORM="*"; \
    fi && \
    WHEEL_URL=$(curl -sf "https://api.github.com/repos/860016/Clawmemory/releases/latest" | \
        python -c "import json,sys; data=json.load(sys.stdin); \
        [print(a['browser_download_url']) for a in data.get('assets',[]) \
        if a['name'].endswith('.whl') and '$PYVER' in a['name'] and 'manylinux' in a['name'].lower()]" 2>/dev/null | head -1) && \
    if [ -n "$WHEEL_URL" ]; then \
        pip install --no-cache-dir "$WHEEL_URL" 2>/dev/null || echo "C wheel install skipped"; \
    fi || echo "C wheel download skipped"

# Copy app source
COPY . .

# Build frontend if not present
RUN if [ ! -f frontend_dist/index.html ]; then \
        echo "Frontend not pre-built, installing Node.js..."; \
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
        apt-get install -y nodejs && \
        cd /tmp && \
        git clone --depth 1 https://github.com/860016/Clawmemory.git clawmemory-frontend 2>/dev/null || true && \
        if [ -d /tmp/clawmemory-frontend/frontend ]; then \
            cd /tmp/clawmemory-frontend/frontend && npm install --silent && npm run build && \
            cp -r /tmp/clawmemory-frontend/backend/frontend_dist /app/frontend_dist; \
        fi && \
        rm -rf /tmp/clawmemory-frontend; \
    fi

# Create data and keys dirs
RUN mkdir -p /app/data /app/keys

EXPOSE 8765

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD curl -sf http://localhost:8765/api/v1/health || exit 1

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8765"]
