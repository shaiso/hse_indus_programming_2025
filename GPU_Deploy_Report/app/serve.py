import io
import os
from contextlib import asynccontextmanager

from app.model import ModelManager
from fastapi import FastAPI, File, HTTPException, UploadFile
from fastapi.responses import JSONResponse
from PIL import Image

model_manager = ModelManager(
    max_batch_size=int(os.environ.get("MAX_BATCH_SIZE", "8")),
    max_wait_ms=float(os.environ.get("MAX_WAIT_MS", "50")),
)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Загрузка модели при старте, graceful shutdown при остановке."""
    await model_manager.load()
    yield
    await model_manager.shutdown()


app = FastAPI(
    title="ResNet-50 Inference API",
    description="GPU-accelerated image classification with dynamic batching",
    lifespan=lifespan,
)


@app.get("/health")
async def health():
    """Health check для Kubernetes probes."""
    if not model_manager.ready:
        return JSONResponse(
            status_code=503,
            content={"status": "loading", "device": str(model_manager.device)},
        )
    return {
        "status": "ready",
        "device": str(model_manager.device),
        "model": "resnet50",
    }


@app.post("/predict")
async def predict(file: UploadFile = File(...)):
    """Классификация изображения. Возвращает top-5 классов."""
    if not model_manager.ready:
        raise HTTPException(status_code=503, detail="Model is loading")

    # Читаем и предобрабатываем изображение
    content = await file.read()
    try:
        image = Image.open(io.BytesIO(content))
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid image")

    tensor = model_manager.preprocess(image)

    # Отправляем в очередь dynamic batching
    results = await model_manager.predict(tensor)

    return {"filename": file.filename, "predictions": results}


@app.get("/info")
async def info():
    """Информация о сервере и устройстве."""
    import torch

    device_info = {
        "device": str(model_manager.device),
        "pytorch_version": torch.__version__,
    }

    if model_manager.device.type == "cuda":
        device_info["gpu_name"] = torch.cuda.get_device_name(0)
        device_info["vram_total_gb"] = round(
            torch.cuda.get_device_properties(0).total_mem / 1e9, 2
        )
        device_info["vram_used_gb"] = round(torch.cuda.memory_allocated(0) / 1e9, 2)
    elif model_manager.device.type == "mps":
        device_info["gpu_name"] = "Apple Silicon (MPS)"

    return device_info
