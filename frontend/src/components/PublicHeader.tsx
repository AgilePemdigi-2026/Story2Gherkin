
import { useState } from 'react'
import { Link, NavLink, useNavigate, useLocation } from 'react-router-dom'
import { UserRound, ChevronDown, LogOut, Menu, X } from 'lucide-react'

import logo from '../assets/story2gherkin-logo.svg'
import './PublicHeader.css'
import { useAuthSession } from '../hooks/useAuthSession'
import { clearAuthSession } from '../services/authSession'

function PublicHeader() {
  const navigate = useNavigate()
  const location = useLocation()

  const returnTo =
    location.pathname + location.search + location.hash
  const session = useAuthSession()

  const [accountOpen, setAccountOpen] = useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  const handleLogout = () => {
    // Review pages may cancel before authentication data is removed.
    const beforeLogout = new Event('story2gherkin:before-logout', { cancelable: true })
    if (!window.dispatchEvent(beforeLogout)) return
    setAccountOpen(false)
    setMobileMenuOpen(false)

    // Leave the protected route before notifying its guard that the session ended.
    navigate('/', { flushSync: true })
    clearAuthSession()
  }

  const menuItems = [
    { label: 'Home', path: '/' },
    ...(session?.role === 'system_analyst' ? [{ label: 'Generate', path: '/generate' }] : []),
    ...(session?.role === 'qa' ? [{ label: 'Evaluation', path: '/evaluation' }] : []),
    { label: 'About', path: '/about' },
    { label: 'Documentation', path: '/documentation' },
  ]

  return (
    <header className="public-header">
      <div className="public-header-container">

        {/* LOGO */}
        <Link to="/" className="public-logo">
          <img src={logo} alt="Story2Gherkin" />
        </Link>

        {/* NAVIGATION */}
        <nav
          className={`public-nav ${mobileMenuOpen ? 'mobile-open' : ''}`}
          aria-label="Main navigation"
        >
          {menuItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              onClick={() => setMobileMenuOpen(false)}
              className={({ isActive }) =>
                `public-nav-link ${isActive || (item.path === '/generate' && location.pathname === '/review') ? 'active' : ''}`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        {/* RIGHT SECTION */}
        <div className="public-header-actions">
          {session ? (

            /* LOGGED IN */
            <div className="public-account">
              <button
                type="button"
                className="public-account-button"
                onClick={() => setAccountOpen(!accountOpen)}
                aria-expanded={accountOpen}
                aria-label="Account menu"
              >
                <span className="public-avatar">
                  <UserRound size={18} />
                </span>

                <span>{session.role === 'system_analyst' ? 'System Analyst' : 'Quality Assurance'}</span>

                <ChevronDown size={16} />
              </button>

              {accountOpen && (
                <div className="public-account-dropdown">
                  <button
                    type="button"
                    onClick={handleLogout}
                  >
                    <LogOut size={16} />
                    Logout
                  </button>
                </div>
              )}
            </div>

          ) : (

            /* NOT LOGGED IN */
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

          {/* MOBILE MENU BUTTON */}
          <button
            type="button"
            className="public-mobile-toggle"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            aria-label="Toggle navigation"
            aria-expanded={mobileMenuOpen}
          >
            {mobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
        </div>

      </div>
    </header>
  )
}

export default PublicHeader
