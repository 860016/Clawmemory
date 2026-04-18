FROM python:3.10-slim

WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 安装 Python 依赖
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 复制应用代码
COPY . .

# 安装 Rust 安全引擎（预编译 wheel，优先尝试 ARM64，其次 x86_64）
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "aarch64" ]; then \
        WHEEL_URL="https://github.com/860016/Clawmemory/releases/latest/download/clawmemory_core-1.0.0-cp310-cp310-manylinux_2_17_aarch64.manylinux2014_aarch64.whl"; \
    else \
        WHEEL_URL="https://github.com/860016/Clawmemory/releases/latest/download/clawmemory_core-1.0.0-cp310-cp310-manylinux_2_17_x86_64.manylinux2014_x86_64.whl"; \
    fi && \
    echo "Downloading Rust engine: $WHEEL_URL" && \
    curl -sfL "$WHEEL_URL" -o /tmp/clawmemory_core.whl 2>/dev/null && \
    pip install --no-cache-dir /tmp/clawmemory_core.whl 2>/dev/null && \
    echo "Rust engine installed" || echo "Rust wheel not available, using Python fallback"; \
    rm -f /tmp/clawmemory_core.whl

# 创建数据目录
RUN mkdir -p /app/data /app/keys

EXPOSE 8765

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8765"]
