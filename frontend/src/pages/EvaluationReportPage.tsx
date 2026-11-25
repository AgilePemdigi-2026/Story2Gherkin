import { useEffect, useRef, useState } from 'react'
import { ArrowLeft, ChevronLeft, ChevronRight, Download, FileText, LoaderCircle, X } from 'lucide-react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import PublicHeader from '../components/PublicHeader'
import CircularPercentage from '../components/CircularPercentage'
import ScenarioSidebar from '../components/ScenarioSidebar'
import { HighlightedGherkin } from '../components/review/GherkinEditor'
import {
  deleteEvaluationRun,
  downloadEvaluationReport,
  getEvaluationReport,
  listEvaluationRuns,
  type EvaluationReport,
  type EvaluationRun,
  type EvaluationRunList,
  type EvaluationRunStatus,
  type MethodResult,
} from '../services/evaluationReportApi'
import { demoReport } from './evaluationReportDemo'
import './GenerateScenarioPage.css'
import './EvaluationReportPage.css'
import './EvaluationCircularMetrics.css'

const REPORTS_PATH = '/evaluation/reports'
const runStatusLabel: Record<EvaluationRunStatus, string> = {
  QUEUED: 'Queued',
  RUNNING: 'Running',
  COMPLETED: 'Completed',
  FAILED: 'Failed',
}

const percentFormatter = new Intl.NumberFormat('en-GB', { maximumFractionDigits: 2 })
const dateFormatter = new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'short' })

function formatDate(value: string) {
  const timestamp = Date.parse(value)
  return Number.isNaN(timestamp) ? '—' : dateFormatter.format(timestamp)
}

function MetricRow({ label, llm, rule }: {
  label: string
  llm?: number | null
  rule?: number | null
}) {
  const difference = llm == null || rule == null ? null : Math.round((llm - rule) * 100) / 100
  return (
    <tr>
      <th scope="row">{label}</th>
      <td className="evaluation-report-metric-value"><CircularPercentage value={llm} label={`${label} — LLM`} variant="llm" size="small" /></td>
      <td className="evaluation-report-metric-value"><CircularPercentage value={rule} label={`${label} — Rule-Based`} variant="rule" size="small" /></td>
      <td className={difference != null && difference > 0 ? 'evaluation-report-positive' : undefined}>
        {difference == null ? '—' : `${difference > 0 ? '+' : ''}${percentFormatter.format(difference)} pp`}
      </td>
    </tr>
  )
}

function ResultStatus({ result, method }: { result: MethodResult | null; method?: string }) {
  const status = result?.status
  return (
    <span className={`evaluation-report-status ${status?.toLowerCase() ?? 'pending'}`}>
      {method && `${method}: `}{status === 'PASS' ? 'Passed' : status === 'WARN' ? 'Needs Review' : status === 'FAIL' ? 'Failed' : 'Not Available'}
    </span>
  )
}

function Coverage({ method, result, variant }: { method: string; result: MethodResult | null; variant: 'llm' | 'rule' }) {
  const coverage = result?.gui_coverage_percent
  return (
    <div className={`evaluation-report-coverage ${variant}`}>
      <span>{method}</span>
      <CircularPercentage value={coverage} label={`GUI Element Recall — ${method}`} variant={variant} />
    </div>
  )
}

function RunStatus({ status }: { status: EvaluationRunStatus }) {
  return <span className={`evaluation-run-status ${status.toLowerCase()}`}>{runStatusLabel[status]}</span>
}

