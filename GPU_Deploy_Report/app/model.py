import asyncio
import time
from collections import deque
from typing import Any

import torch
from app.device import get_device
from PIL import Image
from torchvision import models, transforms

# Стандартная предобработка для ImageNet-моделей
TRANSFORM = transforms.Compose(
    [
        transforms.Resize(256),
        transforms.CenterCrop(224),
        transforms.ToTensor(),
        transforms.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]),
    ]
)


class ModelManager:
    """Управляет загрузкой модели и инференсом с dynamic batching."""

    def __init__(self, max_batch_size: int = 8, max_wait_ms: float = 50.0):
        self.device = get_device()
        self.model: Any = None
        self.labels: list[str] = []
        self.ready = False

        # Dynamic batching
        self.max_batch_size = max_batch_size
        self.max_wait_ms = max_wait_ms
        self._queue: deque[tuple[torch.Tensor, asyncio.Future]] = deque()
        self._batch_task: asyncio.Task | None = None

    async def load(self) -> None:
        """Загружает ResNet-50 и прогревает модель."""
        print(f"[model] Загрузка ResNet-50 на {self.device}...")
        start = time.time()

        self.model = models.resnet50(weights=models.ResNet50_Weights.IMAGENET1K_V2)
        self.model.to(self.device)
        self.model.eval()  # инференс

        # Загружаем названия классов ImageNet
        self.labels = models.ResNet50_Weights.IMAGENET1K_V2.meta["categories"]

        # Прогревочный инференс — первый вызов всегда медленный
        dummy = torch.randn(1, 3, 224, 224).to(self.device)
        with torch.no_grad():
            self.model(dummy)

        elapsed = time.time() - start
        print(f"[model] Модель загружена за {elapsed:.1f}s на {self.device}")
        self.ready = True

        # Запускаем фоновый обработчик батчей
        self._batch_task = asyncio.create_task(self._batch_loop())

    def preprocess(self, image: Image.Image) -> torch.Tensor:
        """Предобработка изображения для ResNet."""
        return TRANSFORM(image.convert("RGB"))

    async def predict(self, tensor: torch.Tensor) -> list[dict]:
        """Ставит запрос в очередь и ждёт результат (dynamic batching)."""
        future: asyncio.Future[list[dict]] = asyncio.get_event_loop().create_future()
        self._queue.append((tensor, future))
        return await future

    async def _batch_loop(self) -> None:
        """Фоновый цикл: собирает батч и запускает инференс."""
        while True:
            if not self._queue:
                await asyncio.sleep(0.005)
                continue

            # Ждём накопления батча или таймаут
            await asyncio.sleep(self.max_wait_ms / 1000)

            if not self._queue:
                continue

            # Собираем батч
            batch_tensors: list[torch.Tensor] = []
            batch_futures: list[asyncio.Future] = []

            while self._queue and len(batch_tensors) < self.max_batch_size:
                tensor, future = self._queue.popleft()
                batch_tensors.append(tensor)
                batch_futures.append(future)

            batch_size = len(batch_tensors)
            print(f"[batch] Обработка батча из {batch_size} запросов")

            # Инференс
            try:
                batch = torch.stack(batch_tensors).to(self.device)

                with torch.no_grad():
                    outputs = self.model(batch)

                probabilities = torch.nn.functional.softmax(outputs, dim=1)

                for i, future in enumerate(batch_futures):
                    probs = probabilities[i]
                    top5_prob, top5_idx = torch.topk(probs, 5)

                    results = [
                        {
                            "class": self.labels[idx.item()],
                            "confidence": round(prob.item(), 4),
                        }
                        for prob, idx in zip(top5_prob, top5_idx)
                    ]
                    future.set_result(results)

            except Exception as e:
                for future in batch_futures:
                    if not future.done():
                        future.set_exception(e)

    async def shutdown(self) -> None:
        """Graceful shutdown: дождаться обработки очереди."""
        print("[model] Завершение работы...")
        if self._batch_task:
            self._batch_task.cancel()
        # Очистка GPU-памяти
        if self.device.type == "cuda":
            torch.cuda.empty_cache()
        self.ready = False
