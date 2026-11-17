
import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  ArrowLeft,
  ArrowRight,
  FileText,
  Layers,
  Code2,
  Clock3,
  AlertCircle,
  LoaderCircle,
} from 'lucide-react'

import PublicHeader from '../components/PublicHeader'
import ScenarioSidebar from '../components/ScenarioSidebar'

import {
  getGenerationHistoryDetail,
  HistoryApiError,
  type HistoryDetail,
} from '../services/historyApi'

import './GenerateScenarioPage.css'
import './HistoryPage.css'
import './HistoryDetailPage.css'

function formatDate(value: string) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) return '-'

  return date.toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  })
}

function formatStatus(value: string) {
  return value
    .replace(/_/g, ' ')
    .toLowerCase()
    .replace(/\b\w/g, (char) => char.toUpperCase())
}

export default function HistoryDetailPage() {
  const { id } = useParams<{ id: string }>()

  const [history, setHistory] = useState<HistoryDetail | null>(
    null
  )
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [retryCount, setRetryCount] = useState(0)

  // ------------------------------------
  // FETCH DETAIL FROM BACKEND
  // ------------------------------------

  useEffect(() => {
    const controller = new AbortController()

    async function loadDetail() {
      setLoading(true)
      setError(null)
      setHistory(null)

      if (!id) {
        setError(new Error('History ID is missing.'))
        setLoading(false)
        return
      }

      try {
        const result = await getGenerationHistoryDetail(
          id,
          controller.signal
        )

        if (!controller.signal.aborted) {
          setHistory(result)
        }
      } catch (err) {
        if (!controller.signal.aborted) {
          setError(
            err instanceof Error
              ? err
              : new Error('Failed to load history detail.')
          )
        }
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false)
        }
      }
    }

    void loadDetail()

    return () => controller.abort()
  }, [id, retryCount])

  const retryLoadDetail = () => {
    setRetryCount((current) => current + 1)
  }

  const status = history?.request_status.toUpperCase() ?? ''

  const statusClass =
    status === 'SUCCESS'
      ? 'valid'
      : status === 'FAILED'
        ? 'invalid'
        : ''

  // ------------------------------------
  // RENDER
  // ------------------------------------

  return (
    <div className="generate-page history-page history-detail-page">
      <PublicHeader />

      <div className="generate-layout">
        <ScenarioSidebar />

        <main className="generate-main">
          {/* BACK BUTTON */}
          <Link
            to="/history"
            className="history-detail-back"
          >
            <ArrowLeft size={18} />
            Back to History
          </Link>

          {/* PAGE HEADING */}
          <div className="generate-page-heading">
            <h1>Generation History Detail</h1>

            <p>
              Review the original inputs and generated Gherkin
              scenario from your previous request.
            </p>
          </div>

          {/* LOADING */}
          {loading && (
            <section
              className="history-panel history-state"
              role="status"
            >
              <LoaderCircle
                size={34}
                className="history-spinner"
              />

              <p>Loading history details...</p>
            </section>
          )}

          {/* ERROR */}
          {!loading && error && (
            <section
              className="history-panel history-state"
              role="alert"
            >
              <AlertCircle
                size={38}
                color="#dc2626"
              />

              <h3>
                {error instanceof HistoryApiError &&
                error.status === 404
                  ? 'History not found'
                  : error instanceof HistoryApiError &&
                      error.status === 403
                    ? 'Access denied'
                    : 'Unable to load history'}
              </h3>

              <p>{error.message}</p>

              <div className="history-detail-actions">
                <Link
                  to="/history"
                  className="history-button outline"
                >
                  <ArrowLeft size={17} />
                  Back to History
                </Link>

                {error instanceof HistoryApiError &&
                error.status === 401 ? (
                  <Link
                    to="/login"
                    state={{ from: `/history/${id}` }}
                    className="history-button"
                  >
                    Login Again
                    <ArrowRight size={17} />
                  </Link>
                ) : !(error instanceof HistoryApiError &&
                    error.status === 403) && (
                  <button
                    type="button"
                    className="history-button"
                    onClick={retryLoadDetail}
                  >
                    Try Again
                  </button>
                )}
              </div>
            </section>
          )}

          {/* HISTORY DETAIL */}
          {!loading && !error && history && (
            <>
              {/* SUMMARY */}
              <section className="history-panel">
                <div className="history-detail-top">
                  <div>
                    <h2>Generation Summary</h2>
                    <p>
                      Details of your previous generation request.
                    </p>
                  </div>

                  <span
                    className={`history-status ${statusClass}`}
                  >
                    {formatStatus(status)}
                  </span>
                </div>

                <div className="history-detail-metadata">
                  <div>
                    <span>Request ID</span>
                    <strong>{history.id}</strong>
                  </div>

                  <div>
                    <span>Created At</span>
                    <strong>
                      {formatDate(history.created_at)}
                    </strong>
                  </div>

                  <div>
                    <span>Last Updated</span>
                    <strong>
                      {formatDate(history.updated_at)}
                    </strong>
                  </div>
                </div>
              </section>

              {/* ORIGINAL INPUTS */}
              <section className="history-panel">
                <div className="history-detail-section-title">
                  <FileText size={22} color="#2563eb" />

                  <div>
                    <h2>Original Input</h2>
                    <p>
                      Information submitted during generation.
                    </p>
                  </div>
                </div>

                {/* USER STORY */}
                <div className="history-detail-field">
                  <h3>User Story</h3>

                  <div className="history-detail-content">
                    {history.user_story || '-'}
                  </div>
                </div>

                {/* GUI ELEMENTS */}
                <div className="history-detail-field">
                  <div className="history-detail-field-heading">
                    <Layers size={18} />
                    <h3>GUI Elements</h3>
                  </div>

                  <div className="history-detail-content">
                    {history.gui_elements.length > 0 ? (
                      <div className="history-detail-elements">
                        {history.gui_elements.map(
                          (element, index) => (
                            <span
                              className="history-detail-element"
                              key={`${element}-${index}`}
                            >
                              {element}
                            </span>
                          )
                        )}
                      </div>
                    ) : (
                      <p>No GUI elements available.</p>
                    )}
                  </div>
                </div>

                {/* FREE SCENARIO */}
                <div className="history-detail-field">
                  <h3>Free Scenario</h3>

                  <div className="history-detail-content preserve-text">
                    {history.free_scenario || '-'}
                  </div>
                </div>
              </section>

              {/* GENERATED GHERKIN */}
              <section className="history-panel">
                <div className="history-detail-section-title">
                  <Code2 size={22} color="#2563eb" />

                  <div>
                    <h2>Generated Gherkin</h2>
                    <p>
                      Gherkin scenario saved for this request.
                    </p>
                  </div>
                </div>

                {history.scenario ? (
                  <>
                    <div className="history-detail-metadata">
                      <div>
                        <span>Scenario Code</span>
                        <strong>
                          {history.scenario.scenario_code}
                        </strong>
                      </div>

                      <div>
                        <span>Feature Name</span>
                        <strong>
                          {history.scenario.feature_name}
                        </strong>
                      </div>

                      <div>
                        <span>Validation Status</span>
                        <strong>
                          {history.scenario.status}
                        </strong>
                      </div>

                      <div>
                        <span>Coverage</span>
                        <strong>
                          {history.scenario.coverage_percent}%
                        </strong>
                      </div>
                    </div>

                    <pre className="history-detail-gherkin">
                      <code>
                        {history.scenario.gherkin}
                      </code>
                    </pre>
                  </>
                ) : (
                  <div className="history-detail-empty">
                    <AlertCircle size={28} />

                    <h3>No Generated Gherkin</h3>

                    <p>
                      No Gherkin scenario was saved for this
                      generation request.
                    </p>
                  </div>
                )}
              </section>

              {/* ERROR MESSAGE */}
              {history.error_message && (
                <section className="history-panel">
                  <div className="history-detail-section-title">
                    <AlertCircle
                      size={22}
                      color="#dc2626"
                    />

                    <div>
                      <h2>Generation Error</h2>
                      <p>
                        The error recorded during generation.
                      </p>
                    </div>
                  </div>

                  <div
                    className="history-detail-error"
                    role="alert"
                  >
                    {history.error_message}
                  </div>
                </section>
              )}

              {/* FOOTER ACTIONS */}
              <div className="history-detail-footer">
                <span>
                  <Clock3 size={16} />
                  Last updated:{' '}
                  {formatDate(history.updated_at)}
                </span>

                <Link
                  to="/history"
                  className="history-button outline"
                >
                  <ArrowLeft size={17} />
                  Back to History
                </Link>
              </div>
            </>
          )}
        </main>
      </div>
    </div>
  )
}
