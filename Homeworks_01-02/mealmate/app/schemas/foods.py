from pydantic import BaseModel, ConfigDict, Field, model_validator


class FoodCreate(BaseModel):
    name: str = Field(min_length=1, max_length=120)
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

    @model_validator(mode="after")
    def check_mode_fields(self):
        if self.mode == "per_100g":
            assert self.kcal_100 is not None, "kcal_100 required for per_100g"
        elif self.mode == "per_serving":
            assert (
                self.kcal_serv is not None and self.serving_grams
            ), "serving_grams & kcal_serv required"
        else:
            raise ValueError("mode must be per_100g or per_serving")
        return self


class FoodOut(BaseModel):
    id: int
    name: str
    mode: str
    model_config = ConfigDict(from_attributes=True)
