from datetime import date

from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from ...repos.database import get_db
from ...schemas.summary import DailyTotals
from ...services.summary_service import SummaryService

router = APIRouter(prefix="/summary", tags=["summary"])


@router.get("", response_model=DailyTotals)
def get_summary(
    day: date = Query(..., description="YYYY-MM-DD"),
    user_id: int = 1,
    db: Session = Depends(get_db),
):
    svc = SummaryService(db)
    return svc.day_totals(user_id=user_id, day=day)
