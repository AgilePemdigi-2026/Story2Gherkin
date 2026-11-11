
import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Search,
  History,
  ArrowRight,
  LoaderCircle,
  AlertCircle,
  FileText,
  ChevronDown,
} from 'lucide-react'

import PublicHeader from '../components/PublicHeader'
import ScenarioSidebar from '../components/ScenarioSidebar'
import {
  listReviewScenarios,
  type ScenarioSummary,
} from '../services/reviewApi'

import './GenerateScenarioPage.css'
import './HistoryPage.css'

type StatusFilter = 'ALL' | 'VALID' | 'INVALID'
const statusOptions: {
  value: StatusFilter
  label: string
}[] = [
  { value: 'ALL', label: 'All Status' },
  { value: 'VALID', label: 'Valid' },
  { value: 'INVALID', label: 'Needs Review' },
]

export default function HistoryPage() {
  const [scenarios, setScenarios] = useState<ScenarioSummary[]>([])
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<StatusFilter>('ALL')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [retry, setRetry] = useState(0)
  const [filterOpen, setFilterOpen] = useState(false)
const filterRef = useRef<HTMLDivElement>(null)

useEffect(() => {
  const handleClickOutside = (event: MouseEvent) => {
    if (
      filterRef.current &&
      !filterRef.current.contains(event.target as Node)
    ) {
      setFilterOpen(false)
    }
  }

  const handleEscape = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      setFilterOpen(false)
    }
  }

  document.addEventListener('mousedown', handleClickOutside)
  document.addEventListener('keydown', handleEscape)

  return () => {
    document.removeEventListener('mousedown', handleClickOutside)
    document.removeEventListener('keydown', handleEscape)
  }
}, [])

  // Mengambil skenario yang tersimpan dari backend
  useEffect(() => {
    const controller = new AbortController()

    setLoading(true)
    setError('')

    listReviewScenarios(controller.signal)
      .then((data) => {
        if (!controller.signal.aborted) {
          setScenarios(data)
        }
      })
      .catch((reason: Error) => {
        if (!controller.signal.aborted) {
          setError(reason.message)
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setLoading(false)
        }
      })

    return () => controller.abort()
  }, [retry])

  // Filter berdasarkan judul, kode, dan status
  const filteredScenarios = scenarios.filter((scenario) => {
    const keyword = search.toLowerCase().trim()

    const matchesSearch =
      scenario.feature_name.toLowerCase().includes(keyword) ||
      (scenario.scenario_code ?? '').toLowerCase().includes(keyword)

    const matchesStatus =
      filter === 'ALL' || scenario.status === filter

    return matchesSearch && matchesStatus
  })

  // Format tanggal
  const formatDate = (date: string) => {
    const parsed = new Date(date)

    if (Number.isNaN(parsed.getTime())) return '-'

    return parsed.toLocaleString('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'short',
    })
  }

  return (
    <div className="generate-page history-page">
      <PublicHeader />

      <div className="generate-layout">
        <ScenarioSidebar />

        <main className="generate-main">
          <div className="generate-page-heading">
            <h1>Generation History</h1>
            <p>
              Browse your previously generated Gherkin scenarios
              and view their details.
            </p>
          </div>

          <section className="history-panel">
            <div className="history-header">
              <div>
                <h2>Scenario History</h2>
                <p>
                  {scenarios.length} saved scenarios
                </p>
              </div>

              <History size={26} color="#2563eb" />
            </div>

            {/* SEARCH AND FILTER */}
            <div className="history-toolbar">
              <div className="history-search">
                <Search size={19} />
                <input
                  type="search"
                  placeholder="Search scenarios..."
                  aria-label="Search scenarios"
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                />
              </div>

              <div className="history-filter-wrap" ref={filterRef}>
  <button
    type="button"
    className="history-filter-trigger"
    aria-label="Filter scenario status"
    aria-expanded={filterOpen}
    onClick={() => setFilterOpen(!filterOpen)}
  >
    <span>
      {statusOptions.find((option) => option.value === filter)?.label}
    </span>

    <ChevronDown size={17} />
  </button>

  {filterOpen && (
    <div className="history-filter-menu">
      {statusOptions.map((option) => (
        <button
          key={option.value}
          type="button"
          className={`history-filter-option ${
            filter === option.value ? 'selected' : ''
          }`}
          onClick={() => {
            setFilter(option.value)
            setFilterOpen(false)
          }}
        >
          {option.label}
        </button>
      ))}
    </div>
  )}
</div>
            </div>

            {/* LOADING STATE */}
            {loading && (
              <div className="history-state" role="status">
                <LoaderCircle className="history-spinner" size={32} />
                <p>Loading your history...</p>
              </div>
            )}

            {/* ERROR STATE */}
            {!loading && error && (
              <div className="history-state" role="alert">
                <AlertCircle size={36} color="#dc2626" />
                <h3>Unable to load history</h3>
                <p>{error}</p>

                <button
                  className="history-button"
                  onClick={() => setRetry((value) => value + 1)}
                >
                  Try Again
                </button>
              </div>
            )}

            {/* EMPTY STATE */}
            {!loading && !error && filteredScenarios.length === 0 && (
              <div className="history-state">
                <FileText size={42} color="#94a3b8" />

                <h3>
                  {scenarios.length === 0
                    ? 'No scenarios yet'
                    : 'No matching scenarios'}
                </h3>

                <p>
                  {scenarios.length === 0
                    ? 'Generate your first scenario to see it here.'
                    : 'Try another search or status filter.'}
                </p>

                {scenarios.length === 0 && (
                  <Link to="/generate" className="history-button">
                    Generate Scenario
                    <ArrowRight size={17} />
                  </Link>
                )}
              </div>
            )}

            {/* HISTORY LIST */}
            {!loading && !error && filteredScenarios.length > 0 && (
              <div className="history-list">
                {filteredScenarios.map((scenario) => (
                  <article className="history-card" key={scenario.id}>
                    <div className="history-card-content">
                      <div className="history-card-title">
                        <h3>{scenario.feature_name}</h3>

                        <span
                          className={`history-status ${
                            scenario.status === 'VALID'
                              ? 'valid'
                              : 'invalid'
                          }`}
                        >
                          {scenario.status === 'VALID'
                            ? 'Valid'
                            : 'Needs Review'}
                        </span>
                      </div>

                      <p className="history-code">
                        {scenario.scenario_code || 'Scenario'}
                      </p>

                      <p className="history-date">
                        Last updated: {formatDate(scenario.updated_at)}
                      </p>
                    </div>

                    <Link
                    to={`/history/${encodeURIComponent(scenario.id)}`}
                    className="history-button outline"
                    >
                    View Detail
                    <ArrowRight size={17} />
                    </Link>
                  </article>
                ))}
              </div>
            )}
          </section>
        </main>
      </div>
    </div>
  )
}
