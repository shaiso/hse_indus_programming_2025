from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ...repos.database import get_db
from ...schemas.meals import MealCreate, MealOut
from ...services.meal_service import MealLogService

router = APIRouter(prefix="/meals", tags=["meals"])


@router.post("", response_model=MealOut, status_code=201)
def add_meal(payload: MealCreate, db: Session = Depends(get_db)):
    try:
        svc = MealLogService(db)
        result = svc.add(payload)
        return result
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
