# utils/cleaners.py
import pandas as pd

def vectorized_numeric_clean(series: pd.Series) -> pd.Series:
    """Векторизованная очистка всей колонки данных"""
    # принудительно переводим в строку
    s = series.astype(str)
    # убираем все виды пробелов
    s = s.str.replace(r'\s+', '', regex=True)
    # удаляем точки, которые служат разделителями тысяч
    s = s.str.replace(r'\.(?=.*,)', '', regex=True)
    # удаляем лишние точки если их несколько
    s = s.str.replace(r'\.(?=.*\.)', '', regex=True)
    # заменяем запятую на точку
    s = s.str.replace(',', '.', regex=False)
    # извлекаем только числовую часть
    s = s.str.extract(r'(-?\d+(?:\.\d+)?)', expand=False)
    return s