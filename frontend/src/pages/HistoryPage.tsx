
import { useEffect, useMemo, useRef, useState } from 'react'
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
  listGenerationHistory,
  type HistorySummary,
} from '../services/historyApi'

import './GenerateScenarioPage.css'
import './HistoryPage.css'

// Format status dari backend
const formatStatus = (status: string) => {
  switch (status.toUpperCase()) {
    case 'SUCCESS':
      return 'Success'
    case 'FAILED':
      return 'Failed'
    case 'PENDING':
      return 'Pending'
    default:
      return status
        .replace(/_/g, ' ')
        .toLowerCase()
        .replace(/\b\w/g, (char) => char.toUpperCase())
  }
}

// Format tanggal
const formatDate = (date: string) => {
  const parsed = new Date(date)

  if (Number.isNaN(parsed.getTime())) return '-'

  return parsed.toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  })
}

export default function HistoryPage() {
  // ------------------------------------
  // HISTORY DATA
  // ------------------------------------

  const [histories, setHistories] = useState<HistorySummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [retryCount, setRetryCount] = useState(0)

  // ------------------------------------
  // SEARCH AND FILTER
  // ------------------------------------

  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState('ALL')
  const [filterOpen, setFilterOpen] = useState(false)

  const filterRef = useRef<HTMLDivElement>(null)

  // ------------------------------------
  // LOAD HISTORY FROM BACKEND
  // ------------------------------------

  useEffect(() => {
    const controller = new AbortController()

    async function loadHistory() {
      setLoading(true)
      setError(null)

      try {
        // Fetch first page
        const first = await listGenerationHistory(
          1,
          100,
          controller.signal
        )

        const records = [...first.history]

        // Fetch remaining pages
        for (let page = 2; page <= first.total_pages; page++) {
          if (controller.signal.aborted) return

          const result = await listGenerationHistory(
            page,
            100,
            controller.signal
          )

          records.push(...result.history)
        }

        if (!controller.signal.aborted) {
          setHistories(records)
        }
      } catch (err) {
        if (!controller.signal.aborted) {
          setError(
            err instanceof Error
              ? err.message
              : 'Failed to load generation history.'
          )
        }
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false)
        }
      }
    }

    void loadHistory()

    return () => controller.abort()
  }, [retryCount])

  // ------------------------------------
  // RETRY
  // ------------------------------------

  const retryLoadHistory = () => {
    setRetryCount((current) => current + 1)
  }

  // ------------------------------------
  // CLOSE DROPDOWN WHEN CLICKING OUTSIDE
  // ------------------------------------

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        filterRef.current &&
        !filterRef.current.contains(event.target as Node)
      ) {
        setFilterOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)

    return () => {
      document.removeEventListener(
        'mousedown',
        handleClickOutside
      )
    }
  }, [])

  // ------------------------------------
  // STATUS OPTIONS
  // ------------------------------------

  const statusOptions = [
  { value: 'ALL', label: 'All Status' },
  { value: 'SUCCESS', label: 'Success' },
  ]

  // ------------------------------------
  // FILTER HISTORY
  // ------------------------------------

  const filteredHistories = useMemo(() => {
    const keyword = search.trim().toLowerCase()

    return histories.filter((history) => {
      const matchesSearch =
        history.user_story.toLowerCase().includes(keyword) ||
        history.id.toLowerCase().includes(keyword)

      const matchesStatus =
        filter === 'ALL' ||
        history.request_status.toUpperCase() === filter

      return matchesSearch && matchesStatus
    })
  }, [histories, search, filter])

  // ------------------------------------
  // RENDER
  // ------------------------------------

  return (
    <div className="generate-page history-page">
      <PublicHeader />

      <div className="generate-layout">
        <ScenarioSidebar />

        <main className="generate-main">
          {/* PAGE HEADING */}
          <div className="generate-page-heading">
            <h1>Generation History</h1>
            <p>
              Browse your previously generated Gherkin scenarios
              and view their details.
            </p>
          </div>

          <section className="history-panel">
            {/* HISTORY HEADER */}
            <div className="history-header">
              <div>
                <h2>Scenario History</h2>
                <p>
                  {histories.length} saved generation requests
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
                  onChange={(event) =>
                    setSearch(event.target.value)
                  }
                />
              </div>

              {/* CUSTOM STATUS DROPDOWN */}
              <div
                className="history-filter-wrap"
                ref={filterRef}
              >
                <button
                  type="button"
                  className="history-filter-trigger"
                  aria-label="Filter scenario status"
                  aria-expanded={filterOpen}
                  onClick={() =>
                    setFilterOpen((current) => !current)
                  }
                >
                  <span>
                    {statusOptions.find(
                      (option) => option.value === filter
                    )?.label ?? 'All Status'}
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
                          filter === option.value
                            ? 'selected'
                            : ''
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
              <div
                className="history-state"
                role="status"
              >
                <LoaderCircle
                  className="history-spinner"
                  size={32}
                />

                <p>Loading your history...</p>
              </div>
            )}

            {/* ERROR STATE */}
            {!loading && error && (
              <div
                className="history-state"
                role="alert"
              >
                <AlertCircle
                  size={36}
                  color="#dc2626"
                />

                <h3>Unable to load history</h3>

                <p>{error}</p>

                <button
                  type="button"
                  className="history-button"
                  onClick={retryLoadHistory}
                >
                  Try Again
                </button>
              </div>
            )}

            {/* EMPTY STATE */}
            {!loading &&
              !error &&
              filteredHistories.length === 0 && (
                <div className="history-state">
                  <FileText
                    size={42}
                    color="#94a3b8"
                  />

                  <h3>
                    {histories.length === 0
                      ? 'No scenarios yet'
                      : 'No matching scenarios'}
                  </h3>

                  <p>
                    {histories.length === 0
                      ? 'Generate your first scenario to see it here.'
                      : 'Try another search or status filter.'}
                  </p>

                  {histories.length === 0 && (
                    <Link
                      to="/generate"
                      className="history-button"
                    >
                      Generate Scenario
                      <ArrowRight size={17} />
                    </Link>
                  )}
                </div>
              )}

            {/* HISTORY LIST */}
            {!loading &&
              !error &&
              filteredHistories.length > 0 && (
                <div className="history-list">
                  {filteredHistories.map((history) => {
                    const status =
                      history.request_status.toUpperCase()

                    const statusClass =
                      status === 'SUCCESS'
                        ? 'valid'
                        : status === 'FAILED'
                          ? 'invalid'
                          : ''

                    return (
                      <article
                        className="history-card"
                        key={history.id}
                      >
                        <div className="history-card-content">
                          <div className="history-card-title">
                            <h3>
                              {history.user_story}
                            </h3>

                            <span
                              className={`history-status ${statusClass}`}
                            >
                              {formatStatus(status)}
                            </span>
                          </div>

                          <p className="history-code">
                            Request ID:{' '}
                            {history.id.slice(0, 8)}
                          </p>

                          <p className="history-date">
                            Last updated:{' '}
                            {formatDate(history.updated_at)}
                          </p>
                        </div>

                        {/* SINGLE VIEW DETAIL BUTTON */}
                        <Link
                          to={`/history/${encodeURIComponent(
                            history.id
                          )}`}
                          className="history-button outline"
                        >
                          View Detail
                          <ArrowRight size={17} />
                        </Link>
                      </article>
                    )
                  })}
                </div>
              )}
          </section>
        </main>
      </div>
    </div>
  )
}
