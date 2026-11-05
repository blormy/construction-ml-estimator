# utils/enricher.py
import pandas as pd
import numpy as np
import logging
from scipy.spatial.distance import cdist
from scipy import stats

logger = logging.getLogger(__name__)


class PumpEnricher:
    """Обогащение данных через поиск k-ближайших соседей (k-NN)"""

    def __init__(self, search_features: list, target_features: list, k_neighbors: int = 3):
        self.search_features = search_features
        self.target_features = target_features
        self.k = k_neighbors

    def enrich(self, df_real: pd.DataFrame, df_base: pd.DataFrame) -> pd.DataFrame:
        logger.info(f"Запуск обогащения данных (k={self.k})")
        df_result = df_real.copy()

        # 1. подготовка эталона
        base_clean = df_base.dropna(subset=self.search_features).copy()
        base_search_vals = base_clean[self.search_features].values

        # нормализация (z-score)
        means = base_search_vals.mean(axis=0)
        stds = base_search_vals.std(axis=0) + 1e-9
        base_norm = (base_search_vals - means) / stds

        # 2. подготовка реальных данных
        real_search_vals = df_result[self.search_features].values
        mask_valid = ~np.any(np.isnan(real_search_vals), axis=1)

        if not np.any(mask_valid):
            return df_result

        real_norm = (real_search_vals[mask_valid] - means) / stds

        # 3. векторный поиск k соседей
        distances = cdist(real_norm, base_norm, metric='euclidean')

        # находим индексы k ближайших соседей для каждой строки
        neighbor_indices = np.argpartition(distances, self.k, axis=1)[:, :self.k]

        # 4. заполнение пропусков
        valid_row_idx = 0
        for i in range(len(df_result)):
            if mask_valid[i]:
                # индексы соседей в базе для текущей строки
                current_neighbors_idx = neighbor_indices[valid_row_idx]
                neighbors_df = base_clean.iloc[current_neighbors_idx]

                for col in self.target_features:
                    if pd.isna(df_result.at[i, col]):
                        # логика: категории — по моде, числа — по среднему
                        if col == 'spec_gravity':
                            # берем самое частое значение (моду)
                            val = neighbors_df[col].mode().iloc[0]
                        else:
                            # берем среднее значение
                            val = neighbors_df[col].mean()

                        df_result.at[i, col] = val

                valid_row_idx += 1

        logger.info(f"Обогащение данных завершено (k={self.k})")
        return df_result