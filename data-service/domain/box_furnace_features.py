import pandas as pd
import logging

logger = logging.getLogger(__name__)


class BoxFurnaceFeatureEngineer:
    """Доменная логика и расчет веса для коробчатых печей (Box Furnace)"""

    def __init__(self, config: dict = None):
        self.config = config or {}

    def calculate_stub_weight(self, duty: float, gas_flow: float) -> float:
        """Оценка веса коробчатой печи (заглушка)"""
        logger.info(f"Расчет заглушки веса печи: duty={duty}, flow={gas_flow}")
        return float(duty * gas_flow)
