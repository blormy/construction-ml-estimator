# pipelines/vessel_training_pipeline.py
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

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class VesselTrainingPipeline(BaseETLPipeline):
    """Пайплайн подготовки данных для обучения моделей сосудов (Vessel)"""
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.feature_engineer = VesselFeatureEngineer(self.config)

    def transform(self, df_features: pd.DataFrame, df_weight: pd.DataFrame) -> pd.DataFrame:
        raw = self.config['raw_names']
        logger.info("Начало трансформации данных сосудов (обучение)")

        # 1. объединение и удаление дублирующего тега
        df_merge = pd.merge(
            df_features, df_weight,
            left_on=raw['tag'], right_on=raw['tag_weight'], how='inner'
        ).drop(raw['tag_weight'], axis=1, errors='ignore')

        # 2. переименование колонок
        df_merge = df_merge.rename(columns=self.get_rename_map())

        # 3. очистка чисел
        for col in self.config['cols_to_convert']:
            df_merge[col] = vectorized_numeric_clean(df_merge[col])
            df_merge[col] = pd.to_numeric(df_merge[col], errors='coerce')

        # 4. валидация
        critical = [raw['diameter'], raw['ss_dist']]
        df_merge = self.feature_engineer.filter_critical_data(df_merge, critical)

        # 5. физика
        df_merge = self.feature_engineer.add_vessel_features(df_merge)

        logger.info("Трансформация данных сосудов (обучение) завершена")
        return df_merge

    def load(self, df: pd.DataFrame):
        """Сохранение датасета для обучения"""
        super().load(df, filename='vessel_dataset_ml.csv')

if __name__ == '__main__':
    vessel_ml_config = config['equipment']['vessel_ml']
    pipeline = VesselTrainingPipeline(
        input_file_path=BASE_DIR / 'data' / 'dataset_v_vessel_ml.xlsx',
        output_folder_path=BASE_DIR / 'datasets',
        config=vessel_ml_config
    )
    pipeline.run()