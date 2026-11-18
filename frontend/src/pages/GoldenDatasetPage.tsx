import { useEffect, useRef, useState, type FormEvent, type ReactNode } from 'react'
import {
  AlertCircle,
  ChevronLeft,
  ChevronRight,
  Database,
  Download,
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
  return new Intl.DateTimeFormat('en-GB', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
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
  const [state, setState] = useState<ListState>({
    data: null,
    error: '',
    loading: true,
  })
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
    listGoldenDatasets({
      search: query,
      page,
      pageSize: PAGE_SIZE,
      signal: controller.signal,
    })
      .then((data) => {
        if (!controller.signal.aborted) {
          setState({ data, error: '', loading: false })
        }
      })
      .catch((reason: Error) => {
        if (!controller.signal.aborted) {
          setState((current) => ({
            ...current,
            error: reason.message,
            loading: false,
          }))
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
          <p>
            Manage Golden Gherkin datasets used as references for evaluation.
          </p>
        </div>
        <button
          type="button"
          className="golden-button primary"
          onClick={() => setUploadOpen(true)}
        >
          <Plus size={18} /> Add Dataset
        </button>
      </div>

      <label className="golden-search">
        <Search size={19} aria-hidden="true" />
        <input
          type="search"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Search datasets..."
          aria-label="Search datasets"
        />
      </label>

      <section className="golden-panel" aria-live="polite">
        {state.loading && (
          <div className="golden-state" role="status">
            <LoaderCircle className="golden-spin" size={28} />
            <p>Loading Golden Datasets…</p>
          </div>
        )}

        {!state.loading && state.error && (
          <div className="golden-state" role="alert">
            <AlertCircle size={30} />
            <p>{state.error}</p>
            <button
              type="button"
              className="golden-button secondary"
              onClick={() => setAttempt((value) => value + 1)}
            >
              Try Again
            </button>
          </div>
        )}

        {!state.loading && !state.error && datasets.length === 0 && (
          <div className="golden-state">
            <Database size={34} />
            <h2>
              {query ? 'No Datasets Found' : 'No Golden Datasets Yet'}
            </h2>
            <p>
              {query
                ? 'Try another keyword or clear the search.'
                : 'Upload a JSON file containing reference scenarios to start evaluating.'}
            </p>
            {!query && (
              <button
                type="button"
                className="golden-button primary"
                onClick={() => setUploadOpen(true)}
              >
                <Plus size={18} /> Add Dataset
              </button>
            )}
          </div>
        )}

        {!state.loading && !state.error && datasets.length > 0 && (
          <>
            <table className="golden-table">
              <thead>
                <tr>
                  <th scope="col">Dataset Name</th>
                  <th scope="col" className="golden-actions-heading">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {datasets.map((dataset) => (
                  <tr key={dataset.id}>
                    <td>
                      <span className="golden-name">{dataset.name}</span>
                      <span className="golden-meta">
                        {dataset.item_count}{' '}
                        {dataset.item_count === 1 ? 'item' : 'items'} &middot; created{' '}
                        {formatDate(dataset.created_at)}
                      </span>
                    </td>
                    <td>
                      <div className="golden-actions">
                        <button
                          type="button"
                          className="golden-action view"
                          onClick={() => setDetailId(dataset.id)}
                        >
                          <Eye size={15} /> View
                        </button>
                        <button
                          type="button"
                          className="golden-action"
                          onClick={() => setEditTarget(dataset)}
                          disabled={Boolean(dataset.used_at)}
                          title={
                            dataset.used_at
                              ? 'Already used for evaluation and cannot be edited'
                              : 'Edit dataset'
                          }
                        >
                          <Pencil size={15} /> Edit
                        </button>
                        <button
                          type="button"
                          className="golden-action delete"
                          onClick={() => setDeleteTarget(dataset)}
                          disabled={Boolean(dataset.used_at)}
                          title={
                            dataset.used_at
                              ? 'Already used for evaluation and cannot be deleted'
                              : 'Delete dataset'
                          }
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
              <nav
                className="golden-pagination"
                aria-label="Dataset pagination"
              >
                <button
                  type="button"
                  className="golden-page-arrow"
                  onClick={() => setPage((value) => Math.max(1, value - 1))}
                  disabled={page <= 1}
                  aria-label="Previous page"
                >
                  <ChevronLeft size={16} />
                </button>
                {Array.from(
                  { length: totalPages },
                  (_, index) => index + 1,
                ).map((number) => (
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
                  aria-label="Next page"
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
      {detailId && (
        <DetailDialog
          key={detailId}
          id={detailId}
          onClose={() => setDetailId('')}
        />
      )}
    </Shell>
  )
}

interface GoldenItemInput {
  item_title: string
  user_story_text: string
  gui_elements: string[]
  free_scenario_text: string
  expected_gherkin: string
}

type ItemDraft = {
  item_title: string
  user_story_text: string
  gui_elements: string
  free_scenario_text: string
  expected_gherkin: string
}

const MAX_FILE_BYTES = 5 * 1024 * 1024

const emptyDraft: ItemDraft = {
  item_title: '',
  user_story_text: '',
  gui_elements: '',
  free_scenario_text: '',
  expected_gherkin: '',
}

const templateItem: GoldenItemInput = {
  item_title: 'Successful login',
  user_story_text:
    'As a user, I want to log in with my email and password so that I can access the dashboard.',
  gui_elements: ['Email field', 'Password field', 'Login button'],
  free_scenario_text:
    'The user enters a valid email and password, then clicks the Login button.',
  expected_gherkin: `Feature: Login
  Scenario: Successful login with valid credentials
    Given the user is on the login page
    When the user enters "user@example.com" in the Email field
    And the user enters "password123" in the Password field
    And the user clicks the Login button
    Then the user is redirected to the dashboard`,
}

function parseElements(value: string) {
  return value
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean)
}

function itemsFile(items: GoldenItemInput[]) {
  return new File([JSON.stringify({ items }, null, 2)], 'golden-dataset.json', {
    type: 'application/json',
  })
}

function downloadTemplate() {
  const blob = new Blob([JSON.stringify({ items: [templateItem] }, null, 2)], {
    type: 'application/json',
  })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'golden-dataset-template.json'
  link.click()
  URL.revokeObjectURL(url)
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
  const [mode, setMode] = useState<'form' | 'file'>('form')
  const [draft, setDraft] = useState<ItemDraft>(emptyDraft)
  const [items, setItems] = useState<GoldenItemInput[]>([])
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const editing = Boolean(dataset)

  useEffect(() => {
    dialog.current?.showModal()
  }, [])

  function updateDraft(field: keyof ItemDraft, value: string) {
    setDraft((current) => ({ ...current, [field]: value }))
  }

  function addItem() {
    const item: GoldenItemInput = {
      item_title: draft.item_title.trim(),
      user_story_text: draft.user_story_text.trim(),
      gui_elements: parseElements(draft.gui_elements),
      free_scenario_text: draft.free_scenario_text.trim(),
      expected_gherkin: draft.expected_gherkin.trim(),
    }

    if (
      !item.item_title ||
      !item.user_story_text ||
      !item.free_scenario_text ||
      !item.expected_gherkin ||
      item.gui_elements.length === 0
    ) {
      return setError(
        'Complete all item fields, including at least one GUI element.',
      )
    }

    const duplicate = items.some(
      (existing) => existing.item_title.toLowerCase() === item.item_title.toLowerCase(),
    )
    if (duplicate) {
      return setError(`Item "${item.item_title}" already exists in the list.`)
    }

    setItems((current) => [...current, item])
    setDraft(emptyDraft)
    setError('')
  }

  function removeItem(index: number) {
    setItems((current) => current.filter((_, position) => position !== index))
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!name.trim()) return setError('Dataset name is required.')

    let upload = file
    if (!editing && mode === 'form') {
      if (items.length === 0) return setError('Add at least one item first.')
      upload = itemsFile(items)
    }
    if (!editing && !upload) return setError('Select a JSON file first.')
    if (upload && upload.size > MAX_FILE_BYTES) {
      return setError('The file exceeds the 5 MB size limit.')
    }

    setSubmitting(true)
    setError('')
    try {
      if (dataset) {
        await updateGoldenDataset(dataset.id, {
          name: name.trim(),
          description,
          file: upload,
        })
      } else {
        await uploadGoldenDataset({
          name: name.trim(),
          description,
          file: upload as File,
        })
      }
      onSaved()
    } catch (reason) {
      setError(
        reason instanceof GoldenDatasetApiError
          ? reason.message
          : 'Failed to save the Golden Dataset.',
      )
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
          <h2 id="golden-upload-title">
            {editing ? 'Edit Golden Dataset' : 'Add Golden Dataset'}
          </h2>
          <button
            type="button"
            className="golden-icon-button"
            onClick={onClose}
            disabled={submitting}
            aria-label="Close"
          >
            <X size={18} />
          </button>
        </div>

        <label className="golden-field">
          <span>Dataset name</span>
          <input
            type="text"
            value={name}
            onChange={(event) => setName(event.target.value)}
            maxLength={200}
            required
            disabled={submitting}
            placeholder="Example: Golden Dataset v1"
          />
        </label>

        <label className="golden-field">
          <span>Description (optional)</span>
          <textarea
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            maxLength={2000}
            rows={2}
            disabled={submitting}
            placeholder="Brief notes about this dataset"
          />
        </label>

        {!editing && (
          <div
            className="golden-mode-toggle"
            role="group"
            aria-label="Dataset input method"
          >
            <button
              type="button"
              className={mode === 'form' ? 'active' : ''}
              aria-pressed={mode === 'form'}
              onClick={() => setMode('form')}
              disabled={submitting}
            >
              Fill Out Form
            </button>
            <button
              type="button"
              className={mode === 'file' ? 'active' : ''}
              aria-pressed={mode === 'file'}
              onClick={() => setMode('file')}
              disabled={submitting}
            >
              Upload JSON File
            </button>
          </div>
        )}

        {!editing && mode === 'form' ? (
          <>
            <div className="golden-item-draft">
              <label className="golden-field">
                <span>Item title</span>
                <input
                  type="text"
                  value={draft.item_title}
                  onChange={(event) => updateDraft('item_title', event.target.value)}
                  maxLength={200}
                  disabled={submitting}
                  placeholder="Example: Successful login"
                />
              </label>

              <label className="golden-field">
                <span>User story</span>
                <textarea
                  value={draft.user_story_text}
                  onChange={(event) => updateDraft('user_story_text', event.target.value)}
                  rows={2}
                  disabled={submitting}
                  placeholder="As a user, I want to … so that …"
                />
              </label>

              <label className="golden-field">
                <span>GUI elements</span>
                <input
                  type="text"
                  value={draft.gui_elements}
                  onChange={(event) => updateDraft('gui_elements', event.target.value)}
                  disabled={submitting}
                  placeholder="Email field, Login button"
                />
                <small>Separate each element with a comma.</small>
              </label>

              <label className="golden-field">
                <span>Free-form scenario</span>
                <textarea
                  value={draft.free_scenario_text}
                  onChange={(event) => updateDraft('free_scenario_text', event.target.value)}
                  rows={2}
                  disabled={submitting}
                  placeholder="A brief description of the flow being tested"
                />
              </label>

              <label className="golden-field">
                <span>Expected Gherkin</span>
                <textarea
                  value={draft.expected_gherkin}
                  onChange={(event) => updateDraft('expected_gherkin', event.target.value)}
                  rows={5}
                  disabled={submitting}
                  placeholder={'Feature: Login\n  Scenario: Successful login'}
                />
              </label>

              <button
                type="button"
                className="golden-button secondary golden-add-item"
                onClick={addItem}
                disabled={submitting}
              >
                <Plus size={16} /> Add Item
              </button>
            </div>

            {items.length > 0 && (
              <ul className="golden-item-list">
                {items.map((item, index) => (
                  <li key={item.item_title}>
                    <span>{item.item_title}</span>
                    <button
                      type="button"
                      className="golden-icon-button"
                      onClick={() => removeItem(index)}
                      disabled={submitting}
                      aria-label={`Remove ${item.item_title}`}
                    >
                      <Trash2 size={15} />
                    </button>
                  </li>
                ))}
              </ul>
            )}

            <small className="golden-item-count">
              {items.length} {items.length === 1 ? 'item' : 'items'} ready to upload.
              At least 1 item is required.
            </small>
          </>
        ) : (
          <>
            <label className="golden-field">
              <span>JSON file {editing ? '(optional)' : ''}</span>
              <input
                type="file"
                accept=".json,application/json"
                onChange={(event) => setFile(event.target.files?.[0] || null)}
                required={!editing}
                disabled={submitting}
              />
              <small>
                {editing
                  ? 'Leave empty when updating only the name or description. Uploading a new file will replace all items.'
                  : 'Maximum file size: 5 MB. Each item is validated against the Gherkin structure and GUI elements.'}
              </small>
            </label>

            {!editing && (
              <button
                type="button"
                className="golden-template-link"
                onClick={downloadTemplate}
                disabled={submitting}
              >
                <Download size={15} /> Download JSON Template
              </button>
            )}
          </>
        )}

        {error && (
          <p className="golden-form-error" role="alert">
            <AlertCircle size={16} /> {error}
          </p>
        )}

        <div className="golden-dialog-actions">
          <button
            type="button"
            className="golden-button secondary"
            onClick={onClose}
            disabled={submitting}
          >
            Cancel
          </button>
          <button
            type="submit"
            className="golden-button primary"
            disabled={submitting}
          >
            {submitting ? (
              <LoaderCircle className="golden-spin" size={17} />
            ) : (
              <Upload size={17} />
            )}
            {submitting
              ? 'Saving…'
              : editing
                ? 'Save Changes'
                : 'Upload Dataset'}
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
      setError(
        reason instanceof GoldenDatasetApiError
          ? reason.message
          : 'Failed to delete the Golden Dataset.',
      )
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
        <h2 id="golden-delete-title">Delete Golden Dataset</h2>
        <button
          type="button"
          className="golden-icon-button"
          onClick={onClose}
          disabled={deleting}
          aria-label="Close"
        >
          <X size={18} />
        </button>
      </div>

      <p className="golden-delete-message">
        Delete <strong>{dataset.name}</strong> and its {dataset.item_count}{' '}
        {dataset.item_count === 1 ? 'item' : 'items'}? This action cannot be undone.
      </p>

      {error && (
        <p className="golden-form-error" role="alert">
          <AlertCircle size={16} /> {error}
        </p>
      )}

      <div className="golden-dialog-actions">
        <button
          type="button"
          className="golden-button secondary"
          onClick={onClose}
          disabled={deleting}
        >
          Cancel
        </button>
        <button
          type="button"
          className="golden-button danger"
          onClick={confirm}
          disabled={deleting}
        >
          {deleting ? (
            <LoaderCircle className="golden-spin" size={17} />
          ) : (
            <Trash2 size={17} />
          )}
          {deleting ? 'Deleting…' : 'Delete Dataset'}
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
        <h2 id="golden-detail-title">
          {detail?.name || 'Golden Dataset Details'}
        </h2>
        <button
          type="button"
          className="golden-icon-button"
          onClick={onClose}
          aria-label="Close"
        >
          <X size={18} />
        </button>
      </div>

      {!detail && !error && (
        <div className="golden-state" role="status">
          <LoaderCircle className="golden-spin" size={26} />
          <p>Loading dataset details…</p>
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
            <span>
              {detail.item_count}{' '}
              {detail.item_count === 1 ? 'item' : 'items'}
            </span>
            <span>Created {formatDate(detail.created_at)}</span>
            <span>
              {detail.used_at
                ? `Used for evaluation on ${formatDate(detail.used_at)}`
                : 'Not yet used for evaluation'}
            </span>
          </div>
          {detail.description && (
            <p className="golden-detail-description">{detail.description}</p>
          )}

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