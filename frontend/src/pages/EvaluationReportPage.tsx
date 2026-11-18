import { useEffect, useRef, useState } from 'react'
import { ArrowLeft, ChevronLeft, ChevronRight, Download } from 'lucide-react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import PublicHeader from '../components/PublicHeader'
import ScenarioSidebar from '../components/ScenarioSidebar'
import { HighlightedGherkin } from '../components/review/GherkinEditor'
import {
  downloadEvaluationReport,
  getEvaluationReport,
  listEvaluationRuns,
  type EvaluationReport,
  type EvaluationRunList,
  type EvaluationRunStatus,
  type MethodResult,
} from '../services/evaluationReportApi'
import { demoReport } from './evaluationReportDemo'
import './GenerateScenarioPage.css'
import './EvaluationReportPage.css'

const REPORTS_PATH = '/evaluation/reports'
const runStatusLabel: Record<EvaluationRunStatus, string> = {
  QUEUED: 'Menunggu',
  RUNNING: 'Sedang berjalan',
  COMPLETED: 'Completed',
  FAILED: 'Gagal',
}

const percentFormatter = new Intl.NumberFormat('id-ID', { maximumFractionDigits: 2 })
const dateFormatter = new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' })

function formatDate(value: string) {
  const timestamp = Date.parse(value)
  return Number.isNaN(timestamp) ? '—' : dateFormatter.format(timestamp)
}

function formatPercent(value: number | null | undefined) {
  return value == null ? '—' : `${percentFormatter.format(value)}%`
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
      <td><strong>{formatPercent(llm)}</strong></td>
      <td><strong>{formatPercent(rule)}</strong></td>
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
      {method && `${method} `}{status ?? '—'}
    </span>
  )
}

function Coverage({ method, result, variant }: { method: string; result: MethodResult | null; variant: 'llm' | 'rule' }) {
  const coverage = result?.gui_coverage_percent
  return (
    <div className={`evaluation-report-coverage ${variant}`}>
      <div><span>{method}</span><strong>{formatPercent(coverage)}</strong></div>
      {coverage != null && <progress aria-label={`GUI Coverage ${method}`} max={100} value={coverage} />}
    </div>
  )
}

