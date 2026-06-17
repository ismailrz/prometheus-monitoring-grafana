import { useState, useEffect, useCallback } from 'react'

const STATUS_COLORS = {
  pending:   'bg-yellow-900/40 text-yellow-400 border-yellow-800',
  confirmed: 'bg-blue-900/40 text-blue-400 border-blue-800',
  shipped:   'bg-purple-900/40 text-purple-400 border-purple-800',
  delivered: 'bg-green-900/40 text-green-400 border-green-800',
  cancelled: 'bg-red-900/40 text-red-400 border-red-800',
}

export default function OrderList({ api }) {
  const [orders, setOrders] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [creating, setCreating] = useState(false)
  const [products, setProducts] = useState([])
  const [form, setForm] = useState({ customer_email: '', items: [{ product_id: '', quantity: 1 }] })

  const loadOrders = useCallback(async () => {
    setLoading(true)
    try {
      const res = await api.orders.list()
      setOrders(res.data || [])
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [api])

  useEffect(() => { loadOrders() }, [loadOrders])

  useEffect(() => {
    if (creating) {
      api.products.list().then(r => setProducts(r.data || []))
    }
  }, [creating, api])

  const handleCreate = async (e) => {
    e.preventDefault()
    try {
      await api.orders.create({
        customer_email: form.customer_email,
        items: form.items.filter(i => i.product_id).map(i => ({
          product_id: i.product_id,
          quantity: parseInt(i.quantity),
        })),
      })
      setForm({ customer_email: '', items: [{ product_id: '', quantity: 1 }] })
      setCreating(false)
      loadOrders()
    } catch (e) {
      setError(e.message)
    }
  }

  const addItem = () => setForm(f => ({ ...f, items: [...f.items, { product_id: '', quantity: 1 }] }))

  const updateItem = (idx, field, val) =>
    setForm(f => ({ ...f, items: f.items.map((item, i) => i === idx ? { ...item, [field]: val } : item) }))

  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold">Orders</h2>
        <button onClick={() => setCreating(v => !v)}
          className="bg-orange-500 hover:bg-orange-600 text-white text-sm px-4 py-1.5 rounded transition-colors">
          {creating ? 'Cancel' : '+ New Order'}
        </button>
      </div>

      {error && <div className="bg-red-900/40 border border-red-700 text-red-300 px-4 py-2 rounded mb-4 text-sm">{error}</div>}

      {creating && (
        <form onSubmit={handleCreate} className="bg-gray-900 border border-gray-800 rounded-lg p-5 mb-6 space-y-4">
          <input required type="email" placeholder="customer@example.com"
            value={form.customer_email} onChange={e => setForm(f => ({...f, customer_email: e.target.value}))}
            className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm" />

          <div className="space-y-2">
            {form.items.map((item, idx) => (
              <div key={idx} className="flex gap-2">
                <select required value={item.product_id} onChange={e => updateItem(idx, 'product_id', e.target.value)}
                  className="flex-1 bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm">
                  <option value="">Select product</option>
                  {products.map(p => <option key={p.id} value={p.id}>{p.name} (${p.price})</option>)}
                </select>
                <input type="number" min="1" value={item.quantity} onChange={e => updateItem(idx, 'quantity', e.target.value)}
                  className="w-24 bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm" />
              </div>
            ))}
          </div>
          <div className="flex gap-2">
            <button type="button" onClick={addItem} className="text-sm text-gray-400 hover:text-white">+ add item</button>
            <button type="submit" className="ml-auto bg-orange-500 hover:bg-orange-600 text-white text-sm px-4 py-2 rounded">Place Order</button>
          </div>
        </form>
      )}

      {loading ? (
        <div className="text-gray-500 text-sm">Loading...</div>
      ) : (
        <div className="space-y-3">
          {orders.length === 0 ? (
            <div className="text-gray-600 text-sm py-8 text-center">
              No orders. Create some or run <code className="bg-gray-800 px-1 rounded">make load-test</code>
            </div>
          ) : orders.map(o => (
            <div key={o.id} className="bg-gray-900 border border-gray-800 rounded-lg p-4 flex items-center justify-between">
              <div>
                <div className="font-mono text-xs text-gray-500">{o.id}</div>
                <div className="text-sm text-white mt-0.5">{o.customer_email}</div>
              </div>
              <div className="flex items-center gap-4">
                <span className={`text-xs border px-2 py-0.5 rounded ${STATUS_COLORS[o.status] || 'bg-gray-800 text-gray-400 border-gray-700'}`}>
                  {o.status}
                </span>
                <span className="text-orange-400 font-semibold text-sm">${Number(o.total_amount).toFixed(2)}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
