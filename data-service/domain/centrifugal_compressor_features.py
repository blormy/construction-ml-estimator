import logging

logger = logging.getLogger(__name__)


class CentrifugalCompressorFeatureEngineer:
    """Доменная логика и расчет веса для центробежных компрессоров (Centrifugal Compressor)"""

    def __init__(self, config: dict = None):
        self.config = config or {}

    def calculate_stub_weight(self, gas_flow_inlet: float, pressure_inlet: float, pressure_outlet: float) -> float:
        """Оценка веса компрессора (заглушка)"""
        logger.info(
            f"Расчет заглушки веса компрессора: flow={gas_flow_inlet}, "
            f"p_in={pressure_inlet}, p_out={pressure_outlet}"
        )
        return float(gas_flow_inlet * pressure_inlet * pressure_outlet)
