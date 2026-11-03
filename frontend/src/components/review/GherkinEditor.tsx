interface Props {
  value: string
  editing: boolean
  disabled: boolean
  onChange: (value: string) => void
}

function HighlightedGherkin({ value }: { value: string }) {
  const lines = value.split(/\r\n|\r|\n/)
  return lines.map((line, index) => {
    const match = line.match(/^(\s*)(Feature:|Scenario:|Given\b|When\b|Then\b|And\b)(.*)$/)
    const newline = index < lines.length - 1 ? '\n' : ''
    if (!match) return <span key={index}>{line}{newline}</span>
    const [, spaces, keyword, rest] = match
    const kind = keyword === 'Feature:' ? 'feature' : keyword === 'Scenario:' ? 'scenario' : 'step'
    return <span className={`review-code-${kind}`} key={index}>{spaces}{keyword}{rest}{newline}</span>
  })
}

export default function GherkinEditor({ value, editing, disabled, onChange }: Props) {
  return (
    <div className="review-code-editor">
      {/* Native text and caret share one surface while editing, including when scrolled. */}
      {editing ? (
        <textarea
          id="review-gherkin"
          aria-label="Generated Gherkin"
          value={value}
          disabled={disabled}
          onChange={(event) => onChange(event.target.value)}
          spellCheck={false}
          autoCapitalize="off"
          autoCorrect="off"
          wrap="off"
        />
      ) : (
        <pre><code><HighlightedGherkin value={value} />{'\n'}</code></pre>
      )}
    </div>
  )
}
