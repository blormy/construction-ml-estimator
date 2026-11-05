"""Модуль предсказания веса оборудования по обученным ML-моделям"""

import numpy as np
import joblib
import logging
from pathlib import Path

logger = logging.getLogger(__name__)

# порядок параметров, которые ожидает модель (должен совпадать с обучением)
FEATURE_COLUMNS = [
    "flow_rate",
    "fluid_head",
    "rpm",
    "spec_gravity",
    "power_kw",
    "diameter_proxy",
    "useful_kw_log",
]

# путь к папке с моделями насосов
MODELS_DIR = Path(__file__).resolve().parent / "models" / "pump"


class PumpPredictor:
    """Предиктор веса насоса с помощью GradientBoosting и PolynomialFeatures"""

    def __init__(self, models_dir: Path = MODELS_DIR):
        model_path = models_dir / "model_gb.joblib"
        poly_path = models_dir / "poly_features.joblib"

        if not model_path.exists():
            raise FileNotFoundError(f"Файл модели не найден: {model_path}")
        if not poly_path.exists():
            raise FileNotFoundError(f"Файл PolynomialFeatures не найден: {poly_path}")

        self.model = joblib.load(model_path)
        self.poly = joblib.load(poly_path)
        logger.info("Модель насоса загружена успешно (model_gb + poly_features)")

    def predict(self, features_dict: dict) -> float:
        # собираем вектор параметров в нужном порядке
        raw_features = []
        for col in FEATURE_COLUMNS:
            value = features_dict.get(col)
            if value is None:
                raise ValueError(f"Отсутствует обязательный параметр для модели: '{col}'")
            raw_features.append(float(value))

        X = np.array([raw_features])

        # полиномиальные и кросс-признаки
        X_poly = self.poly.transform(X)

        # предсказание (модель возвращает weight_log)
        weight_log = self.model.predict(X_poly)[0]

        # обратное преобразование: exp(weight_log) = вес в кг
        weight_kg = np.exp(weight_log)

        logger.info(f"Предсказание (насос): weight_log={weight_log:.4f}, weight_kg={weight_kg:.2f}")
        return float(weight_kg)

# предиктор для сосудов

VESSEL_FEATURE_COLUMNS = [
    "log_liq_volume",
    "log_diameter",
    "log_ss_distance",
    "log_p_abs",
    "log_area_calc",
    "log_thick_proxy",
    "des_temp",
    "sk_height",
    "leg_height",
    "has_skirt",
    "has_legs"
]

VESSEL_MODELS_DIR = Path(__file__).resolve().parent / "models" / "vessel"

class VesselPredictor:
    """Предиктор веса сосуда (Vessel) с помощью GradientBoosting"""

    def __init__(self, models_dir: Path = VESSEL_MODELS_DIR):
        model_path = models_dir / "vessel_model_final.joblib"

        if not model_path.exists():
            # если нет в папке models/vessel, проверяем в корне ml_service
            alt_path = Path(__file__).resolve().parent / "vessel_model_final.joblib"
            if alt_path.exists():
                model_path = alt_path
            else:
                raise FileNotFoundError(f"Файл модели не найден: {model_path}")

        self.model = joblib.load(model_path)
        logger.info("Модель сосуда загружена успешно (vessel_model_final)")

    def predict(self, features_dict: dict) -> float:
        # собираем вектор параметров в нужном порядке
        raw_features = []
        for col in VESSEL_FEATURE_COLUMNS:
            value = features_dict.get(col)
            if value is None:
                raise ValueError(f"Отсутствует обязательный параметр для модели: '{col}'")
            raw_features.append(float(value))

        X = np.array([raw_features])

        # предсказание (модель возвращает weight_log)
        weight_log = self.model.predict(X)[0]

        # обратное преобразование: exp(weight_log) = вес в кг
        weight_kg = np.exp(weight_log)

        logger.info(f"Предсказание (сосуд): weight_log={weight_log:.4f}, weight_kg={weight_kg:.2f}")
        return float(weight_kg)

# предиктор для конвейеров

CONVEYOR_FEATURE_COLUMNS = [
    "log_length_ft",
    "log_width_in",
    "log_flow_tph",
    "log_belt_surface_area",
    "log_load_proxy",
]

CONVEYOR_MODELS_DIR = Path(__file__).resolve().parent / "models" / "conveyer"

class ConveyorPredictor:
    """Предиктор веса конвейера с помощью HistGradientBoosting"""

    def __init__(self, models_dir: Path = CONVEYOR_MODELS_DIR):
        model_path = models_dir / "conveyor_model_final.joblib"

        if not model_path.exists():
            alt_path = Path(__file__).resolve().parent / "conveyor_model_final.joblib"
            if alt_path.exists():
                model_path = alt_path
            else:
                raise FileNotFoundError(f"Файл модели не найден: {model_path}")

        self.model = joblib.load(model_path)
        logger.info("Модель конвейера загружена успешно (conveyor_model_final)")

    def predict(self, features_dict: dict) -> float:
        # собираем вектор параметров в нужном порядке
        raw_features = []
        for col in CONVEYOR_FEATURE_COLUMNS:
            value = features_dict.get(col)
            if value is None:
                raise ValueError(f"Отсутствует обязательный параметр для модели: '{col}'")
            raw_features.append(float(value))

        X = np.array([raw_features])

        # предсказание (модель возвращает weight_kg_log = log1p(weight_kg))
        weight_log = self.model.predict(X)[0]

        # обратное преобразование: expm1(weight_log) = вес в кг
        weight_kg = np.expm1(weight_log)

        logger.info(f"Предсказание (конвейер): weight_log={weight_log:.4f}, weight_kg={weight_kg:.2f}")
        return float(weight_kg)

