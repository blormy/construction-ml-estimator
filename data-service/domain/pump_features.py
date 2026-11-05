# domain/pump_features.py
import pandas as pd
import numpy as np
import logging

logger = logging.getLogger(__name__)


class PumpFeatureEngineer:
    """Доменная логика для насосов"""

    def __init__(self, config: dict):
        self.config = config

    def filter_critical_data(self, df: pd.DataFrame, critical_cols: list) -> pd.DataFrame:
        """Удаление строк, в которых отсутствуют критически важные параметры."""
        before_drop = len(df)
        df = df.dropna(subset=critical_cols, how='any').reset_index(drop=True)
        dropped = before_drop - len(df)

        if dropped > 0:
            logger.info(f"Валидация: удалено {dropped} строк без обязательных параметров")
        return df

    def add_physics_features(self, df_merge: pd.DataFrame, is_inference: bool = False) -> pd.DataFrame:
        """Генерация физических признаков"""
        c_eff = self.config['col_names'].get('pump_eff', 'pump_eff')
        c_head = self.config['col_names']['head']
        c_flow = self.config['col_names']['flow']
        c_weight = self.config['col_names']['weight_kg']
        c_speed = self.config['col_names']['rpm']

        c_useful = self.config['feat_eng']['useful_kw']
        c_diameter = self.config['feat_eng']['diameter_proxy']
        c_weight_log = self.config['feat_eng']['weight_log']

        s_head = df_merge[c_head].astype(float)
        s_speed = df_merge[c_speed].astype(float)

        logger.info("Расчет инженерных признаков для насосов...")

        # вводим полезную мощность (киловатт)
        df_merge[c_useful] = df_merge[c_head] * df_merge[c_flow]

        # геометрический фактор (косвенная оценка диаметра рабочего колеса)
        df_merge[c_diameter] = np.sqrt(s_head) / (s_speed + 1e-6)

        # проверяем безопасность логарифмирования полезной мощности
        invalid_power = df_merge[c_useful] < 0
        if invalid_power.sum() > 0:
            logger.warning(f"Найдено {invalid_power.sum()} строк с отрицательной мощностью, исключаем")
            df_merge = df_merge[~invalid_power]

        cols_to_exclude = [c_weight_log, c_diameter]

        # логарифмирование новых признаков
        new_physics_cols = [col for col in self.config['feat_eng'] if col not in cols_to_exclude]
        for col in new_physics_cols:
            df_merge[f'{col}_log'] = np.log1p(df_merge[col].replace([np.inf, -np.inf], np.nan).fillna(0))
            df_merge = df_merge.drop(columns=[col])

        # безопасная обработка целевой переменной
        if c_weight in df_merge.columns:
            if not is_inference:
                # режим обучения
                invalid_weight = (df_merge[c_weight] <= 0) | df_merge[c_weight].isna()
                if invalid_weight.sum() > 0:
                    logger.warning(f"Удалено {invalid_weight.sum()} строк без целевой переменной (веса)")
                    df_merge = df_merge[~invalid_weight]
                df_merge[c_weight_log] = np.log(df_merge[c_weight])
            else:
                # режим инференса
                invalid_input = df_merge[c_weight] <= 0
                if invalid_input.sum() > 0:
                    df_merge.loc[invalid_input, c_weight] = np.nan
                df_merge[c_weight_log] = np.log(df_merge[c_weight].astype(float))

            # удаляем исходные колонки и ставим weight_log в конец
            df_merge = df_merge.drop(columns=[c_weight, c_eff], errors='ignore')
            cols = df_merge.columns.tolist()
            if c_weight_log in cols:
                cols.remove(c_weight_log)
                cols.append(c_weight_log)
                df_merge = df_merge[cols]
        else:
            logger.info("Колонка веса отсутствует в датасете, шаг пропущен")

        logger.info("Расчет инженерных признаков для насосов завершен")
        return df_merge