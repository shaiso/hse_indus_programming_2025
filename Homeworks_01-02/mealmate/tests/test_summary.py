from datetime import date

from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_daily_summary():
    # create food per_serving
    food = {
        "name": "Yogurt",
        "mode": "per_serving",
        "serving_grams": 125.0,
        "kcal_serv": 80.0,
        "protein_serv": 5.0,
        "fat_serv": 3.0,
        "carb_serv": 9.0,
    }
    r = client.post("/foods", json=food)
    food_id = r.json()["id"]

    # 1.5 servings
    r = client.post("/meals", json={"food_id": food_id, "servings": 1.5})
    assert r.status_code == 201

    today = date.today().isoformat()
    r = client.get(f"/summary?day={today}")
    assert r.status_code == 200
    totals = r.json()
    assert totals["total_kcal"] == 120.0  # 80 * 1.5
