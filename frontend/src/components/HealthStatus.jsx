import { useState, useEffect } from 'react'

const services = [
  { name: 'Backend API',      url: '/health' },
  { name: 'Backend Ready',    url: '/ready' },
  { name: 'Prometheus',       url: 'http://localhost:9090/-/ready',  external: true },
  { name: 'Grafana',          url: 'http://localhost:3001/api/health', external: true },
]

function StatusBadge({ status }) {
  if (status === 'loading') return <span className="text-xs text-gray-500 animate-pulse">checking…</span>
  if (status === 'up')      return <span className="text-xs bg-green-900/40 border border-green-800 text-green-400 px-2 py-0.5 rounded">UP</span>
  return                           <span className="text-xs bg-red-900/40 border border-red-800 text-red-400 px-2 py-0.5 rounded">DOWN</span>
}

export default function HealthStatus() {
  const [checks, setChecks] = useState(services.map(s => ({ ...s, status: 'loading' })))

  const runChecks = async () => {
    const results = await Promise.allSettled(
      services.map(s =>
        fetch(s.url, { mode: s.external ? 'no-cors' : 'same-origin', signal: AbortSignal.timeout(3000) })
      )
    )
    setChecks(services.map((s, i) => {
      const r = results[i]
      const status = r.status === 'fulfilled' ? 'up' : 'down'
      return { ...s, status }
    }))
  }

  useEffect(() => {
    runChecks()
    const id = setInterval(runChecks, 15000)
    return () => clearInterval(id)
  }, [])

  return (
    <div className="max-w-2xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold">Health Status</h2>
        <button onClick={runChecks} className="text-sm text-gray-400 hover:text-white transition-colors">
          Refresh
        </button>
      </div>

      <div className="space-y-3">
        {checks.map(c => (
          <div key={c.name} className="bg-gray-900 border border-gray-800 rounded-lg px-5 py-3 flex items-center justify-between">
            <span className="text-sm">{c.name}</span>
            <StatusBadge status={c.status} />
          </div>
        ))}
      </div>

      <div className="mt-8 bg-gray-900 border border-gray-800 rounded-lg p-5">
        <h3 className="text-sm font-semibold text-gray-300 mb-4">Monitoring Links</h3>
        <div className="grid grid-cols-2 gap-3 text-sm">
          {[
            ['Prometheus Targets', 'http://localhost:9090/targets'],
            ['Prometheus Alerts', 'http://localhost:9090/alerts'],
            ['Grafana Dashboards', 'http://localhost:3001/dashboards'],
            ['AlertManager', 'http://localhost:9093'],
            ['Raw Metrics', 'http://localhost:8080/metrics'],
            ['cAdvisor', 'http://localhost:8081'],
          ].map(([label, href]) => (
            <a key={label} href={href} target="_blank" rel="noreferrer"
              className="flex items-center gap-2 text-orange-400 hover:text-orange-300 hover:underline">
              <span className="text-gray-600">→</span> {label}
            </a>
          ))}
        </div>
      </div>
    </div>
  )
}
