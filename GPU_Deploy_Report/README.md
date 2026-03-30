# Доклад: контейнеризация и деплой GPU-моделей

### Docker + CUDA, Kubernetes GPU

---

## План доклада

1. **Зачем** контейнеризировать ML-модели
2. **CUDA** — как GPU выполняет вычисления
3. **Docker + GPU** — как собирать GPU-образ
4. **Kubernetes** — оркестрация и распределение GPU
5. **Особенности деплоя** нейросетей
6. **Демо** — рабочий сервис классификации изображений

---

## Проблема: "У меня на ноутбуке работает"

ML-модель разработана в Jupyter Notebook. Для продакшена нужно:

- Правильная версия Python, PyTorch, CUDA, cuDNN
- Настроенный GPU-драйвер
- Совпадение версий между всеми компонентами
- Воспроизведение окружения на нескольких серверах

**Контейнер** = код + зависимости + окружение в одном пакете.
Работает одинаково на любом сервере с подходящим GPU-драйвером.

---

## Контейнер vs Виртуальная машина

```
  Виртуальная машина              Контейнер
┌──────────────────┐        ┌──────────────────┐
│     App A        │        │  App A  │  App B │
│   Guest OS       │        │  Libs A │  Libs B│
│   Hypervisor     │        │   Docker Engine  │
│   Host OS        │        │   Host OS (Linux)│
│   Hardware       │        │   Hardware       │
└──────────────────┘        └──────────────────┘
  Минуты запуска               Секунды запуска
  Гигабайты overhead           Мегабайты overhead
  Полная изоляция              Изоляция на уровне ядра
```

Контейнер использует ядро хоста — нет дублирования ОС.

---

## CUDA — зачем GPU для нейросетей

**CPU:** 8-64 мощных ядра — последовательные задачи
**GPU:** 10 000+ простых ядер — параллельные вычисления

Умножение матриц (основа нейросетей) — идеально параллельная задача.

```
CPU:  16 ядер         GPU: 10 000+ ядер
┌──┐┌──┐┌──┐┌──┐     ┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐
│  ││  ││  ││  │     ││││││││││││││││││││││││
└──┘└──┘└──┘└──┘     └┘└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘
Быстрые, но мало      Простые, но очень много
```

---

## Стек CUDA

```
┌──────────────────────────────────────────┐
│  PyTorch / TensorFlow                    │  ← пользовательский код
├──────────────────────────────────────────┤
│  cuDNN, cuBLAS, TensorRT                 │  ← библиотеки NVIDIA
├──────────────────────────────────────────┤
│  CUDA Runtime API                        │  ← высокоуровневый API
├──────────────────────────────────────────┤
│  CUDA Driver API                         │  ← низкоуровневый API
├──────────────────────────────────────────┤
│  NVIDIA Driver                           │  ← драйвер GPU
├──────────────────────────────────────────┤
│  GPU Hardware                            │  ← физическое устройство
└──────────────────────────────────────────┘
```

- **cuDNN** — оптимизированные операции для нейросетей (свёртки, нормализация)
- **TensorRT** — оптимизатор инференса (ускорение 2-5x)

---

## VRAM — главное ограничение

VRAM — память на видеокарте. Всё, с чем GPU работает, должно лежать в VRAM.

```
DDR5 RAM (CPU):     ~50 ГБ/с
GDDR6X (RTX 4090):  ~1000 ГБ/с
HBM3 (H100):        ~3350 ГБ/с
```

**Сколько VRAM нужно модели (FP16):**

| Модель | Параметры | VRAM | GPU |
|--------|-----------|------|-----|
| ResNet-50 | 25M | ~100 МБ | Любой |
| Stable Diffusion | 1B | ~6 ГБ | T4+ |
| LLaMA 8B | 8B | ~16 ГБ | A10 (24 ГБ) |
| LLaMA 70B | 70B | ~140 ГБ | 2× A100-80GB |

`VRAM ≈ параметры × байт_на_параметр + KV-кэш`

