from datetime import datetime

from app.domain.models import Food, MealLog


def test_meal_log_kcal_per_100g():
    food = Food(
        id=1,
        name="Oatmeal",
        mode="per_100g",
        kcal_100=360.0,
    )
    meal = MealLog(
        id=1,
        user_id=1,
        food=food,
        grams=80.0,
        servings=None,
        consumed_at=datetime.utcnow(),
    )

    assert meal.kcal() == 288.0  # 360 * 0.8


def test_meal_log_kcal_per_serving():
    food = Food(
        id=2,
        name="Yogurt",
        mode="per_serving",
        kcal_serv=80.0,
        serving_grams=125.0,
    )
    meal = MealLog(
        id=2,
        user_id=1,
        food=food,
        grams=None,
        servings=1.5,
        consumed_at=datetime.utcnow(),
    )

    assert meal.kcal() == 120.0  # 80 * 1.5
