import { useEffect, useMemo, useState } from 'react'
import type { ReviewScenario } from '../../services/reviewApi'
import {
  listBehatRuns,
  runBehat,
  type BehatAction,
  type BehatRun,
  type BehatStepInput,
} from '../../services/behatRunApi'

const ACTIONS: { value: BehatAction; label: string }[] = [
  { value: 'open', label: 'Open page' },
  { value: 'fill', label: 'Fill field' },
  { value: 'click', label: 'Click' },
  { value: 'assert_text', label: 'Check text' },
  { value: 'assert_visible', label: 'Check visible' },
  { value: 'assert_url', label: 'Check URL' },
]

type Mapping = { action: BehatAction | ''; selector: string; value: string }

function parseSteps(gherkin: string): string[] {
  const steps: string[] = []
  for (const line of gherkin.split('\n')) {
    const trimmed = line.trim()
    if (/^(Given|When|Then|And|But)\s+/i.test(trimmed)) steps.push(trimmed)
  }
  return steps
}

function loadMapping(scenarioId: string, count: number): Mapping[] {
  try {
    const raw = localStorage.getItem(`behat-mapping-${scenarioId}`)
    if (raw) {
      const parsed = JSON.parse(raw) as Mapping[]
      if (Array.isArray(parsed) && parsed.length === count) return parsed
    }
  } catch {
    // ignore malformed storage
  }
  return Array.from({ length: count }, () => ({ action: '', selector: '', value: '' }))
}

function needsSelector(action: BehatAction | '') {
  return action === 'fill' || action === 'click' || action === 'assert_visible'
}

function needsValue(action: BehatAction | '') {
  return action === 'open' || action === 'fill' || action === 'assert_text' || action === 'assert_url'
}

export default function BehatRunPanel({ saved }: { saved: ReviewScenario }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="review-behat">
      <button
        type="button"
        className="review-button secondary"
        aria-expanded={open}
        onClick={() => setOpen((value) => !value)}
      >
        {open ? 'Close end-to-end test' : 'Run end-to-end test'}
      </button>
      {open && <Workspace saved={saved} />}
    </div>
  )
}

function Workspace({ saved }: { saved: ReviewScenario }) {
  const steps = useMemo(() => parseSteps(saved.gherkin), [saved.gherkin])
  const [baseUrl, setBaseUrl] = useState('')
  const [mapping, setMapping] = useState<Mapping[]>(() => loadMapping(saved.id, steps.length))
  const [running, setRunning] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState<BehatRun | null>(null)
  const [history, setHistory] = useState<BehatRun[]>([])

  useEffect(() => {
    localStorage.setItem(`behat-mapping-${saved.id}`, JSON.stringify(mapping))
  }, [mapping, saved.id])

  useEffect(() => {
    const controller = new AbortController()
    listBehatRuns(saved.id)
      .then((runs) => {
        if (!controller.signal.aborted) setHistory(runs)
      })
      .catch(() => {
        // history is optional
      })
    return () => controller.abort()
  }, [saved.id])

  function update(index: number, patch: Partial<Mapping>) {
    setMapping((current) =>
      current.map((item, position) => (position === index ? { ...item, ...patch } : item)),
    )
  }

  const incomplete = steps.some((_, index) => {
    const item = mapping[index]
    if (!item?.action) return true
    if (needsSelector(item.action) && !item.selector.trim()) return true
    if (needsValue(item.action) && !item.value.trim()) return true
    return false
  })

  async function start() {
    if (incomplete || !baseUrl.trim()) return
    setRunning(true)
    setError('')
    setResult(null)
    try {
      const payload: BehatStepInput[] = steps.map((text, index) => ({
        text,
        action: mapping[index].action as BehatAction,
        selector: mapping[index].selector.trim(),
        value: mapping[index].value.trim(),
      }))
      const run = await runBehat(saved.id, baseUrl.trim(), payload)
      setResult(run)
      setHistory((current) => [run, ...current].slice(0, 10))
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'The end-to-end test failed.')
    } finally {
      setRunning(false)
    }
  }

  return (
    <section className="behat-workspace" aria-label="End-to-end test">
      <h3>End-to-end test</h3>

      <label>
        Base URL
        <input
          type="url"
          value={baseUrl}
          onChange={(event) => setBaseUrl(event.target.value)}
          placeholder="https://staging.example.com"
          disabled={running}
        />
      </label>

      <ol className="behat-steps">
        {steps.map((text, index) => (
          <li key={`${index}-${text}`}>
            <p>{text}</p>
            <select
              value={mapping[index]?.action ?? ''}
              onChange={(event) => update(index, { action: event.target.value as BehatAction | '' })}
              disabled={running}
            >
              <option value="">Choose action…</option>
              {ACTIONS.map((action) => (
                <option key={action.value} value={action.value}>{action.label}</option>
              ))}
            </select>
            {needsSelector(mapping[index]?.action ?? '') && (
              <input
                value={mapping[index]?.selector ?? ''}
                onChange={(event) => update(index, { selector: event.target.value })}
                placeholder="CSS selector, e.g. #email"
                disabled={running}
              />
            )}
            {needsValue(mapping[index]?.action ?? '') && (
              <input
                value={mapping[index]?.value ?? ''}
                onChange={(event) => update(index, { value: event.target.value })}
                placeholder={mapping[index]?.action === 'open' ? 'Path, e.g. /login' : 'Expected value'}
                disabled={running}
              />
            )}
          </li>
        ))}
      </ol>

      {error && <p role="alert">{error}</p>}

      <button
        type="button"
        className="review-button primary"
        onClick={start}
        disabled={running || incomplete || !baseUrl.trim()}
      >
        {running ? 'Running…' : 'Run test'}
      </button>

      {result && (
        <div className="behat-result">
          <p><strong>{result.status}</strong> in {result.duration_ms} ms</p>
          <ol>
            {result.steps.map((step) => (
              <li key={step.index} className={`behat-step ${step.status.toLowerCase()}`}>
                <span>{step.status}</span> {step.text}
                {step.message && <em> — {step.message}</em>}
              </li>
            ))}
          </ol>
        </div>
      )}

      {history.length > 0 && (
        <details>
          <summary>Recent runs</summary>
          <ul>
            {history.map((run) => (
              <li key={run.run_id}>
                <span className={`behat-step ${run.status.toLowerCase()}`}>{run.status}</span>{' '}
                {run.created_at ? new Date(run.created_at).toLocaleString('id-ID') : ''} · {run.duration_ms} ms
              </li>
            ))}
          </ul>
        </details>
      )}
    </section>
  )
}
