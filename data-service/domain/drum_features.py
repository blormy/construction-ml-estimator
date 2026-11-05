import pandas as pd
import numpy as np
import logging

logger = logging.getLogger(__name__)


class DrumFeatureEngineer:
    """Доменная логика для горизонтальных емкостей (Drum)"""

    def __init__(self, config: dict):
        self.config = config

    def filter_critical_data(self, df: pd.DataFrame, critical_cols: list) -> pd.DataFrame:
        """Удаление строк без базовой геометрии"""
        return df.dropna(subset=critical_cols, how='any').reset_index(drop=True)

    def add_drum_features(self, df: pd.DataFrame) -> pd.DataFrame:
        """Генерация физических и геометрических признаков для емкости"""
        c_ves_diameter = self.config['col_names']['diameter']
        c_ss_distance = self.config['col_names']['ss_dist']

        logger.info("Расчет инженерных признаков для горизонтальной емкости...")

        # 1. прокси объема
        df['volume_proxy'] = (df[c_ves_diameter] ** 2) * df[c_ss_distance]

        # 2. прокси площади поверхности обечайки
        df['surface_area'] = df[c_ves_diameter] * df[c_ss_distance]

        # 3. прокси площади днищ
        df['heads_proxy'] = df[c_ves_diameter] ** 2

        # 4. отношение длины к диаметру
        df['aspect_ratio'] = np.where(
            df[c_ves_diameter] == 0,
            0,
            df[c_ss_distance] / df[c_ves_diameter]
        )

        # переносим колонку веса в конец
        if 'weight_kg' in df.columns:
            df['weight_kg'] = df.pop('weight_kg')

        logger.info("Расчет инженерных признаков емкости завершен")

        return df