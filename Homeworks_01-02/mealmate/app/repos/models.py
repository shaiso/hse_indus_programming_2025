from datetime import datetime, timezone

from sqlalchemy import CheckConstraint, DateTime, Float, ForeignKey, Integer, String
from sqlalchemy.orm import Mapped, mapped_column, relationship

from .database import Base


class FoodORM(Base):
    __tablename__ = "foods"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    name: Mapped[str] = mapped_column(String(120), unique=True, index=True)
    mode: Mapped[str] = mapped_column(String(20))  # "per_100g" | "per_serving"

    kcal_100: Mapped[float | None] = mapped_column(Float, nullable=True)
    protein_100: Mapped[float | None] = mapped_column(Float, nullable=True)
    fat_100: Mapped[float | None] = mapped_column(Float, nullable=True)
    carb_100: Mapped[float | None] = mapped_column(Float, nullable=True)

    serving_grams: Mapped[float | None] = mapped_column(Float, nullable=True)
    kcal_serv: Mapped[float | None] = mapped_column(Float, nullable=True)
    protein_serv: Mapped[float | None] = mapped_column(Float, nullable=True)
    fat_serv: Mapped[float | None] = mapped_column(Float, nullable=True)
    carb_serv: Mapped[float | None] = mapped_column(Float, nullable=True)

    __table_args__ = (
        CheckConstraint(
            "coalesce(kcal_100,0) >= 0 and coalesce(kcal_serv,0) >= 0",
            name="kcal_nonneg",
        ),
    )


class MealLogORM(Base):
    __tablename__ = "meal_logs"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    user_id: Mapped[int] = mapped_column(Integer, index=True)
    food_id: Mapped[int] = mapped_column(ForeignKey("foods.id"), index=True)
    grams: Mapped[float | None] = mapped_column(Float, nullable=True)
    servings: Mapped[float | None] = mapped_column(Float, nullable=True)

    consumed_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        default=lambda: datetime.now(timezone.utc),
        index=True,
    )

    food = relationship("FoodORM")
