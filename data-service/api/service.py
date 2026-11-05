from pathlib import Path
import logging
import sys

# импортируем наш новый легковесный сервис вместо тяжелого пайплайна
from pipelines.api_pipeline import EquipmentAPIService, VesselAPIService, ConveyorAPIService, DrumAPIService, UTubeAPIService
from domain.box_furnace_features import BoxFurnaceFeatureEngineer
from domain.centrifugal_compressor_features import CentrifugalCompressorFeatureEngineer
from configs.config_loader import config

# добавляем корень проекта в sys.path для импорта ml_service
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

from ml_service.predictor import PumpPredictor, VesselPredictor, ConveyorPredictor, DrumPredictor, UTubePredictor, TowerPredictor

logger = logging.getLogger(__name__)

# определяем пути
BASE_DIR = Path(__file__).resolve().parent.parent
DATASETS_DIR = BASE_DIR / 'datasets'

# инициализируем сервис
pump_service = EquipmentAPIService(
    output_folder_path=DATASETS_DIR,
    config=config['api']
)

# инициализируем предиктор (загрузка модели один раз при старте)
pump_predictor = PumpPredictor()

# инициализируем сервис для сосудов
vessel_service = VesselAPIService(
    output_folder_path=DATASETS_DIR,
    config=config.get('equipment', {}).get('vessel_inference', {}) # Передаем конфиг для инференса
)

vessel_predictor = VesselPredictor()

# инициализируем сервис для конвейеров
conveyor_service = ConveyorAPIService(
    output_folder_path=DATASETS_DIR,
    config=config.get('equipment', {}).get('conveyor_inference', {})
)

conveyor_predictor = ConveyorPredictor()

# инициализируем сервис для горизонтальных емкостей
drum_service = DrumAPIService(
    output_folder_path=DATASETS_DIR,
    config=config.get('equipment', {}).get('drum_ml', {})
)

drum_predictor = DrumPredictor()

# инициализируем сервис для теплообменников
utube_service = UTubeAPIService(
    output_folder_path=DATASETS_DIR,
    config=config.get('equipment', {}).get('utube_ml', {})
)

utube_predictor = UTubePredictor()

# инициализируем предиктор для колонн
tower_predictor = TowerPredictor()

# инициализируем заглушку для коробчатых печей
box_furnace_engineer = BoxFurnaceFeatureEngineer()

# инициализируем заглушку для центробежных компрессоров
compressor_engineer = CentrifugalCompressorFeatureEngineer()

def get_pump_estimation(input_data: dict) -> dict:
    """Прослойка между API и расчетами для насосов"""
    try:
        # 1. прогоняем через пайплайн (очистка + энричинг + FE)
        processed_features = pump_service.process_request(input_data)

        # 2. вызов ML-модели
        predicted_weight = pump_predictor.predict(processed_features)

        return {
            "weight": round(float(predicted_weight), 2),
            "features": processed_features  # возвращаем названия колонок для отладки
        }
    except Exception as e:
        logger.error(f"Ошибка в сервисе оценки насоса: {e}")
        # прокидываем ошибку наверх, чтобы FastAPI мог вернуть понятный 400 статус
        raise ValueError(f"Ошибка обработки данных: {str(e)}")

def get_vessel_estimation(input_data: dict) -> dict:
    """Прослойка между API и расчетами для сосудов (Vessel)"""
    try:
        # маппинг полей из API в формат сервиса (если нужно)
        # У нас API отдает vessel_diameter, vessel_tangent_to_tangent_height и т.д.
        # преобразуем к названиям для сервиса
        mapped_data = {
            "tag": input_data.get("tag"),
            "diameter": input_data.get("vessel_diameter"),
            "ss_distance": input_data.get("vessel_tangent_to_tangent_height"),
            "pressure": input_data.get("design_gauge_pressure", 0),
            "sk_height": input_data.get("skirt_height", 0),
            "leg_height": input_data.get("vessel_leg_height", 0),
            "temp": input_data.get("design_temperature", 0)
        }

        # 1. прогоняем через пайплайн (очистка + FE)
        processed_features = vessel_service.process_request(mapped_data)

        # 2. вызов ML-модели
        predicted_weight = vessel_predictor.predict(processed_features)

        return {
            "weight": round(float(predicted_weight), 2),
            "features": processed_features  # возвращаем названия колонок для отладки
        }
    except Exception as e:
        logger.error(f"Ошибка в сервисе оценки сосуда: {e}")
        raise ValueError(f"Ошибка обработки данных: {str(e)}")

