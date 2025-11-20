from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_add_meal_food_not_found():
    # передаём food_id, которого нет в БД
    r = client.post("/meals", json={"food_id": 9999, "grams": 50})
    # MealLogService.add -> ValueError -> HTTPException 400
    assert r.status_code == 400
    assert r.json()["detail"] == "food not found"
