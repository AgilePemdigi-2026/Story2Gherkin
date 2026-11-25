import { useEffect, useId, useLayoutEffect, useRef, useState, type KeyboardEvent } from 'react'
import { ChevronDown } from 'lucide-react'
import type { GoldenDatasetSummary } from '../services/goldenDatasetApi'

interface Props {
  id: string
  options: GoldenDatasetSummary[]
  value: string
  onChange: (value: string) => void
  disabled?: boolean
  emptyLabel?: string
  invalid?: boolean
  describedBy?: string
}

export default function EvaluationDatasetSelect({
  id, options, value, onChange, disabled = false,
  emptyLabel = 'No datasets available', invalid = false, describedBy,
}: Props) {
  const root = useRef<HTMLDivElement>(null)
  const list = useRef<HTMLDivElement>(null)
  const listId = useId()
  const [open, setOpen] = useState(false)
  const [highlighted, setHighlighted] = useState(0)
  const selectedIndex = options.findIndex(option => option.id === value)
  const selected = options[selectedIndex]
  const unavailable = disabled || options.length === 0
  const expanded = open && !unavailable
  const activeIndex = Math.max(0, Math.min(highlighted, options.length - 1))
  const typeahead = useRef({ text: '', time: 0 })

  useEffect(() => {
    if (!expanded) return
    const outside = (event: PointerEvent) => {
      if (root.current && event.target instanceof Node && !root.current.contains(event.target)) setOpen(false)
    }
    document.addEventListener('pointerdown', outside)
    return () => document.removeEventListener('pointerdown', outside)
  }, [expanded])

  useLayoutEffect(() => {
    if (!expanded) return
    const container = list.current
    const option = container?.querySelector<HTMLElement>('[data-highlighted="true"]')
    if (!container || !option) return
    if (option.offsetTop < container.scrollTop) container.scrollTop = option.offsetTop
    else if (option.offsetTop + option.offsetHeight > container.scrollTop + container.clientHeight) {
      container.scrollTop = option.offsetTop + option.offsetHeight - container.clientHeight
    }
  }, [expanded, activeIndex])

  function choose(index: number) {
    const option = options[index]
    if (!option) return
    onChange(option.id)
    setOpen(false)
  }

  function onKeyDown(event: KeyboardEvent<HTMLButtonElement>) {
    if (unavailable) return
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault()
      const direction = event.key === 'ArrowDown' ? 1 : -1
      setHighlighted(expanded
        ? (activeIndex + direction + options.length) % options.length
        : Math.max(selectedIndex, 0))
      setOpen(true)
    } else if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      if (expanded) choose(activeIndex)
      else { setHighlighted(Math.max(selectedIndex, 0)); setOpen(true) }
    } else if (event.key === 'Escape') {
      if (expanded) { event.preventDefault(); event.stopPropagation(); setOpen(false) }
    } else if (event.key === 'Tab') {
      setOpen(false)
    } else if (expanded && (event.key === 'Home' || event.key === 'End')) {
      event.preventDefault()
      setHighlighted(event.key === 'Home' ? 0 : options.length - 1)
    } else if (event.key.length === 1 && !event.ctrlKey && !event.altKey && !event.metaKey) {
      event.preventDefault()
      const now = Date.now()
      const text = now - typeahead.current.time < 700
        ? typeahead.current.text + event.key.toLowerCase()
        : event.key.toLowerCase()
      typeahead.current = { text, time: now }
      const match = options.findIndex(option => option.name.toLowerCase().startsWith(text))
      if (match >= 0) { setHighlighted(match); setOpen(true) }
    }
  }

  return (
    <div ref={root} className="eval-dataset-select" onBlur={event => {
      if (!event.currentTarget.contains(event.relatedTarget as Node | null)) setOpen(false)
    }}>
      <button
        id={id}
        type="button"
        role="combobox"
        className="eval-dataset-trigger"
        aria-haspopup="listbox"
        aria-expanded={expanded}
        aria-controls={expanded ? listId : undefined}
        aria-activedescendant={expanded ? `${listId}-option-${activeIndex}` : undefined}
        aria-invalid={invalid}
        aria-describedby={describedBy}
        disabled={unavailable}
        onKeyDown={onKeyDown}
        onClick={() => {
          if (!expanded) setHighlighted(Math.max(selectedIndex, 0))
          setOpen(current => !current)
        }}
      >
        <span className="eval-dataset-trigger-text">{selected?.name ?? (options.length === 0 ? emptyLabel : 'Select a Golden Dataset')}</span>
        {selected && <span className="eval-dataset-trigger-count">{selected.item_count} {selected.item_count === 1 ? 'case' : 'cases'}</span>}
        <ChevronDown size={17} aria-hidden="true" />
      </button>

      {expanded && (
        <div ref={list} id={listId} role="listbox" aria-label="Golden Datasets" className="eval-dataset-menu" onMouseLeave={() => setHighlighted(Math.max(selectedIndex, 0))}>
          {options.map((option, index) => (
            <div
              key={option.id}
              id={`${listId}-option-${index}`}
              role="option"
              aria-selected={option.id === value}
              data-highlighted={index === activeIndex}
              className={`eval-dataset-option${option.id === value ? ' is-selected' : ''}${index === activeIndex ? ' is-highlighted' : ''}`}
              onMouseEnter={() => setHighlighted(index)}
              onPointerDown={event => event.preventDefault()}
              onClick={() => choose(index)}
            >
              <span>{option.name}</span>
              <small>{option.item_count} {option.item_count === 1 ? 'case' : 'cases'}</small>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
