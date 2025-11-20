# MealMate (тесты, качество кода и типизация)

## Тесты и покрытие

Тесты пишутся на `pytest`, покрытие считается через `pytest-cov`.
Конфигурация находится в `pytest.ini`:

```ini
[pytest]
testpaths = tests
pythonpath = .
addopts = --cov=app --cov-report=term-missing
```

Запуск тестов:
```
cd Homeworks_01-02/mealmate
pytest
```
В результате:
- выполняются 11 тестов (юнит-тесты для доменной модели + интеграционные тесты для `/foods`, `/meals`, `/summary`);
- считается покрытие по пакету app.

Итоговое покрытие на момент сдачи:
- **TOTAL: ~98%**
- 100% покрытие для:
  - роутеров (`app/api/routers/*.py`);
  - сервисов (`app/services/*.py`);
  - схем (`app/schemas/*.py`);
  - доменной модели (`app/domain/models.py`);
  - репозиториев и ORM-моделей (`app/repos/*.py`).


### Allure-отчёт

Для генерации отчёта используется плагин allure-pytest.

Сбор результатов:
```
pytest --alluredir=allure-results
```

Запуск локального HTML-отчёта (через Allure CLI):
```
allure serve allure-results
```

Скриншот Allure-отчёта прилагается к домашней работе.

### Линтеры и форматирование

В проекте настроены:
- black — автоформатирование кода.
- isort — сортировка импортов.
- flake8 — статический анализ стиля.

`pyproject.toml`:
```
[tool.black]
line-length = 88
target-version = ["py311"]

[tool.isort]
profile = "black"
line_length = 88
known_first_party = ["app"]
```

`.flake8`:
```
[flake8]
max-line-length = 88
extend-ignore = E203, W503
exclude = .git,__pycache__,.venv,.mypy_cache,.pytest_cache
```

Ручной запуск:
```
black app tests
isort app tests
flake8 app tests
```

### pre-commit hooks

Чтобы проверки качества кода запускались автоматически перед каждым коммитом, настроен pre-commit.

Файл `.pre-commit-config.yaml` включает хуки:
- black
- isort
- flake8
- mypy
- end-of-file-fixer
- trailing-whitespace

Установка и запуск:
```
pre-commit install
pre-commit run --all-files
```

Все хуки проходят успешно.

### Статическая типизация (mypy)

Для статической проверки типов используется mypy.
Типы проставлены в:

- сервисах (`app/services/*.py`);
- репозиториях (`app/repos/*.py`);
- доменных моделях (`app/domain/models.py`);
- Pydantic-схемах (`app/schemas/*.py`).

Конфиг `mypy.ini`:
```
[mypy]
python_version = 3.11
ignore_missing_imports = True
strict_optional = True
disallow_untyped_defs = True
check_untyped_defs = True
show_error_codes = True

[mypy-tests.*]
disallow_untyped_defs = False
```

Запуск:
```
mypy app
```

На момент сдачи mypy проходит без ошибок.

### Тесты: краткое содержание

- `tests/test_domain_models.py` — юнит-тесты доменного метода MealLog.kcal
для режима per_100g и per_serving.
- `tests/test_meals.py` — интеграционный тест:
  - создание продукта per_100g;
  - логирование приёма пищи по граммам;
  - проверка корректного расчёта ккал.
- `tests/test_summary.py` — дневная сводка для одного продукта per_serving.
- `tests/test_summary_mixed.py` — сводка по дню, где присутствуют и per_100g, и per_serving продукты, проверка суммарных КБЖУ.
- `tests/test_validation.py` — валидация входных данных (ошибки mode, отсутствующие обязательные поля, негативные значения и конфликт grams/servings).
- `tests/test_meals_errors.py` — обработка ошибки “food not found” через MealLogService и HTTP 400.

Тесты независимы друг от друга: перед каждым тестом БД пересоздаётся через фикстуру в `tests/conftest.py`.
