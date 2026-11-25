import { useEffect, useRef, useState, useSyncExternalStore } from 'react'
import logo from '../assets/story2gherkin-logo.svg'
import './AuthIllustration.css'

const MOTION_QUERY = '(prefers-reduced-motion: reduce)'
const LOOP_MS = 20000
const STORY = 'As a registered user, I want to log in so that I can access my workspace.'
const FREE_SCENARIO = 'Enter an email address and password, then select Log in.'
const CODE = [
  ['Feature:', ' Account access'],
  ['Scenario:', ' Successful login'],
  ['Given', ' the user is on the login page'],
  ['When', ' the user submits valid credentials'],
  ['Then', ' the dashboard is displayed'],
]

function subscribeMotion(callback: () => void) {
  const media = window.matchMedia(MOTION_QUERY)
  media.addEventListener('change', callback)
  return () => media.removeEventListener('change', callback)
}
function motionSnapshot() { return window.matchMedia(MOTION_QUERY).matches }
function serverSnapshot() { return false }
function typed(text: string, time: number, start: number, duration: number) {
  return text.slice(0, Math.floor(text.length * Math.max(0, Math.min(1, (time - start) / duration))))
}

function DocumentMark() {
  return (
    <svg viewBox="0 0 28 32" fill="none" aria-hidden="true">
      <path d="M5 2h11l7 7v19a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2Z" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round" />
      <path d="M16 2v8h7M8 16h10M8 21h10M8 26h6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  )
}

export default function AuthIllustration() {
  const root = useRef<HTMLElement>(null)
  const elapsedRef = useRef(0)
  const [elapsed, setElapsed] = useState(0)
  const [paused, setPaused] = useState(false)
  const reducedMotion = useSyncExternalStore(subscribeMotion, motionSnapshot, serverSnapshot)

  useEffect(() => {
    if (paused || reducedMotion) return
    let frame = 0
    let previous: number | null = null
    let painted = 0
    let visible = true
    const observer = typeof IntersectionObserver !== 'undefined'
      ? new IntersectionObserver(([entry]) => { if (entry) visible = entry.isIntersecting }, { threshold: 0.1 })
      : null
    if (root.current) observer?.observe(root.current)
    const tick = (now: number) => {
      if (previous !== null && visible && !document.hidden) {
        elapsedRef.current = (elapsedRef.current + Math.min(now - previous, 100)) % LOOP_MS
      }
      previous = now
      if (visible && !document.hidden && now - painted >= 60) {
        setElapsed(elapsedRef.current)
        painted = now
      }
      frame = window.requestAnimationFrame(tick)
    }
    frame = window.requestAnimationFrame(tick)
    return () => { window.cancelAnimationFrame(frame); observer?.disconnect() }
  }, [paused, reducedMotion])

  const time = reducedMotion ? 16000 : elapsed
  const story = typed(STORY, time, 1000, 3400)
  const freeScenario = typed(FREE_SCENARIO, time, 4300, 3000)
  const codeStarted = time >= 7600
  const completed = time >= 12400
  const guiCount = time < 2800 ? 0 : time < 3900 ? 1 : time < 5000 ? 2 : 3

  return (
    <section ref={root} className="auth-art" data-paused={paused || reducedMotion} aria-label="Animated preview of a user story becoming a Gherkin scenario">
      <div className="auth-art-scene" aria-hidden="true">
        <div className={`auth-art-card auth-art-file${completed ? ' is-active' : ''}`}>
          <span className="auth-art-file-icon"><DocumentMark /></span>
          <span>.feature</span>
        </div>

        <div className={`auth-art-card auth-art-story${time >= 1000 && time < 5000 ? ' is-active' : ''}`}>
          <div className="auth-art-card-title"><span className="auth-art-label-mark">US</span><strong>User Story</strong></div>
          <p>{story || 'Describe a user goal…'}{time >= 1000 && time < 4400 && <i className="auth-art-caret" />}</p>
        </div>

        <div className={`auth-art-window${codeStarted ? ' is-active' : ''}`}>
          <div className="auth-art-window-header">
            <span className="auth-art-window-dots"><i /><i /><i /></span>
            <span>gherkin_preview.feature</span>
          </div>
          <div className="auth-art-window-body">
            <div className="auth-art-code" role="presentation">
              {CODE.map(([keyword, content], index) => {
                const start = 7600 + index * 760
                const line = typed(content, time, start, 700)
                return (
                  <div key={keyword} className={`auth-art-code-line${time >= start ? ' is-visible' : ''}`}>
                    <span className="auth-art-code-number">{index + 1}</span>
                    <span><b>{keyword}</b>{line}{time >= start && time < start + 700 && <i className="auth-art-caret" />}</span>
                  </div>
                )
              })}
              {!codeStarted && <span className="auth-art-code-placeholder">Your Gherkin draft appears here.</span>}
            </div>
            <img src={logo} alt="" className="auth-art-brand" />
          </div>
          <div className="auth-art-window-footer">
            <span>{completed ? 'Ready to review' : codeStarted ? 'Generating scenario…' : 'User story → Gherkin'}</span>
            <span className={`auth-art-valid${completed ? ' is-visible' : ''}`}>Structure valid</span>
          </div>
        </div>

        <div className={`auth-art-card auth-art-gui${time >= 2800 && time < 7600 ? ' is-active' : ''}`}>
          <div className="auth-art-card-title"><span className="auth-art-label-mark">GUI</span><strong>GUI Elements</strong></div>
          {['Email field', 'Password field', 'Log in button'].map((label, index) => (
            <div key={label} className={`auth-art-gui-row${guiCount > index ? ' is-checked' : ''}`}>
              <span className="auth-art-checkbox">{guiCount > index ? '✓' : ''}</span><span>{label}</span>
            </div>
          ))}
        </div>

        <div className={`auth-art-card auth-art-free${time >= 4300 && time < 7600 ? ' is-active' : ''}`}>
          <div className="auth-art-card-title"><span className="auth-art-label-mark">FS</span><strong>Free Scenario</strong></div>
          <p>{freeScenario || 'Describe the intended flow…'}{time >= 4300 && time < 7300 && <i className="auth-art-caret" />}</p>
        </div>
      </div>

      {!reducedMotion && (
        <button
          className="auth-art-motion-control"
          type="button"
          aria-pressed={paused}
          onClick={() => setPaused(value => !value)}
        >
          {paused ? 'Resume preview animation' : 'Pause preview animation'}
        </button>
      )}
    </section>
  )
}
