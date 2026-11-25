import { useEffect, useMemo, useState } from 'react'
import type { ReviewScenario } from '../../services/reviewApi'
import { listBehatRuns, runBehat, type BehatRun } from '../../services/behatRunApi'

type ParsedScenario = { name: string; steps: string[] }

function parseScenarios(gherkin: string): ParsedScenario[] {
  const scenarios: ParsedScenario[] = []
  let current: ParsedScenario | null = null
  for (const line of gherkin.split('\n')) {
    const trimmed = line.trim()
    const heading = trimmed.match(/^Scenario(?: Outline)?:\s*(.+)$/i)
    if (heading) {
      current = { name: heading[1].trim(), steps: [] }
      scenarios.push(current)
      continue
    }
    if (current && /^(Given|When|Then|And|But)\s+/i.test(trimmed)) {
      current.steps.push(trimmed)
    }
  }
  return scenarios
}

function placeholders(steps: string[]): string[] {
  const found = new Set<string>()
  for (const step of steps) {
    for (const match of step.matchAll(/<([A-Za-z0-9_]+)>/g)) found.add(match[1])
  }
  return [...found]
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
  const scenarios = useMemo(() => parseScenarios(saved.gherkin), [saved.gherkin])
  const [scenarioIndex, setScenarioIndex] = useState(0)
  const scenario = scenarios[scenarioIndex]
  const steps = scenario?.steps ?? []
  const keys = useMemo(() => placeholders(steps), [steps])

  const [baseUrl, setBaseUrl] = useState('')
  const [values, setValues] = useState<Record<string, string>>({})
  const [running, setRunning] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState<BehatRun | null>(null)
  const [history, setHistory] = useState<BehatRun[]>([])

  useEffect(() => {
    setResult(null)
    setError('')
  }, [scenarioIndex])

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

  const missing = keys.some((key) => !values[key]?.trim())

  async function start() {
    if (!scenario || !baseUrl.trim() || missing) return
    setRunning(true)
    setError('')
    setResult(null)
    try {
      const run = await runBehat(saved.id, baseUrl.trim(), scenario.name, values)
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
      <p className="behat-help">
        Isi Base URL dan nilai placeholder. Langkah Gherkin dijalankan otomatis.
      </p>

      <label>
        Scenario
        <select
          value={scenarioIndex}
          onChange={(event) => setScenarioIndex(Number(event.target.value))}
          disabled={running}
        >
          {scenarios.map((item, index) => (
            <option key={`${index}-${item.name}`} value={index}>{item.name}</option>
          ))}
        </select>
      </label>

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

      {keys.length > 0 && (
        <fieldset className="behat-values">
          <legend>Values</legend>
          {keys.map((key) => (
            <label key={key}>
              {key}
              <input
                value={values[key] ?? ''}
                onChange={(event) => setValues((current) => ({ ...current, [key]: event.target.value }))}
                disabled={running}
              />
            </label>
          ))}
        </fieldset>
      )}

      <ol className="behat-steps">
        {steps.map((text, index) => (
          <li key={`${index}-${text}`}>{text}</li>
        ))}
      </ol>

      {error && <p role="alert">{error}</p>}

      <button
        type="button"
        className="review-button primary"
        onClick={start}
        disabled={running || missing || !baseUrl.trim()}
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
