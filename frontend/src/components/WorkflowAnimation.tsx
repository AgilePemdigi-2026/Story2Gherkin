import {
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from 'react'

import saCharacter from '../assets/workflow-sa.png'
import qaCharacter from '../assets/workflow-qa.png'

import './WorkflowAnimation.css'

const REDUCED_MOTION_QUERY = '(prefers-reduced-motion: reduce)'
const LOOP_DURATION = 30000

const USER_STORY =
  'As a registered user, I want to reset my password so that I can regain access to my account.'

const SA_GOAL =
  'I want to turn my user stories into clear Gherkin scenarios.'

const SA_REFINEMENT_GOAL =
  'I want each scenario to reflect the complete user journey.'

const QA_GOAL =
  'I want to assess scenario quality against the golden dataset.'

const QA_COMPARISON_GOAL =
  'I want to compare LLM and rule-based results with confidence.'

const DRAFT_LINES = [
  'Feature: Password reset',
  '',
  '  Scenario: Request a reset link',
  '    Given I am on the password reset page',
  '    When I enter my registered email address',
  '    Then I should receive a password reset link',
]

const ADDED_LINE = '    And I select "Send reset link"'
const DRAFT = DRAFT_LINES.join('\n')

function subscribeReducedMotion(callback: () => void) {
  const media = window.matchMedia(REDUCED_MOTION_QUERY)

  media.addEventListener('change', callback)

  return () => {
    media.removeEventListener('change', callback)
  }
}

function getReducedMotionSnapshot() {
  return window.matchMedia(REDUCED_MOTION_QUERY).matches
}

function getServerSnapshot() {
  return false
}

function typeText(
  text: string,
  time: number,
  start: number,
  duration: number,
) {
  const progress = Math.max(
    0,
    Math.min(1, (time - start) / duration),
  )

  return text.slice(0, Math.floor(text.length * progress))
}

function getPhase(time: number) {
  if (time < 1200 || time >= 26500) return 'idle'
  if (time < 7600) return 'story'
  if (time < 9200) return 'generating'
  if (time < 12600) return 'draft'
  if (time < 16400) return 'feedback'
  if (time < 21800) return 'revision'

  return 'complete'
}

function CodeLine({
  line,
  added,
}: {
  line: string
  added: boolean
}) {
  const match = line.match(
    /^(\s*)(Feature:|Scenario:|Given|When|Then|And)(.*)$/,
  )

  return (
    <span className={`workflow-code-line${added ? ' is-added' : ''}`}>
      {match ? (
        <>
          {match[1]}
          <span className="workflow-code-keyword">
            {match[2]}
          </span>
          {match[3]}
        </>
      ) : (
        line || '\u00a0'
      )}
    </span>
  )
}

function WorkflowPerson({
  role,
  active,
  message,
  messageId,
  typing,
}: {
  role: 'sa' | 'qa'
  active: boolean
  message: string
  messageId: string
  typing: boolean
}) {
  const isSA = role === 'sa'

  return (
    <div
      className={`workflow-person workflow-person-${role}${
        active ? ' is-active' : ''
      }`}
    >
      <div className="workflow-person-heading">
        <div className="workflow-character">
          <img
            src={isSA ? saCharacter : qaCharacter}
            alt=""
            width={104}
            height={104}
            draggable={false}
          />
        </div>

        <div className="workflow-person-label">
          <strong>{isSA ? 'System Analyst' : 'QA'}</strong>

          <span>
            {isSA
              ? 'Clear scenarios from user requirements'
              : 'Confidence in scenario quality'}
          </span>
        </div>
      </div>

      <div className="workflow-bubble-space">
        {message && (
          <div
            key={messageId}
            className="workflow-chat-bubble"
          >
            <span className="workflow-chat-author">
              {isSA ? 'System Analyst' : 'QA'}
            </span>

            <div className="workflow-chat-message">
              {message}

              {typing && (
                <span className="workflow-caret" />
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default function WorkflowAnimation() {
  const rootRef = useRef<HTMLElement>(null)
  const elapsedRef = useRef(0)

  const [elapsed, setElapsed] = useState(0)
  const [paused] = useState(false)

  const reducedMotion = useSyncExternalStore(
    subscribeReducedMotion,
    getReducedMotionSnapshot,
    getServerSnapshot,
  )

  useEffect(() => {
    if (paused || reducedMotion) return

    let frame = 0
    let previousTime: number | null = null
    let previousPaint = 0
    let visible = true

    const root = rootRef.current

    const observer =
      root && typeof IntersectionObserver !== 'undefined'
        ? new IntersectionObserver(
            ([entry]) => {
              if (entry) {
                visible = entry.isIntersecting
              }
            },
            { threshold: 0.1 },
          )
        : null

    if (root) {
      observer?.observe(root)
    }

    const tick = (now: number) => {
      if (
        previousTime !== null &&
        visible &&
        !document.hidden
      ) {
        const delta = Math.min(now - previousTime, 100)

        elapsedRef.current =
          (elapsedRef.current + delta) % LOOP_DURATION
      }

      previousTime = now

      if (
        visible &&
        !document.hidden &&
        now - previousPaint >= 40
      ) {
        setElapsed(elapsedRef.current)
        previousPaint = now
      }

      frame = window.requestAnimationFrame(tick)
    }

    frame = window.requestAnimationFrame(tick)

    return () => {
      window.cancelAnimationFrame(frame)
      observer?.disconnect()
    }
  }, [paused, reducedMotion])

  const time = reducedMotion ? 25000 : elapsed
  const phase = getPhase(time)

  const isIdle = phase === 'idle'
  const isGenerating = phase === 'generating'
  const isRevision = phase === 'revision'
  const isComplete = phase === 'complete'

  const saActive =
    phase === 'story' ||
    phase === 'generating' ||
    phase === 'draft' ||
    phase === 'revision'

  const qaActive =
    phase === 'feedback' ||
    phase === 'complete'

  const story = isIdle
    ? ''
    : typeText(USER_STORY, time, 3400, 4200)

  let code = ''

  if (time >= 9200 && time < 18800) {
    code = typeText(DRAFT, time, 9200, 3200)
  } else if (time >= 18800 && time < 26500) {
    const updatedLines = [...DRAFT_LINES]
    const addedLine = typeText(ADDED_LINE, time, 18800, 2600)

    if (addedLine) {
      updatedLines.splice(updatedLines.length - 1, 0, addedLine)
    }

    code = updatedLines.join('\n')
  }

  // Keep each goal visible until the next goal replaces it.
  const saMessage =
    time >= 1200 && time < 16400
      ? typeText(SA_GOAL, time, 1200, 2800)
      : time >= 16400 && time < 26500
        ? typeText(SA_REFINEMENT_GOAL, time, 16400, 2800)
        : ''

  const qaMessage =
    time >= 12600 && time < 21800
      ? typeText(QA_GOAL, time, 12600, 2800)
      : time >= 21800 && time < 26500
        ? typeText(QA_COMPARISON_GOAL, time, 21800, 2800)
        : ''

  const storyTyping = time >= 3400 && time < 7600

  const codeTyping =
    (time >= 9200 && time < 12400) ||
    (time >= 18800 && time < 21400)

  const saTyping =
    (time >= 1200 && time < 4000) ||
    (time >= 16400 && time < 19200)

  const qaTyping =
    (time >= 12600 && time < 15400) ||
    (time >= 21800 && time < 24600)

  const statusText = isIdle
    ? 'Ready for a user story'
    : isGenerating
      ? 'Generating scenario'
      : phase === 'feedback'
        ? 'Checking intended behavior'
        : isRevision
          ? 'Refining scenario'
          : isComplete
            ? 'Changes saved'
            : phase === 'draft'
              ? 'Draft generated'
              : 'Writing user story'

  return (
    <section
      ref={rootRef}
      className="workflow-animation"
      data-phase={phase}
      data-paused={paused || reducedMotion}
      aria-label="Illustrated System Analyst and QA workflow"
    >
      <p className="workflow-sr-only">
        System Analysts want clear Gherkin scenarios that reflect
        user requirements. QA teams want confidence in scenario
        quality and comparisons between LLM and rule-based results.
        The illustration shows a user story becoming a Gherkin
        draft, followed by refinement of its steps.
      </p>

      <div className="workflow-stage" aria-hidden="true">
        <WorkflowPerson
          role="sa"
          active={saActive}
          message={saMessage}
          messageId={time >= 16400 ? 'sa-refinement' : 'sa-goal'}
          typing={saTyping}
        />

        <div
          className={`workflow-connector workflow-connector-sa${
            saActive ? ' is-active' : ''
          }`}
        />

        <div className="workflow-app">
          <div className="workflow-app-bar">
            <strong>Story2Gherkin</strong>
            <span>Scenario workspace</span>
          </div>

          <div className="workflow-app-body">
            <div className="workflow-app-heading">
              <span>
                {isRevision || isComplete
                  ? 'Review Scenario'
                  : 'Generate Scenario'}
              </span>

              <span className="workflow-app-state">
                {isComplete ? 'Saved' : isIdle ? 'Ready' : 'Draft'}
              </span>
            </div>

            <div className="workflow-story-field">
              <div className="workflow-field-label">
                User story
              </div>

              <div
                className={`workflow-story-content${
                  story ? ' has-content' : ''
                }`}
              >
                {story || 'Describe what the user wants to achieve...'}

                {storyTyping && (
                  <span className="workflow-caret" />
                )}
              </div>
            </div>

            <div className="workflow-editor">
              <div className="workflow-editor-bar">
                <span>Gherkin scenario</span>

                {code && !codeTyping && (
                  <span className="workflow-valid">
                    Structure valid
                  </span>
                )}
              </div>

              <div className="workflow-editor-content">
                {code ? (
                  <pre>
                    <code>
                      {code.split('\n').map((line, index) => (
                        <CodeLine
                          key={index}
                          line={line}
                          added={
                            (isRevision || isComplete) &&
                            line.trimStart().startsWith('And')
                          }
                        />
                      ))}
                    </code>
                  </pre>
                ) : (
                  <div className="workflow-empty-editor">
                    {isGenerating
                      ? 'Preparing your Gherkin draft…'
                      : 'Your generated scenario will appear here.'}
                  </div>
                )}
              </div>
            </div>

            <div className="workflow-app-footer">
              <span className="workflow-process-label">
                {statusText}
              </span>

              <span
                className={`workflow-action-preview${
                  isGenerating ? ' is-processing' : ''
                }`}
              >
                {isComplete
                  ? 'Saved'
                  : isRevision
                    ? 'Save changes'
                    : isGenerating
                      ? 'Generating…'
                      : 'Generate'}
              </span>
            </div>
          </div>
        </div>

        <div
          className={`workflow-connector workflow-connector-qa${
            qaActive ? ' is-active' : ''
          }`}
        />

        <WorkflowPerson
          role="qa"
          active={qaActive}
          message={qaMessage}
          messageId={time >= 21800 ? 'qa-comparison' : 'qa-goal'}
          typing={qaTyping}
        />
      </div>
    </section>
  )
}