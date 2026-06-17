const BASE = '/api/v1'

async function request(method, path, body) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' },
  }
  if (body) opts.body = JSON.stringify(body)

  const res = await fetch(BASE + path, opts)
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  products: {
    list:   (category) => request('GET', `/products${category ? `?category=${category}` : ''}`),
    get:    (id)       => request('GET', `/products/${id}`),
    create: (data)     => request('POST', '/products', data),
    update: (id, data) => request('PUT', `/products/${id}`, data),
    delete: (id)       => request('DELETE', `/products/${id}`),
  },
  orders: {
    list:   (status) => request('GET', `/orders${status ? `?status=${status}` : ''}`),
    get:    (id)     => request('GET', `/orders/${id}`),
    create: (data)   => request('POST', '/orders', data),
  },
  health: () => fetch('/health').then(r => r.json()),
  ready:  () => fetch('/ready').then(r => r.json()),
}
