# ============================================================
# Стадия 1: сборка зависимостей (devel — есть компилятор nvcc)
# ============================================================
FROM nvidia/cuda:12.2.0-devel-ubuntu22.04 AS builder

RUN apt-get update && apt-get install -y \
    python3 python3-dev python3-pip python3-venv \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY requirements.txt .

# Устанавливаем зависимости в venv — потом скопируем целиком
RUN python3 -m venv /opt/venv && \
    /opt/venv/bin/pip install --no-cache-dir -r requirements.txt


# ============================================================
# Стадия 2: финальный образ (runtime — без компилятора, легче)
# ============================================================
FROM nvidia/cuda:12.2.0-runtime-ubuntu22.04

RUN apt-get update && apt-get install -y \
    python3 python3-venv \
    && rm -rf /var/lib/apt/lists/*

# Копируем venv из builder-стадии
COPY --from=builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

WORKDIR /app

# Сначала копируем requirements (кэширование слоёв Docker):
# если requirements не изменились — слои выше берутся из кэша
COPY requirements.txt .

# Затем копируем код приложения (меняется чаще)
COPY app/ ./app/

# Непривилегированный пользователь (безопасность)
RUN useradd -m appuser && chown -R appuser /app
USER appuser

EXPOSE 8080

# Health check для Docker и Kubernetes
HEALTHCHECK --interval=30s --timeout=5s --start-period=60s \
    CMD python3 -c "import urllib.request; urllib.request.urlopen('http://localhost:8080/health')"

# Запуск сервера
CMD ["python3", "-m", "uvicorn", "app.serve:app", "--host", "0.0.0.0", "--port", "8080"]
