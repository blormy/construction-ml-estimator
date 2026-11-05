# pipelines/conveyor_training_pipeline.py
import pandas as pd
import logging
from pathlib import Path
import sys

BASE_DIR = Path(__file__).resolve().parent.parent
sys.path.append(str(BASE_DIR))

from configs.config_loader import config
from pipelines.base_etl import BaseETLPipeline
from domain.conveyor_features import ConveyorFeatureEngineer
from utils.cleaners import vectorized_numeric_clean

# настройка логгера
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)

class ConveyorTrainingPipeline(BaseETLPipeline):
    """Пайплайн подготовки данных для обучения моделей конвейеров (Conveyor)"""
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.feature_engineer = ConveyorFeatureEngineer(self.config)

    def transform(self, df_features: pd.DataFrame, df_weight: pd.DataFrame) -> pd.DataFrame:
        raw = self.config['raw_names']
        logger.info("Начало трансформации данных конвейера (обучение)")

        # 1. объединение (inner join)
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

        # 4. удаление строк с пропусками
        df_merge = df_merge.dropna(subset=self.config['cols_to_convert']).reset_index(drop=True)

        # 5. физика
        df_merge = self.feature_engineer.add_conveyor_features(df_merge)

        logger.info("Трансформация данных конвейера (обучение) завершена")
        return df_merge

    def load(self, df: pd.DataFrame):
        """Сохранение датасета для обучения"""
        super().load(df, filename='conveyor_dataset_ml.csv')

if __name__ == '__main__':
    conveyor_ml_config = config['equipment']['conveyor_ml']
    pipeline = ConveyorTrainingPipeline(
        input_file_path=BASE_DIR / 'data' / 'Belt conveyor-open.xlsx',
        output_folder_path=BASE_DIR / 'datasets',
        config=conveyor_ml_config
    )
    pipeline.run()