# предиктор для горизонтальных емкостей

DRUM_FEATURE_COLUMNS = [
    "log_liq_volume",
    "log_ves_diameter",
    "log_ss_distance",
    "log_gauge_pres",
    "log_volume_proxy",
    "log_surface_area",
    "aspect_ratio",
]

DRUM_MODELS_DIR = Path(__file__).resolve().parent / "models" / "drum"

class DrumPredictor:
    """Предиктор веса горизонтальной емкости (Drum) с помощью TransformedTargetRegressor"""

    def __init__(self, models_dir: Path = DRUM_MODELS_DIR):
        model_path = models_dir / "drum_model_final.joblib"

        if not model_path.exists():
            alt_path = Path(__file__).resolve().parent / "drum_model_final.joblib"
            if alt_path.exists():
                model_path = alt_path
            else:
                raise FileNotFoundError(f"Файл модели не найден: {model_path}")

        self.model = joblib.load(model_path)
        logger.info("Модель горизонтальной емкости загружена успешно (drum_model_final)")

    def predict(self, features_dict: dict) -> float:
        # собираем вектор параметров в нужном порядке
        raw_features = []
        for col in DRUM_FEATURE_COLUMNS:
            value = features_dict.get(col)
            if value is None:
                raise ValueError(f"Отсутствует обязательный параметр для модели: '{col}'")
            raw_features.append(float(value))

        X = np.array([raw_features])

        # предсказание (модель возвращает вес в кг напрямую)
        weight_kg = self.model.predict(X)[0]

        logger.info(f"Предсказание (емкость): weight_kg={weight_kg:.2f}")
        return float(weight_kg)

# предиктор для теплообменников

UTUBE_FEATURE_COLUMNS = [
    "heat_area",
    "shell_diameter",
    "tube_out_diameter",
    "tube_des_pres",
    "tube_len",
]

UTUBE_MODELS_DIR = Path(__file__).resolve().parent / "models"

class UTubePredictor:
    """Предиктор веса теплообменников (U-Tube) с помощью RandomForestRegressor"""

    def __init__(self, models_dir: Path = UTUBE_MODELS_DIR):
        model_path = models_dir / "utube_weight_model.joblib"

        if not model_path.exists():
            alt_path = Path(__file__).resolve().parent / "utube_weight_model.joblib"
            if alt_path.exists():
                model_path = alt_path
            else:
                raise FileNotFoundError(f"Файл модели не найден: {model_path}")

        self.model = joblib.load(model_path)
        logger.info("Модель теплообменника загружена успешно (utube_weight_model)")

    def predict(self, features_dict: dict) -> float:
        # собираем вектор параметров в нужном порядке
        raw_features = []
        for col in UTUBE_FEATURE_COLUMNS:
            value = features_dict.get(col)
            if value is None:
                raise ValueError(f"Отсутствует обязательный параметр для модели: '{col}'")
            raw_features.append(float(value))

        X = np.array([raw_features])

        # предсказание
        weight_kg = self.model.predict(X)[0]

        logger.info(f"Предсказание (теплообменник): weight_kg={weight_kg:.2f}")
        return float(weight_kg)

# предиктор для колонн

TOWER_FEATURE_COLUMNS = [
    "diameter",
    "ss_dist",
    "pressure",
    "trays_num"
]

TOWER_MODELS_DIR = Path(__file__).resolve().parent / "models"

class TowerPredictor:
    """Предиктор веса колонн (Tower) с помощью RandomForestRegressor Pipeline"""

    def __init__(self, models_dir: Path = TOWER_MODELS_DIR):
        model_path = models_dir / "tower_weight_model.joblib"

        if not model_path.exists():
            alt_path = Path(__file__).resolve().parent / "tower_weight_model.joblib"
            if alt_path.exists():
                model_path = alt_path
            else:
                raise FileNotFoundError(f"Файл модели не найден: {model_path}")

        self.model = joblib.load(model_path)
        logger.info("Модель колонны загружена успешно (tower_weight_model)")

    def predict(self, features_dict: dict) -> float:
        # собираем вектор параметров в нужном порядке
        raw_features = []
        for col in TOWER_FEATURE_COLUMNS:
            value = features_dict.get(col)
            if value is None:
                raise ValueError(f"Отсутствует обязательный параметр для модели: '{col}'")
            raw_features.append(float(value))

        import pandas as pd
        # модель была обучена на DataFrame с именами колонок, поэтому создаем DataFrame
        X = pd.DataFrame([raw_features], columns=TOWER_FEATURE_COLUMNS)

        # предсказание
        weight_kg = self.model.predict(X)[0]

        logger.info(f"Предсказание (колонна): weight_kg={weight_kg:.2f}")
        return float(weight_kg)
