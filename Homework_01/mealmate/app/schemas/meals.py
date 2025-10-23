from pydantic import BaseModel, field_validator
from datetime import datetime

class MealCreate(BaseModel):
    user_id: int = 1
    food_id: int
    grams: float | None = None
    servings: float | None = None
    consumed_at: datetime | None = None

    @field_validator("grams", "servings")
    def non_negative(cls, v):
        if v is not None and v <= 0:
            raise ValueError("value must be > 0")
        return v

    @field_validator("servings")
    def xor_grams_servings(cls, v, info):
        grams = info.data.get("grams")
        if (grams is None and v is None) or (grams is not None and v is not None):
            raise ValueError("provide either grams or servings")
        return v

class MealOut(BaseModel):
    id: int
    user_id: int
    food_name: str
    kcal: float
    consumed_at: datetime