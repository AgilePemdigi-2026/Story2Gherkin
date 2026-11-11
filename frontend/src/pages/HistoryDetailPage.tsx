
import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  AlertCircle,
  ArrowLeft,
  ArrowRight,
  Code2,
  LoaderCircle,
} from 'lucide-react'

import PublicHeader from '../components/PublicHeader'
import ScenarioSidebar from '../components/ScenarioSidebar'

import {
  getReviewScenario,
  listReviewScenarios,
  type ReviewScenario,
  type ScenarioSummary,
} from '../services/reviewApi'

import './GenerateScenarioPage.css'
import './HistoryDetailPage.css'

export default function HistoryDetailPage() {
  const { id } = useParams<{ id: string }>()

  const [summary, setSummary] = useState<ScenarioSummary | null>(null)
  const [detail, setDetail] = useState<ReviewScenario | null>(null)

  const [loading, setLoading] = useState(true)
  const [detailLoading, setDetailLoading] = useState(true)

  const [error, setError] = useState('')
  const [detailError, setDetailError] = useState('')
  const [retry, setRetry] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    const scenarioId = id

    async function loadScenario() {
      try {
        // Endpoint list tidak memerlukan validator
        const scenarios = await listReviewScenarios(
          controller.signal
        )

        if (controller.signal.aborted) return

        const selected = scenarios.find(
          (item) => item.id === scenarioId
        )

        if (!selected) {
          setError('Skenario tidak ditemukan dalam History.')
          setDetailLoading(false)
          return
        }

        setSummary(selected)
        setLoading(false)

        // Endpoint detail saat ini menggunakan validator
        try {
          const result = await getReviewScenario(
            scenarioId!,
            controller.signal
          )

          if (!controller.signal.aborted) {
            setDetail(result)
          }
        } catch (reason) {
          if (!controller.signal.aborted) {
            setDetailError(
              reason instanceof Error
                ? reason.message
                : 'Gagal memuat detail skenario.'
            )
          }
        }
      } catch (reason) {
        if (!controller.signal.aborted) {
          setError(
            reason instanceof Error
              ? reason.message
              : 'Gagal memuat History.'
          )
        }
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false)
          setDetailLoading(false)
        }
      }
    }

    void loadScenario()

    return () => controller.abort()
  }, [id, retry])

  function retryLoad() {
    setSummary(null)
    setDetail(null)
    setLoading(true)
    setDetailLoading(true)
    setError('')
    setDetailError('')
    setRetry((value) => value + 1)
  }

  const formatDate = (date: string) => {
    const parsed = new Date(date)

    if (Number.isNaN(parsed.getTime())) return '-'

    return parsed.toLocaleString('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'short',
    })
  }

  return (
    <div className="generate-page history-detail-page">
      <PublicHeader />

      <div className="generate-layout">
        <ScenarioSidebar />

        <main className="generate-main">
          <Link to="/history" className="history-detail-back">
            <ArrowLeft size={18} />
            Back to History
          </Link>

          <div className="generate-page-heading">
            <h1>Scenario Detail</h1>
            <p>
              View the information and saved Gherkin
              scenario from your generation history.
            </p>
          </div>

          {loading && (
            <section className="history-detail-card history-detail-state">
              <LoaderCircle
                className="history-detail-spinner"
                size={32}
              />
              <p>Loading scenario...</p>
            </section>
          )}

          {!loading && error && (
            <section
              className="history-detail-card history-detail-state"
              role="alert"
            >
              <AlertCircle size={36} />
              <h2>Unable to load scenario</h2>
              <p>{error}</p>

              <button
                className="history-detail-button"
                onClick={retryLoad}
              >
                Try Again
              </button>
            </section>
          )}

          {!loading && !error && summary && (
            <section className="history-detail-card">
              {/* SCENARIO INFORMATION */}
              <div className="history-detail-header">
                <div>
                  <p className="history-detail-code">
                    {summary.scenario_code || 'Scenario'}
                  </p>

                  <h2>{summary.feature_name}</h2>
                </div>

                <span
                  className={`history-detail-status ${
                    summary.status === 'VALID'
                      ? 'valid'
                      : 'invalid'
                  }`}
                >
                  {summary.status === 'VALID'
                    ? 'Valid'
                    : 'Needs Review'}
                </span>
              </div>

              <div className="history-detail-meta">
                <span>
                  Source: {summary.source_type}
                </span>
                <span>
                  Updated: {formatDate(summary.updated_at)}
                </span>
              </div>

              <div className="history-detail-divider" />

              {/* ORIGINAL INPUTS */}
              <div className="history-detail-section">
                <h3>Original Inputs</h3>

                <div className="history-detail-field">
                  <h4>User Story</h4>

                  <div className="history-detail-value">
                    <p>
                      Belum tersedia pada respons API
                      History saat ini.
                    </p>
                  </div>
                </div>

                <div className="history-detail-field">
                  <h4>GUI Elements</h4>

                  <div className="history-detail-value">
                    {detailLoading ? (
                      <p>Loading...</p>
                    ) : detail?.gui_elements?.length ? (
                      <div className="history-detail-tags">
                        {detail.gui_elements.map(
                          (element, index) => (
                            <span key={index}>
                              {element}
                            </span>
                          )
                        )}
                      </div>
                    ) : (
                      <p>
                        {detail
                          ? 'No GUI elements available.'
                          : 'GUI elements could not be loaded.'}
                      </p>
                    )}
                  </div>
                </div>

                <div className="history-detail-field">
                  <h4>Free Scenario</h4>

                  <div className="history-detail-value">
                    <p>
                      Belum tersedia pada respons API
                      History saat ini.
                    </p>
                  </div>
                </div>
              </div>

              <div className="history-detail-divider" />

              {/* GENERATED GHERKIN */}
              <div className="history-detail-section">
                <div className="history-detail-section-title">
                  <Code2 size={21} />
                  <h3>Generated Gherkin</h3>
                </div>

                {detailLoading && (
                  <div className="history-detail-value">
                    <p>Loading Gherkin...</p>
                  </div>
                )}

                {!detailLoading && detail && (
                  <pre className="history-detail-gherkin">
                    <code>{detail.gherkin}</code>
                  </pre>
                )}

                {!detailLoading && !detail && (
                  <div
                    className="history-detail-alert"
                    role="status"
                  >
                    <AlertCircle size={22} />

                    <div>
                      <strong>
                        Gherkin could not be loaded
                      </strong>

                      <p>
                        {detailError ||
                          'Scenario detail is currently unavailable.'}
                      </p>

                      <p>
                        Your saved scenario has not changed.
                        The current detail API requires
                        the Gherkin validator.
                      </p>

                      <button
                        className="history-detail-retry"
                        onClick={retryLoad}
                      >
                        Try Again
                      </button>
                    </div>
                  </div>
                )}
              </div>

              {/* NAVIGATION TO REVIEW */}
              {detail && (
                <div className="history-detail-actions">
                  <Link
                    className="history-detail-button"
                    to={`/review?scenario=${encodeURIComponent(
                      summary.id
                    )}`}
                  >
                    Open in Review
                    <ArrowRight size={18} />
                  </Link>
                </div>
              )}
            </section>
          )}
        </main>
      </div>
    </div>
  )
}
