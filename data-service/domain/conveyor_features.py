# domain/conveyor_features.py
import pandas as pd
import numpy as np
import logging

logger = logging.getLogger(__name__)


class ConveyorFeatureEngineer:
    """Доменная логика для ленточных конвейеров"""

    def __init__(self, config: dict):
        self.config = config

    def filter_critical_data(self, df: pd.DataFrame, critical_cols: list) -> pd.DataFrame:
        """Удаление строк без базовой геометрии"""
        return df.dropna(subset=critical_cols, how='any').reset_index(drop=True)

    def add_conveyor_features(self, df: pd.DataFrame) -> pd.DataFrame:
        """Генерация инженерных признаков для конвейера"""
        c_len = self.config['col_names']['length']
        c_width = self.config['col_names']['width']
        c_weight = self.config['col_names']['weight_kg']
        c_weight_log = self.config['feat_eng']['weight_log']

        logger.info("Расчет инженерных признаков для конвейера...")

        # расчет условной площади ленты
        df[self.config['feat_eng']['belt_area']] = df[c_len] * df[c_width]

        # логарифмирование целевой переменной
        if c_weight in df.columns:
            df[c_weight_log] = np.log(df[c_weight].astype(float).replace(0, np.nan))
            df = df.drop(columns=[c_weight])

        logger.info("Расчет инженерных признаков конвейера завершен")

        return df