import type { ScenarioSummary } from '../../services/reviewApi'

interface Props {
  scenarios: ScenarioSummary[]
  selected: string
  onSelect: (id: string) => void
}

export default function ScenarioSelector({ scenarios, selected, onSelect }: Props) {
  return (
    <div className="review-selector">
      <label htmlFor="review-scenario-select">Pilih Skenario</label>
      <select id="review-scenario-select" value={selected} onChange={(event) => onSelect(event.target.value)}>
        {!scenarios.some((item) => item.id === selected) && <option value={selected}>Skenario yang dipilih</option>}
        {scenarios.map((item) => <option key={item.id} value={item.id}>{item.id} — {item.feature_name}</option>)}
      </select>
    </div>
  )
}
