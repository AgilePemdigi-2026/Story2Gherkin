import { useLayoutEffect, useRef, useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import './AnimatedQANav.css'

const items = [
  { label: 'Golden Dataset', path: '/evaluation/golden-dataset', end: false },
  { label: 'Run Evaluation', path: '/evaluation', end: true },
  { label: 'Evaluation Report', path: '/evaluation/reports', end: false },
]

let previousActiveIndex: number | null = null

export default function AnimatedQANav() {
  const location = useLocation()
  const activeIndex = items.findIndex(item =>
    item.end
      ? location.pathname === item.path || location.pathname === `${item.path}/`
      : location.pathname === item.path || location.pathname.startsWith(`${item.path}/`),
  )
  const marker = useRef<HTMLSpanElement>(null)
  const [startingIndex] = useState(() => previousActiveIndex ?? Math.max(activeIndex, 0))

  useLayoutEffect(() => {
    const element = marker.current
    if (!element) return
    if (activeIndex < 0) {
      element.style.opacity = '0'
      return
    }
    let secondFrame = 0
    element.style.opacity = '1'
    const firstFrame = window.requestAnimationFrame(() => {
      secondFrame = window.requestAnimationFrame(() => {
        element.style.transform = `translate3d(0, ${activeIndex * 54}px, 0)`
      })
    })
    previousActiveIndex = activeIndex
    return () => {
      window.cancelAnimationFrame(firstFrame)
      window.cancelAnimationFrame(secondFrame)
    }
  }, [activeIndex])

  return (
    <div className="qa-sidebar-section">
      <p className="qa-sidebar-heading">EVALUATION</p>
      <div className="qa-sidebar-track">
        <span ref={marker} className="qa-sidebar-marker" aria-hidden="true" style={{ transform: `translate3d(0, ${startingIndex * 54}px, 0)`, opacity: activeIndex < 0 ? 0 : 1 }} />
        {items.map(item => (
          <NavLink key={item.path} to={item.path} end={item.end} className={({ isActive }) => `qa-sidebar-link${isActive ? ' is-active' : ''}`}><span>{item.label}</span></NavLink>
        ))}
      </div>
    </div>
  )
}
