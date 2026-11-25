import { useEffect, useId, useRef, useState, type FormEvent, type ReactNode, type TextareaHTMLAttributes } from 'react'
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
import './GoldenDatasetForm.css'
import { useAutoGrowTextarea } from '../hooks/useAutoGrowTextarea'

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
    <div className="generate-page qa-page golden-page">
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
  const searchInputRef = useRef<HTMLInputElement>(null)
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


      <section className="golden-panel">
        <div className="golden-panel-heading">
          <h2>Reference Datasets</h2>
          <p>Browse and manage your Golden Gherkin datasets.</p>
        </div>
        <div className="golden-search">
          <Search size={19} aria-hidden="true" />
          <input ref={searchInputRef} type="search" value={search}
            onChange={event => setSearch(event.target.value)}
            placeholder="Search datasets..." aria-label="Search datasets" />
          <button type="button" className="golden-search-clear"
            aria-label="Clear dataset search" disabled={search.length === 0}
            onClick={() => { setSearch(''); setQuery(''); setPage(1); searchInputRef.current?.focus() }}>
            <X size={17} aria-hidden="true" />
          </button>
        </div>
        
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
            <div className="golden-table-wrap" role="region" aria-label="Golden Dataset table" tabIndex={0}>
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
            </div>

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
            setUploadOpen(false)
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

type GoldenFormFieldKey = 'name' | 'file' | keyof ItemDraft
type GoldenFormErrors = Partial<Record<GoldenFormFieldKey, string>>

function GoldenInputField({ id, label, error, children }: {
  id: string
  label: string
  error?: string
  children: ReactNode
}) {
  return (
    <div className="golden-field">
      <label htmlFor={id}>{label}</label>
      <div className={`golden-input-shell${error ? ' has-error' : ''}`}>
        {children}
        {error && <p id={`${id}-error`} className="golden-inline-error" role="alert">{error}</p>}
      </div>
    </div>
  )
}

function GoldenAutoTextarea({ value, ...props }: Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'value' | 'defaultValue'> & { value: string }) {
  const ref = useAutoGrowTextarea(value)
  return <textarea {...props} ref={ref} value={value} />
}

