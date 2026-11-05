# pipelines/vessel_inference_pipeline.py
import pandas as pd
import logging
from pathlib import Path
import sys

BASE_DIR = Path(__file__).resolve().parent.parent
sys.path.append(str(BASE_DIR))

from configs.config_loader import config
from pipelines.base_etl import BaseETLPipeline
from domain.vessel_features import VesselFeatureEngineer
from utils.cleaners import vectorized_numeric_clean

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)

logger = logging.getLogger(__name__)

class VesselInferencePipeline(BaseETLPipeline):
    """Пайплайн инференса для вертикальных сосудов (Vessel)"""
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.feature_engineer = VesselFeatureEngineer(self.config)

    def transform(self, df_features: pd.DataFrame, df_weight: pd.DataFrame = None) -> pd.DataFrame:
        raw = self.config['raw_names']
        logger.info("Начало трансформации данных сосудов (инференс)")

        # 1. если есть веса, объединяем таблицы
        if df_weight is not None and not df_weight.empty:
            df_merge = pd.merge(
                df_features, df_weight,
                left_on=raw['tag'], right_on=raw['weight_tag'], how='outer'
            )
            df_merge = df_merge.dropna(subset=[raw['tag']])
            df_merge[raw['tag']] = df_merge[raw['tag']].fillna(df_merge[raw['weight_tag']])
            df_merge = df_merge.drop(columns=[raw['weight_tag']] + self.config.get('weight_col_to_drop', []), errors='ignore')
        else:
            df_merge = df_features.copy()

        # 2. переименование
        df_merge = df_merge.rename(columns=self.get_rename_map())

        # 3. очистка чисел
        for col in self.config.get('cols_to_convert', []):
            if col in df_merge.columns:
                df_merge[col] = vectorized_numeric_clean(df_merge[col])
                df_merge[col] = pd.to_numeric(df_merge[col], errors='coerce')

        # 4. фильтрация
        # для сосудов важен диаметр и высота (ss_distance)
        critical_cols = [self.config['col_names']['diameter'], self.config['col_names']['ss_distance']]
        df_merge = self.feature_engineer.filter_critical_data(df_merge, critical_cols)

        # 5. обогащение (при необходимости)

        # 6. физика
        df_merge = self.feature_engineer.add_physics_features(df_merge, is_inference=True)

        logger.info("Трансформация данных сосудов (инференс) завершена")
        return df_merge

    def load(self, df: pd.DataFrame, filename: str = 'vessel_dataset_inference.csv') -> None:
        """Сохранение датасета инференса"""
        super().load(df, filename=filename)

if __name__ == '__main__':
    vessel_inf_config = config['equipment']['vessel_inference']
    pipeline = VesselInferencePipeline(
        input_file_path=BASE_DIR / 'data' / 'Vessel_Data.xlsx', # пример файла
        output_folder_path=BASE_DIR / 'datasets',
        config=vessel_inf_config
    )
    pipeline.run()