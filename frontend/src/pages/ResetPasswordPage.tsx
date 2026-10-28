
import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { LockKeyhole, Eye, EyeOff, KeyRound } from 'lucide-react'
import { resetPassword } from '../services/authApi'

import logo from '../assets/story2gherkin-logo.svg'
import illustration from '../assets/register-illustration.png'

import './RegisterPage.css'
import './PasswordPages.css'

function ResetPasswordPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  const urlToken = searchParams.get('token')?.trim() || ''

  const [manualToken, setManualToken] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (
    event: FormEvent<HTMLFormElement>
  ) => {
    event.preventDefault()
    if (loading) return

    setError('')

    const token = urlToken || manualToken.trim()

    if (!token) {
      setError('Please enter your password reset token.')
      return
    }

    if (
      password.length < 8 ||
      !/[A-Z]/.test(password) ||
      !/[a-z]/.test(password)
    ) {
      setError(
        'Password must have at least 8 characters, including uppercase and lowercase letters.'
      )
      return
    }

    if (new TextEncoder().encode(password).length > 72) {
      setError('Password must not exceed 72 bytes.')
      return
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match.')
      return
    }

    setLoading(true)

    try {
      await resetPassword(token, password)
      navigate('/login', { replace: true })
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'Failed to reset password.'
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="register-page password-page">
      <header className="register-header">
        <img
          src={logo}
          alt="Story2Gherkin"
          className="register-logo"
        />
      </header>

      <main className="register-layout">
        <aside className="register-left">
          <img
            src={illustration}
            alt="Story2Gherkin illustration"
            className="register-illustration"
          />
        </aside>

        <section className="register-content">
          <div className="register-form-container">
            <h1>Reset Password</h1>

            <p className="register-description">
              Enter your new password.
            </p>

            <form
              className="password-form reset-form"
              onSubmit={handleSubmit}
            >
              {/* Shown only when URL has no token */}
              {!urlToken && (
                <div className="form-group">
                  <label htmlFor="reset-token">
                    Reset Token
                  </label>

                  <div className="password-input-wrapper">
                    <KeyRound size={22} className="field-icon" />
                    <input
                      id="reset-token"
                      type="text"
                      placeholder="Enter the token from your email"
                      value={manualToken}
                      onChange={(event) =>
                        setManualToken(event.target.value)
                      }
                      autoComplete="off"
                      required
                    />
                  </div>
                </div>
              )}

              {/* PASSWORD */}
              <div className="form-group">
                <label htmlFor="new-password">
                  Password
                </label>

                <div className="password-input-wrapper">
                  <LockKeyhole size={22} className="field-icon" />

                  <input
                    id="new-password"
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Enter your new password"
                    autoComplete="new-password"
                    value={password}
                    onChange={(event) =>
                      setPassword(event.target.value)
                    }
                    required
                  />

                  <button
                    type="button"
                    className="password-eye-button"
                    onClick={() =>
                      setShowPassword(!showPassword)
                    }
                    aria-label={
                      showPassword
                        ? 'Hide password'
                        : 'Show password'
                    }
                  >
                    {showPassword
                      ? <Eye size={22} />
                      : <EyeOff size={22} />}
                  </button>
                </div>
              </div>

              {/* CONFIRM PASSWORD */}
              <div className="form-group">
                <label htmlFor="confirm-new-password">
                  Confirm Password
                </label>

                <div className="password-input-wrapper">
                  <LockKeyhole size={22} className="field-icon" />

                  <input
                    id="confirm-new-password"
                    type={showConfirmPassword ? 'text' : 'password'}
                    placeholder="Confirm your new password"
                    autoComplete="new-password"
                    value={confirmPassword}
                    onChange={(event) =>
                      setConfirmPassword(event.target.value)
                    }
                    required
                  />

                  <button
                    type="button"
                    className="password-eye-button"
                    onClick={() =>
                      setShowConfirmPassword(!showConfirmPassword)
                    }
                    aria-label={
                      showConfirmPassword
                        ? 'Hide password'
                        : 'Show password'
                    }
                  >
                    {showConfirmPassword
                      ? <Eye size={22} />
                      : <EyeOff size={22} />}
                  </button>
                </div>
              </div>

              {error && (
                <p className="password-feedback error" role="alert">
                  {error}
                </p>
              )}

              <button
                type="submit"
                className="password-submit"
                disabled={loading}
              >
                {loading ? 'Resetting...' : 'Reset Password'}
              </button>
            </form>

            {!urlToken && (
              <p className="password-help">
                Don't have a token?{' '}
                <Link to="/forgot-password">
                  Request password reset
                </Link>
              </p>
            )}
          </div>
        </section>
      </main>
    </div>
  )
}

export default ResetPasswordPage