---

## Docker: ключевые концепции

**Image** — неизменяемый шаблон, состоит из слоёв:
```
Layer 4: COPY app.py           ← код
Layer 3: RUN pip install torch  ← зависимости
Layer 2: RUN apt-get install    ← системные пакеты
Layer 1: nvidia/cuda:12.2-base  ← базовый образ
```

**Container** — запущенный экземпляр образа
**Volume** — хранение данных за пределами контейнера
**Registry** — хранилище образов (Docker Hub, NGC)

Слои кэшируются — если изменился только код, остальное берётся из кэша.

---

## NVIDIA Container Toolkit

Docker не знает о GPU. Toolkit пробрасывает GPU-драйвер внутрь контейнера.

```
┌────────────────────────────┐
│       Container            │
│  App → CUDA Runtime        │
│          ↓                 │
│  nvidia-container-toolkit  │
│          ↓                 │
│  Host NVIDIA Driver        │
│          ↓                 │
│  Physical GPU              │
└────────────────────────────┘
```

```bash
docker run --gpus all my-model:latest
```

Версия CUDA в контейнере должна быть ≤ версии драйвера на хосте.

---

## Выбор базового образа

NVIDIA предоставляет три типа:

| Образ | Содержит | Размер | Когда |
|-------|---------|--------|-------|
| `base` | Минимальный runtime | ~200 МБ | Модель не требует CUDA |
| `runtime` | CUDA runtime libs | ~3.5 ГБ | **Инференс** |
| `devel` | + компилятор nvcc | ~7.5 ГБ | Обучение / компиляция |

```
nvidia/cuda:12.2.0-runtime-ubuntu22.04   ← для продакшена
nvidia/cuda:12.2.0-devel-ubuntu22.04     ← для сборки
```

**Правило:** в продакшене всегда `runtime`, не `devel`.

---

## Multi-stage build

**Проблема:** для сборки нужен `devel` (nvcc), для запуска — `runtime`.

```dockerfile
# Стадия 1: сборка (devel ~8 ГБ)
FROM nvidia/cuda:12.2.0-devel-ubuntu22.04 AS builder
COPY requirements.txt .
RUN pip install --prefix=/install -r requirements.txt

# Стадия 2: финальный образ (runtime ~4 ГБ)
FROM nvidia/cuda:12.2.0-runtime-ubuntu22.04
COPY --from=builder /install /usr/local
COPY app/ ./app/
CMD ["python3", "app/serve.py"]
```

Экономия: **4-5 ГБ** на размере образа.

---

## Порядок слоёв и кэширование

От редко меняющихся к часто меняющимся:

```dockerfile
# 1. Системные пакеты (раз в месяц)
RUN apt-get install python3

# 2. Python-зависимости (раз в неделю)
COPY requirements.txt .
RUN pip install -r requirements.txt

# 3. Код приложения (каждый день)
COPY app/ ./app/
```

Изменился только код → слои 1, 2 из кэша → сборка за секунды.

---

## Веса модели — вне образа!

```dockerfile
# ПЛОХО: образ 20+ ГБ, пересборка при каждом обновлении весов
COPY model_weights/ /app/weights/
```

```bash
# ХОРОШО: веса монтируются при запуске
docker run -v /data/weights:/app/weights my-model
```

Варианты хранения:
- **Volume mount** — один сервер
- **PersistentVolume** — Kubernetes
- **S3/GCS** — скачиваются init-контейнером
- **Model Registry** (MLflow, W&B) — версионирование

---

## Kubernetes: зачем

Docker = один сервер. Kubernetes = кластер серверов.

| Задача | Решение K8s |
|--------|------------|
| Распределить контейнеры по серверам | Scheduler |
| Перезапустить упавший контейнер | Self-healing |
| Масштабировать при нагрузке | HPA |
| Обновить без простоя | Rolling Update |
| Балансировать трафик | Service |

---

## Архитектура Kubernetes

