import { useState, useEffect } from 'react'
import { api } from './services/api'
import ProductList from './components/ProductList'
import OrderList from './components/OrderList'
import HealthStatus from './components/HealthStatus'

const NAV_ITEMS = ['Products', 'Orders', 'Health']

export default function App() {
  const [tab, setTab] = useState('Products')

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="bg-gray-900 border-b border-gray-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-orange-500 rounded-lg flex items-center justify-center font-bold text-sm">P</div>
          <h1 className="text-lg font-semibold text-white">3-Tier Demo App</h1>
          <span className="text-xs bg-gray-800 text-gray-400 px-2 py-0.5 rounded">Prometheus + Grafana</span>
        </div>
        <div className="flex gap-1">
          {NAV_ITEMS.map(item => (
            <button
              key={item}
              onClick={() => setTab(item)}
              className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${
                tab === item
                  ? 'bg-orange-500 text-white'
                  : 'text-gray-400 hover:text-white hover:bg-gray-800'
              }`}
            >
              {item}
            </button>
          ))}
        </div>
      </header>

      {/* Links bar */}
      <div className="bg-gray-900 border-b border-gray-800 px-6 py-2 flex gap-4 text-xs text-gray-500">
        <span>Stack:</span>
        <a href="http://localhost:9090" target="_blank" rel="noreferrer" className="text-orange-400 hover:underline">Prometheus :9090</a>
        <a href="http://localhost:3001" target="_blank" rel="noreferrer" className="text-orange-400 hover:underline">Grafana :3001</a>
        <a href="http://localhost:9093" target="_blank" rel="noreferrer" className="text-orange-400 hover:underline">AlertManager :9093</a>
        <a href="http://localhost:8081" target="_blank" rel="noreferrer" className="text-orange-400 hover:underline">cAdvisor :8081</a>
        <a href="http://localhost:8080/metrics" target="_blank" rel="noreferrer" className="text-orange-400 hover:underline">Raw Metrics</a>
      </div>

      {/* Content */}
      <main className="flex-1 p-6">
        {tab === 'Products' && <ProductList api={api} />}
        {tab === 'Orders'   && <OrderList api={api} />}
        {tab === 'Health'   && <HealthStatus />}
      </main>
    </div>
  )
}
