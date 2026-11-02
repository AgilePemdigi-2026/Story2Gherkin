import { useRef, useState, type FormEvent } from 'react'
import { Link, NavLink } from 'react-router-dom'
import {
  Plus,
  X,
  AlertCircle,
  FileText,
  FileWarning,
  LoaderCircle,
  Check,
  ArrowRight,
} from 'lucide-react'
import PublicHeader from '../components/PublicHeader'
import {
  generateScenario,
  type GenerateScenarioResult,
} from '../services/scenarioApi'
import './GenerateScenarioPage.css'

interface GuiElement {
  id: number
  value: string
}

interface FormErrors {
  userStory?: string
  guiElements?: string
  freeScenario?: string
}

// Only used when preview=success; real responses always use backend Gherkin.
const previewGherkin = `Feature: Delete User Data
Scenario: Delete user successfully
Given I am on the user list page
When I click the delete button
And I confirm the delete action
Then the user data is deleted`

function renderGherkin(source: string) {
  const lines = source.split(/\r?\n/)

  return lines.map((line, index) => {
    const match = line.match(
      /^(\s*)(Feature:|Scenario(?: Outline)?:|Given|When|Then|And|But)(.*)$/
    )
    const newline = index < lines.length - 1 ? '\n' : ''

    if (!match) return <span key={index}>{line}{newline}</span>

    const [, spaces, keyword, rest] = match
    const keywordClass =
      keyword === 'Feature:'
        ? 'gherkin-feature'
        : keyword.startsWith('Scenario')
          ? 'gherkin-scenario'
          : 'gherkin-step'

    return (
      <span key={index}>
        {spaces}<span className={keywordClass}>{keyword}</span>{rest}{newline}
      </span>
    )
  })
}

