import './CircularPercentage.css'

type Props = {
  value: number | null | undefined
  label: string
  variant?: 'llm' | 'rule' | 'default'
  size?: 'small' | 'medium' | 'large'
  kind?: 'meter' | 'progressbar'
  locale?: string
}

export default function CircularPercentage({ value, label, variant = 'default', size = 'medium', kind = 'meter', locale = 'en-GB' }: Props) {
  const available = typeof value === 'number' && Number.isFinite(value)
  // The ring has a fixed visual range; the text keeps the supplied percentage.
  const bounded = available ? Math.min(100, Math.max(0, value)) : 0
  const display = available
    ? `${new Intl.NumberFormat(locale, { maximumFractionDigits: 2 }).format(value)}%`
    : '—'

  return (
    <div
      className={`circular-percentage circular-percentage--${variant} circular-percentage--${size}${available ? '' : ' is-unavailable'}`}
      role={available ? kind : 'img'}
      aria-label={available ? label : `${label}: not available`}
      aria-valuemin={available ? 0 : undefined}
      aria-valuemax={available ? 100 : undefined}
      aria-valuenow={available ? bounded : undefined}
      aria-valuetext={available ? display : undefined}
    >
      <svg viewBox="0 0 100 100" aria-hidden="true" focusable="false">
        <circle className="circular-percentage-track" cx="50" cy="50" r="43" />
        <circle
          className="circular-percentage-fill"
          cx="50" cy="50" r="43"
          pathLength="100"
          strokeDasharray="100 100"
          strokeDashoffset={100 - bounded}
          style={{ opacity: bounded > 0 ? 1 : 0 }}
        />
      </svg>
      <span className="circular-percentage-value" aria-hidden="true">{display}</span>
    </div>
  )
}
