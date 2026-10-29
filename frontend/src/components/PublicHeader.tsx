
import { useState } from 'react'
import { Link, NavLink, useNavigate, useLocation } from 'react-router-dom'
import { UserRound, ChevronDown, LogOut, Menu, X } from 'lucide-react'

import logo from '../assets/story2gherkin-logo.svg'
import './PublicHeader.css'

function PublicHeader() {
  const navigate = useNavigate()
  const location = useLocation()

  const returnTo =
    location.pathname + location.search + location.hash
  const [isLoggedIn, setIsLoggedIn] = useState(() => {
    return Boolean(
      localStorage.getItem('access_token') ||
      sessionStorage.getItem('access_token')
    )
  })

  const [accountOpen, setAccountOpen] = useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  const handleLogout = () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    sessionStorage.removeItem('access_token')
    sessionStorage.removeItem('refresh_token')

    setIsLoggedIn(false)
    setAccountOpen(false)
    setMobileMenuOpen(false)

    navigate('/')
  }

  const menuItems = [
    { label: 'Home', path: '/' },
    { label: 'Generate', path: '/generate' },
    { label: 'Evaluation', path: '/evaluation' },
    { label: 'About Project', path: '/about' },
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
                `public-nav-link ${isActive ? 'active' : ''}`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        {/* RIGHT SECTION */}
        <div className="public-header-actions">
          {isLoggedIn ? (

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

                <span>My Account</span>

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
