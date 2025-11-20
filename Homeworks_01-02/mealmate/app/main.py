from contextlib import asynccontextmanager

from fastapi import FastAPI

from .api.routers import foods, meals, summary
from .repos.database import engine
from .repos.models import Base


@asynccontextmanager
async def lifespan(app: FastAPI):
    Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(title="MealMate API (MVP)", lifespan=lifespan)

app.include_router(foods.router)
app.include_router(meals.router)
app.include_router(summary.router)