```
┌────────────────── Control Plane ──────────────────┐
│  API Server │ Scheduler │ etcd │ Controller Mgr   │
└──────────────────────┬────────────────────────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
   ┌────────────┐┌────────────┐┌────────────┐
   │  Node 1    ││  Node 2    ││  Node 3    │
   │ Pod A,B    ││ Pod C,D    ││ Pod E,F    │
   │ kubelet    ││ kubelet    ││ kubelet    │
   └────────────┘└────────────┘└────────────┘
```

- **Pod** — минимальная единица, один или несколько контейнеров
- **Deployment** — "хочу 3 реплики" → K8s обеспечивает
- **Service** — стабильный адрес для группы подов

---

## Путь запроса в Kubernetes

```
        Пользователь
            │
            ▼
        ┌────────┐
        │Ingress │  маршрутизация по URL
        └───┬────┘
            ▼
        ┌────────┐
        │Service │  балансировка нагрузки
        └───┬────┘
            ▼
      ┌─────┼─────┐
      ▼     ▼     ▼
   Pod 1  Pod 2  Pod 3  ← реплики (Deployment)
      │     │     │
      ▼     ▼     ▼
   ┌──────────────────┐
   │   PVC (веса)     │  общие веса модели
   └──────────────────┘
```

---

## GPU в Kubernetes

K8s не знает о GPU из коробки. **NVIDIA Device Plugin** регистрирует GPU как ресурс:

```yaml
resources:
  requests:
    nvidia.com/gpu: 1      # запрашиваем 1 GPU
  limits:
    nvidia.com/gpu: 1      # requests = limits обязательно!
```

**GPU нельзя "оверкоммитить"** — 1 GPU = 1 контейнер.

```
Node "gpu-worker-01":
  Capacity:
    nvidia.com/gpu: 4      ← 4 GPU на ноде
  Allocatable:
    nvidia.com/gpu: 2      ← 2 свободны
```

---

## Taints и Tolerations — защита GPU-нод

GPU-ноды дорогие. Без защиты на них попадут обычные поды (API, БД).

```bash
# Пометить ноду: "только для GPU-задач"
kubectl taint nodes gpu-01 nvidia.com/gpu=present:NoSchedule
```

```yaml
# Под с GPU должен иметь toleration
tolerations:
  - key: nvidia.com/gpu
    operator: Exists
    effect: NoSchedule
```

Обычные поды → отталкиваются от GPU-нод
GPU-поды с toleration → размещаются

---

## Стратегии разделения GPU

```
 1 GPU = 1 Pod (по умолчанию)
 Просто, изолированно, но дорого

 MIG (A100/H100) — аппаратное разделение
 GPU: [MIG-3g 40GB | MIG-2g 20GB | MIG-2g 20GB]
      [  Pod A     |   Pod B     |   Pod C    ]

 Time-Slicing — разделение по времени
 GPU: Pod A ─ Pod B ─ Pod A ─ Pod B (чередование)
 Нет изоляции памяти!

 Multi-GPU — несколько GPU на один под
 Pod A → GPU 0 + GPU 1 + GPU 2 + GPU 3
 Для моделей, не помещающихся в 1 GPU
```

---

## Холодный старт — главная боль GPU-деплоя

| Этап | Время |
|------|-------|
| Запуск контейнера | 1-5 сек |
| Загрузка Python + зависимости | 5-15 сек |
| Загрузка весов в RAM | 10-60 сек |
| Перенос на GPU + прогрев | 5-30 сек |
| **Итого** | **30-120 сек** |

Это критично для масштабирования, обновлений и отказоустойчивости.

---

## Решение: Kubernetes Probes

```yaml
# Не убивать долго стартующий под
startupProbe:
  httpGet: { path: /health, port: 8080 }
  failureThreshold: 30     # до 5 минут на старт
  periodSeconds: 10

# Не слать трафик пока модель не загружена
readinessProbe:
  httpGet: { path: /health, port: 8080 }
  periodSeconds: 10

# Перезапустить зависший под
livenessProbe:
  httpGet: { path: /health, port: 8080 }
  periodSeconds: 30
```

