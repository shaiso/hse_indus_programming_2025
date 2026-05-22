const express = require('express');

const app = express();
app.use(express.json());

// In-memory store for demo purposes.
const products = new Map();
let nextId = 1;

// GET /products — list all products with optional category filter.
app.get('/products', (req, res) => {
  const category = req.query.category;
  const items = [...products.values()].filter(p =>
    !category || p.category === category
  );
  res.json(items);
});

// GET /products/:id — fetch a single product.
app.get('/products/:id', (req, res) => {
  const id = Number(req.params.id);
  const product = products.get(id);
  if (!product) return res.status(404).json({ error: 'not found' });
  res.json(product);
});

// POST /products — create a new product.
app.post('/products', (req, res) => {
  const { name, price, category } = req.body || {};
  if (!name || typeof price !== 'number') {
    return res.status(400).json({ error: 'name and numeric price are required' });
  }
  const product = { id: nextId++, name, price, category: category || 'misc' };
  products.set(product.id, product);
  res.status(201).json(product);
});

// PATCH /products/:id — partial update.
app.patch('/products/:id', (req, res) => {
  const id = Number(req.params.id);
  const product = products.get(id);
  if (!product) return res.status(404).json({ error: 'not found' });
  Object.assign(product, req.body || {});
  res.json(product);
});

app.listen(3000, () => console.log('listening on :3000'));
