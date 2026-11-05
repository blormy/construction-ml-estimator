# pipelines/pump_training_pipeline.py
import pandas as pd
import logging
from pathlib import Path
import sys

BASE_DIR = Path(__file__).resolve().parent.parent
sys.path.append(str(BASE_DIR))

from configs.config_loader import config
from pipelines.base_etl import BaseETLPipeline
from domain.pump_features import PumpFeatureEngineer
from utils.cleaners import vectorized_numeric_clean

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)

logger = logging.getLogger(__name__)

class PumpTrainingPipeline(BaseETLPipeline):
    """Пайплайн подготовки данных для обучения моделей насосов"""
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        # подключаем доменную логику
        self.feature_engineer = PumpFeatureEngineer(self.config)

    def transform(self, df_features: pd.DataFrame, df_weight: pd.DataFrame) -> pd.DataFrame:
        raw = self.config['raw_names']
        logger.info("Начало трансформации данных насосов (обучение)")

        # 1. объединение (inner join)
        df_merge = pd.merge(
            df_features, df_weight,
            left_on=raw['tag'], right_on=raw['tag_weight'], how='inner'
        ).drop('Tag No', axis=1, errors='ignore')

        # 2. переименование колонок
        df_merge = df_merge.rename(columns=self.get_rename_map())

        # 3. очистка чисел
        for col in self.config['cols_to_convert']:
            df_merge[col] = vectorized_numeric_clean(df_merge[col])
            df_merge[col] = pd.to_numeric(df_merge[col], errors='coerce')

        # 4. удаление пропусков и валидация
        df_merge = df_merge.dropna(subset=self.config['cols_to_convert']).reset_index(drop=True)
        critical_cols = [raw['flow'], raw['head']]
        df_merge = self.feature_engineer.filter_critical_data(df_merge, critical_cols)

        # 5. физика
        df_merge = self.feature_engineer.add_physics_features(df_merge, is_inference=False)

        logger.info("Трансформация данных насосов (обучение) завершена")
        return df_merge

    def load(self, df: pd.DataFrame):
        """Сохранение датасета для обучения"""
        super().load(df, filename='pump_dataset_ml.csv')

if __name__ == '__main__':
    pump_ml_config = config['equipment']['pump_ml']
    pipeline = PumpTrainingPipeline(
        input_file_path=BASE_DIR / 'data' / 'Data.xlsx',
        output_folder_path=BASE_DIR / 'datasets',
        config=pump_ml_config
    )
    pipeline.run()