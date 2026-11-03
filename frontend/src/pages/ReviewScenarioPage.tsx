import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react'
import { Link, Navigate, useBeforeUnload, useBlocker, useLocation, useSearchParams } from 'react-router-dom'
import { AlertCircle, ArrowRight, FileText, LoaderCircle } from 'lucide-react'
import PublicHeader from '../components/PublicHeader'
import ScenarioSidebar from '../components/ScenarioSidebar'
import GherkinEditor from '../components/review/GherkinEditor'
import ScenarioSelector from '../components/review/ScenarioSelector'
import ValidationPanel from '../components/review/ValidationPanel'
import { downloadScenario, getReviewScenario, listReviewScenarios, ReviewApiError, saveCorrection, type ReviewScenario, type ScenarioSummary, type ScenarioValidation } from '../services/reviewApi'
import './GenerateScenarioPage.css'
import './ReviewScenarioPage.css'

type Feedback = { kind: 'success' | 'error'; message: string } | null

function ReviewShell({ children }: { children: ReactNode }) {
  return <div className="generate-page review-page"><PublicHeader /><div className="generate-layout"><ScenarioSidebar /><main className="generate-main"><div className="generate-page-heading review-heading"><h1>Review &amp; Koreksi Skenario</h1><p>Periksa hasil skenario, lakukan koreksi bila diperlukan, simpan perubahan, lalu unduh file .feature.</p></div>{children}</main></div></div>
}

function ErrorState({ error, retry }: { error: Error; retry: () => void }) {
  const status = error instanceof ReviewApiError ? error.status : 0
  return <section className="review-panel review-state" role="alert"><AlertCircle size={36} /><h2>{status === 403 ? 'Akses terbatas' : status === 404 ? 'Skenario tidak ditemukan' : 'Skenario belum dapat dimuat'}</h2><p>{error.message}</p>{status === 401 ? <Link to="/login" state={{ from: window.location.pathname + window.location.search }}>Login kembali <ArrowRight size={17} /></Link> : status !== 403 && <button className="review-button primary" onClick={retry}>Coba lagi</button>}</section>
}

function ReviewWorkspace() {
  const [params, setParams] = useSearchParams()
  const [scenarios, setScenarios] = useState<ScenarioSummary[] | null>(null)
  const [error, setError] = useState<Error | null>(null)
  const [attempt, setAttempt] = useState(0)
  useEffect(() => {
    const controller = new AbortController()
    listReviewScenarios(controller.signal).then(setScenarios).catch((reason: Error) => {
      if (!controller.signal.aborted) setError(reason)
    })
    return () => controller.abort()
  }, [attempt])

  const selected = params.get('scenario') || scenarios?.[0]?.id || ''
  function updateSummary(item: ReviewScenario) {
    setScenarios((current) => current?.map((row) => row.id === item.id ? item : row) || null)
  }
  if (error) return <ReviewShell><ErrorState error={error} retry={() => { setError(null); setScenarios(null); setAttempt((n) => n + 1) }} /></ReviewShell>
  if (!scenarios) return <ReviewShell><div className="review-panel review-state" role="status"><LoaderCircle className="review-spin" size={32} /><p>Memuat daftar skenario…</p></div></ReviewShell>
  if (!selected) return <ReviewShell><section className="review-panel review-state"><FileText size={42} /><h2>Belum ada skenario untuk ditinjau</h2><p>Generate skenario terlebih dahulu untuk melakukan review dan koreksi.</p><Link className="review-button primary" to="/generate">Generate Scenario <ArrowRight size={18} /></Link></section></ReviewShell>
  return <ReviewShell><ScenarioReview key={selected} id={selected} selector={<ScenarioSelector scenarios={scenarios} selected={selected} onSelect={(id) => setParams({ scenario: id })} />} onSaved={updateSummary} /></ReviewShell>
}

