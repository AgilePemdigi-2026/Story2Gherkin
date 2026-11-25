import { useLayoutEffect, useRef } from 'react'
import { NavLink, useLocation } from 'react-router-dom'

import './AnimatedSANav.css'

const items = [
  { label: 'Generate Scenario', path: '/generate' },
  { label: 'Review Scenario', path: '/review' },
  { label: 'History', path: '/history' },
]

// Preserve the previous position when switching between page components.
let previousActiveIndex: number | null = null

export default function AnimatedSANav() {
  const location = useLocation()

  const activeIndex = items.findIndex(
    item =>
      location.pathname === item.path ||
      location.pathname.startsWith(`${item.path}/`),
  )

  const markerRef = useRef<HTMLSpanElement>(null)

  const startingIndex = useRef(
    previousActiveIndex ?? Math.max(activeIndex, 0),
  )

  useLayoutEffect(() => {
    const marker = markerRef.current

    if (!marker) return

    if (activeIndex < 0) {
      marker.style.opacity = '0'
      return
    }

    let firstFrame = 0
    let secondFrame = 0

    marker.style.opacity = '1'

    firstFrame = window.requestAnimationFrame(() => {
      secondFrame = window.requestAnimationFrame(() => {
        marker.style.transform =
          `translate3d(0, ${activeIndex * 54}px, 0)`
      })
    })

    previousActiveIndex = activeIndex

    return () => {
      window.cancelAnimationFrame(firstFrame)
      window.cancelAnimationFrame(secondFrame)
    }
  }, [activeIndex])

  return (
    <div className="sa-sidebar-section">
      <p className="sa-sidebar-heading">GENERATE</p>

      <div className="sa-sidebar-track">
        <span
          ref={markerRef}
          className="sa-sidebar-marker"
          aria-hidden="true"
          style={{
            transform:
              `translate3d(0, ${startingIndex.current * 54}px, 0)`,
            opacity: activeIndex < 0 ? 0 : 1,
          }}
        />

        {items.map(item => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) =>
              `sa-sidebar-link${isActive ? ' is-active' : ''}`
            }
          >
            <span>{item.label}</span>
          </NavLink>
        ))}
      </div>
    </div>
  )
}