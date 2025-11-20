from datetime import date

from sqlalchemy.orm import Session

from ..repos.food_repo import FoodRepo
from ..repos.meal_repo import MealRepo
from ..schemas.summary import DailyTotals


class SummaryService:
    def __init__(self, db: Session):
        self.meal_repo = MealRepo(db)
        self.food_repo = FoodRepo(db)

    def day_totals(self, user_id: int, day: date) -> DailyTotals:
        meals = self.meal_repo.get_by_user_and_day(user_id, day)
        total_kcal = total_p = total_f = total_c = 0.0
        for m in meals:
            food = self.food_repo.get(m.food_id)
            if not food:
                continue
            if m.grams:
                factor = m.grams / 100.0
                total_kcal += (food.kcal_100 or 0) * factor
                total_p += (food.protein_100 or 0) * factor
                total_f += (food.fat_100 or 0) * factor
                total_c += (food.carb_100 or 0) * factor
            else:
                factor = m.servings or 0
                total_kcal += (food.kcal_serv or 0) * factor
                total_p += (food.protein_serv or 0) * factor
                total_f += (food.fat_serv or 0) * factor
                total_c += (food.carb_serv or 0) * factor

        return DailyTotals(
            date=day,
            total_kcal=round(total_kcal, 2),
            total_protein=round(total_p, 2),
            total_fat=round(total_f, 2),
            total_carb=round(total_c, 2),
        )
