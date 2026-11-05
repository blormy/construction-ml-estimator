import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np
from pathlib import Path


def generate_report(file_path='../datasets/pump_dataset_inference.csv'):
    """Генерация графиков для разведочного анализа данных (EDA)"""
    # загрузка и настройка стилей
    df = pd.read_csv(file_path)
    sns.set_theme(style="whitegrid")

    # анализ пропусков
    plt.figure(figsize=(10, 5))
    missing_pct = (df.isnull().sum() / len(df)) * 100
    missing_pct = missing_pct.sort_values(ascending=False)
    sns.barplot(x=missing_pct.values, y=missing_pct.index, palette='Reds_r')
    plt.title('Percentage of Missing Data (AFTER Enrichment)')
    plt.xlabel('% Missing')
    plt.tight_layout()
    plt.savefig('ds_inf_02_missing_analysis.png')

    # матрица корреляций
    plt.figure(figsize=(12, 10))
    numeric_df = df.select_dtypes(include=[np.number])
    corr = numeric_df.corr()
    sns.heatmap(corr, annot=True, cmap='coolwarm', center=0, fmt='.2f')
    plt.title('Feature Correlation Matrix')
    plt.tight_layout()
    plt.savefig('ds_inf_03_correlations.png')

    # распределение целевой переменной
    plt.figure(figsize=(8, 6))
    sns.histplot(df['weight_log'].dropna(), kde=True, color='purple')
    plt.title('Target Variable Distribution: log(Weight)')
    plt.savefig('ds_inf_01_target_dist.png')

    print("Графики для отчета успешно сгенерированы.")


if __name__ == "__main__":
    generate_report()