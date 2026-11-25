import { useLayoutEffect, useRef, useState } from 'react'
import {
  Link,
  NavLink,
  useLocation,
  useNavigate,
} from 'react-router-dom'
import {
  UserRound,
  ChevronDown,
  LogOut,
  Menu,
  X,
} from 'lucide-react'

import logo from '../assets/story2gherkin-logo.svg'
import { useAuthSession } from '../hooks/useAuthSession'
import { clearAuthSession } from '../services/authSession'

import './PublicHeader.css'
import '../styles/WorkspaceTheme.css'

interface HeaderPillPosition {
  x: number
  y: number
  width: number
  height: number
  viewportWidth: number
}

let previousHeaderPill: HeaderPillPosition | null = null

function PublicHeader() {
  const navigate = useNavigate()
  const location = useLocation()
  const session = useAuthSession()

  const headerNavRef = useRef<HTMLElement>(null)
  const headerPillRef = useRef<HTMLSpanElement>(null)

  const [accountOpen, setAccountOpen] = useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  const returnTo =
    location.pathname + location.search + location.hash

  const isSAWorkspace = ['/generate', '/review', '/history'].some(
    path =>
      location.pathname === path ||
      location.pathname.startsWith(`${path}/`),
  )

  const menuItems = [
    { label: 'Home', path: '/' },

    ...(session?.role === 'system_analyst'
      ? [{ label: 'Generate', path: '/generate' }]
      : []),

    ...(session?.role === 'qa'
      ? [{ label: 'Evaluation', path: '/evaluation' }]
      : []),

    { label: 'Documentation', path: '/documentation' },
  ]

  useLayoutEffect(() => {
    const nav = headerNavRef.current
    const pill = headerPillRef.current

    if (!nav || !pill) return

    let animation: Animation | null = null
    let disposed = false

    const updateIndicator = (animate: boolean) => {
      if (disposed) return

      const activeLink = nav.querySelector<HTMLAnchorElement>(
        '.public-nav-link.active',
      )

      if (
        !activeLink ||
        !window.matchMedia('(min-width: 1001px)').matches
      ) {
        animation?.cancel()
        pill.style.opacity = '0'
        return
      }

      const next: HeaderPillPosition = {
        x: activeLink.offsetLeft,
        y: activeLink.offsetTop,
        width: activeLink.offsetWidth,
        height: activeLink.offsetHeight,
        viewportWidth: window.innerWidth,
      }

      const previous = previousHeaderPill

      animation?.cancel()

      Object.assign(pill.style, {
        width: `${next.width}px`,
        height: `${next.height}px`,
        transform: `translate3d(${next.x}px, ${next.y}px, 0)`,
        opacity: '1',
      })

      const reducedMotion = window.matchMedia(
        '(prefers-reduced-motion: reduce)',
      ).matches

      const positionChanged =
        previous !== null &&
        (
          previous.x !== next.x ||
          previous.y !== next.y ||
          previous.width !== next.width ||
          previous.height !== next.height
        )

      if (
        animate &&
        previous &&
        positionChanged &&
        previous.viewportWidth === next.viewportWidth &&
        !reducedMotion &&
        typeof pill.animate === 'function'
      ) {
        animation = pill.animate(
          [
            {
              width: `${previous.width}px`,
              height: `${previous.height}px`,
              transform:
                `translate3d(${previous.x}px, ${previous.y}px, 0)`,
            },
            {
              width: `${next.width}px`,
              height: `${next.height}px`,
              transform:
                `translate3d(${next.x}px, ${next.y}px, 0)`,
            },
          ],
          {
            duration: 280,
            easing: 'cubic-bezier(0.22, 1, 0.36, 1)',
          },
        )
      }

      previousHeaderPill = next
    }

    updateIndicator(true)

    const handleResize = () => {
      updateIndicator(false)
    }

    window.addEventListener('resize', handleResize)

    // Keep the indicator aligned after the web font loads.
    void document.fonts.ready.then(() => {
      updateIndicator(false)
    })

    return () => {
      disposed = true
      animation?.cancel()
      window.removeEventListener('resize', handleResize)
    }
  }, [location.pathname, session?.role])

  const handleLogout = () => {
    const beforeLogout = new Event(
      'story2gherkin:before-logout',
      { cancelable: true },
    )

    if (!window.dispatchEvent(beforeLogout)) return

    setAccountOpen(false)
    setMobileMenuOpen(false)

    navigate('/', { flushSync: true })
    clearAuthSession()
  }

  return (
    <header className="public-header">
      <div className="public-header-container">
        <Link to="/" className="public-logo">
          <img src={logo} alt="Story2Gherkin" />
        </Link>

        <nav
          ref={headerNavRef}
          id="public-main-navigation"
          className={`public-nav${
            mobileMenuOpen ? ' mobile-open' : ''
          }`}
          aria-label="Main navigation"
        >
          <span
            ref={headerPillRef}
            className="public-nav-indicator"
            aria-hidden="true"
          />

          {menuItems.map(item => (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              onClick={() => {
                setMobileMenuOpen(false)
                setAccountOpen(false)
              }}
              className={({ isActive }) => {
                const active =
                  item.path === '/generate'
                    ? isSAWorkspace
                    : isActive

                return `public-nav-link${
                  active ? ' active' : ''
                }`
              }}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="public-header-actions">
          {session ? (
            <div className="public-account">
              <button
                type="button"
                className="public-account-button"
                onClick={() => {
                  setAccountOpen(current => !current)
                  setMobileMenuOpen(false)
                }}
                aria-expanded={accountOpen}
                aria-controls="public-account-dropdown"
                aria-label="Account menu"
              >
                <span className="public-avatar">
                  <UserRound size={18} aria-hidden="true" />
                </span>

                <span>
                  {session.role === 'system_analyst'
                    ? 'System Analyst'
                    : 'Quality Assurance'}
                </span>

                <ChevronDown size={16} aria-hidden="true" />
              </button>

              {accountOpen && (
                <div
                  id="public-account-dropdown"
                  className="public-account-dropdown"
                >
                  <button
                    type="button"
                    onClick={handleLogout}
                  >
                    <LogOut size={16} aria-hidden="true" />
                    Logout
                  </button>
                </div>
              )}
            </div>
          ) : (
            <div className="public-auth-actions">
              <Link
                to="/login"
                className="public-login-button"
                state={{ from: returnTo }}
              >
                Login
              </Link>

              <Link
                to="/register"
                className="public-signup-button"
              >
                Sign Up
              </Link>
            </div>
          )}

          <button
            type="button"
            className="public-mobile-toggle"
            onClick={() => {
              setMobileMenuOpen(current => !current)
              setAccountOpen(false)
            }}
            aria-label={
              mobileMenuOpen
                ? 'Close navigation'
                : 'Open navigation'
            }
            aria-expanded={mobileMenuOpen}
            aria-controls="public-main-navigation"
          >
            {mobileMenuOpen ? (
              <X size={24} aria-hidden="true" />
            ) : (
              <Menu size={24} aria-hidden="true" />
            )}
          </button>
        </div>
      </div>
    </header>
  )
}

export default PublicHeader