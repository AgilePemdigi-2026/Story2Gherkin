import type { ScenarioSummary } from '../../services/reviewApi'

interface Props {
  scenarios: ScenarioSummary[]
  selected: string
  onSelect: (id: string) => void
}

export default function ScenarioSelector({ scenarios, selected, onSelect }: Props) {
  return (
    <div className="review-selector">
      <label htmlFor="review-scenario-select">Select Scenario</label>
      <select id="review-scenario-select" value={selected} onChange={(event) => onSelect(event.target.value)}>
        {!scenarios.some((item) => item.id === selected) && (
  <option value={selected}>Selected scenario</option>
)}
        {scenarios.map((item) => <option key={item.id} value={item.id}>{item.scenario_code ? `${item.scenario_code} — ${item.feature_name}` : item.feature_name}</option>)}
      </select>
    </div>
  )
}