function DeleteReportDialog({ run, onClose, onDeleted }: {
  run: EvaluationRun
  onClose: () => void
  onDeleted: () => void
}) {
  const dialog = useRef<HTMLDialogElement>(null)
  const mounted = useRef(true)
  const closingRef = useRef(false)
  const deletingRef = useRef(false)
  const exitAnimation = useRef<Animation | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [closing, setClosing] = useState(false)
  const [error, setError] = useState('')
  const busy = deleting || closing

  useEffect(() => {
    mounted.current = true
    const element = dialog.current
    if (element && !element.open) element.showModal()
    return () => { mounted.current = false; exitAnimation.current?.cancel() }
  }, [])

  async function finishClose(callback: () => void) {
    if (closingRef.current) return
    closingRef.current = true
    setClosing(true)
    const element = dialog.current
    if (element && !window.matchMedia('(prefers-reduced-motion: reduce)').matches && typeof element.animate === 'function') {
      const computed = window.getComputedStyle(element)
      const from = { opacity: computed.opacity, transform: computed.transform }
      element.getAnimations().forEach(animation => animation.cancel())
      exitAnimation.current = element.animate([from, { opacity: '0', transform: 'translateY(5px) scale(0.985)' }], {
        duration: 130, easing: 'ease-in', fill: 'forwards',
      })
      try { await exitAnimation.current.finished } catch { /* Unmount cancellation. */ }
    }
    if (mounted.current) callback()
  }

  function cancel() {
    if (!deletingRef.current && !closingRef.current) void finishClose(onClose)
  }

  async function confirm() {
    if (deletingRef.current || closingRef.current) return
    deletingRef.current = true
    setDeleting(true)
    setError('')
    try {
      await deleteEvaluationRun(run.run_id)
      if (mounted.current) await finishClose(onDeleted)
    } catch (cause) {
      if (mounted.current) setError(cause instanceof Error ? cause.message : 'Unable to delete the evaluation report.')
    } finally {
      deletingRef.current = false
      if (mounted.current) setDeleting(false)
    }
  }

  return (
    <dialog ref={dialog} className="evaluation-report-delete-dialog" data-closing={closing} aria-labelledby="evaluation-report-delete-title" onCancel={event => {
      if (event.target !== event.currentTarget) return
      event.preventDefault()
      cancel()
    }}>
      <div className="evaluation-report-dialog-heading">
        <h2 id="evaluation-report-delete-title">Delete Evaluation Report</h2>
        <button type="button" className="evaluation-report-icon-button" onClick={cancel} disabled={busy} aria-label="Close delete confirmation"><X size={18} aria-hidden="true" /></button>
      </div>
      <p>Delete <strong>{run.report_code}</strong> and its evaluation results? This action cannot be undone.</p>
      {error && <p className="evaluation-report-error" role="alert">{error}</p>}
      <div className="evaluation-report-dialog-actions">
        <button type="button" className="evaluation-report-button secondary" onClick={cancel} disabled={busy}>Cancel</button>
        <button type="button" className="evaluation-report-button danger" onClick={() => { void confirm() }} disabled={busy}>
          {deleting && <LoaderCircle className="evaluation-report-spinner" size={17} aria-hidden="true" />}
          {deleting ? 'Deleting…' : 'Delete Report'}
        </button>
      </div>
    </dialog>
  )
}

