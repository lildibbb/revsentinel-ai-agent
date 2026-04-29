import { useEffect, useMemo, useState } from 'react'
import './App.css'

type LeakCase = {
  id: string
  tenant_id: string
  case_type: string
  status: string
  severity: string
  title: string
  summary: string
  exposure_amount?: number | null
  currency: string
  confidence?: number | null
  created_at: string
}

type CaseListResponse = { items: LeakCase[] }

type IngestResponse = {
  event_id: string
  cases_created: string[]
}

async function fetchJSON<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const res = await fetch(input, init)
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }
  return (await res.json()) as T
}

function App() {
  const [cases, setCases] = useState<LeakCase[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const totalExposure = useMemo(() => {
    return cases.reduce((sum, c) => sum + (c.exposure_amount || 0), 0)
  }, [cases])

  async function refresh() {
    setLoading(true)
    setError(null)
    try {
      const data = await fetchJSON<CaseListResponse>('/api/cases')
      setCases(data.items)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }

  async function createSampleDiscountBreach() {
    setLoading(true)
    setError(null)
    try {
      await fetchJSON<IngestResponse>('/api/ingest', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          event_type: 'discount_event',
          occurred_at: new Date().toISOString(),
          payload: {
            customer_id: 'cust_123',
            contract_id: 'ctr_001',
            invoice_id: 'inv_1001',
            amount: 12000,
            currency: 'USD',
            discount_pct: 0.22,
            allowed_discount_pct: 0.1,
          },
        }),
      })
      await refresh()
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
      setLoading(false)
    }
  }

  useEffect(() => {
    void refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div style={{ maxWidth: 1100, margin: '0 auto', padding: 24 }}>
      <h1>LeakGuard AI (MVP scaffold)</h1>
      <p style={{ opacity: 0.8 }}>
        Minimal end-to-end: ingest event → rules evaluate → case created → UI lists cases.
      </p>

      <div style={{ display: 'flex', gap: 12, marginBottom: 16 }}>
        <button disabled={loading} onClick={() => void refresh()}>
          {loading ? 'Loading…' : 'Refresh'}
        </button>
        <button disabled={loading} onClick={() => void createSampleDiscountBreach()}>
          Create sample discount breach
        </button>
      </div>

      {error ? (
        <pre style={{ background: '#2a1b1b', color: '#ffd6d6', padding: 12, borderRadius: 8 }}>
          {error}
        </pre>
      ) : null}

      <div style={{ margin: '12px 0', opacity: 0.9 }}>
        <strong>Open cases:</strong> {cases.length} &nbsp;|&nbsp; <strong>Total exposure:</strong>{' '}
        {totalExposure.toLocaleString(undefined, { maximumFractionDigits: 2 })}
      </div>

      <div style={{ overflowX: 'auto' }}>
        <table width="100%" cellPadding={8} style={{ borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ textAlign: 'left', borderBottom: '1px solid #333' }}>
              <th>Severity</th>
              <th>Type</th>
              <th>Title</th>
              <th>Status</th>
              <th>Exposure</th>
              <th>Created</th>
            </tr>
          </thead>
          <tbody>
            {cases.map((c) => (
              <tr key={c.id} style={{ borderBottom: '1px solid #222' }}>
                <td>{c.severity}</td>
                <td>{c.case_type}</td>
                <td>{c.title}</td>
                <td>{c.status}</td>
                <td>
                  {c.exposure_amount != null
                    ? `${c.exposure_amount.toLocaleString(undefined, {
                        maximumFractionDigits: 2,
                      })} ${c.currency}`
                    : '-'}
                </td>
                <td>{new Date(c.created_at).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export default App
