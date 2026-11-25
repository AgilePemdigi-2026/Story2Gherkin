import { useEffect, useMemo, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { CheckCircle2, Loader2 } from 'lucide-react'

import PublicHeader from '../components/PublicHeader'
import ScenarioSidebar from '../components/ScenarioSidebar'
import EvaluationDatasetSelect from '../components/EvaluationDatasetSelect'
import CircularPercentage from '../components/CircularPercentage'
import { listGoldenDatasets, type GoldenDatasetSummary } from '../services/goldenDatasetApi'
import {
  createEvaluationRun,
  getEvaluationRunProgress,
  type EvaluationRunCreated,
  type EvaluationRunProgress,
} from '../services/evaluationRunApi'

import './GenerateScenarioPage.css'
import './EvaluationRunPage.css'
import '../styles/EvaluationStatusColors.css'

const POLL_INTERVAL_MS = 1500
const POLL_RETRY_MS = 5000

const STATUS_LABELS = {
  QUEUED: 'Queued',
  RUNNING: 'Running',
  COMPLETED: 'Completed',
  FAILED: 'Failed',
} as const

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'Not available'
  return new Intl.DateTimeFormat('en-GB', {
    day: '2-digit', month: 'short', year: 'numeric',
  }).format(date)
}

function messageFrom(cause: unknown, fallback: string) {
  return cause instanceof Error ? cause.message : fallback
}

