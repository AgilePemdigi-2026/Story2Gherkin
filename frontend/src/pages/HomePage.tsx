import { useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'

import PublicHeader from '../components/PublicHeader'
import { useAuthSession } from '../hooks/useAuthSession'

import './HomePage.css'
import WorkflowAnimation from '../components/WorkflowAnimation'


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
            <video
  className="s2g-hook-video"
  src="/videos/story2gherkin-hook.mp4"
  poster="/videos/story2gherkin-hook-poster.png"
  autoPlay
  muted
  loop
  playsInline
  preload="metadata"
  aria-label="Story2Gherkin: System Analyst and QA workflow"
/>
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
              <video
  className="s2g-sa-demo-video"
  src="/videos/system-analyst-actual-ui-no-mouse.mp4"
  poster="/videos/system-analyst-actual-ui-poster.png"
  autoPlay
  muted
  loop
  playsInline
  preload="metadata"
  aria-label="System Analyst demo"
/>
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
              <video
  className="s2g-qa-demo-video"
  src="/videos/qa-actual-ui-no-mouse.mp4"
  poster="/videos/qa-actual-ui-poster.png"
  autoPlay
  muted
  loop
  playsInline
  preload="metadata"
  aria-label="QA workflow demo"
/>
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