function DatasetFormDialog({ dataset, onClose, onSaved }: {
  dataset?: GoldenDatasetSummary
  onClose: () => void
  onSaved: () => void
}) {
  const dialog = useRef<HTMLDialogElement>(null)
  const fileInput = useRef<HTMLInputElement>(null)
  const dragDepth = useRef(0)
  const mounted = useRef(true)
  const closingRef = useRef(false)
  const submittingRef = useRef(false)
  const closeAnimation = useRef<Animation | null>(null)
  const prefix = useId()

  const [name, setName] = useState(dataset?.name ?? '')
  const [description, setDescription] = useState(dataset?.description ?? '')
  const [file, setFile] = useState<File | null>(null)
  const [mode, setMode] = useState<'form' | 'file'>('form')
  const [draft, setDraft] = useState<ItemDraft>(emptyDraft)
  const [items, setItems] = useState<GoldenItemInput[]>([])
  const [submitting, setSubmitting] = useState(false)
  const [closing, setClosing] = useState(false)
  const [dragging, setDragging] = useState(false)
  const [fieldErrors, setFieldErrors] = useState<GoldenFormErrors>({})
  const [itemError, setItemError] = useState('')
  const [error, setError] = useState('')
  const editing = Boolean(dataset)
  const busy = submitting || closing

  useEffect(() => {
    mounted.current = true
    const element = dialog.current
    if (element && !element.open) element.showModal()

    const resetDrag = () => {
      dragDepth.current = 0
      setDragging(false)
    }
    const preventFileNavigation = (event: globalThis.DragEvent) => {
      if (Array.from(event.dataTransfer?.types ?? []).includes('Files')) {
        event.preventDefault()
      }
    }
    const handleOutsideDrop = (event: globalThis.DragEvent) => {
      preventFileNavigation(event)
      resetDrag()
    }
    window.addEventListener('dragover', preventFileNavigation)
    window.addEventListener('drop', handleOutsideDrop)
    window.addEventListener('dragend', resetDrag)

    return () => {
      mounted.current = false
      closeAnimation.current?.cancel()
      window.removeEventListener('dragover', preventFileNavigation)
      window.removeEventListener('drop', handleOutsideDrop)
      window.removeEventListener('dragend', resetDrag)
    }
  }, [])

  function fieldProps(field: GoldenFormFieldKey) {
    const id = `${prefix}-${field}`
    return {
      id,
      'aria-invalid': Boolean(fieldErrors[field]),
      'aria-describedby': fieldErrors[field] ? `${id}-error` : undefined,
    }
  }

  function clearFieldError(field: GoldenFormFieldKey) {
    setFieldErrors(current => {
      const next = { ...current }
      delete next[field]
      return next
    })
  }

  function updateDraft(field: keyof ItemDraft, value: string) {
    setDraft(current => ({ ...current, [field]: value }))
    clearFieldError(field)
    setItemError('')
    setError('')
  }

  function validateDraft(): GoldenFormErrors {
    const errors: GoldenFormErrors = {}
    if (!draft.item_title.trim()) errors.item_title = 'Item title is required.'
    if (!draft.user_story_text.trim()) errors.user_story_text = 'User story is required.'
    if (parseElements(draft.gui_elements).length === 0) errors.gui_elements = 'Enter at least one GUI element.'
    if (!draft.free_scenario_text.trim()) errors.free_scenario_text = 'Free-form scenario is required.'
    if (!draft.expected_gherkin.trim()) errors.expected_gherkin = 'Expected Gherkin is required.'
    return errors
  }

  function addItem() {
    if (submittingRef.current || closingRef.current) return
    const errors = validateDraft()
    setError('')
    setItemError('')
    if (Object.keys(errors).length > 0) {
      setFieldErrors(current => ({ ...current, ...errors }))
      return
    }
    const item: GoldenItemInput = {
      item_title: draft.item_title.trim(),
      user_story_text: draft.user_story_text.trim(),
      gui_elements: parseElements(draft.gui_elements),
      free_scenario_text: draft.free_scenario_text.trim(),
      expected_gherkin: draft.expected_gherkin.trim(),
    }
    if (items.some(existing => existing.item_title.toLowerCase() === item.item_title.toLowerCase())) {
      setFieldErrors(current => ({ ...current, item_title: 'An item with this title is already in the list.' }))
      return
    }
    setItems(current => [...current, item])
    setDraft(emptyDraft)
    setFieldErrors(current => {
      const next = { ...current }
      for (const field of Object.keys(emptyDraft) as (keyof ItemDraft)[]) delete next[field]
      return next
    })
  }

  function removeItem(index: number) {
    if (submittingRef.current || closingRef.current) return
    setItems(current => current.filter((_, position) => position !== index))
    setItemError('')
  }

  function receiveFile(candidate: File) {
    if (submittingRef.current || closingRef.current) return
    setError('')
    let message = ''
    if (!/\.json$/i.test(candidate.name)) message = 'Choose a JSON file with the .json extension.'
    else if (candidate.size === 0) message = 'The JSON file is empty.'
    else if (candidate.size > MAX_FILE_BYTES) message = 'The file exceeds the 5 MB size limit.'
    if (fileInput.current) fileInput.current.value = ''
    if (message) {
      setFile(null)
      setFieldErrors(current => ({ ...current, file: message }))
      return
    }
    setFile(candidate)
    clearFieldError('file')
  }

  async function finishClose(callback: () => void) {
    if (closingRef.current) return
    closingRef.current = true
    setClosing(true)
    const element = dialog.current
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    if (element && !reduced && typeof element.animate === 'function') {
      const styles = window.getComputedStyle(element)
      const from = { opacity: styles.opacity, transform: styles.transform }
      element.getAnimations().forEach(animation => animation.cancel())
      closeAnimation.current = element.animate(
        [from, { opacity: '0', transform: 'translateY(5px) scale(0.985)' }],
        { duration: 130, easing: 'ease-in', fill: 'forwards' },
      )
      try { await closeAnimation.current.finished } catch { /* Unmount cancellation. */ }
    }
    if (mounted.current) callback()
  }

  function requestClose() {
    if (submittingRef.current || closingRef.current) return
    void finishClose(onClose)
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (submittingRef.current || closingRef.current) return
    const errors: GoldenFormErrors = {}
    if (!name.trim()) errors.name = 'Dataset name is required.'
    setError('')
    setItemError('')

    let upload = file
    if (!editing && mode === 'form') {
      if (items.length === 0) {
        Object.assign(errors, validateDraft())
        setItemError('Add at least one item using the Add Item button before uploading.')
      } else upload = itemsFile(items)
    } else if (!editing && !upload) {
      errors.file = 'Select or drop a JSON file first.'
    }
    const oversized = Boolean(upload && upload.size > MAX_FILE_BYTES)
    if (oversized) {
      if (!editing && mode === 'form') setItemError('The dataset exceeds the 5 MB size limit. Reduce the number or size of the items.')
      else errors.file = 'The file exceeds the 5 MB size limit.'
    }
    setFieldErrors(errors)
    if (oversized || Object.keys(errors).length > 0 || (!editing && mode === 'form' && items.length === 0)) return

    submittingRef.current = true
    setSubmitting(true)
    try {
      if (dataset) {
        await updateGoldenDataset(dataset.id, { name: name.trim(), description, file: upload })
      } else {
        await uploadGoldenDataset({ name: name.trim(), description, file: upload as File })
      }
      if (mounted.current) await finishClose(onSaved)
    } catch (reason) {
      if (mounted.current) setError(reason instanceof GoldenDatasetApiError ? reason.message : 'Failed to save the Golden Dataset.')
    } finally {
      submittingRef.current = false
      if (mounted.current) setSubmitting(false)
    }
  }

  return (
    <dialog
      ref={dialog}
      className="golden-dialog golden-form-dialog"
      data-closing={closing}
      aria-labelledby={`${prefix}-title`}
      onCancel={event => {
        // A canceled file picker can emit a bubbling cancel event.
        // Only a cancel event originating on this dialog should close it.
        if (event.target !== event.currentTarget) return
        event.preventDefault()
        requestClose()
      }}
    >
      <form onSubmit={submit} className="golden-upload-form" noValidate>
        <div className="golden-dialog-heading">
          <h2 id={`${prefix}-title`}>{editing ? 'Edit Golden Dataset' : 'Add Golden Dataset'}</h2>
          <button type="button" className="golden-icon-button" onClick={requestClose} disabled={busy} aria-label="Close dataset form"><X size={18} aria-hidden="true" /></button>
        </div>

        <GoldenInputField id={`${prefix}-name`} label="Dataset name" error={fieldErrors.name}>
          <input {...fieldProps('name')} type="text" value={name} onChange={event => { setName(event.target.value); clearFieldError('name'); setError('') }} maxLength={200} required disabled={busy} placeholder="Example: Golden Dataset v1" />
        </GoldenInputField>

        <GoldenInputField id={`${prefix}-description`} label="Description (optional)">
          <GoldenAutoTextarea id={`${prefix}-description`} value={description} onChange={event => setDescription(event.target.value)} maxLength={2000} rows={2} disabled={busy} placeholder="Brief notes about this dataset" />
        </GoldenInputField>

        {!editing && (
          <div className="golden-mode-toggle" role="group" aria-label="Dataset input method">
            {(['form', 'file'] as const).map(value => (
              <button key={value} type="button" className={mode === value ? 'active' : ''} aria-pressed={mode === value} disabled={busy} onClick={() => { setMode(value); setFieldErrors({}); setItemError(''); setError(''); dragDepth.current = 0; setDragging(false) }}>
                {value === 'form' ? 'Fill Out Form' : 'Upload JSON File'}
              </button>
            ))}
          </div>
        )}

        {!editing && mode === 'form' ? (
          <>
            <div className="golden-item-draft">
              <GoldenInputField id={`${prefix}-item_title`} label="Item title" error={fieldErrors.item_title}>
                <input {...fieldProps('item_title')} type="text" value={draft.item_title} onChange={event => updateDraft('item_title', event.target.value)} maxLength={200} disabled={busy} placeholder="Example: Successful login" />
              </GoldenInputField>
              <GoldenInputField id={`${prefix}-user_story_text`} label="User story" error={fieldErrors.user_story_text}>
                <GoldenAutoTextarea {...fieldProps('user_story_text')} value={draft.user_story_text} onChange={event => updateDraft('user_story_text', event.target.value)} rows={2} disabled={busy} placeholder="As a user, I want to … so that …" />
              </GoldenInputField>
              <GoldenInputField id={`${prefix}-gui_elements`} label="GUI elements" error={fieldErrors.gui_elements}>
                <input {...fieldProps('gui_elements')} type="text" value={draft.gui_elements} onChange={event => updateDraft('gui_elements', event.target.value)} disabled={busy} placeholder="Email field, Login button" />
              </GoldenInputField>
              <small className="golden-field-help">Separate each GUI element with a comma.</small>
              <GoldenInputField id={`${prefix}-free_scenario_text`} label="Free-form scenario" error={fieldErrors.free_scenario_text}>
                <GoldenAutoTextarea {...fieldProps('free_scenario_text')} value={draft.free_scenario_text} onChange={event => updateDraft('free_scenario_text', event.target.value)} rows={2} disabled={busy} placeholder="A brief description of the flow being tested" />
              </GoldenInputField>
              <GoldenInputField id={`${prefix}-expected_gherkin`} label="Expected Gherkin" error={fieldErrors.expected_gherkin}>
                <GoldenAutoTextarea {...fieldProps('expected_gherkin')} className="golden-code-textarea" value={draft.expected_gherkin} onChange={event => updateDraft('expected_gherkin', event.target.value)} rows={5} disabled={busy} placeholder={'Feature: Login\n  Scenario: Successful login'} />
              </GoldenInputField>
              {itemError && <p className="golden-form-error" role="alert">{itemError}</p>}
              {error && <p className="golden-form-error" role="alert">{error}</p>}
              <button type="button" className="golden-button secondary golden-add-item" onClick={addItem} disabled={busy}><Plus size={16} aria-hidden="true" />Add Item</button>
            </div>
            {items.length > 0 && (
              <ul className="golden-item-list">
                {items.map((item, index) => (
                  <li key={item.item_title}>
                    <span>{item.item_title}</span>
                    <button type="button" className="golden-icon-button" onClick={() => removeItem(index)} disabled={busy} aria-label={`Remove ${item.item_title}`}><Trash2 size={15} aria-hidden="true" /></button>
                  </li>
                ))}
              </ul>
            )}
            <small className="golden-item-count">{items.length} {items.length === 1 ? 'item' : 'items'} ready to upload. At least 1 item is required.</small>
          </>
        ) : (
          <div className="golden-file-section">
            <div className="golden-field">
              <span id={`${prefix}-file-label`}>JSON file {editing ? '(optional)' : ''}</span>
              <div className={`golden-file-shell${fieldErrors.file ? ' has-error' : ''}`}>
                <div
                  className={`golden-dropzone${dragging ? ' is-dragging' : ''}${file ? ' has-file' : ''}${busy ? ' is-disabled' : ''}`}
                  role="group"
                  aria-labelledby={`${prefix}-file-label`}
                  aria-describedby={fieldErrors.file ? `${prefix}-file-error` : `${prefix}-file-help`}
                  onDragEnter={event => {
                    event.preventDefault()
                    event.stopPropagation()
                    if (busy || !Array.from(event.dataTransfer.types).includes('Files')) return
                    dragDepth.current += 1
                    setDragging(true)
                  }}
                  onDragOver={event => { event.preventDefault(); event.stopPropagation(); event.dataTransfer.dropEffect = busy ? 'none' : 'copy' }}
                  onDragLeave={event => {
                    event.preventDefault()
                    event.stopPropagation()
                    dragDepth.current = Math.max(0, dragDepth.current - 1)
                    if (dragDepth.current === 0) setDragging(false)
                  }}
                  onDrop={event => {
                    event.preventDefault()
                    event.stopPropagation()
                    dragDepth.current = 0
                    setDragging(false)
                    if (busy) return
                    if (event.dataTransfer.files.length !== 1) {
                      setFieldErrors(current => ({ ...current, file: 'Drop one JSON file at a time.' }))
                      return
                    }
                    const candidate = event.dataTransfer.files[0]
                    if (candidate) receiveFile(candidate)
                  }}
                >
                  <input ref={fileInput} type="file" accept=".json,application/json" hidden disabled={busy} onChange={event => { const candidate = event.target.files?.[0]; if (candidate) receiveFile(candidate) }} />
                  <Upload className="golden-dropzone-icon" size={28} aria-hidden="true" />
                  <strong>{dragging ? 'Release to attach your JSON file' : 'Drag your JSON file here'}</strong>
                  <span>One JSON file, up to 5 MB</span>
                  <button type="button" className="golden-button secondary" disabled={busy} aria-invalid={Boolean(fieldErrors.file)} aria-describedby={fieldErrors.file ? `${prefix}-file-error` : undefined} onClick={() => fileInput.current?.click()}>Choose JSON File</button>
                </div>
                {fieldErrors.file && <p id={`${prefix}-file-error`} className="golden-inline-error" role="alert">{fieldErrors.file}</p>}
              </div>
              <small id={`${prefix}-file-help`}>{editing ? 'Leave empty to keep the existing items. A new file replaces all items.' : 'The backend validates the uploaded dataset and its Gherkin scenarios.'}</small>
            </div>
            {file && (
              <div className="golden-file-preview" key={`${file.name}-${file.size}-${file.lastModified}`}>
                <div role="status"><strong>{file.name}</strong><span>{(file.size / 1024).toFixed(1)} KB · Ready to upload</span></div>
                <button type="button" className="golden-icon-button" disabled={busy} aria-label="Remove selected JSON file" onClick={() => { setFile(null); clearFieldError('file'); if (fileInput.current) fileInput.current.value = '' }}><X size={17} aria-hidden="true" /></button>
              </div>
            )}
            {error && <p className="golden-form-error" role="alert">{error}</p>}
            {!editing && (
              <button type="button" className="golden-template-link" disabled={busy} onClick={event => {
                const button = event.currentTarget
                if (!window.matchMedia('(prefers-reduced-motion: reduce)').matches && typeof button.animate === 'function') {
                  button.getAnimations().forEach(animation => animation.cancel())
                  button.animate([{ transform: 'scale(1)' }, { transform: 'scale(0.96)' }, { transform: 'scale(1.01)' }, { transform: 'scale(1)' }], { duration: 240, easing: 'ease-out' })
                }
                downloadTemplate()
              }}><Download size={17} aria-hidden="true" />Download JSON Template</button>
            )}
          </div>
        )}

        <div className="golden-dialog-actions">
          <button type="button" className="golden-button secondary" onClick={requestClose} disabled={busy}>Cancel</button>
          <button type="submit" className="golden-button primary" disabled={busy}>
            {submitting ? <LoaderCircle className="golden-spin" size={17} aria-hidden="true" /> : <Upload size={17} aria-hidden="true" />}
            {submitting ? 'Saving…' : editing ? 'Save Changes' : 'Upload Dataset'}
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