# pipelines/conveyor_inference_pipeline.py
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
    """Пайплайн инференса для конвейеров (Conveyor)"""
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.feature_engineer = ConveyorFeatureEngineer(self.config)

    def transform(self, df_features: pd.DataFrame, df_weight: None) -> pd.DataFrame:
        raw = self.config['raw_names']
        logger.info("Начало трансформации данных конвейера (инференс)")

        # 1. считываем данные листа
        df_merge = df_features.copy()
        # выводим колонки для контроля скрытых пробелов
        logger.info(f"Колонки датасета: {df_merge.columns.tolist()}")

        # удаляем служебную колонку parent area
        df_merge = df_merge.drop(columns=[raw['parent_area']], errors='ignore')

        # 2. переименование
        df_merge = df_merge.rename(columns=self.get_rename_map())

        # очистка от дубликатов
        duplicated_cols = df_merge.columns[df_merge.columns.duplicated()].unique().tolist()
        if duplicated_cols:
            logger.warning(f"Найдены дублирующиеся колонки: {duplicated_cols}, оставляем первые вхождения")
            df_merge = df_merge.loc[:, ~df_merge.columns.duplicated()]

        # 3. очистка чисел
        cols_to_convert = self.config.get('cols_to_convert', [])
        for col in cols_to_convert:
            if col in df_merge.columns:
                df_merge[col] = vectorized_numeric_clean(df_merge[col])
                df_merge[col] = pd.to_numeric(df_merge[col], errors='coerce')
            else:
                logger.debug(f"Пропуск очистки: {col} отсутствует.")

        logger.info("Трансформация данных конвейера (инференс) завершена")
        return df_merge

    def load(self, df: pd.DataFrame):
        """Сохранение датасета инференса"""
        super().load(df, filename='conveyor_dataset_inference.csv')

if __name__ == '__main__':
    conveyor_inference_config = config['equipment']['conveyor_inference']
    pipeline = ConveyorTrainingPipeline(
        input_file_path=BASE_DIR / 'data' / 'Реальные HorizDrum и BeltConveyor.xlsx',
        output_folder_path=BASE_DIR / 'datasets',
        config=conveyor_inference_config
    )
    pipeline.run()