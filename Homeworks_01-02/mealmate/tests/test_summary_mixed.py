from datetime import date

from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_daily_summary_mixed_foods():
    # 1) per_100g продукт
    food1 = {
        "name": "Chicken",
        "mode": "per_100g",
        "kcal_100": 200.0,
        "protein_100": 30.0,
        "fat_100": 5.0,
        "carb_100": 0.0,
    }
    r = client.post("/foods", json=food1)
    assert r.status_code == 201
    food1_id = r.json()["id"]

    # 2) per_serving продукт
    food2 = {
        "name": "Rice",
        "mode": "per_serving",
        "serving_grams": 150.0,
        "kcal_serv": 180.0,
        "protein_serv": 4.0,
        "fat_serv": 1.0,
        "carb_serv": 38.0,
    }
    r = client.post("/foods", json=food2)
    assert r.status_code == 201
    food2_id = r.json()["id"]

    # Приём пищи: 150 г курицы -> 1.5 * (200,30,5,0)
    r = client.post("/meals", json={"food_id": food1_id, "grams": 150})
    assert r.status_code == 201

    # Приём пищи: 2 порции риса
    r = client.post("/meals", json={"food_id": food2_id, "servings": 2})
    assert r.status_code == 201

    today = date.today().isoformat()
    r = client.get(f"/summary?day={today}")
    assert r.status_code == 200
    totals = r.json()

    # Курочка: фактор 1.5
    # kcal: 200 * 1.5 = 300
    # protein: 30 * 1.5 = 45
    # fat: 5 * 1.5 = 7.5
    # carb: 0

    # Рис: 2 порции
    # kcal: 180 * 2 = 360
    # protein: 4 * 2 = 8
    # fat: 1 * 2 = 2
    # carb: 38 * 2 = 76

    assert totals["total_kcal"] == 660.0
    assert totals["total_protein"] == 53.0  # 45 + 8
    assert totals["total_fat"] == 9.5  # 7.5 + 2
    assert totals["total_carb"] == 76.0