function ReportList({ data, page }: { data: EvaluationRunList; page: number }) {
  return (
    <section className="evaluation-report-card" aria-labelledby="evaluation-report-list-title">
      <h2 id="evaluation-report-list-title">Riwayat Evaluasi</h2>
      {data.runs.length === 0 ? <p>Belum ada hasil evaluasi pada halaman ini.</p> : (
        <div className="evaluation-report-table-wrap" tabIndex={0} aria-label="Daftar hasil evaluasi">
          <table className="evaluation-report-table">
            <thead>
              <tr>
                <th scope="col">Report Code</th>
                <th scope="col">Dataset</th>
                <th scope="col">Waktu</th>
                <th scope="col">Total Kasus</th>
                <th scope="col">Status</th>
                <th scope="col">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {data.runs.map((run) => (
                <tr key={run.run_id}>
                  <th scope="row">{run.report_code}</th>
                  <td>{run.dataset_name}</td>
                  <td>{formatDate(run.created_at)}</td>
                  <td>{run.total_cases}</td>
                  <td>{runStatusLabel[run.status]}</td>
                  <td><Link to={`${REPORTS_PATH}/${encodeURIComponent(run.run_id)}`} aria-label={`Lihat laporan ${run.report_code} untuk ${run.dataset_name}`}>Lihat Report</Link></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {(page > 1 || data.total_pages > 1) && (
        <nav className="evaluation-report-pagination" aria-label="Halaman riwayat evaluasi">
          {page > 1 && <Link to={`?page=${page - 1}`}><ChevronLeft size={16} aria-hidden="true" /> Sebelumnya</Link>}
          <span>Halaman {page}</span>
          {page < data.total_pages && <Link to={`?page=${page + 1}`}>Berikutnya <ChevronRight size={16} aria-hidden="true" /></Link>}
        </nav>
      )}
    </section>
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
        setDownloadError(error instanceof Error ? error.message : 'Laporan gagal diunduh.')
      }
    } finally {
      currentDownload.current = null
      if (!controller.signal.aborted) setDownloading(false)
    }
  }

  return (
    <div className="evaluation-report-grid">
      <div className="evaluation-report-main">
        <Link className="evaluation-report-back" to={REPORTS_PATH}><ArrowLeft size={16} aria-hidden="true" /> Riwayat Evaluasi</Link>
        <div className="evaluation-report-heading">
          <div className="generate-page-heading">
            <h1>Evaluation Report {report.report_code}</h1>
            <p>Ringkasan hasil evaluasi LLM dan Rule-Based terhadap Golden Dataset.</p>
          </div>
          <button className="evaluation-report-download" type="button" disabled={!canDownload || downloading} onClick={handleDownload}>
            <Download size={18} aria-hidden="true" /> {downloading ? 'Mengunduh…' : 'Unduh Laporan'}
          </button>
        </div>
        {demo && <p className="evaluation-report-note">Data contoh untuk pengembangan. Unduhan tersedia pada hasil evaluasi sebenarnya.</p>}
        {downloadError && <p className="evaluation-report-error" role="alert">{downloadError}</p>}
        <dl className="evaluation-report-summary">
          <div className="evaluation-report-card"><dt>Dataset</dt><dd>{report.dataset_name}</dd><dd className="evaluation-report-caption">{formatDate(report.created_at)}</dd></div>
          <div className="evaluation-report-card"><dt>Total Cases</dt><dd className="evaluation-report-number">{report.total_cases}</dd><dd className="evaluation-report-caption">Skenario pengujian</dd></div>
          <div className="evaluation-report-card"><dt>Status</dt><dd>{runStatusLabel[report.status]}</dd><dd className={report.status === 'COMPLETED' ? 'evaluation-report-positive evaluation-report-caption' : 'evaluation-report-caption'}>
            {pending ? 'Evaluasi belum selesai' : report.status === 'COMPLETED' ? 'Evaluasi selesai' : 'Evaluasi gagal'}
          </dd></div>
        </dl>
        {pending && <p className="evaluation-report-note">Hasil akan tersedia setelah evaluasi selesai. <button type="button" onClick={reload}>Muat ulang hasil</button></p>}
        {report.status === 'FAILED' && <p className="evaluation-report-error" role="alert">{report.error_message || 'Evaluasi gagal diproses.'} <button type="button" onClick={reload}>Muat ulang hasil</button></p>}
        <section className="evaluation-report-card" aria-labelledby="evaluation-report-metrics-title">
          <h2 id="evaluation-report-metrics-title">Evaluation Metrics</h2>
          <div className="evaluation-report-table-wrap" tabIndex={0} aria-label="Perbandingan metrik">
            <table className="evaluation-report-table">
              <thead><tr><th scope="col">Metric</th><th scope="col">LLM</th><th scope="col">Rule-Based</th><th scope="col">Selisih</th></tr></thead>
              <tbody>
                <MetricRow label="Validation Metric" llm={llm?.validation_percent} rule={rule?.validation_percent} />
                <MetricRow label="GUI Coverage" llm={llm?.gui_coverage_percent} rule={rule?.gui_coverage_percent} />
              </tbody>
            </table>
          </div>
          {report.metrics === null
            ? <p className="evaluation-report-caption">Metrik belum tersedia.</p>
            : <p className="evaluation-report-caption">Selisih LLM dan Rule-Based dalam poin persentase (pp).</p>}
        </section>
        <section className="evaluation-report-card" aria-labelledby="evaluation-report-cases-title">
          <h2 id="evaluation-report-cases-title">Hasil Perbandingan Skenario</h2>
          {report.cases.length === 0 ? <p>Belum ada hasil per kasus.</p> : (
            <div className="evaluation-report-table-wrap" tabIndex={0} aria-label="Hasil per kasus">
              <table className="evaluation-report-table evaluation-report-cases">
                <thead><tr><th scope="col">ID</th><th scope="col">Skenario</th><th scope="col">Golden</th><th scope="col">LLM</th><th scope="col">Rule-Based</th><th scope="col">Aksi</th></tr></thead>
                <tbody>
                  {report.cases.map((item) => (
                    <tr key={item.case_id} className={selected?.case_id === item.case_id ? 'selected' : undefined}>
                      <td>{item.case_id}</td><th scope="row">{item.title}</th>
                      <td>{item.expected_gherkin.trim() ? <span aria-label="Golden Gherkin tersedia" className="evaluation-report-positive">✓</span> : '—'}</td>
                      <td><ResultStatus result={item.llm} /></td>
                      <td><ResultStatus result={item.rule_based} /></td>
                      <td><button type="button" aria-controls="evaluation-report-case-detail" aria-pressed={selected?.case_id === item.case_id} aria-label={`Lihat detail ${item.title}`} onClick={() => {
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
              : <p>Golden Gherkin belum tersedia.</p>}
            <h3>Scenario Status</h3>
            <div className="evaluation-report-badges"><ResultStatus result={selected.llm} method="LLM" /><ResultStatus result={selected.rule_based} method="Rule-Based" /></div>
            <h3>GUI Coverage</h3>
            <Coverage method="LLM" result={selected.llm} variant="llm" />
            <Coverage method="Rule-Based" result={selected.rule_based} variant="rule" />
          </>
        ) : <p>Detail akan tampil setelah hasil kasus tersedia.</p>}
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
  const [params] = useSearchParams()
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
          setSettled({ key: requestKey, kind: 'error', message: error instanceof Error ? error.message : 'Laporan gagal dimuat.' })
        }
      }
    }
    void load()
    return () => controller.abort()
  }, [demo, page, requestKey, runId])

  const report = demo ? demoReport : state?.kind === 'detail' ? state.data : null
  return (
    <div className="generate-page">
      <PublicHeader />
      <div className="generate-layout">
        <ScenarioSidebar />
        <main className="generate-main evaluation-report">
          {report ? <ReportDetail key={requestKey} report={report} demo={demo} reload={reload} /> : (
            <>
              {runId && <Link className="evaluation-report-back" to={REPORTS_PATH}><ArrowLeft size={16} aria-hidden="true" /> Riwayat Evaluasi</Link>}
              <div className="generate-page-heading">
                <h1>Evaluation Report</h1>
                <p>Ringkasan hasil evaluasi LLM dan Rule-Based terhadap Golden Dataset.</p>
                {import.meta.env.DEV && !runId && <Link to={`${REPORTS_PATH}/demo`}>Buka contoh laporan</Link>}
              </div>
              {!state && <p className="evaluation-report-card" role="status">Memuat hasil evaluasi…</p>}
              {state?.kind === 'error' && <section className="evaluation-report-card"><p role="alert">{state.message}</p><button type="button" onClick={reload}>Coba Lagi</button></section>}
              {state?.kind === 'list' && <ReportList data={state.data} page={page} />}
            </>
          )}
        </main>
      </div>
    </div>
  )
}
