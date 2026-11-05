# Construction Materials Estimator

## Project Description
Prototype of an internal tool for estimating physical volumes of construction materials based on object parameters using machine learning.

## Repository Structure
- `backend/` — backend service
- `ml-service/` — ML service
- `data-service/` - data service
- `docs/` — project documentation

## Team
- Backend developer
- Backend developer
- ML engineer
- Data engineer
- Generalist

## Микросервисная архитектура
```mermaid
graph LR
    Client[Клиент] -->|JSON| Go[Go Backend]
    Go -->|HTTP/gRPC| Python[Python ML Service<br/>FastAPI]
    Python -->|Загрузка модели| Model[ML Model]
    Python -->|Результат| Go
    Go -->|Ответ| Client
