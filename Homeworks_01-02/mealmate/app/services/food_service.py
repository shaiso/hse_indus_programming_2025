from sqlalchemy.orm import Session

from ..repos.food_repo import FoodRepo
from ..repos.models import FoodORM
from ..schemas.foods import FoodCreate, FoodOut


class FoodService:
    def __init__(self, db: Session):
        self.repo = FoodRepo(db)

    def create(self, payload: FoodCreate) -> FoodOut:
        orm = FoodORM(**payload.model_dump())
        orm = self.repo.create(orm)
        return FoodOut.model_validate(orm)