function GenerateScenarioPage() {
  // FORM STATE
  const [userStory, setUserStory] = useState('')
  const [freeScenario, setFreeScenario] = useState('')

  const [guiElements, setGuiElements] = useState<GuiElement[]>([
    { id: 1, value: '' },
  ])

  const nextGuiId = useRef(4)

  // VALIDATION
  const [errors, setErrors] = useState<FormErrors>({})
  const [formMessage, setFormMessage] = useState('')
  
  // API GENERATE STATE
  const [isGenerating, setIsGenerating] = useState(false)

  const [generatedResult, setGeneratedResult] =
    useState<GenerateScenarioResult | null>(null)

  const [generateError, setGenerateError] = useState('')


  // ADD GUI ELEMENT
  const addGuiElement = () => {
    if (isGenerating || guiElements.length >= 30) return
    setGeneratedResult(null)
    setGenerateError('')

    const id = nextGuiId.current
    nextGuiId.current += 1

    setGuiElements((prev) => [...prev, { id, value: '' }])
    setErrors((prev) => ({ ...prev, guiElements: undefined }))
    setFormMessage('')
  }

  // UPDATE GUI ELEMENT
  const updateGuiElement = (id: number, value: string) => {
    if (isGenerating) return
    setGeneratedResult(null)
    setGenerateError('')
    setGuiElements((prev) =>
      prev.map((element) =>
        element.id === id ? { ...element, value } : element
      )
    )

    setErrors((prev) => ({ ...prev, guiElements: undefined }))
    setFormMessage('')
  }

  // REMOVE GUI ELEMENT
  const removeGuiElement = (id: number) => {
    if (isGenerating) return
    setGeneratedResult(null)
    setGenerateError('')
    if (guiElements.length === 1) {
      setGuiElements((prev) =>
        prev.map((element) => ({ ...element, value: '' }))
      )
    } else {
      setGuiElements((prev) =>
        prev.filter((element) => element.id !== id)
      )
    }

    setErrors((prev) => ({ ...prev, guiElements: undefined }))
    setFormMessage('')
  }

  
  // VALIDATE FORM & GENERATE SCENARIO
  const handleGenerate = async (
    event: FormEvent<HTMLFormElement>
  ) => {
    event.preventDefault()

    // Prevent duplicate requests
    if (isGenerating) return

    const newErrors: FormErrors = {}

    // Validate User Story
    if (!userStory.trim()) {
      newErrors.userStory = 'User story is required.'
    } else if (userStory.trim().length < 10) {
      newErrors.userStory =
        'User story must contain at least 10 characters.'
    }

    // Validate GUI Elements
    if (
      guiElements.length === 0 ||
      guiElements.some((element) => !element.value.trim())
    ) {
      newErrors.guiElements =
        'Please fill in all GUI elements or remove empty rows.'
    }

    // Validate Free Scenario
    if (!freeScenario.trim()) {
      newErrors.freeScenario = 'Free scenario is required.'
    } else if (freeScenario.trim().length < 10) {
      newErrors.freeScenario =
        'Free scenario must contain at least 10 characters.'
    }

    setErrors(newErrors)
    setFormMessage('')
    setGenerateError('')
    setGeneratedResult(null)

    // Stop if validation fails
    if (Object.keys(newErrors).length > 0) {
      return
    }

    // Send request to backend
    setIsGenerating(true)

    try {
      const result = await generateScenario({
        user_story: userStory.trim(),
        gui_elements: guiElements.map((element) =>
          element.value.trim()
        ),
        free_scenario: freeScenario.trim(),
      })

      // Save generated result
      setGeneratedResult(result)
    } catch (error: unknown) {
      const message =
        error instanceof Error
          ? error.message
          : 'Failed to generate Gherkin scenarios.'

      setGenerateError(message)
    } finally {
      setIsGenerating(false)
    }
  }

  const hasErrors = Object.values(errors).some(Boolean) || Boolean(generateError)

  // Preview states are only enabled in the development environment.
  const preview = import.meta.env.DEV
    ? new URLSearchParams(window.location.search).get('preview')
    : null
  const isProcessingPreview = preview === 'processing'
  const isSuccessPreview = preview === 'success'
  // Processing state: real API request or development preview
  const isProcessing = isGenerating || isProcessingPreview
  const isSuccess = generatedResult !== null || isSuccessPreview

  return (
    <div className="generate-page">
      <PublicHeader />

      <div className="generate-layout">

        {/* SIDEBAR */}
        <aside className="generate-sidebar">
          <nav
            className="generate-sidebar-nav"
            aria-label="Workspace navigation"
          >
            <div className="generate-sidebar-section">
              <p className="generate-sidebar-heading">
                GENERATE
              </p>

              <NavLink
                to="/generate"
                className={({ isActive }) =>
                  `generate-sidebar-link ${isActive ? 'active' : ''}`
                }
              >
                Generate Scenario
              </NavLink>

              <span className="generate-sidebar-link disabled">
                Review Scenario
              </span>
            </div>

            <div className="generate-sidebar-divider" />

            <div className="generate-sidebar-section">
              <p className="generate-sidebar-heading">
                EVALUATION
              </p>

              <span className="generate-sidebar-link disabled">
                Golden Dataset
              </span>

              <span className="generate-sidebar-link disabled">
                Run Evaluation
              </span>

              <span className="generate-sidebar-link disabled">
                Evaluation Report
              </span>
            </div>
          </nav>
        </aside>

        {/* MAIN CONTENT */}
        <main className="generate-main">

          {/* TITLE */}
          <div className="generate-page-heading">
            <h1>Generate Skenario Gherkin</h1>
            <p>
              Masukkan user story, elemen GUI, dan skenario bebas
              untuk menghasilkan skenario Gherkin secara otomatis.
            </p>
          </div>

          <div className="generate-panels">

            {/* LEFT PANEL: FORM */}
            <section className="generate-panel generate-input-panel">
              <form
                className="generate-form"
                onSubmit={handleGenerate}
                noValidate
              >

                {/* ERROR BANNER */}
                {hasErrors && (
                  <div className="generate-error-banner" role="alert">
                    <AlertCircle size={19} />
                    <span>
                      {generateError || 'Generate failed. Please complete all required fields.'}
                    </span>
                  </div>
                )}

                {/* USER STORY */}
                <div className="generate-form-group">
                  <label htmlFor="user-story">
                    USER STORY <span>*</span>
                  </label>

                  <textarea
                    id="user-story"
                    className={`generate-textarea ${
                      errors.userStory ? 'has-error' : ''
                    }`}
                    placeholder="Contoh: As an admin, I want to delete user data"
                    value={userStory}
                    maxLength={8000}
                    disabled={isProcessing}
                    onChange={(event) => {
                      setUserStory(event.target.value)
                      setGeneratedResult(null)
                      setGenerateError('')
                      setErrors((prev) => ({
                        ...prev,
                        userStory: undefined,
                      }))
                      setFormMessage('')
                    }}
                    aria-invalid={Boolean(errors.userStory)}
                    aria-describedby={
                      errors.userStory ? 'user-story-error' : undefined
                    }
                  />

                  {errors.userStory && (
                    <p
                      id="user-story-error"
                      className="generate-field-error"
                      role="alert"
                    >
                      <AlertCircle size={15} />
                      {errors.userStory}
                    </p>
                  )}
                </div>

                {/* GUI ELEMENTS */}
                <div className="generate-form-group">
                  <label>
                    GUI ELEMENTS <span>*</span>
                  </label>

                  <div className="generate-gui-list">
                    {guiElements.map((element, index) => (
                      <div
                        className="generate-gui-row"
                        key={element.id}
                      >
                        <input
                          type="text"
                          aria-label={`GUI Element ${index + 1}`}
                          aria-invalid={Boolean(errors.guiElements)}
                          placeholder={`GUI Element ${index + 1}`}
                          value={element.value}
                          maxLength={200}
                          disabled={isProcessing}
                          onChange={(event) =>
                            updateGuiElement(
                              element.id,
                              event.target.value
                            )
                          }
                        />

                        <button
                          type="button"
                          className="generate-remove-button"
                          aria-label={`Remove GUI element ${index + 1}`}
                          disabled={isProcessing}
                          onClick={() => removeGuiElement(element.id)}
                        >
                          <X size={19} />
                        </button>
                      </div>
                    ))}
                  </div>

                  <button
                    type="button"
                    className="generate-add-button"
                    onClick={addGuiElement}
                    disabled={isProcessing || guiElements.length >= 30}
                  >
                    <Plus size={19} />
                    Tambah Elemen
                  </button>

                  {errors.guiElements && (
                    <p className="generate-field-error" role="alert">
                      <AlertCircle size={15} />
                      {errors.guiElements}
                    </p>
                  )}
                </div>

                {/* FREE SCENARIO */}
                <div className="generate-form-group">
                  <label htmlFor="free-scenario">
                    FREE SCENARIO <span>*</span>
                  </label>

                  <textarea
                    id="free-scenario"
                    className={`generate-textarea generate-free-textarea ${
                      errors.freeScenario ? 'has-error' : ''
                    }`}
                    placeholder="Admin login, membuka halaman, klik delete..."
                    value={freeScenario}
                    maxLength={4000}
                    disabled={isProcessing}
                    onChange={(event) => {
                      setFreeScenario(event.target.value)
                      setGeneratedResult(null)
                      setGenerateError('')
                      setErrors((prev) => ({
                        ...prev,
                        freeScenario: undefined,
                      }))
                      setFormMessage('')
                    }}
                    aria-invalid={Boolean(errors.freeScenario)}
                    aria-describedby={
                      errors.freeScenario
                        ? 'free-scenario-error'
                        : undefined
                    }
                  />

                  {errors.freeScenario && (
                    <p
                      id="free-scenario-error"
                      className="generate-field-error"
                      role="alert"
                    >
                      <AlertCircle size={15} />
                      {errors.freeScenario}
                    </p>
                  )}
                </div>

                {/* SUBMIT / GENERATE AGAIN */}
                {isSuccess ? (
                  generatedResult ? (
                    <button
                      type="button"
                      className="generate-submit-button"
                      onClick={() => {
                        setGeneratedResult(null)
                        setGenerateError('')
                        setErrors({})
                      }}
                    >
                      Generate Skenario Lain
                    </button>
                  ) : (
                    <Link
                      to="/generate"
                      className="generate-submit-button generate-again-link"
                    >
                      Generate Skenario Lain
                    </Link>
                  )
                ) : (
                  <button
                    type="submit"
                    className="generate-submit-button"
                    disabled={isProcessing}
                  >
                    {isProcessing ? (
                      <>
                        <LoaderCircle
                          size={19}
                          className="generate-loading-spinner"
                        />
                        Generating scenario...
                      </>
                    ) : (
                      'Generate Skenario'
                    )}
                  </button>
                )}

                {formMessage && (
                  <p className="generate-form-message" role="status">
                    {formMessage}
                  </p>
                )}

              </form>
            </section>

        
            {/* RIGHT PANEL: RESULT */}
            <section className="generate-panel generate-result-panel">

            {/* RESULT HEADER */}
            <div className="generate-result-heading">
                <h2>Hasil Generate</h2>

            <span
              className={`generate-status ${
                hasErrors
                  ? 'failed'
                  : isProcessing
                    ? 'processing'
                    : isSuccess
                      ? 'success'
                      : 'ready'
              }`}
            >
              {hasErrors
                ? 'Gagal'
                : isProcessing
                  ? 'Sedang proses...'
                  : isSuccess
                    ? 'Berhasil'
                    : 'Siap generate'}
            </span>
            </div>

            {hasErrors ? (

                /* ERROR STATE */
                <div className="generate-result-error">

                <div className="generate-result-error-icon">
                    <FileWarning size={45} strokeWidth={1.8} />
                    <span className="generate-error-icon-badge">
                    <AlertCircle size={23} />
                    </span>
                </div>

                <h3>Gagal melakukan generate</h3>

                <p>
                     {generateError
                       ? 'Skenario belum dapat dihasilkan. Periksa pesan kesalahan dan coba kembali.'
                       : 'Mohon lengkapi semua field yang diperlukan untuk menghasilkan skenario Gherkin.'}
                </p>

                {/* ERROR ALERT */}
                <div className="generate-result-alert" role="alert">
                    <AlertCircle size={19} />
                    <span>
                     {generateError || 'Please complete all required fields.'}
                    </span>
                </div>

                
  </div>

            ) : isProcessing ? (

            /* PROCESSING STATE */
            <div className="generate-processing-state" role="status">

                {/* LOADING PROGRESS */}
                <div className="generate-processing-card">

                <div className="generate-processing-info">
                    <LoaderCircle
                    size={30}
                    className="generate-loading-spinner"
                    />

                    <div className="generate-processing-text">
                    <h3>Generating Gherkin scenario...</h3>
                    <p>
                        AI sedang memproses user story, elemen GUI,
                        dan skenario Anda. Mohon tunggu beberapa saat.
                    </p>
                    </div>
                </div>

                {/* INDETERMINATE PROGRESS BAR */}
                <div
                    className="generate-progress-track"
                    role="progressbar"
                    aria-label="Generating Gherkin scenario"
                    aria-valuetext="Sedang memproses"
                >
                    <div className="generate-progress-indicator" />
                </div>

                </div>

                    {/* GH ERKIN RESULT SKELETON */}
                    <div className="generate-skeleton-result" aria-hidden="true">
                    <div className="generate-skeleton-line long" />
                    <div className="generate-skeleton-row">
                        <div className="generate-skeleton-line short" />
                        <div className="generate-skeleton-line medium" />
                    </div>
                    <div className="generate-skeleton-line large" />
                    <div className="generate-skeleton-line full" />
                    <div className="generate-skeleton-line medium" />
                    </div>

                    {/* VALIDATION SKELETON */}
                    <div className="generate-skeleton-cards" aria-hidden="true">
                    <div className="generate-skeleton-box">
                        <div className="generate-skeleton-square" />
                        <div className="generate-skeleton-line medium" />
                        <div className="generate-skeleton-line full" />
                        <div className="generate-skeleton-line short" />
                    </div>

                    <div className="generate-skeleton-box">
                        <div className="generate-skeleton-square" />
                        <div className="generate-skeleton-line medium" />
                        <div className="generate-skeleton-line full" />
                        <div className="generate-skeleton-line short" />
                    </div>
                    </div>

                    {/* COVERAGE SKELETON */}
                    <div className="generate-skeleton-warning" aria-hidden="true">
                    <div className="generate-skeleton-circle" />

                    <div className="generate-skeleton-warning-lines">
                        <div className="generate-skeleton-line full" />
                        <div className="generate-skeleton-line medium" />
                    </div>
                    </div>

                </div>

            ) : isSuccess ? (

  /* SUCCESS STATE */
  <div className="generate-success-state">

    {/* GHERKIN OUTPUT - data backend atau preview development */}
    <div className="generate-gherkin-output">
      <pre className="generate-gherkin-code">
        <code>{renderGherkin(generatedResult?.gherkin ?? previewGherkin)}</code>
      </pre>
    </div>

    {/* VALIDATION - contoh desain HANYA untuk preview success. */}
    {/* API Generate belum mengembalikan persentase GUI Coverage. */}
    {!generatedResult && (
    <div className="generate-validation-section">
      <h3>Validation</h3>

      <div className="generate-validation-grid">

        {/* GHERKIN STRUCTURE */}
        <div className="generate-validation-card structure-valid">
          <div className="generate-validation-check">
            <Check size={20} />
          </div>

          <div>
            <strong>Struktur Gherkin valid</strong>
            <p>Given / When / Then lengkap</p>
          </div>
        </div>

        {/* GUI COVERAGE */}
        <div className="generate-validation-card coverage-card">
          <div className="generate-coverage-header">
            <strong>GUI Coverage</strong>
            <span>100%</span>
          </div>

          <div
            className="generate-coverage-track"
            role="progressbar"
            aria-label="GUI Coverage"
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={100}
          >
            <div
              className="generate-coverage-fill"
              style={{ width: '100%' }}
            />
          </div>
        </div>
      </div>

      {/* COVERAGE RESULT */}
      <div className="generate-coverage-success">
        <div className="generate-coverage-success-icon">
          <Check size={19} />
        </div>

        <div>
          <strong>
            Semua elemen GUI sudah tercakup dalam skenario
          </strong>
          <p>3 dari 3 elemen GUI telah ditambahkan</p>
        </div>
      </div>
    </div>
    )}

    {generatedResult?.feature.warnings?.length ? (
      <div className="generate-tips" role="note">
        <div className="generate-tips-heading">
          <AlertCircle size={18} />
          <strong>Peringatan dari generator</strong>
        </div>
        {generatedResult.feature.warnings.map((warning, index) => (
          <p key={index}>{warning}</p>
        ))}
      </div>
    ) : null}

    {/* REVIEW BUTTON */}
    <button
      type="button"
      className="generate-review-button"
      disabled
      title="Halaman Review Scenario belum tersedia"
    >
      Review Skenario
      <ArrowRight size={18} />
    </button>

  </div>
            ) : (
              /* INITIAL STATE */
              <div className="generate-empty-state">
                <div className="generate-empty-icon">
                  <FileText size={43} strokeWidth={1.8} />
                </div>
                <h3>Belum ada hasil generate</h3>
                <p>
                  Lengkapi form di sebelah kiri, lalu klik tombol
                  “Generate Skenario” untuk melihat hasilnya.
                </p>
              </div>
            )}

            </section>



          </div>
        </main>
      </div>
    </div>
  )
}

export default GenerateScenarioPage
