import { useEffect, useRef, useState } from 'react'
import { legalDocument, policyDetails } from '../content/legalDocuments'
import './LegalDocumentDialog.css'

type Props = { onClose: () => void }

export default function LegalDocumentDialog({ onClose }: Props) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const bodyRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [closing, setClosing] = useState(false)
  const [canClose, setCanClose] = useState(false)
  const document = legalDocument

  useEffect(() => {
    const dialog = dialogRef.current
    const body = bodyRef.current
    const content = contentRef.current
    const previousFocus = window.document.activeElement
    const previousOverflow = window.document.body.style.overflow
    window.document.body.style.overflow = 'hidden'
    if (dialog && !dialog.open) dialog.showModal()
    if (body) body.scrollTop = 0
    body?.focus()

    // A short document that already fits is also fully visible.
    const checkEnd = () => {
      if (body && body.clientHeight > 0 && body.scrollHeight - body.clientHeight - body.scrollTop <= 2) setCanClose(true)
    }
    const frame = window.requestAnimationFrame(checkEnd)
    const observer = typeof ResizeObserver !== 'undefined' ? new ResizeObserver(checkEnd) : null
    if (body) observer?.observe(body)
    if (content) observer?.observe(content)
    window.addEventListener('resize', checkEnd)

    return () => {
      window.cancelAnimationFrame(frame)
      observer?.disconnect()
      window.removeEventListener('resize', checkEnd)
      if (timerRef.current !== null) clearTimeout(timerRef.current)
      if (dialog?.open) dialog.close()
      window.document.body.style.overflow = previousOverflow
      if (previousFocus instanceof HTMLElement && previousFocus.isConnected) previousFocus.focus()
    }
  }, [])

  const requestClose = () => {
    if (!canClose || timerRef.current !== null) return
    setClosing(true)
    const duration = window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 130
    timerRef.current = setTimeout(onClose, duration)
  }

  return (
    <dialog
      ref={dialogRef}
      className={`legal-document-dialog${closing ? ' is-closing' : ''}`}
      aria-labelledby="legal-document-title"
      aria-describedby="legal-document-date legal-document-scroll-hint"
      onCancel={event => { event.preventDefault(); requestClose() }}
      onClick={event => {
        if (event.target !== event.currentTarget) return
        const rect = event.currentTarget.getBoundingClientRect()
        if (event.clientX < rect.left || event.clientX > rect.right || event.clientY < rect.top || event.clientY > rect.bottom) requestClose()
      }}
    >
      <header className="legal-document-header">
        <div>
          <h2 id="legal-document-title">{document.title}</h2>
          <p id="legal-document-date">Last updated: {policyDetails.effectiveDate}</p>
        </div>
      </header>
      <div
        ref={bodyRef}
        className="legal-document-body"
        tabIndex={0}
        role="region"
        aria-label={`${document.title} content`}
        onScroll={event => {
          const body = event.currentTarget
          if (body.scrollHeight - body.clientHeight - body.scrollTop <= 2) setCanClose(true)
        }}
      >
        <div ref={contentRef}>
          {document.sections.map(section => (
            <section key={section.title}>
              <h3>{section.title}</h3>
              <ol>{section.items.map(item => <li key={item}>{item}</li>)}</ol>
            </section>
          ))}
        </div>
      </div>
      <footer className="legal-document-footer">
        <p id="legal-document-scroll-hint" role="status">
          {canClose ? 'You can now close this document.' : 'Scroll to the bottom to enable Close.'}
        </p>
        <button type="button" disabled={!canClose || closing} onClick={requestClose} aria-describedby="legal-document-scroll-hint">
          Close
        </button>
      </footer>
    </dialog>
  )
}
