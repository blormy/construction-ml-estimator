# pipelines/pump_inference_pipeline.py
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
from utils.enricher import PumpEnricher

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)

logger = logging.getLogger(__name__)

class PumpInferencePipeline(BaseETLPipeline):
    """Пайплайн для очистки и обогащения реальных данных"""
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.feature_engineer = PumpFeatureEngineer(self.config)

    def transform(self, df_features: pd.DataFrame, df_weight: pd.DataFrame) -> pd.DataFrame:
        raw = self.config['raw_names']
        logger.info("Начало трансформации данных насосов (инференс)")

        # 1. объединение (outer join)
        logger.warning("Теги в файлах различаются")
        df_merge = pd.merge(
            df_features, df_weight,
            left_on=raw['tag'], right_on=raw['weight_tag'], how='outer'
        )

        df_merge = df_merge.dropna(subset=[raw['tag']])
        df_merge[raw['tag']] = df_merge[raw['tag']].fillna(df_merge[raw['weight_tag']])
        df_merge = df_merge.drop(columns=[raw['weight_tag']] + self.config['weight_col_to_drop'], errors='ignore')

        # 2. переименование
        df_merge = df_merge.rename(columns=self.get_rename_map())

        # 3. очистка чисел
        for col in self.config['cols_to_convert']:
            if col in df_merge.columns:
                df_merge[col] = vectorized_numeric_clean(df_merge[col])
                df_merge[col] = pd.to_numeric(df_merge[col], errors='coerce')

        # 4. фильтрация
        critical_cols = [raw['flow'], raw['head']]
        df_merge = self.feature_engineer.filter_critical_data(df_merge, critical_cols)

        # 5. обогащение данных
        try:
            base_path = self.output_folder / 'dataset_ml.csv'
            if base_path.exists():
                df_base = pd.read_csv(base_path)
                enricher = PumpEnricher(
                    search_features=['flow_rate', 'fluid_head'],
                    target_features=['rpm', 'spec_gravity', 'power_kw']
                )
                df_merge = enricher.enrich(df_merge, df_base)
            else:
                logger.warning("Эталонный датасет не найден, пропуск шага обогащения")
        except Exception as e:
            logger.error(f"Ошибка при обогащении данных: {e}")

        # 6. физика
        df_merge = self.feature_engineer.add_physics_features(df_merge, is_inference=True)

        logger.info("Трансформация данных насосов (инференс) завершена")
        return df_merge

    def load(self, df: pd.DataFrame, filename: str = 'pump_dataset_inference.csv') -> None:
        """Сохранение датасета инференса"""
        super().load(df, filename=filename)

if __name__ == '__main__':
    pump_inf_config = config['equipment']['pump_inference']
    pipeline = PumpInferencePipeline(
        input_file_path=BASE_DIR / 'data' / 'Реальные насосные.xlsx',
        output_folder_path=BASE_DIR / 'datasets',
        config=pump_inf_config,
        weight_file_path=BASE_DIR / 'data' / 'Реалные_насосные_альтернативный_лист_с_весом.xlsx'
    )
    pipeline.run()
