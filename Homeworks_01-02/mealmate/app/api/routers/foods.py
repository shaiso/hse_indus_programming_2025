from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ...repos.database import get_db
from ...schemas.foods import FoodCreate, FoodOut
from ...services.food_service import FoodService

router = APIRouter(prefix="/foods", tags=["foods"])


@router.post("", response_model=FoodOut, status_code=201)
def create_food(payload: FoodCreate, db: Session = Depends(get_db)):
    svc = FoodService(db)
    return svc.create(payload)
