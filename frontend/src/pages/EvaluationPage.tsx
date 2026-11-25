import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight, CheckCircle2, Loader2 } from 'lucide-react'
import PublicHeader from '../components/PublicHeader'
import ScenarioSidebar from '../components/ScenarioSidebar'
import { listGoldenDatasets, type GoldenDatasetSummary } from '../services/goldenDatasetApi'
import {
  createEvaluationRun,
  getEvaluationRunProgress,
  type EvaluationRunCreated,
  type EvaluationRunProgress,
} from '../services/evaluationRunApi'
import './GenerateScenarioPage.css'
import './DocumentationPage.css'
import './EvaluationRunPage.css'

const POLL_INTERVAL_MS = 1500

function formatDate(value: string) {
  return new Date(value).toLocaleDateString('id-ID', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

export default function EvaluationPage() {
  const [datasets, setDatasets] = useState<GoldenDatasetSummary[]>([])
  const [selectedId, setSelectedId] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [run, setRun] = useState<EvaluationRunCreated | null>(null)
  const [progress, setProgress] = useState<EvaluationRunProgress | null>(null)
  const [starting, setStarting] = useState(false)

  useEffect(() => {
    const controller = new AbortController()
    listGoldenDatasets({ pageSize: 100, signal: controller.signal })
      .then((result) => {
        setDatasets(result.datasets)
        setSelectedId((current) => current || result.datasets[0]?.id || '')
      })
      .catch((cause: unknown) => {
        if (controller.signal.aborted) return
        setError(cause instanceof Error ? cause.message : 'Gagal memuat Golden Dataset.')
      })
      .finally(() => setLoading(false))
    return () => controller.abort()
  }, [])

  useEffect(() => {
    if (!run || progress?.status === 'COMPLETED' || progress?.status === 'FAILED') return

    const controller = new AbortController()
    const timer = setInterval(() => {
      getEvaluationRunProgress(run.run_id, controller.signal)
        .then(setProgress)
        .catch((cause: unknown) => {
          if (controller.signal.aborted) return
          setError(cause instanceof Error ? cause.message : 'Gagal memantau evaluasi.')
        })
    }, POLL_INTERVAL_MS)

    return () => {
      clearInterval(timer)
      controller.abort()
    }
  }, [run, progress?.status])

  const selected = useMemo(
    () => datasets.find((dataset) => dataset.id === selectedId) ?? null,
    [datasets, selectedId],
  )

  const finished = progress?.status === 'COMPLETED' || progress?.status === 'FAILED'
  const active = starting || (progress !== null && !finished)
  const total = progress?.total_cases ?? selected?.item_count ?? 0
  const processed = progress?.processed_cases ?? 0

  async function start() {
    if (!selected) return
    setStarting(true)
    setError('')
    setProgress(null)
    try {
      const created = await createEvaluationRun(selected.id)
      setRun(created)
      setProgress({
        run_id: created.run_id,
        status: created.status,
        total_cases: created.total_cases,
        processed_cases: 0,
        error_message: null,
      })
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Gagal memulai evaluasi.')
    } finally {
      setStarting(false)
    }
  }

  return (
    <div className="generate-page">
      <PublicHeader />
      <div className="generate-layout">
        <ScenarioSidebar />
        <main className="generate-main">
          <div className="generate-page-heading">
            <h1>Run Evaluation</h1>
            <p>Pilih Golden Dataset lalu jalankan evaluasi untuk membandingkan hasil LLM dan Rule-Based.</p>
          </div>

          <section className="documentation-card" aria-labelledby="run-evaluation-title">
            <h2 id="run-evaluation-title">Pilih Golden Dataset</h2>

            <label className="eval-field">
              <span>Golden Dataset</span>
              <select
                value={selectedId}
                onChange={(event) => setSelectedId(event.target.value)}
                disabled={loading || active}
              >
                {loading && <option value="">Memuat dataset…</option>}
                {!loading && datasets.length === 0 && <option value="">Belum ada dataset</option>}
                {datasets.map((dataset) => (
                  <option key={dataset.id} value={dataset.id}>
                    {dataset.name} — {dataset.item_count} Cases
                  </option>
                ))}
              </select>
            </label>

            <div className="eval-grid">
              <div className="eval-panel">
                <h3>Informasi Dataset</h3>
                {selected ? (
                  <>
                    <dl className="eval-details">
                      <div>
                        <dt>Jumlah Kasus</dt>
                        <dd>{selected.item_count}</dd>
                      </div>
                      <div>
                        <dt>Tanggal Dibuat</dt>
                        <dd>{formatDate(selected.created_at)}</dd>
                      </div>
                      <div className="eval-details-wide">
                        <dt>Deskripsi</dt>
                        <dd>{selected.description?.trim() || 'Tidak ada deskripsi.'}</dd>
                      </div>
                    </dl>
                  </>
                ) : (
                  <p>Pilih dataset untuk melihat detail.</p>
                )}
              </div>

              <div className="eval-panel">
                <h3>Evaluation Metrics</h3>
                <ul className="eval-metrics">
                  <li><CheckCircle2 size={16} /> Gherkin Validity Rate</li>
                  <li><CheckCircle2 size={16} /> GUI Element Recall</li>
                  <li><CheckCircle2 size={16} /> Acceptance-Criteria Precision</li>
                  <li><CheckCircle2 size={16} /> Acceptance-Criteria Recall</li>
                  <li><CheckCircle2 size={16} /> Acceptance-Criteria F1 Score</li>
                </ul>
                <p className="eval-note">Proses akan:</p>
                <ul className="eval-steps">
                  <li>Generate via LLM</li>
                  <li>Generate via Rule-Based</li>
                  <li>Compare with Golden acceptance criteria</li>
                  <li>Save evaluation result</li>
                </ul>
              </div>
            </div>

            <p className="eval-note">Metode Evaluasi</p>
            <div className="eval-methods">
              <div className="eval-method">
                <span className="eval-badge">AI</span>
                <div>
                  <strong>LLM Generator</strong>
                  <span>Menggunakan LLM</span>
                </div>
              </div>
              <span className="eval-vs">VS</span>
              <div className="eval-method eval-method-rule">
                <span className="eval-badge">R</span>
                <div>
                  <strong>Rule-Based Generator</strong>
                  <span>Berdasarkan rule</span>
                </div>
              </div>
            </div>

            {error && <p className="eval-error" role="alert">{error}</p>}

            {progress && (
              <div className="eval-progress" role="status">
                {progress.status === 'COMPLETED' ? (
                  <>
                    <p>Evaluation completed successfully.</p>
                    <Link className="eval-report-link" to={`/evaluation/reports/${progress.run_id}`}>
                      Lihat Evaluation Report <ArrowRight size={16} />
                    </Link>
                  </>
                ) : progress.status === 'FAILED' ? (
                  <p>{progress.error_message || 'Evaluasi gagal diproses.'}</p>
                ) : (
                  <>
                    <p>
                      <Loader2 className="eval-spinner" size={16} /> Memproses {processed} dari {total} kasus…
                    </p>
                    <progress max={total || 1} value={processed} />
                  </>
                )}
              </div>
            )}

            <button
              type="button"
              className="eval-primary"
              onClick={start}
              disabled={!selected || active || finished}
            >
              {active ? 'Menjalankan evaluasi…' : 'Jalankan Evaluasi'}
            </button>

            {!loading && datasets.length === 0 && (
              <p className="eval-note">Belum ada Golden Dataset. Unggah dataset terlebih dahulu.</p>
            )}
          </section>
        </main>
      </div>
    </div>
  )
}
