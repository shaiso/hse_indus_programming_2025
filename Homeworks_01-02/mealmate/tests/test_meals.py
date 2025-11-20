from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_create_food_and_meal_and_calc_kcal():
    # create food per_100g
    food = {
        "name": "Oatmeal",
        "mode": "per_100g",
        "kcal_100": 360.0,
        "protein_100": 12.0,
        "fat_100": 6.0,
        "carb_100": 60.0,
    }
    r = client.post("/foods", json=food)
    assert r.status_code == 201
    food_id = r.json()["id"]

    # add meal 80 g
    meal = {"food_id": food_id, "grams": 80}
    r = client.post("/meals", json=meal)
    assert r.status_code == 201
    data = r.json()
    assert data["kcal"] == 288.0  # 360 * 0.8
