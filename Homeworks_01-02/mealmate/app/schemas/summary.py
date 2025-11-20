from datetime import date

from pydantic import BaseModel


class DailyTotals(BaseModel):
    date: date
    total_kcal: float
    total_protein: float
    total_fat: float
    total_carb: float
