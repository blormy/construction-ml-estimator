from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from typing import Optional, Any, Dict

from .service import get_pump_estimation, get_vessel_estimation, get_conveyor_estimation, get_drum_estimation, get_utube_estimation, get_tower_estimation, get_box_furnace_estimation, get_centrifugal_compressor_estimation

app = FastAPI(title="Pump Data/ML Service", version="1.0.0")

MODEL_VERSION = "v1"


# модели запросов

class PumpRequest(BaseModel):
    tag: str = Field(..., min_length=1)
    flow_rate: float = Field(..., gt=0)      # обязательное
    fluid_head: float = Field(..., gt=0)     # обязательное

    # опциональные
    rpm: Optional[float] = Field(default=None, gt=0)
    spec_gravity: Optional[float] = Field(default=None, gt=0)
    power_kw: Optional[float] = Field(default=None, gt=0)


class DrumRequest(BaseModel):
    tag: str = Field(..., min_length=1)
    vessel_diameter: float = Field(..., gt=0)                     # обязательное
    design_tangent_to_tangent_length: float = Field(..., gt=0)    # обязательное

    # опциональные
    design_gauge_pressure: Optional[float] = Field(default=None, gt=0)
    design_temperature: Optional[float] = Field(default=None)
    operating_temperature: Optional[float] = Field(default=None)
    corossion_allowance: Optional[float] = Field(default=None, ge=0)  # допускаем 0


class VesselRequest(BaseModel):
    tag: str = Field(..., min_length=1)
    vessel_diameter: float = Field(..., gt=0)                       # обязательное
    vessel_tangent_to_tangent_height: float = Field(..., gt=0)      # обязательное

    # опциональные
    design_gauge_pressure: Optional[float] = Field(default=None, gt=0)
    design_temperature: Optional[float] = Field(default=None)
    operating_temperature: Optional[float] = Field(default=None)
    skirt_height: Optional[float] = Field(default=None, ge=0)        # допускаем 0
    vessel_leg_height: Optional[float] = Field(default=None, ge=0)   # допускаем 0


class ConveyorRequest(BaseModel):
    tag: str = Field(..., min_length=1)
    conveyor_length: float = Field(..., gt=0)     # обязательное
    belt_width: float = Field(..., gt=0)          # обязательное

    # опциональные
    conveyor_flow_rate: Optional[float] = Field(default=None, ge=0)          # допускаем 0
    driver_power_per_second: Optional[float] = Field(default=None, ge=0)     # допускаем 0
    conveyor_speed: Optional[float] = Field(default=None, ge=0)              # допускаем 0
    number_of_walkways: Optional[float] = Field(default=None, ge=0)          # допускаем 0


class UTubeRequest(BaseModel):
    tag: str = Field(..., min_length=1)
    shell_diameter: float = Field(..., gt=0)      # обязательное
    tube_out_diameter: float = Field(..., gt=0)   # обязательное
    tube_len: float = Field(..., gt=0)            # обязательное

    # опциональные
    tube_des_pres: Optional[float] = Field(default=None, ge=0)
    heat_area: Optional[float] = Field(default=None, ge=0)


class TowerRequest(BaseModel):
    tag: str = Field(..., min_length=1)
    vessel_diameter: float = Field(..., gt=0)     # обязательное
    number_of_trays: float = Field(..., ge=0)     # обязательное
    
    # опциональные
    design_tangent_to_tangent_length: Optional[float] = Field(default=None, ge=0)
    design_gauge_pressure: Optional[float] = Field(default=None)


class BoxFurnaceRequest(BaseModel):
    tag: str = Field(..., min_length=1)
    duty: float = Field(..., gt=0)                          # обязательное, > 0
    standard_gas_flow_rate: float = Field(..., gt=0)         # обязательное, > 0
    design_gauge_pressure: Optional[float] = Field(default=None, ge=0)
    design_temperature: Optional[float] = Field(default=None)


class CentrifugalCompressorRequest(BaseModel):
    tag: str = Field(..., min_length=1)
    actual_gas_flow_rate_inlet: float = Field(..., gt=0)      # обязательное, > 0
    design_gauge_pressure_inlet: float = Field(..., gt=0)      # обязательное, > 0
    design_gauge_pressure_outlet: float = Field(..., gt=0)     # обязательное, > 0
    driver_power: Optional[float] = Field(default=None, ge=0)

# модель ответа

class Response(BaseModel):
    model_version: str
    weight: float
    debug_features: Dict[str, Any]


# маршруты API

@app.get("/health")
def health():
    """Проверка доступности сервиса"""
    return {"status": "ok", "model_version": MODEL_VERSION}


@app.post("/pump/estimate", response_model=Response)
def pump_estimate(req: PumpRequest):
    """Оценка веса насоса"""
    try:
        input_data = req.model_dump()
        result = get_pump_estimation(input_data)

        return Response(
            model_version=MODEL_VERSION,
            weight=result["weight"],
            debug_features=result["features"]
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/drum/estimate", response_model=Response)
def drum_estimate(req: DrumRequest):
    """Оценка веса горизонтальной емкости"""
    try:
        input_data = req.model_dump()
        result = get_drum_estimation(input_data)

        return Response(
            model_version=MODEL_VERSION,
            weight=result["weight"],
            debug_features=result["features"]
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/vessel/estimate", response_model=Response)
def vessel_estimate(req: VesselRequest):
    """Оценка веса вертикального сосуда"""
    try:
        input_data = req.model_dump()
        result = get_vessel_estimation(input_data)
        
        return Response(
            model_version=MODEL_VERSION,
            weight=result["weight"],
            debug_features=result["features"]
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/conveyor/estimate", response_model=Response)
def conveyor_estimate(req: ConveyorRequest):
    """Оценка веса конвейера"""
    try:
        input_data = req.model_dump()
        result = get_conveyor_estimation(input_data)

        return Response(
            model_version=MODEL_VERSION,
            weight=result["weight"],
            debug_features=result["features"]
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.post("/utube/estimate", response_model=Response)
def utube_estimate(req: UTubeRequest):
    """Оценка веса теплообменника (U-Tube)"""
    try:
        input_data = req.model_dump()
        result = get_utube_estimation(input_data)

        return Response(
            model_version=MODEL_VERSION,
            weight=result["weight"],
            debug_features=result["features"]
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.post("/tower/estimate", response_model=Response)
def tower_estimate(req: TowerRequest):
    """Оценка веса колонны (Tower)"""
    try:
        input_data = req.model_dump()
        result = get_tower_estimation(input_data)

        return Response(
            model_version=MODEL_VERSION,
            weight=result["weight"],
            debug_features=result["features"]
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.post("/box_furnace/estimate", response_model=Response)
def box_furnace_estimate(req: BoxFurnaceRequest):
    """Оценка веса коробчатой печи (Box Furnace)"""
    try:
        input_data = req.model_dump()
        result = get_box_furnace_estimation(input_data)

        return Response(
            model_version=MODEL_VERSION,
            weight=result["weight"],
            debug_features=result["features"]
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.post("/centrifugal_compressor/estimate", response_model=Response)
def centrifugal_compressor_estimate(req: CentrifugalCompressorRequest):
    """Оценка веса центробежного компрессора"""
    try:
        input_data = req.model_dump()
        result = get_centrifugal_compressor_estimation(input_data)

        return Response(
            model_version=MODEL_VERSION,
            weight=result["weight"],
            debug_features=result["features"]
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))
