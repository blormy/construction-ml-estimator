# configs/config_loader.py
import os
import yaml
from dotenv import load_dotenv
from pathlib import Path

# загружаем .env из корня проекта
BASE_DIR = Path(__file__).resolve().parent.parent
load_dotenv(dotenv_path=BASE_DIR / '.env')


def load_config(config_path: Path) -> dict:
    """Загрузка конфигурации из YAML-файла"""
    if not config_path.exists():
        raise FileNotFoundError(f"Файл конфигурации не найден: {config_path}")

    with open(config_path, "r", encoding="utf-8") as file:
        return yaml.safe_load(file)


# создаем глобальный объект конфига
yaml_path = BASE_DIR / 'configs' / 'config.yaml'
config = load_config(yaml_path)