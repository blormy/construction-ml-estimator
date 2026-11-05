import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns

plt.style.use('seaborn-v0_8-darkgrid')
sns.set_context("talk")


def plot_presentation_charts(raw_weight_path: str, ml_dataset_path: str):
    """Построение аналитических графиков для презентации"""
    print("Загрузка данных для визуализации...")

    df_raw_weight = pd.read_csv(raw_weight_path)
    # очищаем вес от единиц измерения
    df_raw_weight['Equip Weight Clean'] = df_raw_weight['Equip Weight'].astype(str).str.replace(' KG', '', regex=False)
    df_raw_weight['Equip Weight Clean'] = pd.to_numeric(df_raw_weight['Equip Weight Clean'], errors='coerce')

    # загружаем готовый датасет
    df_ml = pd.read_csv(ml_dataset_path)

    # 1. распределение целевой переменной (до и после)
    fig1, (ax1, ax2) = plt.subplots(1, 2, figsize=(16, 7))
    fig1.suptitle('Трансформация целевой переменной', fontsize=20, fontweight='bold')

    # исходное распределение
    sns.histplot(df_raw_weight['Equip Weight Clean'].dropna(), bins=50, kde=True, color='indianred', ax=ax1)
    ax1.set_title('ДО: Исходное распределение веса (кг)\n(Смещение вправо)', fontsize=14)
    ax1.set_xlabel('Вес (кг)')
    ax1.set_ylabel('Количество насосов')

    # логарифмированное распределение
    sns.histplot(df_ml['weight_log'].dropna(), bins=50, kde=True, color='mediumseagreen', ax=ax2)
    ax2.set_title('ПОСЛЕ: Логарифмированный вес\n(Нормальное распределение)', fontsize=14)
    ax2.set_xlabel('Log(Вес)')
    ax2.set_ylabel('Количество насосов')

    plt.tight_layout()
    plt.savefig('ds_ml_01_weight_distribution.png', dpi=300, bbox_inches='tight')
    print("Сохранен график: 01_weight_distribution.png")
    plt.close()

    # 2. инженерные признаки (полезная мощность)
    fig2 = plt.figure(figsize=(10, 8))

    sns.regplot(data=df_ml, x='useful_kw_log', y='weight_log',
                scatter_kws={'alpha': 0.5, 's': 40, 'color': 'steelblue'},
                line_kws={'color': 'darkorange', 'linewidth': 3})

    plt.title('Создание нового признака: Полезная мощность', fontsize=18, fontweight='bold', pad=20)
    plt.xlabel('Log(Полезная мощность) [Напор * Расход]')
    plt.ylabel('Log(Вес насоса)')

    plt.tight_layout()
    plt.savefig('ds_ml_02_feature_engineering.png', dpi=300, bbox_inches='tight')
    print("Сохранен график: 02_feature_engineering.png")

    # 3. матрица корреляций признаков
    fig3 = plt.figure(figsize=(12, 10))

    numeric_df = df_ml.select_dtypes(include=[np.number])
    corr_matrix = numeric_df.corr()

    # строим тепловую карту без маски
    sns.heatmap(corr_matrix, annot=True, fmt=".2f", cmap='coolwarm',
                vmin=-1, vmax=1, center=0, square=True, linewidths=.5,
                cbar_kws={"shrink": .8})

    plt.title('Взаимосвязь параметров (Матрица корреляций)', fontsize=18, fontweight='bold', pad=20)

    plt.tight_layout()
    plt.savefig('ds_ml_03_correlation_matrix.png', dpi=300, bbox_inches='tight')
    print("Сохранен график: 03_correlation_matrix.png")
    plt.close()

    print("Все графики успешно сгенерированы")


if __name__ == '__main__':
    RAW_WEIGHT_CSV = '../data/.cache/tag_weight.csv'
    ML_DATASET_CSV = '../datasets/pump_dataset_ml.csv'

    plot_presentation_charts(RAW_WEIGHT_CSV, ML_DATASET_CSV)