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
  return (
    <section className="review-panel review-validation" aria-labelledby="review-validation-title">
      <h2 id="review-validation-title">Validation Result</h2>
      {draft && <p className="review-draft-note">{validatedDraft ? 'Hasil validasi draft. Perubahan belum disimpan.' : 'Draft belum disimpan. Panel menampilkan validasi versi tersimpan; simpan untuk memvalidasi perubahan.'}</p>}
      <div className={`review-structure ${validation.valid ? 'valid' : 'invalid'}`}>
        <div className="review-structure-heading"><h3>Struktur Gherkin</h3><span className="review-status">{validation.valid ? 'VALID' : 'INVALID'}</span></div>
        <ul className="review-checks">
          {Object.entries(labels).map(([key, label]) => {
            const valid = validation.checks[key as keyof typeof labels]
            return <li key={key}>{valid ? <Check className="review-check" size={19} /> : <X className="review-cross" size={19} />}Memiliki {label}</li>
          })}
        </ul>
      </div>
      {validation.issues.length > 0 && <ul className="review-issues" aria-label="Kesalahan validasi">{validation.issues.map((issue, index) => <li key={index}>Baris {issue.line}, kolom {issue.column}: {issue.message}</li>)}</ul>}
      <div className="review-coverage-heading"><h3>GUI Coverage</h3><strong>{validation.coverage_percent}%</strong></div>
      <div className="review-progress" role="progressbar" aria-label="GUI Coverage" aria-valuemin={0} aria-valuemax={100} aria-valuenow={validation.coverage_percent}><span style={{ width: `${validation.coverage_percent}%` }} /></div>
      <h4>USED ({validation.used_elements.length})</h4>
      <ul className="review-elements">{validation.used_elements.map((element, index) => <li key={index}><Check className="review-check" size={19} /><span>{element}</span></li>)}</ul>
      <h4>NOT COVERED ({validation.not_covered_elements.length})</h4>
      <ul className="review-uncovered">{validation.not_covered_elements.map((element, index) => <li key={index}><AlertTriangle size={19} /><span>{element}</span></li>)}</ul>
      {validation.not_covered_elements.length > 0 && <p className="review-coverage-note">Coverage di bawah 100% tidak menghalangi penyimpanan skenario yang valid.</p>}
      {feedback && <div className={`review-feedback ${feedback.kind}`} role={feedback.kind === 'error' ? 'alert' : 'status'}>{feedback.kind === 'success' ? <Check size={21} /> : <AlertTriangle size={21} />}<span>{feedback.message}</span></div>}
    </section>
  )
}
