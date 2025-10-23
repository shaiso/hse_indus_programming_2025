from dataclasses import dataclass
from datetime import datetime

@dataclass
class Food:
    id: int
    name: str
    mode: str
    # per_100g
    kcal_100: float | None = None
    protein_100: float | None = None
    fat_100: float | None = None
    carb_100: float | None = None
    # per_serving
    serving_grams: float | None = None
    kcal_serv: float | None = None
    protein_serv: float | None = None
    fat_serv: float | None = None
    carb_serv: float | None = None

@dataclass
class MealLog:
    id: int | None
    user_id: int
    food: Food
    grams: float | None
    servings: float | None
    consumed_at: datetime

    # инвариант: ровно один из grams/servings
    def kcal(self) -> float:
        if self.grams:
            assert self.food.kcal_100 is not None
            return self.grams * (self.food.kcal_100 / 100.0)
        else:
            assert self.food.kcal_serv is not None
            return self.servings * self.food.kcal_serv