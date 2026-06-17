import { useState, useEffect, useCallback } from 'react'

const CATEGORIES = ['', 'electronics', 'laptops', 'phones', 'audio', 'tablets', 'wearables', 'general']

export default function ProductList({ api }) {
  const [products, setProducts] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [category, setCategory] = useState('')
  const [creating, setCreating] = useState(false)
  const [form, setForm] = useState({ name: '', description: '', price: '', stock: '10', category: 'general' })

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await api.products.list(category)
      setProducts(res.data || [])
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [api, category])

  useEffect(() => { load() }, [load])

  const handleCreate = async (e) => {
    e.preventDefault()
    try {
      await api.products.create({ ...form, price: parseFloat(form.price), stock: parseInt(form.stock) })
      setForm({ name: '', description: '', price: '', stock: '10', category: 'general' })
      setCreating(false)
      load()
    } catch (e) {
      setError(e.message)
    }
  }

  const handleDelete = async (id) => {
    if (!confirm('Delete this product?')) return
    try {
      await api.products.delete(id)
      load()
    } catch (e) {
      setError(e.message)
    }
  }

  return (
    <div className="max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold">Products</h2>
        <div className="flex items-center gap-3">
          <select
            value={category}
            onChange={e => setCategory(e.target.value)}
            className="bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm"
          >
            {CATEGORIES.map(c => <option key={c} value={c}>{c || 'All categories'}</option>)}
          </select>
          <button
            onClick={() => setCreating(v => !v)}
            className="bg-orange-500 hover:bg-orange-600 text-white text-sm px-4 py-1.5 rounded transition-colors"
          >
            {creating ? 'Cancel' : '+ New Product'}
          </button>
        </div>
      </div>

      {error && <div className="bg-red-900/40 border border-red-700 text-red-300 px-4 py-2 rounded mb-4 text-sm">{error}</div>}

      {creating && (
        <form onSubmit={handleCreate} className="bg-gray-900 border border-gray-800 rounded-lg p-5 mb-6 grid grid-cols-2 gap-4">
          <input required placeholder="Product name" value={form.name}
            onChange={e => setForm(f => ({...f, name: e.target.value}))}
            className="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm col-span-2" />
          <input placeholder="Description" value={form.description}
            onChange={e => setForm(f => ({...f, description: e.target.value}))}
            className="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm col-span-2" />
          <input required type="number" step="0.01" min="0.01" placeholder="Price (USD)" value={form.price}
            onChange={e => setForm(f => ({...f, price: e.target.value}))}
            className="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm" />
          <input required type="number" min="0" placeholder="Stock quantity" value={form.stock}
            onChange={e => setForm(f => ({...f, stock: e.target.value}))}
            className="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm" />
          <select value={form.category} onChange={e => setForm(f => ({...f, category: e.target.value}))}
            className="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm">
            {CATEGORIES.filter(Boolean).map(c => <option key={c} value={c}>{c}</option>)}
          </select>
          <button type="submit" className="bg-orange-500 hover:bg-orange-600 text-white text-sm px-4 py-2 rounded transition-colors">
            Create Product
          </button>
        </form>
      )}

      {loading ? (
        <div className="text-gray-500 text-sm">Loading...</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {products.length === 0 ? (
            <div className="col-span-3 text-gray-600 text-sm py-8 text-center">
              No products. Add some or run <code className="bg-gray-800 px-1 rounded">make seed</code>
            </div>
          ) : products.map(p => (
            <div key={p.id} className="bg-gray-900 border border-gray-800 rounded-lg p-4 flex flex-col gap-2">
              <div className="flex justify-between items-start">
                <h3 className="font-medium text-white">{p.name}</h3>
                <span className="text-xs bg-gray-800 text-gray-400 px-2 py-0.5 rounded">{p.category}</span>
              </div>
              {p.description && <p className="text-xs text-gray-500 line-clamp-2">{p.description}</p>}
              <div className="flex justify-between items-center mt-auto pt-2">
                <div>
                  <span className="text-orange-400 font-semibold">${p.price}</span>
                  <span className="text-xs text-gray-600 ml-2">stock: {p.stock}</span>
                </div>
                <button onClick={() => handleDelete(p.id)} className="text-xs text-red-500 hover:text-red-400 transition-colors">
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
