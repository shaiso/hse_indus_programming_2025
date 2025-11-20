from datetime import date, datetime, timedelta

from sqlalchemy.orm import Session

from .models import MealLogORM


class MealRepo:
    def __init__(self, db: Session):
        self.db = db

    def create(self, orm: MealLogORM) -> MealLogORM:
        self.db.add(orm)
        self.db.commit()
        self.db.refresh(orm)
        return orm

    def get_by_user_and_day(
        self,
        user_id: int,
        day: date,
        tz_offset_hours: int = 0,
    ):
        start = datetime(day.year, day.month, day.day)
        end = start + timedelta(days=1)
        q = (
            self.db.query(MealLogORM)
            .filter(MealLogORM.user_id == user_id)
            .filter(MealLogORM.consumed_at >= start)
            .filter(MealLogORM.consumed_at < end)
            .order_by(MealLogORM.consumed_at.asc())
        )
        return q.all()
