from sqlalchemy.orm import Session

from ..repos.food_repo import FoodRepo
from ..repos.meal_repo import MealRepo
from ..repos.models import FoodORM, MealLogORM
from ..schemas.meals import MealCreate, MealOut


class MealLogService:
    def __init__(self, db: Session):
        self.meal_repo = MealRepo(db)
        self.food_repo = FoodRepo(db)
        self.db = db

    def add(self, payload: MealCreate) -> MealOut:
        food: FoodORM | None = self.food_repo.get(payload.food_id)
        if not food:
            raise ValueError("food not found")

        # Инвариант: либо grams, либо servings (уже проверено Pydantic)
        orm = MealLogORM(
            user_id=payload.user_id,
            food_id=payload.food_id,
            grams=payload.grams,
            servings=payload.servings,
            consumed_at=payload.consumed_at,
        )
        orm = self.meal_repo.create(orm)

        # Domain -> DTO mapping
        kcal = (
            (payload.grams or 0) * (food.kcal_100 or 0) / 100.0
            if payload.grams is not None
            else (payload.servings or 0) * (food.kcal_serv or 0)
        )
        return MealOut(
            id=orm.id,
            user_id=orm.user_id,
            food_name=food.name,
            kcal=round(float(kcal), 2),
            consumed_at=orm.consumed_at,
        )
