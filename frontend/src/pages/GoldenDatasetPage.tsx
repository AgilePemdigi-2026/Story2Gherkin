import { useEffect, useRef, useState, type FormEvent, type ReactNode } from 'react'
import {
  AlertCircle,
  ChevronLeft,
  ChevronRight,
  Database,
  Eye,
  LoaderCircle,
  Pencil,
  Plus,
  Search,
  Trash2,
  Upload,
  X,
} from 'lucide-react'

import PublicHeader from '../components/PublicHeader'
import ScenarioSidebar from '../components/ScenarioSidebar'
import {
  GoldenDatasetApiError,
  deleteGoldenDataset,
  getGoldenDataset,
  listGoldenDatasets,
  updateGoldenDataset,
  uploadGoldenDataset,
  type GoldenDatasetDetail,
  type GoldenDatasetList,
  type GoldenDatasetSummary,
} from '../services/goldenDatasetApi'

import './GenerateScenarioPage.css'
import './GoldenDatasetPage.css'

const PAGE_SIZE = 10

type ListState = { data: GoldenDatasetList | null; error: string; loading: boolean }

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(date)
}

function Shell({ children }: { children: ReactNode }) {
  return (
    <div className="generate-page">
      <PublicHeader />
      <div className="generate-layout">
        <ScenarioSidebar />
        <main className="generate-main">{children}</main>
      </div>
    </div>
  )
}

