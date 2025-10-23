from sqlalchemy.orm import Session
from .models import FoodORM

class FoodRepo:
    def __init__(self, db: Session):
        self.db = db

    def create(self, orm: FoodORM) -> FoodORM:
        self.db.add(orm)
        self.db.commit()
        self.db.refresh(orm)
        return orm

    def get(self, food_id: int) -> FoodORM | None:
        return self.db.get(FoodORM, food_id)

    def get_by_name(self, name: str) -> FoodORM | None:
        return self.db.query(FoodORM).filter(FoodORM.name == name).first()