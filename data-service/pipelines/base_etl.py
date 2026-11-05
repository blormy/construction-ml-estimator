# pipelines/base_etl.py
import pandas as pd
from pathlib import Path
import logging
import os
from dotenv import load_dotenv
from sqlalchemy import create_engine
from configs.config_loader import config as global_config

logger = logging.getLogger(__name__)

class BaseETLPipeline:
    """Базовый класс для io-операций (выгрузка и загрузка)"""
    def __init__(self, input_file_path: str | Path, output_folder_path: str | Path, config: dict, weight_file_path: str | Path | None = None):
        self.input_file = Path(input_file_path)
        self.weight_file = Path(weight_file_path) if weight_file_path else self.input_file
        self.output_folder = Path(output_folder_path)
        self.config = config

    def extract(self) -> tuple[pd.DataFrame, pd.DataFrame | None]:
        """Чтение данных с кэшированием в CSV"""
        logger.info(f"Чтение данных из источника: {self.input_file}")
        input_dir = self.input_file.parent
        cache_dir = input_dir / '.cache'
        cache_dir.mkdir(parents=True, exist_ok=True)

        features_sheet = self.config['sheets'].get('features')
        weight_sheet = self.config['sheets'].get('weight')

        # 1. читаем основной лист с признаками
        df_features = self._read_with_cache(
            self.input_file, features_sheet, cache_dir, "features"
        )

        # 2. проверка: требуется ли повторное чтение
        if not weight_sheet or (self.input_file == self.weight_file and features_sheet == weight_sheet):
            logger.info("Характеристики и вес находятся в одном источнике, повторное чтение не требуется")
            df_weight = None
        else:
            df_weight = self._read_with_cache(
                self.weight_file, weight_sheet, cache_dir, "weight"
            )

        logger.info("Данные успешно загружены")
        return df_features, df_weight

    def _read_with_cache(self, excel_path: Path, sheet_name: str | int | None, cache_dir: Path, suffix: str) -> pd.DataFrame:
        """Чтение Excel с кэшированием в CSV"""

        # формируем путь к файлу кэша
        cache_file = cache_dir / f'{excel_path.stem}_{suffix}.csv'

        try:
            # проверяем наличие xlsx файла
            if not excel_path.exists():
                raise FileNotFoundError()

            # флаг для обновления кэша
            need_update = True

            # если файл кэша есть, проверяем изменения
            if cache_file.exists():
                # получаем время последнего изменения
                excel_mtime = excel_path.stat().st_mtime
                cache_mtime = cache_file.stat().st_mtime

                # если кэш новее файла, то изменения не требуются
                if cache_mtime > excel_mtime:
                    need_update = False

            if need_update:
                # если второй файл без явного листа, используем первый лист
                target_sheet = sheet_name if sheet_name else 0
                logger.info(f"Чтение Excel-файла {excel_path.name} (лист: {target_sheet})")

                header_param = self.config['sheets'].get('header', 0)

                # читаем лист Excel
                df_excel = pd.read_excel(excel_path, sheet_name=target_sheet, header=header_param)

                if isinstance(df_excel.columns, pd.MultiIndex):
                    flattened_cols = []
                    for col_tuple in df_excel.columns:
                        lvl0 = str(col_tuple[0]).strip()
                        lvl1 = str(col_tuple[1]).strip()

                        # проверка валидности названия колонки
                        def is_valid_name(name):
                            return bool(name) and 'Unnamed' not in name and name.lower() != 'nan'

                        # логика приоритета уровней
                        if is_valid_name(lvl0):
                            final_name = lvl0
                        elif is_valid_name(lvl1):
                            final_name = lvl1
                        else:
                            final_name = f"Empty_Col_{len(flattened_cols)}"

                        flattened_cols.append(final_name)

                    df_excel.columns = flattened_cols

                logger.info(f"Создание или обновление кэша: {cache_file.name}")

                # сохраняем результат в csv
                df_excel.to_csv(cache_file, index=False, encoding='utf-8-sig')
                df = df_excel
            else:
                logger.info(f"Найден актуальный кэш для {suffix}, быстрое чтение")
                # читаем данные из csv
                df = pd.read_csv(cache_file, encoding='utf-8-sig')

            return df

        # обработка ошибок
        except FileNotFoundError:
            logger.error(f"Файл '{excel_path}' не найден")
            logger.info("Проверьте правильность пути и наличие файла")
            raise

        except ValueError as e:
            logger.error(f"Не найден нужный лист в Excel-файле '{excel_path.name}': {e}")
            raise

        except Exception as e:
            logger.error(f"Ошибка при извлечении данных из '{excel_path.name}': {e}")
            raise

    def transform(self, df_features: pd.DataFrame, df_weight: pd.DataFrame) -> pd.DataFrame:
        raise NotImplementedError("Этот метод реализуется в дочернем пайплайне")

    def load(self, df: pd.DataFrame, filename: str = 'dataset_ml.csv') -> None:
        """Сохранение готовых данных в CSV"""
        self.output_folder.mkdir(parents=True, exist_ok=True)
        df.to_csv(self.output_folder / filename, index=False)
        logger.info(f"Файлы успешно сохранены: {filename}")

    def load_to_db(self, df: pd.DataFrame) -> None:
        """Выгрузка готовых датасетов в PostgreSQL"""
        logger.info("Подключение к базе данных PostgreSQL...")

        table_prefix = self.config['db_params']['table_prefix']

        # получаем параметры подключения из переменных окружения
        db_host = os.getenv('DB_HOST')
        db_port = os.getenv('DB_PORT')
        db_name = os.getenv('DB_NAME')
        db_user = os.getenv('DB_USER')
        db_pass = os.getenv('DB_PASSWORD')

        # проверяем наличие всех обязательных параметров
        if not all([db_host, db_port, db_name, db_user, db_pass]):
            logger.error("Не найдены параметры подключения к БД в переменных окружения")
            return

        # формируем строку подключения
        engine_url = f"postgresql+psycopg2://{db_user}:{db_pass}@{db_host}:{db_port}/{db_name}?sslmode=require"

        try:
            engine = create_engine(engine_url)

            logger.info(f"Создание таблицы {table_prefix}_ml и выгрузка данных...")
            df.to_sql(f"{table_prefix}_ml", engine, if_exists='replace', index=False)

            logger.info("Данные успешно сохранены в БД")

        except Exception as e:
            logger.error(f"Ошибка при записи данных в БД: {e}")

    def run(self):
        """Запуск конвейера ETL по шагам"""
        logger.info("Запуск ETL-пайплайна")
        df_features, df_weight = self.extract()
        df_transformed = self.transform(df_features, df_weight)
        self.load(df_transformed)
        self.load_to_db(df_transformed)
        logger.info("ETL-пайплайн успешно завершен")

    def get_rename_map(self) -> dict:
        """Автоматически собирает словарь для переименования колонок"""
        raw = self.config.get('raw_names', {})
        cols = self.config.get('col_names', {})

        return {raw[k]: cols[k] for k in raw.keys() if k in cols}