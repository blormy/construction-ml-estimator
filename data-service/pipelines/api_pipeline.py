# pipelines/api_pipeline.py
import pandas as pd
import logging
from pathlib import Path
import sys

BASE_DIR = Path(__file__).resolve().parent.parent
sys.path.append(str(BASE_DIR))

from utils.cleaners import vectorized_numeric_clean
from domain.pump_features import PumpFeatureEngineer
from utils.enricher import PumpEnricher
from domain.vessel_features import VesselFeatureEngineer
from domain.conveyor_features import ConveyorFeatureEngineer
from domain.drum_features import DrumFeatureEngineer

logger = logging.getLogger(__name__)


class EquipmentAPIService:
    """Легковесный сервис для обработки одиночных запросов из API"""

    def __init__(self, output_folder_path: Path, config: dict):
        self.output_folder = output_folder_path
        self.config = config
        # в будущем здесь можно сделать фабрику под нужный тип оборудования
        self.feature_engineer = PumpFeatureEngineer(config)

    def process_request(self, input_dict: dict) -> dict:
        """Главный метод для API"""
        logger.info("Обработка запроса из API для насосов")

        # 1. из словаря в DataFrame
        df = pd.DataFrame([input_dict])

        # 2. валидация
        critical_cols = self.config['critical_cols']
        df = self.feature_engineer.filter_critical_data(df, critical_cols)
        if df.empty:
            raise ValueError("Данные не прошли валидацию: отсутствуют подача или напор")

        # 3. очистка
        for col in self.config['cols_to_clean']:
            if col in df.columns:
                df[col] = vectorized_numeric_clean(df[col])
                df[col] = pd.to_numeric(df[col], errors='coerce')

        # 4. обогащение
        try:
            base_path = self.output_folder / 'dataset_ml.csv'
            if base_path.exists():
                df_base = pd.read_csv(base_path)
                enricher = PumpEnricher(
                    search_features=self.config['search_features'],
                    target_features=self.config['target_features']
                )
                df = enricher.enrich(df, df_base)
            else:
                logger.warning("Эталонный датасет не найден, пропуск шага обогащения.")
        except Exception as e:
            logger.error(f"Ошибка при обогащении данных: {e}")

        # 5. физика
        df = self.feature_engineer.add_physics_features(df, is_inference=True)

        # 6. возвращаем чистый словарь
        return df.to_dict(orient='records')[0]

class ConveyorAPIService:
    """Легковесный сервис для обработки одиночных запросов из API для конвейеров"""

    def __init__(self, output_folder_path: Path, config: dict):
        self.output_folder = output_folder_path
        self.config = config

    def process_request(self, input_dict: dict) -> dict:
        logger.info("Обработка запроса из API для конвейеров")

        import numpy as np

        # 1. из словаря в DataFrame
        df = pd.DataFrame([input_dict])

        # 2. валидация критических полей
        critical_cols = ['length_ft', 'width_in']
        for col in critical_cols:
            if col not in df.columns or pd.isna(df[col].iloc[0]):
                raise ValueError(f"Отсутствует обязательный параметр: {col}")

        # 3. приведение типов
        for col in ['length_ft', 'width_in', 'flow_tph']:
            if col in df.columns:
                df[col] = pd.to_numeric(df[col], errors='coerce')

        # 4. заполнение пропусков в производительности
        if 'flow_tph' not in df.columns or pd.isna(df['flow_tph'].iloc[0]):
            df['flow_tph'] = 0.0

        # 5. генерация инженерных признаков
        df['belt_surface_area'] = df['length_ft'] * df['width_in']
        df['load_proxy'] = df['flow_tph'] / (df['width_in'] + 1)

        # 6. логарифмирование
        cols_to_log = ['length_ft', 'width_in', 'flow_tph', 'belt_surface_area', 'load_proxy']
        for col in cols_to_log:
            df[f'log_{col}'] = np.log1p(df[col])

        # 7. возвращаем чистый словарь
        result = df.to_dict(orient='records')[0]
        logger.info(f"Обработанные признаки конвейера: {result}")
        return result