function ReportList({ data, page, onDeleted }: { data: EvaluationRunList; page: number; onDeleted: () => void }) {
  const [deleteTarget, setDeleteTarget] = useState<EvaluationRun | null>(null)
  return (
    <>
      <section className="evaluation-report-card" aria-labelledby="evaluation-report-list-title">
        <div className="evaluation-report-card-heading">
          <h2 id="evaluation-report-list-title">Evaluation History</h2>
          <p>Browse saved evaluation runs and inspect their results.</p>
        </div>

        {data.runs.length === 0 ? (
          <div className="evaluation-report-state">
            <FileText size={42} aria-hidden="true" />
            <h3>{page > 1 ? 'No Reports on This Page' : 'No Evaluation Reports Yet'}</h3>
            <p>{page > 1 ? 'Return to the first page to browse available reports.' : 'Run an evaluation to compare both methods and save your first report.'}</p>
            <Link className="evaluation-report-button primary" to={page > 1 ? REPORTS_PATH : '/evaluation'}>{page > 1 ? 'Back to First Page' : 'Run Evaluation'}</Link>
          </div>
        ) : (
          <div className="evaluation-report-table-wrap" role="region" tabIndex={0} aria-label="Evaluation report history">
            <table className="evaluation-report-table evaluation-report-list-table">
              <thead><tr><th scope="col">Report Code</th><th scope="col">Dataset</th><th scope="col">Created</th><th scope="col">Test Cases</th><th scope="col">Status</th><th scope="col">Actions</th></tr></thead>
              <tbody>
                {data.runs.map(run => (
                  <tr key={run.run_id}>
                    <th scope="row">{run.report_code}</th>
                    <td>{run.dataset_name}</td>
                    <td>{formatDate(run.created_at)}</td>
                    <td>{run.total_cases}</td>
                    <td><RunStatus status={run.status} /></td>
                    <td><div className="evaluation-report-row-actions">
                      <Link className="evaluation-report-button secondary compact" to={`${REPORTS_PATH}/${encodeURIComponent(run.run_id)}`} aria-label={`View report ${run.report_code} for ${run.dataset_name}`}>View Report</Link>
                      {run.can_delete === true && <button type="button" className="evaluation-report-button danger-outline compact" onClick={() => setDeleteTarget(run)} aria-label={`Delete report ${run.report_code}`}>Delete</button>}
                    </div></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {(page > 1 || data.total_pages > 1) && (
          <nav className="evaluation-report-pagination" aria-label="Evaluation report pagination">
            {page > 1 && <Link to={`?page=${page - 1}`}><ChevronLeft size={16} aria-hidden="true" />Previous</Link>}
            <span>Page {page}{data.total_pages > 0 ? ` of ${data.total_pages}` : ''}</span>
            {page < data.total_pages && <Link to={`?page=${page + 1}`}>Next<ChevronRight size={16} aria-hidden="true" /></Link>}
          </nav>
        )}
      </section>
      {deleteTarget && <DeleteReportDialog run={deleteTarget} onClose={() => setDeleteTarget(null)} onDeleted={() => { setDeleteTarget(null); onDeleted() }} />}
    </>
  )
}

function ReportDetail({ report, demo, reload }: { report: EvaluationReport; demo: boolean; reload: () => void }) {
  const [selectedId, setSelectedId] = useState('')
  const [downloading, setDownloading] = useState(false)
  const [downloadError, setDownloadError] = useState('')
  const currentDownload = useRef<AbortController | null>(null)
  const caseDetail = useRef<HTMLElement>(null)
  const selected = report.cases.find((item) => item.case_id === selectedId) ?? report.cases[0]
  const canDownload = report.status === 'COMPLETED' && !demo
  const pending = report.status === 'QUEUED' || report.status === 'RUNNING'
  const llm = report.metrics?.llm
  const rule = report.metrics?.rule_based

  useEffect(() => () => currentDownload.current?.abort(), [])

  async function handleDownload() {
    if (!canDownload || currentDownload.current) return
    const controller = new AbortController()
    currentDownload.current = controller
    setDownloading(true)
    setDownloadError('')
    try {
      await downloadEvaluationReport(
        report.run_id,
        controller.signal,
        report.report_code,
      )
    } catch (error) {
      if (!controller.signal.aborted) {
        setDownloadError(error instanceof Error ? error.message : 'Unable to download the evaluation report.')
      }
    } finally {
      currentDownload.current = null
      if (!controller.signal.aborted) setDownloading(false)
    }
  }

  return (
    <div className="evaluation-report-grid">
      <div className="evaluation-report-main">
        <Link className="evaluation-report-back" to={REPORTS_PATH}><ArrowLeft size={16} aria-hidden="true" /> Back to Reports</Link>
        <div className="evaluation-report-heading">
          <div className="generate-page-heading">
            <h1>Evaluation Report {report.report_code}</h1>
            <p>Compare LLM and rule-based results against the Golden Dataset.</p>
          </div>
          <button className="evaluation-report-download" type="button" disabled={!canDownload || downloading} onClick={handleDownload}>
            <Download size={18} aria-hidden="true" /> {downloading ? 'Downloading…' : 'Download CSV'}
          </button>
        </div>
        {demo && <p className="evaluation-report-note">Illustrative report for development. CSV downloads are available for completed evaluation runs.</p>}
        {downloadError && <p className="evaluation-report-error" role="alert">{downloadError}</p>}
        <dl className="evaluation-report-summary">
          <div className="evaluation-report-card"><dt>Dataset</dt><dd>{report.dataset_name}</dd><dd className="evaluation-report-caption">{formatDate(report.created_at)}</dd></div>
          <div className="evaluation-report-card"><dt>Total Cases</dt><dd className="evaluation-report-number">{report.total_cases}</dd><dd className="evaluation-report-caption">Test scenarios</dd></div>
          <div className="evaluation-report-card"><dt>Status</dt><dd><RunStatus status={report.status} /></dd><dd className={report.status === 'COMPLETED' ? 'evaluation-report-positive evaluation-report-caption' : 'evaluation-report-caption'}>
            {pending ? 'Evaluation is still in progress' : report.status === 'COMPLETED' ? 'Evaluation completed' : 'Evaluation failed'}
          </dd></div>
        </dl>
        {pending && <p className="evaluation-report-note">Results will be available when the evaluation is complete. <button type="button" onClick={reload}>Reload Results</button></p>}
        {report.status === 'FAILED' && <p className="evaluation-report-error" role="alert">{report.error_message || 'The evaluation could not be completed.'} <button type="button" onClick={reload}>Reload Results</button></p>}
        <section className="evaluation-report-card" aria-labelledby="evaluation-report-metrics-title">
          <h2 id="evaluation-report-metrics-title">Evaluation Metrics</h2>
          <div className="evaluation-report-table-wrap" tabIndex={0} aria-label="Evaluation metric comparison">
            <table className="evaluation-report-table">
              <thead><tr><th scope="col">Metric</th><th scope="col" className="evaluation-report-metric-method">LLM</th><th scope="col" className="evaluation-report-metric-method">Rule-Based</th><th scope="col">Difference</th></tr></thead>
              <tbody>
                <MetricRow label="Gherkin Validity Rate" llm={llm?.validation_percent} rule={rule?.validation_percent} />
                <MetricRow label="GUI Element Recall" llm={llm?.gui_coverage_percent} rule={rule?.gui_coverage_percent} />
                <MetricRow label="Acceptance-Criteria Precision" llm={llm?.acceptance_criteria_precision_percent} rule={rule?.acceptance_criteria_precision_percent} />
                <MetricRow label="Acceptance-Criteria Recall" llm={llm?.acceptance_criteria_recall_percent} rule={rule?.acceptance_criteria_recall_percent} />
                <MetricRow label="Acceptance-Criteria F1 Score" llm={llm?.acceptance_criteria_f1_percent} rule={rule?.acceptance_criteria_f1_percent} />
              </tbody>
            </table>
          </div>
          {report.metrics === null
            ? <p className="evaluation-report-caption">Metrics are not available yet.</p>
            : <p className="evaluation-report-caption">LLM minus Rule-Based, expressed in percentage points (pp). Acceptance-criteria metrics are shown when available for the dataset.</p>}
        </section>
        <section className="evaluation-report-card" aria-labelledby="evaluation-report-cases-title">
          <h2 id="evaluation-report-cases-title">Scenario Comparison</h2>
          {report.cases.length === 0 ? <p>No case results are available yet.</p> : (
            <div className="evaluation-report-table-wrap" tabIndex={0} aria-label="Evaluation case results">
              <table className="evaluation-report-table evaluation-report-cases">
                <thead><tr><th scope="col">ID</th><th scope="col">Scenario</th><th scope="col">Golden</th><th scope="col">LLM</th><th scope="col">Rule-Based</th><th scope="col">Actions</th></tr></thead>
                <tbody>
                  {report.cases.map((item) => (
                    <tr key={item.case_id} className={selected?.case_id === item.case_id ? 'selected' : undefined}>
                      <td>{item.case_id}</td><th scope="row">{item.title}</th>
                      <td>{item.expected_gherkin.trim() ? <span aria-label="Golden Gherkin available" className="evaluation-report-positive">✓</span> : '—'}</td>
                      <td><ResultStatus result={item.llm} /></td>
                      <td><ResultStatus result={item.rule_based} /></td>
                      <td><button type="button" aria-controls="evaluation-report-case-detail" aria-pressed={selected?.case_id === item.case_id} aria-label={`View details for ${item.title}`} onClick={() => {
                        setSelectedId(item.case_id)
                        caseDetail.current?.focus()
                      }}>View</button></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </div>
      <aside className="evaluation-report-card evaluation-report-case-detail" id="evaluation-report-case-detail" ref={caseDetail} tabIndex={-1} aria-labelledby="evaluation-report-detail-title">
        <h2 id="evaluation-report-detail-title">Case Detail</h2>
        {selected ? (
          <>
            <p className="evaluation-report-case-name" aria-live="polite">{selected.case_id} • {selected.title}</p>
            <h3>Golden Gherkin</h3>
            {selected.expected_gherkin.trim()
              ? <pre tabIndex={0} aria-label="Golden Gherkin"><code><HighlightedGherkin value={selected.expected_gherkin} /></code></pre>
              : <p>Golden Gherkin is not available yet.</p>}
            <h3>Scenario Status</h3>
            <div className="evaluation-report-badges"><ResultStatus result={selected.llm} method="LLM" /><ResultStatus result={selected.rule_based} method="Rule-Based" /></div>
            <h3>GUI Element Recall</h3>
            <Coverage method="LLM" result={selected.llm} variant="llm" />
            <Coverage method="Rule-Based" result={selected.rule_based} variant="rule" />
          </>
        ) : <p>Case details will appear when case results are available.</p>}
      </aside>
    </div>
  )
}

type ReportState =
  | { key: string; kind: 'list'; data: EvaluationRunList }
  | { key: string; kind: 'detail'; data: EvaluationReport }
  | { key: string; kind: 'error'; message: string }

export default function EvaluationReportPage() {
  const { runId } = useParams()
  const [params, setParams] = useSearchParams()
  const requestedPage = Number(params.get('page'))
  const page = !runId && Number.isSafeInteger(requestedPage) && requestedPage > 0 ? requestedPage : 1
  const demo = import.meta.env.DEV && runId === 'demo'
  const [attempt, setAttempt] = useState(0)
  const [settled, setSettled] = useState<ReportState>()
  const requestKey = JSON.stringify([runId, page, attempt])
  const state = settled?.key === requestKey ? settled : undefined
  const reload = () => setAttempt((current) => current + 1)

  useEffect(() => {
    if (demo) return
    const controller = new AbortController()
    async function load() {
      try {
        const next: ReportState = runId
          ? { key: requestKey, kind: 'detail', data: await getEvaluationReport(runId, controller.signal) }
          : { key: requestKey, kind: 'list', data: await listEvaluationRuns(page, controller.signal) }
        if (!controller.signal.aborted) setSettled(next)
      } catch (error) {
        if (!controller.signal.aborted) {
          setSettled({ key: requestKey, kind: 'error', message: error instanceof Error ? error.message : 'Unable to load the evaluation report.' })
        }
      }
    }
    void load()
    return () => controller.abort()
  }, [demo, page, requestKey, runId])

  const report = demo ? demoReport : state?.kind === 'detail' ? state.data : null
  return (
    <div className="generate-page qa-page evaluation-report-page">
      <PublicHeader />
      <div className="generate-layout">
        <ScenarioSidebar />
        <main className="generate-main evaluation-report">
          {report ? <ReportDetail key={requestKey} report={report} demo={demo} reload={reload} /> : (
            <>
              {runId && <Link className="evaluation-report-back" to={REPORTS_PATH}><ArrowLeft size={16} aria-hidden="true" /> Back to Reports</Link>}
              <div className="generate-page-heading">
                <h1>Evaluation Report</h1>
                <p>Compare LLM and rule-based results against the Golden Dataset.</p>
                             </div>
              {!state && <section className="evaluation-report-card evaluation-report-state" role="status"><LoaderCircle className="evaluation-report-spinner" size={32} aria-hidden="true" /><p>Loading evaluation reports…</p></section>}
              {state?.kind === 'error' && <section className="evaluation-report-card evaluation-report-state"><h2>Unable to Load Evaluation Reports</h2><p className="evaluation-report-error" role="alert">{state.message}</p><button type="button" className="evaluation-report-button primary" onClick={reload}>Try Again</button></section>}
              {state?.kind === 'list' && (
                <ReportList data={state.data} page={page} onDeleted={() => {
                  if (page > 1 && state.data.runs.length === 1) {
                    setParams({ page: String(page - 1) }, { replace: true })
                  } else {
                    reload()
                  }
                }} />
              )}
            </>
          )}
        </main>
      </div>
    </div>
  )
}
