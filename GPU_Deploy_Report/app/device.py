import torch


def get_device() -> torch.device:
    """Автоматически определяет лучшее доступное устройство."""
    if torch.cuda.is_available():
        return torch.device("cuda")
    if torch.backends.mps.is_available():
        return torch.device("mps")
    return torch.device("cpu")