Health endpoint возвращает 503 при загрузке, 200 когда готов.

---

## Dynamic Batching

GPU эффективен при обработке батчей. Один запрос — GPU простаивает на 80-90%.

```
Без батчинга:              С батчингом:
Req1 → GPU → Resp1        Req1 ─┐
Req2 → GPU → Resp2        Req2 ─┼→ GPU → [R1, R2, R3, R4]
Req3 → GPU → Resp3        Req3 ─┤
Req4 → GPU → Resp4        Req4 ─┘
Время: 4 × T              Время: ~1.2 × T
GPU util: 10-20%           GPU util: 60-90%
```

Накапливаем запросы N мс → обрабатываем батчем → раздаём результаты.

---

## Стратегии обновления

**Rolling Update** — постепенная замена подов (`maxSurge: 1`, `maxUnavailable: 0`):

```
Шаг 1:  [v1] [v1] [v1]              ← исходное состояние
Шаг 2:  [v1] [v1] [v1] [v2 ...]     ← создан новый под, грузит модель
Шаг 3:  [v1] [v1] [v1] [v2 ✓]      ← v2 прошёл readiness probe
Шаг 4:  [v1] [v1] [──] [v2 ✓]      ← один v1 удалён
Шаг 5:  [v1] [v1] [v2 ✓] [v2 ...]  ← следующий v2 стартует
  ...
Итог:   [v2] [v2] [v2]              ← все обновлены, 0 downtime
```

Нужен **1 дополнительный GPU** на время обновления.

**Canary** — 10% трафика на v2, 90% на v1. Проверяем метрики → постепенно переключаем.
**Blue-Green** — полная копия кластера, мгновенное переключение (2x ресурсов).

---

## Мониторинг GPU

```
GPU → DCGM Exporter → Prometheus → Grafana

Ключевые метрики:
  - GPU utilization (%)
  - VRAM usage (МБ)
  - Температура (°C)
  - Inference latency (мс)
  - Throughput (req/sec)
```

**HPA** (автомасштабирование) — добавляет реплики при росте нагрузки.
Проблема: новая реплика стартует 30-120 сек → нужно предиктивное масштабирование.

---

## Демо: наш проект

**FastAPI + ResNet-50 + Dynamic Batching**

```
app/
├── device.py     ← автоопределение: cuda / mps / cpu
├── model.py      ← ResNet-50 + dynamic batching
└── serve.py      ← /health, /predict, /info

Dockerfile        ← multi-stage: devel → runtime
docker-compose.yml
k8s/              ← deployment, service, configmap, hpa
```

Отправляем изображение → модель классифицирует (top-5 ImageNet классов).

---

## Демо: результат

```bash
curl -X POST http://localhost:8080/predict -F "file=@cat.jpg"
```

```json
{
  "filename": "cat.jpg",
  "predictions": [
    { "class": "tabby cat",     "confidence": 0.72 },
    { "class": "tiger cat",     "confidence": 0.15 },
    { "class": "Egyptian cat",  "confidence": 0.08 },
    { "class": "lynx",          "confidence": 0.02 },
    { "class": "Persian cat",   "confidence": 0.01 }
  ]
}
```

---

## Итоги

| Аспект | Решение |
|--------|---------|
| **GPU-образ** | `runtime` + multi-stage build |
| **Веса модели** | Volumes / S3, не в образе |
| **GPU в K8s** | Device Plugin, taints, tolerations |
| **Разделение GPU** | 1:1, MIG, time-slicing |
| **Холодный старт** | Probes + прогрев модели |
| **Эффективность** | Dynamic batching |
| **Оптимизация** | FP16 / INT8, TensorRT |
| **Обновления** | Rolling Update, Canary |
| **Мониторинг** | DCGM → Prometheus → Grafana |

---

## Спасибо!
