import { AlertTriangle, Check, X } from 'lucide-react'
import type { ScenarioValidation } from '../../services/reviewApi'

interface Props {
  validation: ScenarioValidation
  draft: boolean
  validatedDraft: boolean
  feedback: { kind: 'success' | 'error'; message: string } | null
}

export default function ValidationPanel({ validation, draft, validatedDraft, feedback }: Props) {
  const labels = { feature: 'Feature', scenario: 'Scenario', given: 'Given', when: 'When', then: 'Then' } as const
  const coverageAvailable = validation.coverage_available !== false
  return (
    <section className="review-panel review-validation" aria-labelledby="review-validation-title">
      <h2 id="review-validation-title">Validation Result</h2>
      {draft && (
  <p className="review-draft-note">
    {validatedDraft
      ? 'Draft validation results. Changes have not been saved.'
      : 'This draft has not been saved. The panel shows validation results for the saved version; save your changes to validate them.'}
  </p>
)}
      <div className={`review-structure ${validation.valid ? 'valid' : 'invalid'}`}>
        <div className="review-structure-heading"><h3>Gherkin Structure</h3><span className="review-status">{validation.valid ? 'VALID' : 'INVALID'}</span></div>
        <ul className="review-checks">
          {Object.entries(labels).map(([key, label]) => {
            const valid = validation.checks[key as keyof typeof labels]
            const description = valid ? `Includes ${label}` : key === 'feature' || key === 'scenario'
              ? `${label} is incomplete or invalid`
              : `${label} is incomplete; check the step and its description`
            return <li key={key}>{valid ? <Check className="review-check" size={19} /> : <X className="review-cross" size={19} />}<span>{description}</span></li>
          })}
        </ul>
      </div>
      {validation.issues.length > 0 && (
          <ul className="review-issues" aria-label="Validation errors">
            {validation.issues.map((issue, index) => (
              <li key={index}>
                Line {issue.line}, column {issue.column}: {issue.message}
              </li>
            ))}
          </ul>
        )}
      {coverageAvailable ? <>
      <div className="review-coverage-heading"><h3>GUI Coverage</h3><strong>{validation.coverage_percent}%</strong></div>
      <div className="review-progress" role="progressbar" aria-label="GUI Coverage" aria-valuemin={0} aria-valuemax={100} aria-valuenow={validation.coverage_percent}><span style={{ width: `${validation.coverage_percent}%` }} /></div>
      <h4>USED ({validation.used_elements.length})</h4>
      <ul className="review-elements">{validation.used_elements.map((element, index) => <li key={index}><Check className="review-check" size={19} /><span>{element}</span></li>)}</ul>
      <h4>NOT COVERED ({validation.not_covered_elements.length})</h4>
      <ul className="review-uncovered">{validation.not_covered_elements.map((element, index) => <li key={index}><AlertTriangle size={19} /><span>{element}</span></li>)}</ul>
      {validation.not_covered_elements.length > 0 && <p className="review-coverage-note">Valid scenarios can still be saved with coverage below 100%.</p>}
      </> : <div className="review-coverage-unavailable"><h3>GUI Coverage</h3><p>Coverage cannot be calculated yet. Fix the Gherkin structure first so that GUI elements can be checked.</p></div>}
      {feedback && <div className={`review-feedback ${feedback.kind}`} role={feedback.kind === 'error' ? 'alert' : 'status'}>{feedback.kind === 'success' ? <Check size={21} /> : <AlertTriangle size={21} />}<span>{feedback.message}</span></div>}
    </section>
  )
}
