import pytest

from app.repos.database import Base, engine


@pytest.fixture(autouse=True)
def reset_db():
    """
    Перед КАЖДЫМ тестом полностью пересоздаём схему БД.
    """
    Base.metadata.drop_all(bind=engine)
    Base.metadata.create_all(bind=engine)
    yield