class VesselAPIService:
    """Легковесный сервис для обработки одиночных запросов из API для сосудов (Vessel)"""

    def __init__(self, output_folder_path: Path, config: dict):
        self.output_folder = output_folder_path
        self.config = config
        self.feature_engineer = VesselFeatureEngineer(config)

    def process_request(self, input_dict: dict) -> dict:
        logger.info("Обработка запроса из API для сосудов")

        # 1. из словаря в DataFrame
        df = pd.DataFrame([input_dict])

        # 2. очистка
        cols_to_clean = self.config.get('cols_to_clean', ['diameter', 'ss_distance', 'pressure', 'sk_height', 'leg_height', 'temp'])
        for col in cols_to_clean:
            if col in df.columns:
                df[col] = vectorized_numeric_clean(df[col])
                df[col] = pd.to_numeric(df[col], errors='coerce')

        # 3. валидация
        critical_cols = self.config.get('critical_cols', ['diameter', 'ss_distance'])
        df = self.feature_engineer.filter_critical_data(df, critical_cols)
        if df.empty:
            raise ValueError("Данные не прошли валидацию: отсутствуют диаметр или высота")

        # 4. физика
        df = self.feature_engineer.add_physics_features(df, is_inference=True)

        return df.to_dict(orient='records')[0]

class DrumAPIService:
    """Легковесный сервис для обработки одиночных запросов из API для горизонтальных емкостей (Drum)"""

    def __init__(self, output_folder_path: Path, config: dict):
        self.output_folder = output_folder_path
        self.config = config

    def process_request(self, input_dict: dict) -> dict:
        logger.info("Обработка запроса из API для горизонтальных емкостей")

        import numpy as np

        # 1. из словаря в DataFrame
        df = pd.DataFrame([input_dict])

        # 2. валидация критических полей
        critical_cols = ['ves_diameter', 'ss_distance']
        for col in critical_cols:
            if col not in df.columns or pd.isna(df[col].iloc[0]):
                raise ValueError(f"Отсутствует обязательный параметр: {col}")

        # 3. приведение типов
        for col in ['ves_diameter', 'ss_distance', 'gauge_pres']:
            if col in df.columns:
                df[col] = pd.to_numeric(df[col], errors='coerce')

        # 4. заполнение пропусков давления
        if 'gauge_pres' not in df.columns or pd.isna(df['gauge_pres'].iloc[0]):
            df['gauge_pres'] = 0.0

        # 5. генерация инженерных признаков
        # объем по геометрии
        df['liq_volume'] = np.pi * (df['ves_diameter'] / 2) ** 2 * df['ss_distance']

        # прокси объема
        df['volume_proxy'] = (df['ves_diameter'] ** 2) * df['ss_distance']

        # площадь поверхности обечайки
        df['surface_area'] = df['ves_diameter'] * df['ss_distance']

        # расчет отношения длины к диаметру
        df['aspect_ratio'] = np.where(
            df['ves_diameter'] == 0,
            0,
            df['ss_distance'] / df['ves_diameter']
        )

        # 6. логарифмирование
        cols_to_log = ['liq_volume', 'ves_diameter', 'ss_distance', 'gauge_pres', 'volume_proxy', 'surface_area']
        for col in cols_to_log:
            df[f'log_{col}'] = np.log1p(df[col].clip(lower=0))

        # 7. возвращаем чистый словарь
        result = df.to_dict(orient='records')[0]
        logger.info(f"Обработанные признаки емкости: {result}")
        return result

class UTubeAPIService:
    """Легковесный сервис для обработки одиночных запросов из API для теплообменников (U-Tube)"""

    def __init__(self, output_folder_path: Path, config: dict):
        self.output_folder = output_folder_path
        self.config = config

    def process_request(self, input_dict: dict) -> dict:
        logger.info("Обработка запроса из API для теплообменников")

        import numpy as np

        # 1. из словаря в DataFrame
        df = pd.DataFrame([input_dict])

        # 2. валидация критических полей
        critical_cols = ['shell_diameter', 'tube_out_diameter', 'tube_len']
        for col in critical_cols:
            if col not in df.columns or pd.isna(df[col].iloc[0]):
                raise ValueError(f"Отсутствует обязательный параметр: {col}")

        # 3. приведение типов
        for col in ['heat_area', 'shell_diameter', 'tube_out_diameter', 'tube_des_pres', 'tube_len']:
            if col in df.columns:
                df[col] = pd.to_numeric(df[col], errors='coerce')

        # 4. заполнение пропусков
        if 'tube_des_pres' not in df.columns or pd.isna(df['tube_des_pres'].iloc[0]):
            df['tube_des_pres'] = 0.0

        if 'heat_area' not in df.columns or pd.isna(df['heat_area'].iloc[0]):
            # грубая оценка площади, если не задана
            df['heat_area'] = 0.0

        # 5. возвращаем чистый словарь
        result = df.to_dict(orient='records')[0]
        logger.info(f"Обработанные признаки теплообменника: {result}")
        return result
