# app/config.py
import os
try:
    import yaml
except ImportError as e:
    raise RuntimeError(
        "PyYAML is required. Please add PyYAML to requirements.txt"
    ) from e

from functools import lru_cache

CONFIG_PATH = os.getenv("AP_CONTROLLER_CONFIG", "/app/config/controller.yaml")


@lru_cache()
def load_config() -> dict:
    with open(CONFIG_PATH, "r") as f:
        return yaml.safe_load(f)