export default function EvaluationPage() {
  const [datasets, setDatasets] = useState<GoldenDatasetSummary[]>([])
  const [selectedId, setSelectedId] = useState('')
  const [loading, setLoading] = useState(true)
  const [datasetError, setDatasetError] = useState('')
  const [datasetAttempt, setDatasetAttempt] = useState(0)
  const [selectionError, setSelectionError] = useState('')
  const [runError, setRunError] = useState('')
  const [monitorError, setMonitorError] = useState('')
  const [monitorAttempt, setMonitorAttempt] = useState(0)
  const [run, setRun] = useState<EvaluationRunCreated | null>(null)
  const [progress, setProgress] = useState<EvaluationRunProgress | null>(null)
  const [starting, setStarting] = useState(false)
  const startController = useRef<AbortController | null>(null)
  const startLock = useRef(false)

  useEffect(() => {
    const controller = new AbortController()
    listGoldenDatasets({ pageSize: 100, signal: controller.signal })
      .then(result => {
        if (controller.signal.aborted) return
        setDatasets(result.datasets)
        setDatasetError('')
        setSelectedId(current =>
          result.datasets.some(dataset => dataset.id === current)
            ? current
            : result.datasets[0]?.id ?? '',
        )
      })
      .catch((cause: unknown) => {
        if (!controller.signal.aborted) {
          setDatasetError(messageFrom(cause, 'Unable to load Golden Datasets.'))
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false)
      })
    return () => controller.abort()
  }, [datasetAttempt])

  useEffect(() => () => startController.current?.abort(), [])

  const runId = run?.run_id
  const runStatus = progress?.status

  useEffect(() => {
    if (!runId || runStatus === 'COMPLETED' || runStatus === 'FAILED') return
    const controller = new AbortController()
    let timer: number | undefined

    // Schedule after each response, so slow requests do not overlap.
    async function poll() {
      let delay = POLL_INTERVAL_MS
      try {
        const result = await getEvaluationRunProgress(runId!, controller.signal)
        if (controller.signal.aborted) return
        setProgress(result)
        setMonitorError('')
        if (result.status === 'COMPLETED' || result.status === 'FAILED') return
      } catch (cause) {
        if (controller.signal.aborted) return
        setMonitorError(messageFrom(cause, 'Unable to check evaluation progress.'))
        delay = POLL_RETRY_MS
      }
      if (!controller.signal.aborted) timer = window.setTimeout(() => { void poll() }, delay)
    }

    void poll()
    return () => {
      controller.abort()
      window.clearTimeout(timer)
    }
  }, [runId, runStatus, monitorAttempt])

  const selected = useMemo(
    () => datasets.find(dataset => dataset.id === selectedId) ?? null,
    [datasets, selectedId],
  )
  const finished = runStatus === 'COMPLETED' || runStatus === 'FAILED'
  const active = starting || Boolean(run && !finished)
  const total = Math.max(0, progress?.total_cases ?? selected?.item_count ?? 0)
  const processed = Math.max(0, Math.min(total, progress?.processed_cases ?? 0))

  const completionPercent = runStatus === 'COMPLETED' ? 100 : total > 0 ? processed / total * 100 : null

  async function start() {
    if (startLock.current || !selected || active) return
    if (selected.item_count <= 0) {
      setSelectionError('Choose a dataset containing at least one test case.')
      return
    }

    const controller = new AbortController()
    startController.current = controller
    startLock.current = true
    setStarting(true)
    setSelectionError('')
    setRunError('')
    setMonitorError('')
    setRun(null)
    setProgress(null)

    try {
      const created = await createEvaluationRun(selected.id, controller.signal)
      if (controller.signal.aborted) return
      setRun(created)
      setProgress({
        run_id: created.run_id,
        status: created.status,
        total_cases: created.total_cases,
        processed_cases: created.status === 'COMPLETED' ? created.total_cases : 0,
        error_message: null,
      })
    } catch (cause) {
      if (!controller.signal.aborted) {
        setRunError(messageFrom(cause, 'Unable to start the evaluation.'))
      }
    } finally {
      startLock.current = false
      if (!controller.signal.aborted) setStarting(false)
    }
  }

  return (
    <div className="generate-page qa-page evaluation-run-page">
      <PublicHeader />
      <div className="generate-layout">
        <ScenarioSidebar />
        <main className="generate-main">
          <div className="generate-page-heading">
            <h1>Run Evaluation</h1>
            <p>Select a Golden Dataset to compare LLM and rule-based scenario generation.</p>
          </div>

          <section className="eval-card" aria-labelledby="run-evaluation-title">
            <div className="eval-card-heading">
              <h2 id="run-evaluation-title">Configure Evaluation</h2>
              <p>Choose the reference dataset for this evaluation run.</p>
            </div>

            {datasetError && (
              <div className="eval-error-group">
                <p className="eval-error" role="alert">{datasetError}</p>
                <button type="button" className="eval-secondary" disabled={active} onClick={() => {
                  setLoading(true)
                  setDatasetError('')
                  setDatasetAttempt(value => value + 1)
                }}>Try Again</button>
              </div>
            )}

            <div className="eval-field">
              <label htmlFor="evaluation-dataset">Golden Dataset</label>
              <div className={`eval-select-shell${selectionError ? ' has-error' : ''}`}>
                <EvaluationDatasetSelect
                  id="evaluation-dataset"
                  options={datasets}
                  value={selectedId}
                  disabled={loading || active || Boolean(datasetError)}
                  emptyLabel={loading ? 'Loading datasets…' : datasetError ? 'Datasets unavailable' : 'No datasets available'}
                  invalid={Boolean(selectionError)}
                  describedBy={selectionError ? 'evaluation-dataset-error' : undefined}
                  onChange={value => {
                    setSelectedId(value)
                    setSelectionError('')
                    setRunError('')
                    setMonitorError('')
                    setRun(null)
                    setProgress(null)
                  }}
                />
                {selectionError && <p id="evaluation-dataset-error" className="eval-field-error" role="alert">{selectionError}</p>}
              </div>
            </div>

            {!loading && !datasetError && datasets.length === 0 && (
              <div className="eval-empty-state">
                <h3>No Golden Datasets Yet</h3>
                <p>Create or upload a reference dataset before running an evaluation.</p>
                <Link className="eval-secondary" to="/evaluation/golden-dataset">Manage Golden Datasets</Link>
              </div>
            )}

            <div className="eval-grid">
              <section className="eval-panel" aria-labelledby="evaluation-dataset-info">
                <h3 id="evaluation-dataset-info">Dataset Information</h3>
                {selected ? (
                  <dl className="eval-details">
                    <div><dt>Test Cases</dt><dd>{selected.item_count}</dd></div>
                    <div><dt>Created</dt><dd>{formatDate(selected.created_at)}</dd></div>
                    <div className="eval-details-wide"><dt>Description</dt><dd>{selected.description?.trim() || 'No description provided.'}</dd></div>
                  </dl>
                ) : <p className="eval-muted">Select a dataset to view its details.</p>}
              </section>

              <section className="eval-panel" aria-labelledby="evaluation-metrics-title">
                <h3 id="evaluation-metrics-title">Evaluation Metrics</h3>
                <ul className="eval-metrics">
                  <li><CheckCircle2 size={16} aria-hidden="true" />Gherkin Validity Rate</li>
                  <li><CheckCircle2 size={16} aria-hidden="true" />GUI Element Recall</li>
                  <li><CheckCircle2 size={16} aria-hidden="true" />Acceptance-Criteria Precision</li>
                  <li><CheckCircle2 size={16} aria-hidden="true" />Acceptance-Criteria Recall</li>
                  <li><CheckCircle2 size={16} aria-hidden="true" />Acceptance-Criteria F1 Score</li>
                </ul>
                <p className="eval-note">The evaluation will:</p>
                <ul className="eval-steps">
                  <li>Generate scenarios using the LLM.</li>
                  <li>Generate scenarios using the rule-based method.</li>
                  <li>Compare the results against the Golden Dataset.</li>
                  <li>Save the results in an evaluation report.</li>
                </ul>
              </section>
            </div>

            <section className="eval-method-section" aria-labelledby="evaluation-methods-title">
              <h3 id="evaluation-methods-title">Evaluation Methods</h3>
              <div className="eval-methods">
                <div className="eval-method"><strong>LLM Generator</strong><span>Language model-based scenario generation</span></div>
                <span className="eval-vs">vs</span>
                <div className="eval-method"><strong>Rule-Based Generator</strong><span>Scenario generation using predefined rules</span></div>
              </div>
            </section>

            {runError && <p className="eval-error" role="alert">{runError}</p>}

            {progress && (
              <section className={`eval-progress eval-progress-${progress.status.toLowerCase()}`} aria-labelledby="evaluation-progress-title">
                <div className="eval-progress-heading">
                  <h3 id="evaluation-progress-title">Evaluation Progress</h3>
                  <span className={`eval-status eval-status-${progress.status.toLowerCase()}`}>{STATUS_LABELS[progress.status]}</span>
                </div>
                {run?.report_code && <p className="eval-run-code">Report {run.report_code}</p>}

                <div className="eval-progress-overview">
                  <CircularPercentage value={completionPercent} label="Evaluation progress" kind="progressbar" size="large" />
                  <div className="eval-progress-copy">
                    {progress.status === 'COMPLETED' ? (
                      <>
                        <p role="status">Evaluation completed successfully.</p>
                        <p className="eval-progress-count">{processed} of {total} cases processed</p>
                        <Link className="eval-report-link" to={`/evaluation/reports/${progress.run_id}`}>View Evaluation Report</Link>
                      </>
                    ) : progress.status === 'FAILED' ? (
                      <>
                        <p className="eval-failure-message" role="alert">{progress.error_message || 'The evaluation could not be completed. Try again or open the evaluation reports for more details.'}</p>
                        <p className="eval-progress-count">{processed} of {total} cases processed</p>
                        <Link className="eval-secondary" to={`/evaluation/reports/${progress.run_id}`}>View Evaluation Report</Link>
                      </>
                    ) : (
                      <>
                        <p role="status">{progress.status === 'QUEUED' ? 'Your evaluation is queued.' : `Processing ${processed} of ${total} cases…`}</p>
                        <p className="eval-progress-count">The results will be saved in an evaluation report.</p>
                      </>
                    )}
                  </div>
                </div>

                {monitorError && !finished && (
                  <div className="eval-monitor-error">
                    <p className="eval-error" role="alert">{monitorError} Status checks will retry automatically.</p>
                    <button type="button" className="eval-secondary" onClick={() => setMonitorAttempt(value => value + 1)}>Retry Status Check</button>
                  </div>
                )}
              </section>
            )}

            <button type="button" className="eval-primary" onClick={() => { void start() }} disabled={loading || !selected || active || Boolean(datasetError)}>
              {active && <Loader2 className="eval-spinner" size={18} aria-hidden="true" />}
              {starting ? 'Starting Evaluation…' : active ? 'Evaluation in Progress…' : finished ? 'Run Another Evaluation' : 'Run Evaluation'}
            </button>
          </section>
        </main>
      </div>
    </div>
  )
}
