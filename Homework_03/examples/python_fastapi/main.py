from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from typing import List, Optional

app = FastAPI(title="Items Service", version="1.0.0")


class Item(BaseModel):
    id: int
    name: str
    price: float
    in_stock: bool = True


class CreateItemRequest(BaseModel):
    name: str = Field(..., min_length=1)
    price: float = Field(..., gt=0)
    in_stock: Optional[bool] = True


@app.get("/items", response_model=List[Item])
def list_items(limit: int = 10, offset: int = 0):
    """Return a paginated list of items."""
    return []


@app.get("/items/{item_id}", response_model=Item)
def get_item(item_id: int):
    """Return a single item by id."""
    if item_id <= 0:
        raise HTTPException(status_code=404, detail="not found")
    return Item(id=item_id, name="demo", price=1.0)


@app.post("/items", response_model=Item, status_code=201)
def create_item(req: CreateItemRequest):
    """Create a new item."""
    return Item(id=1, name=req.name, price=req.price, in_stock=req.in_stock)


@app.put("/items/{item_id}", response_model=Item)
def update_item(item_id: int, req: CreateItemRequest):
    """Update an existing item."""
    return Item(id=item_id, name=req.name, price=req.price, in_stock=req.in_stock)


@app.delete("/items/{item_id}", status_code=204)
def delete_item(item_id: int):
    """Remove an item."""
    return None
