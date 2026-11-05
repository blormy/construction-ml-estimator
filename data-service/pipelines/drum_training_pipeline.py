# pipelines/drum_training_pipeline.py
import pandas as pd
import logging
from pathlib import Path
import sys

BASE_DIR = Path(__file__).resolve().parent.parent
sys.path.append(str(BASE_DIR))

from configs.config_loader import config
from pipelines.base_etl import BaseETLPipeline
from domain.drum_features import DrumFeatureEngineer
from utils.cleaners import vectorized_numeric_clean

# настройка логгера
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)

class DrumInferencePipeline(BaseETLPipeline):
    """Пайплайн подготовки данных для горизонтальных емкостей (Drum)"""
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.feature_engineer = DrumFeatureEngineer(self.config)

    def transform(self, df_features: pd.DataFrame, df_weight: None) -> pd.DataFrame:
        raw = self.config['raw_names']
        logger.info("Начало трансформации данных емкостей (обучение)")

        # 1. считываем данные листа
        df_merge = df_features.copy()

        # 2. переименование колонок
        df_merge = df_merge.rename(columns=self.get_rename_map())

        # 3. очистка чисел
        cols_to_convert = self.config.get('cols_to_convert', [])
        for col in cols_to_convert:
            if col in df_merge.columns:
                df_merge[col] = vectorized_numeric_clean(df_merge[col])
                df_merge[col] = pd.to_numeric(df_merge[col], errors='coerce')
            else:
                logger.debug(f"Пропуск очистки: {col} отсутствует.")

        logger.info(f"Колонки датасета: {df_merge.columns.tolist()}")

        # 4. генерация инженерных признаков
        df_merge = self.feature_engineer.add_drum_features(df_merge)

        logger.info("Трансформация данных емкостей (обучение) завершена")
        return df_merge.iloc[1:]

    def load(self, df: pd.DataFrame):
        """Сохранение датасета для обучения"""
        super().load(df, filename='drum_dataset_ml.csv')

if __name__ == "__main__":
    drum_inference_config = config['equipment']['drum_ml']
    pipeline = DrumInferencePipeline(
        input_file_path=BASE_DIR / 'data' / 'Horizontal drum.xlsx',
        output_folder_path=BASE_DIR / 'datasets',
        config=drum_inference_config
    )
    pipeline.run()