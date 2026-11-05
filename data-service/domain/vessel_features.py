# domain/vessel_features.py
import pandas as pd
import numpy as np
import logging

logger = logging.getLogger(__name__)


class VesselFeatureEngineer:
    """Доменная логика для сосудов (vessels)"""

    def __init__(self, config: dict):
        self.config = config

    def filter_critical_data(self, df: pd.DataFrame, critical_cols: list) -> pd.DataFrame:
        """Удаление строк, в которых отсутствуют критически важные параметры."""
        before_drop = len(df)
        df = df.dropna(subset=critical_cols, how='any').reset_index(drop=True)
        dropped = before_drop - len(df)

        if dropped > 0:
            logger.info(f"Валидация: удалено {dropped} строк из-за отсутствия критических параметров.")
        return df

    def add_physics_features(self, df: pd.DataFrame, is_inference: bool = False) -> pd.DataFrame:
        """Генерация физических признаков для сосудов"""
        logger.info("Расчет инженерных признаков для сосуда...")
        
        # получаем названия колонок из конфига или используем значения по умолчанию
        col_names = self.config.get('col_names', {})
        c_diameter = col_names.get('diameter', 'diameter')
        c_ss_dist = col_names.get('ss_distance', 'ss_distance')
        c_pressure = col_names.get('pressure', 'pressure')
        c_sk_height = col_names.get('sk_height', 'sk_height')
        c_leg_height = col_names.get('leg_height', 'leg_height')
        c_weight = col_names.get('weight_kg', 'weight_kg')
        c_temp = col_names.get('temp', 'temp')

        df_out = df.copy()

        # приводим типы данных
        for col in [c_diameter, c_ss_dist, c_pressure, c_sk_height, c_leg_height, c_temp]:
            if col in df_out.columns:
                df_out[col] = pd.to_numeric(df_out[col], errors='coerce')
            else:
                if col in [c_sk_height, c_leg_height]:
                    df_out[col] = np.nan
                else:
                    df_out[col] = 0

        # 1. коррекция давления (МПа -> кПа)
        mask_mpa = (df_out[c_pressure] > 0) & (df_out[c_pressure] < 20)
        df_out.loc[mask_mpa, c_pressure] *= 1000
        
        # 2. абсолютное давление и заполнение пропусков
        df_out['p_abs'] = df_out[c_pressure].fillna(0) + 101.3
        df_out.loc[df_out['p_abs'] < 50, 'p_abs'] = 101.3
        
        # 3. восстановление объема по геометрии
        if 'liq_volume' not in df_out.columns:
            df_out['liq_volume'] = np.nan
            
        mask_vol = df_out['liq_volume'].isna() | (df_out['liq_volume'] <= 0)
        df_out.loc[mask_vol, 'liq_volume'] = (np.pi * (df_out[c_diameter]**2) / 4) * df_out[c_ss_dist]

        # 4. новые физические признаки
        # площадь поверхности (стенки + днища)
        df_out['area_calc'] = (np.pi * df_out[c_diameter] * df_out[c_ss_dist]) + (1.5 * np.pi * (df_out[c_diameter]**2) / 4)
        # прокси толщины стенки
        df_out['thick_proxy'] = df_out['p_abs'] * df_out[c_diameter]
        
        # 5. опоры (бинарные признаки)
        df_out['has_skirt'] = df_out[c_sk_height].notna().astype(int)
        df_out['has_legs'] = df_out[c_leg_height].notna().astype(int)
        df_out[c_sk_height] = df_out[c_sk_height].fillna(0)
        df_out[c_leg_height] = df_out[c_leg_height].fillna(0)
        
        # маппинг температуры
        df_out['des_temp'] = df_out[c_temp].fillna(0)

        # 6. логарифмирование сильно смещенных величин
        for col in ['liq_volume', c_diameter, c_ss_dist, 'p_abs', 'area_calc', 'thick_proxy']:
            df_out[f'log_{col}'] = np.log1p(df_out[col])

        # обработка целевой переменной (только для режима обучения)
        if c_weight in df_out.columns:
            if not is_inference:
                invalid_weight = (df_out[c_weight] <= 0) | df_out[c_weight].isna()
                if invalid_weight.sum() > 0:
                    logger.warning(f"Удалено {invalid_weight.sum()} строк без целевой переменной (веса)")
                    df_out = df_out[~invalid_weight]
                df_out['weight_kg_log'] = np.log(df_out[c_weight])
                # перевод веса из фунтов в кг для обучения
                df_out['weight_kg_log'] = df_out['weight_kg_log'] - np.log(2.20462)
            else:
                invalid_input = df_out[c_weight] <= 0
                if invalid_input.sum() > 0:
                    df_out.loc[invalid_input, c_weight] = np.nan
                df_out['weight_kg_log'] = np.log(df_out[c_weight].astype(float))
                
            df_out = df_out.drop(columns=[c_weight], errors='ignore')

        # оставляем только нужные колонки
        if f'log_{c_diameter}' != 'log_diameter':
            df_out['log_diameter'] = df_out[f'log_{c_diameter}']
        if f'log_{c_ss_dist}' != 'log_ss_distance':
            df_out['log_ss_distance'] = df_out[f'log_{c_ss_dist}']

        required_cols = [
            'log_liq_volume', 'log_diameter', 'log_ss_distance', 'log_p_abs', 
            'log_area_calc', 'log_thick_proxy', 'des_temp', 
            'sk_height', 'leg_height', 'has_skirt', 'has_legs'
        ]
        
        if 'tag' in df_out.columns:
            required_cols = ['tag'] + required_cols
            
        if 'weight_kg_log' in df_out.columns:
            required_cols.append('weight_kg_log')
            
        for col in required_cols:
            if col not in df_out.columns:
                df_out[col] = 0

        result_cols = [col for col in required_cols if col in df_out.columns]
        df_out = df_out[result_cols]

        logger.info("Расчет инженерных признаков для сосуда завершен")
        return df_out