function ScenarioReview({ id, selector, onSaved }: { id: string; selector: ReactNode; onSaved: (item: ReviewScenario) => void }) {
  const [scenario, setScenario] = useState<ReviewScenario | null>(null)
  const [error, setError] = useState<Error | null>(null)
  const [draft, setDraft] = useState('')
  const [editing, setEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [downloading, setDownloading] = useState(false)
  const [feedback, setFeedback] = useState<Feedback>(null)
  const [draftValidation, setDraftValidation] = useState<ScenarioValidation | null>(null)
  const [attempt, setAttempt] = useState(0)
  const dialog = useRef<HTMLDialogElement>(null)
  const confirmedLogout = useRef(false)
  const dirty = scenario !== null && draft !== scenario.gherkin
  const blocker = useBlocker(({ currentLocation, nextLocation }) => dirty && !confirmedLogout.current && (currentLocation.pathname !== nextLocation.pathname || currentLocation.search !== nextLocation.search))

  useEffect(() => {
    function beforeLogout(event: Event) {
      if (dirty && !window.confirm('Perubahan belum disimpan. Tetap logout dan tinggalkan perubahan?')) {
        event.preventDefault()
      } else {
        confirmedLogout.current = true
      }
    }
    window.addEventListener('story2gherkin:before-logout', beforeLogout)
    return () => window.removeEventListener('story2gherkin:before-logout', beforeLogout)
  }, [dirty])

  useBeforeUnload(useCallback((event) => {
    if (dirty) { event.preventDefault(); event.returnValue = '' }
  }, [dirty]))

  useEffect(() => {
    const controller = new AbortController()
    getReviewScenario(id, controller.signal).then((item) => {
      if (!controller.signal.aborted) { setScenario(item); setDraft(item.gherkin) }
    }).catch((reason: Error) => { if (!controller.signal.aborted) setError(reason) })
    return () => controller.abort()
  }, [id, attempt])

  useEffect(() => {
    if (blocker.state === 'blocked') dialog.current?.showModal()
    else dialog.current?.close()
  }, [blocker.state])

  async function handleSave() {
    if (!scenario || !editing || !dirty || saving) return
    setSaving(true)
    setFeedback(null)
    try {
      const result = await saveCorrection(scenario, draft)
      setScenario(result)
      setDraft(result.gherkin)
      setDraftValidation(null)
      setEditing(false)
      setFeedback({ kind: 'success', message: 'Correction saved successfully!' })
      onSaved(result)
    } catch (reason) {
      if (reason instanceof ReviewApiError && reason.validation) setDraftValidation(reason.validation)
      setFeedback({ kind: 'error', message: reason instanceof Error ? reason.message : 'Perubahan gagal disimpan. Coba lagi.' })
    } finally { setSaving(false) }
  }

  async function handleDownload() {
    if (!scenario || !scenario.validation.valid || scenario.status !== 'VALID' || downloading) return
    setDownloading(true)
    try { await downloadScenario(scenario) }
    catch (reason) { setFeedback({ kind: 'error', message: reason instanceof Error ? reason.message : 'Download gagal. Coba lagi.' }) }
    finally { setDownloading(false) }
  }

  if (error) return <>{selector}<ErrorState error={error} retry={() => { setError(null); setAttempt((n) => n + 1) }} /></>
  if (!scenario) return <div className="review-panel review-state" role="status"><LoaderCircle className="review-spin" size={32} /><p>Memuat skenario dan validasi…</p></div>

  return <>
    <div className="review-panels">
      <section className="review-panel review-editor-panel" aria-label="Review skenario">
        {selector}
        <div className="review-editor-heading"><h2>Generated Gherkin</h2>{dirty && <span className="review-unsaved">Belum disimpan</span>}</div>
        <GherkinEditor value={draft} editing={editing} disabled={saving} onChange={(value) => { setDraft(value); setDraftValidation(null); setFeedback(null) }} />
        <div className="review-actions">
          <button className="review-button secondary" disabled={saving} onClick={() => {
            if (editing) {
              if (dirty && !window.confirm('Batalkan perubahan yang belum disimpan?')) return
              setDraft(scenario.gherkin)
              setDraftValidation(null)
              setEditing(false)
            } else { setEditing(true) }
            setFeedback(null)
          }}>{editing ? 'Batal Edit' : 'Edit Skenario'}</button>
          <button className="review-button primary" disabled={!editing || !dirty || saving} onClick={handleSave}>{saving ? <><LoaderCircle className="review-spin" size={19} />Menyimpan…</> : 'Simpan Perubahan'}</button>
          <button className="review-button download" disabled={!scenario.validation.valid || scenario.status !== 'VALID' || saving || downloading} onClick={handleDownload}>{downloading ? 'Mengunduh…' : 'Download .feature'}</button>
        </div>
        {dirty && <p className="review-saved-note">Download menggunakan versi valid terakhir yang tersimpan, bukan draft di editor.</p>}
        {feedback?.kind === 'error' && <div className="review-editor-error" role="alert">{feedback.message}</div>}
      </section>
      <ValidationPanel validation={draftValidation || scenario.validation} draft={dirty} validatedDraft={draftValidation !== null} feedback={feedback} />
    </div>
    <dialog ref={dialog} className="review-leave-dialog" aria-labelledby="review-leave-title" onCancel={(event) => { event.preventDefault(); if (blocker.state === 'blocked') blocker.reset() }}>
      <h2 id="review-leave-title">Perubahan belum disimpan</h2><p>Perubahan di editor akan hilang jika Anda berpindah skenario atau meninggalkan halaman ini.</p>
      <div><button className="review-button secondary" autoFocus onClick={() => { if (blocker.state === 'blocked') blocker.reset() }}>Tetap di sini</button><button className="review-button primary" onClick={() => { if (blocker.state === 'blocked') blocker.proceed() }}>Tinggalkan perubahan</button></div>
    </dialog>
  </>
}

export default function ReviewScenarioPage() {
  const location = useLocation()
  const token = sessionStorage.getItem('access_token') || localStorage.getItem('access_token')
  if (!token) return <Navigate to="/login" replace state={{ from: location.pathname + location.search }} />
  return <ReviewWorkspace />
}
