from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_create_food_invalid_mode():
    # mode не из ["per_100g", "per_serving"] -> 422 от FastAPI
    payload = {
        "name": "BadFood",
        "mode": "invalid_mode",
    }
    r = client.post("/foods", json=payload)
    assert r.status_code == 422


def test_create_food_per_100g_missing_kcal():
    # per_100g, но нет kcal_100 -> ошибка валидации -> 422
    payload = {
        "name": "NoKcalFood",
        "mode": "per_100g",
        # "kcal_100" отсутствует
    }
    r = client.post("/foods", json=payload)
    assert r.status_code == 422


def test_create_food_per_serving_missing_fields():
    # per_serving, но нет serving_grams и kcal_serv -> 422
    payload = {
        "name": "BadServingFood",
        "mode": "per_serving",
    }
    r = client.post("/foods", json=payload)
    assert r.status_code == 422


def test_add_meal_negative_grams():
    # сначала создаём валидный food
    food = {
        "name": "Apple",
        "mode": "per_100g",
        "kcal_100": 52.0,
    }
    r = client.post("/foods", json=food)
    assert r.status_code == 201
    food_id = r.json()["id"]

    # grams < 0 -> валидатор MealCreate.non_negative -> 422
    meal = {"food_id": food_id, "grams": -10}
    r = client.post("/meals", json=meal)
    assert r.status_code == 422


def test_add_meal_both_grams_and_servings():
    # создаём ещё раз food
    food = {
        "name": "Bread",
        "mode": "per_100g",
        "kcal_100": 250.0,
    }
    r = client.post("/foods", json=food)
    assert r.status_code == 201
    food_id = r.json()["id"]

    # и grams, и servings → должен сработать xor-валидатор -> 422
    meal = {"food_id": food_id, "grams": 50, "servings": 1.0}
    r = client.post("/meals", json=meal)
    assert r.status_code == 422