export default function GoldenDatasetPage() {
  const [search, setSearch] = useState('')
  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [attempt, setAttempt] = useState(0)
  const [state, setState] = useState<ListState>({ data: null, error: '', loading: true })
  const [uploadOpen, setUploadOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<GoldenDatasetSummary | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<GoldenDatasetSummary | null>(null)
  const [detailId, setDetailId] = useState('')

  // Debounce typing so the list is not refetched on every keystroke.
  useEffect(() => {
    const timer = window.setTimeout(() => {
      setPage(1)
      setQuery(search)
    }, 300)
    return () => window.clearTimeout(timer)
  }, [search])

  useEffect(() => {
    const controller = new AbortController()
    listGoldenDatasets({ search: query, page, pageSize: PAGE_SIZE, signal: controller.signal })
      .then((data) => {
        if (!controller.signal.aborted) setState({ data, error: '', loading: false })
      })
      .catch((reason: Error) => {
        if (!controller.signal.aborted) {
          setState((current) => ({ ...current, error: reason.message, loading: false }))
        }
      })
    return () => controller.abort()
  }, [query, page, attempt])

  const datasets = state.data?.datasets || []
  const totalPages = state.data?.total_pages || 0

  return (
    <Shell>
      <div className="golden-heading-row">
        <div className="generate-page-heading">
          <h1>Golden Dataset</h1>
          <p>Kelola dataset Golden Gherkin yang digunakan sebagai acuan dalam proses evaluasi.</p>
        </div>
        <button type="button" className="golden-button primary" onClick={() => setUploadOpen(true)}>
          <Plus size={18} /> Tambah Dataset
        </button>
      </div>

      <label className="golden-search">
        <Search size={19} aria-hidden="true" />
        <input
          type="search"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Cari dataset..."
          aria-label="Cari dataset"
        />
      </label>

      <section className="golden-panel" aria-live="polite">
        {state.loading && (
          <div className="golden-state" role="status">
            <LoaderCircle className="golden-spin" size={28} />
            <p>Memuat Golden Dataset…</p>
          </div>
        )}

        {!state.loading && state.error && (
          <div className="golden-state" role="alert">
            <AlertCircle size={30} />
            <p>{state.error}</p>
            <button type="button" className="golden-button secondary" onClick={() => setAttempt((value) => value + 1)}>
              Coba lagi
            </button>
          </div>
        )}

        {!state.loading && !state.error && datasets.length === 0 && (
          <div className="golden-state">
            <Database size={34} />
            <h2>{query ? 'Dataset tidak ditemukan' : 'Belum ada Golden Dataset'}</h2>
            <p>
              {query
                ? 'Coba kata kunci lain atau kosongkan pencarian.'
                : 'Unggah file JSON berisi skenario acuan untuk memulai evaluasi.'}
            </p>
            {!query && (
              <button type="button" className="golden-button primary" onClick={() => setUploadOpen(true)}>
                <Plus size={18} /> Tambah Dataset
              </button>
            )}
          </div>
        )}

        {!state.loading && !state.error && datasets.length > 0 && (
          <>
            <table className="golden-table">
              <thead>
                <tr>
                  <th scope="col">Nama Dataset</th>
                  <th scope="col" className="golden-actions-heading">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {datasets.map((dataset) => (
                  <tr key={dataset.id}>
                    <td>
                      <span className="golden-name">{dataset.name}</span>
                      <span className="golden-meta">
                        {dataset.item_count} item &middot; dibuat {formatDate(dataset.created_at)}
                      </span>
                    </td>
                    <td>
                      <div className="golden-actions">
                        <button type="button" className="golden-action view" onClick={() => setDetailId(dataset.id)}>
                          <Eye size={15} /> View
                        </button>
                        <button
                          type="button"
                          className="golden-action"
                          onClick={() => setEditTarget(dataset)}
                          disabled={Boolean(dataset.used_at)}
                          title={dataset.used_at ? 'Sudah dipakai evaluasi, tidak dapat diubah' : 'Edit dataset'}
                        >
                          <Pencil size={15} /> Edit
                        </button>
                        <button
                          type="button"
                          className="golden-action delete"
                          onClick={() => setDeleteTarget(dataset)}
                          disabled={Boolean(dataset.used_at)}
                          title={dataset.used_at ? 'Sudah dipakai evaluasi, tidak dapat dihapus' : 'Hapus dataset'}
                        >
                          <Trash2 size={15} /> Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {totalPages > 1 && (
              <nav className="golden-pagination" aria-label="Navigasi halaman dataset">
                <button
                  type="button"
                  className="golden-page-arrow"
                  onClick={() => setPage((value) => Math.max(1, value - 1))}
                  disabled={page <= 1}
                  aria-label="Halaman sebelumnya"
                >
                  <ChevronLeft size={16} />
                </button>
                {Array.from({ length: totalPages }, (_, index) => index + 1).map((number) => (
                  <button
                    key={number}
                    type="button"
                    className={`golden-page-number ${number === page ? 'active' : ''}`}
                    onClick={() => setPage(number)}
                    aria-current={number === page ? 'page' : undefined}
                  >
                    {number}
                  </button>
                ))}
                <button
                  type="button"
                  className="golden-page-arrow"
                  onClick={() => setPage((value) => Math.min(totalPages, value + 1))}
                  disabled={page >= totalPages}
                  aria-label="Halaman berikutnya"
                >
                  <ChevronRight size={16} />
                </button>
              </nav>
            )}
          </>
        )}
      </section>

      {uploadOpen && (
        <DatasetFormDialog
          onClose={() => setUploadOpen(false)}
          onSaved={() => {
            setPage(1)
            setAttempt((value) => value + 1)
          }}
        />
      )}
      {editTarget && (
        <DatasetFormDialog
          key={editTarget.id}
          dataset={editTarget}
          onClose={() => setEditTarget(null)}
          onSaved={() => {
            setEditTarget(null)
            setAttempt((value) => value + 1)
          }}
        />
      )}
      {deleteTarget && (
        <DeleteDialog
          key={deleteTarget.id}
          dataset={deleteTarget}
          onClose={() => setDeleteTarget(null)}
          onDeleted={() => {
            setDeleteTarget(null)
            setAttempt((value) => value + 1)
          }}
        />
      )}
      {detailId && <DetailDialog key={detailId} id={detailId} onClose={() => setDetailId('')} />}
    </Shell>
  )
}

function DatasetFormDialog({
  dataset,
  onClose,
  onSaved,
}: {
  dataset?: GoldenDatasetSummary
  onClose: () => void
  onSaved: () => void
}) {
  const dialog = useRef<HTMLDialogElement>(null)
  const [name, setName] = useState(dataset?.name ?? '')
  const [description, setDescription] = useState(dataset?.description ?? '')
  const [file, setFile] = useState<File | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const editing = Boolean(dataset)

  useEffect(() => {
    dialog.current?.showModal()
  }, [])

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!name.trim()) return setError('Nama dataset wajib diisi.')
    if (!editing && !file) return setError('Pilih file JSON terlebih dahulu.')
    if (file && file.size > 5 * 1024 * 1024) return setError('Ukuran file melebihi batas 5 MB.')

    setSubmitting(true)
    setError('')
    try {
      if (dataset) {
        await updateGoldenDataset(dataset.id, { name: name.trim(), description, file })
      } else {
        await uploadGoldenDataset({ name: name.trim(), description, file: file as File })
      }
      onSaved()
    } catch (reason) {
      setError(reason instanceof GoldenDatasetApiError ? reason.message : 'Golden Dataset gagal disimpan.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <dialog
      ref={dialog}
      className="golden-dialog"
      aria-labelledby="golden-upload-title"
      onCancel={(event) => {
        event.preventDefault()
        if (!submitting) onClose()
      }}
    >
      <form onSubmit={submit} className="golden-upload-form">
        <div className="golden-dialog-heading">
          <h2 id="golden-upload-title">{editing ? 'Edit Golden Dataset' : 'Tambah Golden Dataset'}</h2>
          <button type="button" className="golden-icon-button" onClick={onClose} disabled={submitting} aria-label="Tutup">
            <X size={18} />
          </button>
        </div>

        <label className="golden-field">
          <span>Nama dataset</span>
          <input
            type="text"
            value={name}
            onChange={(event) => setName(event.target.value)}
            maxLength={200}
            required
            disabled={submitting}
            placeholder="Contoh: Golden Dataset v1"
          />
        </label>

        <label className="golden-field">
          <span>Deskripsi (opsional)</span>
          <textarea
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            maxLength={2000}
            rows={2}
            disabled={submitting}
            placeholder="Catatan singkat tentang dataset ini"
          />
        </label>

        <label className="golden-field">
          <span>File JSON {editing ? '(opsional)' : ''}</span>
          <input
            type="file"
            accept=".json,application/json"
            onChange={(event) => setFile(event.target.files?.[0] || null)}
            required={!editing}
            disabled={submitting}
          />
          <small>
            {editing
              ? 'Kosongkan bila hanya mengubah nama atau deskripsi. Mengunggah file baru akan mengganti seluruh item.'
              : 'Maksimal 5 MB. Setiap item divalidasi terhadap struktur Gherkin dan GUI Elements.'}
          </small>
        </label>

        {error && (
          <p className="golden-form-error" role="alert">
            <AlertCircle size={16} /> {error}
          </p>
        )}

        <div className="golden-dialog-actions">
          <button type="button" className="golden-button secondary" onClick={onClose} disabled={submitting}>
            Batal
          </button>
          <button type="submit" className="golden-button primary" disabled={submitting}>
            {submitting ? <LoaderCircle className="golden-spin" size={17} /> : <Upload size={17} />}
            {submitting ? 'Menyimpan…' : editing ? 'Simpan Perubahan' : 'Unggah Dataset'}
          </button>
        </div>
      </form>
    </dialog>
  )
}

function DeleteDialog({
  dataset,
  onClose,
  onDeleted,
}: {
  dataset: GoldenDatasetSummary
  onClose: () => void
  onDeleted: () => void
}) {
  const dialog = useRef<HTMLDialogElement>(null)
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    dialog.current?.showModal()
  }, [])

  async function confirm() {
    setDeleting(true)
    setError('')
    try {
      await deleteGoldenDataset(dataset.id)
      onDeleted()
    } catch (reason) {
      setError(reason instanceof GoldenDatasetApiError ? reason.message : 'Golden Dataset gagal dihapus.')
    } finally {
      setDeleting(false)
    }
  }

  return (
    <dialog
      ref={dialog}
      className="golden-dialog"
      aria-labelledby="golden-delete-title"
      onCancel={(event) => {
        event.preventDefault()
        if (!deleting) onClose()
      }}
    >
      <div className="golden-dialog-heading">
        <h2 id="golden-delete-title">Hapus Golden Dataset</h2>
        <button type="button" className="golden-icon-button" onClick={onClose} disabled={deleting} aria-label="Tutup">
          <X size={18} />
        </button>
      </div>

      <p className="golden-delete-message">
        Hapus <strong>{dataset.name}</strong> beserta {dataset.item_count} item di dalamnya? Tindakan ini tidak dapat
        dibatalkan.
      </p>

      {error && (
        <p className="golden-form-error" role="alert">
          <AlertCircle size={16} /> {error}
        </p>
      )}

      <div className="golden-dialog-actions">
        <button type="button" className="golden-button secondary" onClick={onClose} disabled={deleting}>
          Batal
        </button>
        <button type="button" className="golden-button danger" onClick={confirm} disabled={deleting}>
          {deleting ? <LoaderCircle className="golden-spin" size={17} /> : <Trash2 size={17} />}
          {deleting ? 'Menghapus…' : 'Hapus Dataset'}
        </button>
      </div>
    </dialog>
  )
}

function DetailDialog({ id, onClose }: { id: string; onClose: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null)
  const [detail, setDetail] = useState<GoldenDatasetDetail | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    dialog.current?.showModal()
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    getGoldenDataset(id, controller.signal)
      .then((result) => {
        if (!controller.signal.aborted) setDetail(result)
      })
      .catch((reason: Error) => {
        if (!controller.signal.aborted) setError(reason.message)
      })
    return () => controller.abort()
  }, [id])

  return (
    <dialog
      ref={dialog}
      className="golden-dialog golden-detail-dialog"
      aria-labelledby="golden-detail-title"
      onCancel={(event) => {
        event.preventDefault()
        onClose()
      }}
    >
      <div className="golden-dialog-heading">
        <h2 id="golden-detail-title">{detail?.name || 'Detail Golden Dataset'}</h2>
        <button type="button" className="golden-icon-button" onClick={onClose} aria-label="Tutup">
          <X size={18} />
        </button>
      </div>

      {!detail && !error && (
        <div className="golden-state" role="status">
          <LoaderCircle className="golden-spin" size={26} />
          <p>Memuat detail dataset…</p>
        </div>
      )}

      {error && (
        <div className="golden-state" role="alert">
          <AlertCircle size={26} />
          <p>{error}</p>
        </div>
      )}

      {detail && !error && (
        <>
          <div className="golden-detail-meta">
            <span>{detail.item_count} item</span>
            <span>Dibuat {formatDate(detail.created_at)}</span>
            <span>{detail.used_at ? `Dipakai ${formatDate(detail.used_at)}` : 'Belum dipakai evaluasi'}</span>
          </div>
          {detail.description && <p className="golden-detail-description">{detail.description}</p>}

          <ol className="golden-items">
            {detail.items.map((item) => (
              <li key={item.id}>
                <h3>{item.item_title}</h3>
                <p className="golden-item-story">{item.user_story_text}</p>
                <ul className="golden-chips">
                  {item.gui_elements.map((element) => (
                    <li key={element}>{element}</li>
                  ))}
                </ul>
                <p className="golden-item-free">{item.free_scenario_text}</p>
                <details>
                  <summary>Expected Gherkin</summary>
                  <pre>{item.expected_gherkin}</pre>
                </details>
              </li>
            ))}
          </ol>
        </>
      )}
    </dialog>
  )
}