def get_conveyor_estimation(input_data: dict) -> dict:
    """Прослойка между API и расчетами для конвейеров (Conveyor)"""
    try:
        # маппинг полей из API в формат сервиса
        mapped_data = {
            "tag": input_data.get("tag"),
            "length_ft": input_data.get("conveyor_length"),
            "width_in": input_data.get("belt_width"),
            "flow_tph": input_data.get("conveyor_flow_rate", 0),
        }

        # 1. прогоняем через пайплайн (очистка + FE)
        processed_features = conveyor_service.process_request(mapped_data)

        # 2. вызов ML-модели
        predicted_weight = conveyor_predictor.predict(processed_features)

        return {
            "weight": round(float(predicted_weight), 2),
            "features": processed_features
        }
    except Exception as e:
        logger.error(f"Ошибка в сервисе оценки конвейера: {e}")
        raise ValueError(f"Ошибка обработки данных: {str(e)}")

def get_drum_estimation(input_data: dict) -> dict:
    """Прослойка между API и расчетами для горизонтальных емкостей (Drum)"""
    try:
        # маппинг полей из API в формат сервиса
        mapped_data = {
            "tag": input_data.get("tag"),
            "ves_diameter": input_data.get("vessel_diameter"),
            "ss_distance": input_data.get("design_tangent_to_tangent_length"),
            "gauge_pres": input_data.get("design_gauge_pressure", 0),
        }

        # 1. прогоняем через пайплайн (очистка + FE)
        processed_features = drum_service.process_request(mapped_data)

        # 2. вызов ML-модели
        predicted_weight = drum_predictor.predict(processed_features)

        return {
            "weight": round(float(predicted_weight), 2),
            "features": processed_features
        }
    except Exception as e:
        logger.error(f"Ошибка в сервисе оценки емкости: {e}")
        raise ValueError(f"Ошибка обработки данных: {str(e)}")

def get_utube_estimation(input_data: dict) -> dict:
    """Прослойка между API и расчетами для теплообменников (U-Tube)"""
    try:
        # маппинг полей из API в формат сервиса
        mapped_data = {
            "tag": input_data.get("tag"),
            "shell_diameter": input_data.get("shell_diameter"),
            "tube_out_diameter": input_data.get("tube_out_diameter"),
            "tube_len": input_data.get("tube_len"),
            "tube_des_pres": input_data.get("tube_des_pres", 0),
            "heat_area": input_data.get("heat_area", 0),
        }

        # 1. прогоняем через пайплайн
        processed_features = utube_service.process_request(mapped_data)

        # 2. вызов ML-модели
        predicted_weight = utube_predictor.predict(processed_features)

        return {
            "weight": round(float(predicted_weight), 2),
            "features": processed_features
        }
    except Exception as e:
        logger.error(f"Ошибка в сервисе оценки теплообменника: {e}")
        raise ValueError(f"Ошибка обработки данных: {str(e)}")

def get_tower_estimation(input_data: dict) -> dict:
    """Прослойка между API и расчетами для колонн (Tower) без использования пайплайна"""
    try:
        # маппинг полей из API в формат предиктора
        features = {
            "diameter": input_data.get("vessel_diameter", 0.0),
            "ss_dist": input_data.get("design_tangent_to_tangent_length", 0.0),
            "pressure": input_data.get("design_gauge_pressure", 0.0),
            "trays_num": input_data.get("number_of_trays", 0.0)
        }

        # так как опциональные параметры могут быть None, заменяем их на 0.0 для надежности
        for k, v in features.items():
            if v is None:
                features[k] = 0.0

        # вызов ML-модели напрямую (без pipeline)
        predicted_weight = tower_predictor.predict(features)

        return {
            "weight": round(float(predicted_weight), 2),
            "features": features
        }
    except Exception as e:
        logger.error(f"Ошибка в сервисе оценки колонны: {e}")
        raise ValueError(f"Ошибка обработки данных: {str(e)}")

def get_box_furnace_estimation(input_data: dict) -> dict:
    """Прослойка между API и заглушкой расчета печи (Box Furnace)"""
    try:
        duty = input_data.get("duty", 0.0) or 0.0
        gas_flow = input_data.get("standard_gas_flow_rate", 0.0) or 0.0

        predicted_weight = box_furnace_engineer.calculate_stub_weight(duty, gas_flow)

        return {
            "weight": round(float(predicted_weight), 2),
            "features": input_data
        }
    except Exception as e:
        logger.error(f"Ошибка в сервисе оценки печи: {e}")
        raise ValueError(f"Ошибка обработки данных: {str(e)}")

def get_centrifugal_compressor_estimation(input_data: dict) -> dict:
    """Прослойка между API и заглушкой расчета компрессора (Centrifugal Compressor)"""
    try:
        gas_flow = input_data.get("actual_gas_flow_rate_inlet", 0.0) or 0.0
        p_inlet = input_data.get("design_gauge_pressure_inlet", 0.0) or 0.0
        p_outlet = input_data.get("design_gauge_pressure_outlet", 0.0) or 0.0

        predicted_weight = compressor_engineer.calculate_stub_weight(gas_flow, p_inlet, p_outlet)

        return {
            "weight": round(float(predicted_weight), 2),
            "features": input_data
        }
    except Exception as e:
        logger.error(f"Ошибка в сервисе оценки компрессора: {e}")
        raise ValueError(f"Ошибка обработки данных: {str(e)}")
