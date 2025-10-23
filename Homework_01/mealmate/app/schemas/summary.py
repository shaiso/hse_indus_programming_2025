from pydantic import BaseModel
from datetime import date

class DailyTotals(BaseModel):
    date: date
    total_kcal: float
    total_protein: float
    total_fat: float
    total_carb: float