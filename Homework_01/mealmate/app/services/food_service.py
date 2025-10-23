from sqlalchemy.orm import Session
from ..schemas.foods import FoodCreate,FoodOut
from ..repos.models import FoodORM
from ..repos.food_repo import FoodRepo

class FoodService:
    def __init__(self, db: Session):
        self.repo = FoodRepo(db)

    def create(self, payload: FoodCreate) -> FoodOut:
        orm = FoodORM(**payload.model_dump())
        orm = self.repo.create(orm)
        return FoodOut.model_validate(orm)