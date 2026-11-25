import { useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'

import PublicHeader from '../components/PublicHeader'
import { useAuthSession } from '../hooks/useAuthSession'

import './HomePage.css'
import WorkflowAnimation from '../components/WorkflowAnimation'

const scenarioLines = [
  { keyword: 'Feature:', text: ' Password reset' },
  { keyword: 'Scenario:', text: ' Request a reset link' },
  { keyword: 'Given', text: ' I am on the password reset page' },
  { keyword: 'When', text: ' I submit my registered email address' },
  { keyword: 'Then', text: ' I should receive a password reset link' },
]

type PreviewMode = 'generate' | 'review' | 'report'

function GherkinPreview() {
  return (
    <pre className="s2g-code">
      <code>
        {scenarioLines.map((line, index) => (
          <span
            className="s2g-code-line"
            key={line.keyword}
            style={{ animationDelay: `${index * 0.65}s` }}
          >
            <span className="s2g-code-number" aria-hidden="true">
              {index + 1}
            </span>

            <span>
              <span className="s2g-code-keyword">{line.keyword}</span>
              {line.text}
            </span>
          </span>
        ))}
      </code>
    </pre>
  )
}

function StoryPreview() {
  return (
    <div className="s2g-story-preview">
      <div className="s2g-preview-label">User story</div>

      <p>
        As a registered user, I want to reset my password so that I can
        regain access to my account.
      </p>

      <div className="s2g-preview-label s2g-gui-label">GUI context</div>

      <div className="s2g-mini-form">
        <strong>Reset password</strong>
        <div className="s2g-mini-input">Email address</div>
        <span className="s2g-mini-submit">Send reset link</span>
      </div>
    </div>
  )
}

function ReportPreview() {
  return (
    <div className="s2g-report-preview">
      <div className="s2g-report-toolbar">
        <span>Scenario comparison</span>
        <span className="s2g-preview-download">Download CSV</span>
      </div>

      <div className="s2g-table-scroll">
        <table className="s2g-report-table">
          <caption className="s2g-sr-only">
            Illustrative comparison of LLM and rule-based scenarios
          </caption>

          <thead>
            <tr>
              <th scope="col">Scenario</th>
              <th scope="col">LLM</th>
              <th scope="col">Rule-based</th>
            </tr>
          </thead>

          <tbody>
            <tr>
              <th scope="row">Password reset</th>
              <td><span className="s2g-status">Valid</span></td>
              <td><span className="s2g-status">Valid</span></td>
            </tr>

            <tr>
              <th scope="row">Change email</th>
              <td><span className="s2g-status">Valid</span></td>
              <td><span className="s2g-status">Valid</span></td>
            </tr>

            <tr>
              <th scope="row">Update profile</th>
              <td><span className="s2g-status">Valid</span></td>
              <td>
                <span className="s2g-status s2g-status-warning">
                  Needs review
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}

function ProductWindow({ mode = 'generate' }: { mode?: PreviewMode }) {
  const isReport = mode === 'report'

  const title = isReport
    ? 'Evaluation Report'
    : mode === 'review'
      ? 'Review Scenario'
      : 'Generate Scenario'

  const navigation = isReport
    ? ['Golden Dataset', 'Run Evaluation', 'Evaluation Report']
    : ['Generate Scenario', 'Review Scenario', 'History']

  return (
    <div className={`s2g-product-window s2g-product-${mode}`}>
      <div className="s2g-window-bar">
        <span className="s2g-window-controls" aria-hidden="true">
          <i />
          <i />
          <i />
        </span>

        <span>Story2Gherkin</span>
        <span className="s2g-window-label">Workspace</span>
      </div>

      <div className="s2g-product-layout">
        <div className="s2g-product-sidebar">
          <strong>Story2Gherkin</strong>

          <span className="s2g-sidebar-caption">
            {isReport ? 'EVALUATION' : 'SCENARIOS'}
          </span>

          {navigation.map(item => (
            <span
              key={item}
              className={
                item === title
                  ? 's2g-product-nav is-current'
                  : 's2g-product-nav'
              }
            >
              {item}
            </span>
          ))}
        </div>

        <div className="s2g-product-content">
          <div className="s2g-product-heading">
            <strong>{title}</strong>
            <span className="s2g-preview-tag">
              {isReport ? 'Completed' : 'Draft'}
            </span>
          </div>

          {isReport ? (
            <ReportPreview />
          ) : (
            <>
              <div className="s2g-editor-layout">
                <StoryPreview />

                <div className="s2g-preview-editor">
                  <div className="s2g-editor-toolbar">
                    <span>Gherkin scenario</span>
                    <span className="s2g-status">Structure valid</span>
                  </div>

                  <GherkinPreview />
                </div>
              </div>

              {mode === 'review' && (
                <div className="s2g-validation-summary">
                  <div>
                    <span>Structure validation</span>
                    <strong>Valid Gherkin</strong>
                  </div>

                  <div>
                    <span>GUI coverage</span>
                    <strong>Review matched elements</strong>
                  </div>

                  <span className="s2g-preview-save">Save changes</span>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  )
}

function MagneticCTA({
  to,
  children,
}: {
  to: string
  children: string
}) {
  const areaRef = useRef<HTMLSpanElement>(null)

  useEffect(() => {
    const area = areaRef.current

    if (!area) return

    const pointerMedia = window.matchMedia(
      '(hover: hover) and (pointer: fine)',
    )

    const motionMedia = window.matchMedia(
      '(prefers-reduced-motion: reduce)',
    )

    let frame = 0

    const resetPosition = () => {
      window.cancelAnimationFrame(frame)
      frame = 0

      area.style.setProperty('--magnetic-x', '0px')
      area.style.setProperty('--magnetic-y', '0px')
    }

    const handlePointerMove = (event: PointerEvent) => {
      if (
        event.pointerType !== 'mouse' ||
        !pointerMedia.matches ||
        motionMedia.matches
      ) {
        resetPosition()
        return
      }

      const pointerX = event.clientX
      const pointerY = event.clientY

      window.cancelAnimationFrame(frame)

      frame = window.requestAnimationFrame(() => {
        const bounds = area.getBoundingClientRect()

        const distanceX = Math.max(
          bounds.left - pointerX,
          0,
          pointerX - bounds.right,
        )

        const distanceY = Math.max(
          bounds.top - pointerY,
          0,
          pointerY - bounds.bottom,
        )

        const distanceFromButton = Math.hypot(distanceX, distanceY)

        const magneticReach = 56
        const maximumMovement = 10

        if (distanceFromButton > magneticReach) {
          area.style.setProperty('--magnetic-x', '0px')
          area.style.setProperty('--magnetic-y', '0px')
          return
        }

        const centerX = bounds.left + bounds.width / 2
        const centerY = bounds.top + bounds.height / 2

        const normalizedX =
          (pointerX - centerX) / (bounds.width / 2 + magneticReach)

        const normalizedY =
          (pointerY - centerY) / (bounds.height / 2 + magneticReach)

        const offsetX =
          Math.max(-1, Math.min(1, normalizedX)) * maximumMovement

        const offsetY =
          Math.max(-1, Math.min(1, normalizedY)) * maximumMovement

        area.style.setProperty('--magnetic-x', `${offsetX.toFixed(2)}px`)
        area.style.setProperty('--magnetic-y', `${offsetY.toFixed(2)}px`)
      })
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Tab') {
        resetPosition()
      }
    }

    window.addEventListener('pointermove', handlePointerMove, {
      passive: true,
    })

    window.addEventListener('scroll', resetPosition, {
      passive: true,
      capture: true,
    })

    window.addEventListener('blur', resetPosition)
    window.addEventListener('keydown', handleKeyDown)

    document.documentElement.addEventListener(
      'mouseleave',
      resetPosition,
    )

    pointerMedia.addEventListener('change', resetPosition)
    motionMedia.addEventListener('change', resetPosition)

    return () => {
      window.cancelAnimationFrame(frame)

      window.removeEventListener('pointermove', handlePointerMove)
      window.removeEventListener('scroll', resetPosition, true)
      window.removeEventListener('blur', resetPosition)
      window.removeEventListener('keydown', handleKeyDown)

      document.documentElement.removeEventListener(
        'mouseleave',
        resetPosition,
      )

      pointerMedia.removeEventListener('change', resetPosition)
      motionMedia.removeEventListener('change', resetPosition)

      area.style.removeProperty('--magnetic-x')
      area.style.removeProperty('--magnetic-y')
    }
  }, [])

  return (
    <span ref={areaRef} className="s2g-magnetic-area">
      <Link className="s2g-button s2g-button-large" to={to}>
        {children}
      </Link>
    </span>
  )
}

function MediaPlaceholder({
  variant,
}: {
  variant: 'hero' | 'workflow'
}) {
  return (
    <div
      className={`s2g-media-placeholder s2g-media-placeholder-${variant}`}
      role="img"
      aria-label={
        variant === 'hero'
          ? 'Reserved space for the Story2Gherkin product demonstration'
          : 'Reserved space for the System Analyst and QA workflow animation'
      }
    />
  )
}

export default function HomePage() {
  const session = useAuthSession()

  const workspacePath =
    session?.role === 'system_analyst'
      ? '/generate'
      : session?.role === 'qa'
        ? '/evaluation/reports'
        : '/register'

  const actionLabel = session ? 'Open Workspace' : 'Get Started'

  return (
    <div className="s2g-home">
      <PublicHeader />

      <main>
        <section className="s2g-hero s2g-container">
          <div className="s2g-hero-copy">
            <h1>From user stories to test-ready Gherkin.</h1>

            <p>
              Help System Analysts create and review Gherkin scenarios and Quality Assurance teams evaluate their quality in one platform. 
            </p>

            <MagneticCTA to={workspacePath}>
              {actionLabel}
            </MagneticCTA>
          </div>

          <div className="s2g-hero-visual">
            <MediaPlaceholder variant="hero" />
          </div>
        </section>

        <section
          className="s2g-workflow s2g-container"
          aria-labelledby="s2g-workflow-title"
        >
          <div className="s2g-section-heading">
            <h2 id="s2g-workflow-title">
              Turn clear requirements into structured scenarios
            </h2>

            <p>
              Move from a user story to a Gherkin draft, bring in QA
              feedback, and refine the scenario with confidence.
            </p>
          </div>

          <WorkflowAnimation />

          <div className="s2g-benefits">
            <div>
              <h3>Less repetitive writing</h3>

              <p>
                Create a structured first draft from your requirements
                and spend more time refining the details.
              </p>
            </div>

            <div>
              <h3>Visible structure and coverage</h3>

              <p>
                Check Gherkin structure and identify GUI elements that
                still need attention.
              </p>
            </div>

            <div>
              <h3>LLM and rule-based comparison</h3>

              <p>
                Inspect both approaches through evaluation reports
                and downloadable results.
              </p>
            </div>
          </div>
        </section>

        <section
          className="s2g-roles s2g-container"
          aria-labelledby="s2g-roles-title"
        >
          <div className="s2g-section-heading">
            <h2 id="s2g-roles-title">
              Built around how your team works
            </h2>
          </div>

          <div className="s2g-role-row">
            <div className="s2g-role-copy">
              <p className="s2g-role-label">System Analyst</p>

              <h3>Shape the scenario before it ships.</h3>

              <p>
                Generate a draft, refine the steps, and revisit previous
                requests without losing context.
              </p>

              <p className="s2g-role-features">
                Generate <span>/</span> Review <span>/</span> History
              </p>
            </div>

            <div className="s2g-role-visual">
              <ProductWindow mode="review" />
            </div>
          </div>

          <div className="s2g-role-row s2g-role-row-reversed">
            <div className="s2g-role-copy">
              <p className="s2g-role-label">QA</p>

              <h3>Evaluate with evidence.</h3>

              <p>
                Organize golden datasets and inspect LLM and rule-based
                results side by side.
              </p>

              <p className="s2g-role-features">
                Golden Dataset <span>/</span> Reports
              </p>
            </div>

            <div className="s2g-role-visual">
              <ProductWindow mode="report" />
            </div>
          </div>

          <p className="s2g-preview-disclaimer">
            Product previews are illustrative.
          </p>
        </section>

        <section className="s2g-final s2g-container">
          <h2>Make your next story test-ready.</h2>

          <p>
            Bring more clarity and consistency to your scenario testing
            workflow.
          </p>

          <div className="s2g-final-actions">
            <MagneticCTA to={workspacePath}>
              {actionLabel}
            </MagneticCTA>

            <Link className="s2g-text-link" to="/documentation">
              Read Documentation
            </Link>
          </div>
        </section>
      </main>

      <footer className="s2g-footer s2g-container">

        <nav aria-label="Footer navigation">
          <Link to="/documentation">Documentation</Link>
          <Link to="/login">Log in</Link>
        </nav>
      </footer>
    </div>
  )
}