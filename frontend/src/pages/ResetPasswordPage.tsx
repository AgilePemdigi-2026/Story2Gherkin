import { useState, type FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import {
  CheckCircle2,
  Eye,
  EyeOff,
  LockKeyhole,
} from 'lucide-react'

import AuthHeader from '../components/AuthHeader'
import { resetPassword } from '../services/authApi'
import illustration from '../assets/register-illustration.png'

import './RegisterPage.css'
import './PasswordPages.css'

function ResetPasswordForm({ token }: { token: string }) {
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] =
    useState(false)

  const [loading, setLoading] = useState(false)
  const [completed, setCompleted] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (
    event: FormEvent<HTMLFormElement>
  ) => {
    event.preventDefault()

    if (loading || completed) return

    setError('')

    if (/[\p{White_Space}\uFEFF]/u.test(password)) {
      setError(
        'Password must not contain spaces, tabs, or line breaks.'
      )
      return
    }

    if (
      Array.from(password).length < 8 ||
      !/[A-Z]/.test(password) ||
      !/[a-z]/.test(password)
    ) {
      setError(
        'Password must contain at least 8 characters, including uppercase and lowercase letters.'
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

      setPassword('')
      setConfirmPassword('')
      setCompleted(true)
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'Unable to reset your password. Please try again.'
      )
    } finally {
      setLoading(false)
    }
  }

  if (completed) {
    return (
      <>
        <h1>Password Updated</h1>

        <p className="register-description">
          You can now sign in with your new password.
        </p>

        <div
          className="password-success-panel"
          role="status"
        >
          <CheckCircle2 size={24} aria-hidden="true" />

          <p>
            Your password has been reset successfully.
          </p>
        </div>

        <Link
          to="/login"
          replace
          className="password-submit password-next-link"
        >
          Back to Login
        </Link>
      </>
    )
  }

  return (
    <>
      <h1>Reset Password</h1>

      <p className="register-description">
        Enter and confirm your new password.
      </p>

      <form
        className="password-form reset-form"
        onSubmit={handleSubmit}
        aria-busy={loading}
      >
        <div className="form-group">
          <label htmlFor="new-password">
            New Password
          </label>

          <div className="password-input-wrapper">
            <LockKeyhole
              size={22}
              className="field-icon"
              aria-hidden="true"
            />

            <input
              id="new-password"
              name="new_password"
              type={showPassword ? 'text' : 'password'}
              placeholder="Enter your new password"
              autoComplete="new-password"
              value={password}
              disabled={loading}
              aria-describedby="reset-password-guidance"
              onChange={(event) => {
                setPassword(event.target.value)
                setError('')
              }}
              required
            />

            <button
              type="button"
              className="password-eye-button"
              disabled={loading}
              onClick={() => setShowPassword(!showPassword)}
              aria-label={
                showPassword
                  ? 'Hide password'
                  : 'Show password'
              }
              aria-controls="new-password"
            >
              {showPassword ? (
                <EyeOff size={22} aria-hidden="true" />
              ) : (
                <Eye size={22} aria-hidden="true" />
              )}
            </button>
          </div>

          <p
            id="reset-password-guidance"
            className="password-requirements"
          >
            Use at least 8 characters, including uppercase and
            lowercase letters. Maximum 72 bytes. Spaces, tabs,
            and line breaks are not allowed.
          </p>
        </div>

        <div className="form-group">
          <label htmlFor="confirm-new-password">
            Confirm Password
          </label>

          <div className="password-input-wrapper">
            <LockKeyhole
              size={22}
              className="field-icon"
              aria-hidden="true"
            />

            <input
              id="confirm-new-password"
              name="confirm_password"
              type={
                showConfirmPassword ? 'text' : 'password'
              }
              placeholder="Confirm your new password"
              autoComplete="new-password"
              value={confirmPassword}
              disabled={loading}
              onChange={(event) => {
                setConfirmPassword(event.target.value)
                setError('')
              }}
              required
            />

            <button
              type="button"
              className="password-eye-button"
              disabled={loading}
              onClick={() =>
                setShowConfirmPassword(!showConfirmPassword)
              }
              aria-label={
                showConfirmPassword
                  ? 'Hide confirmation password'
                  : 'Show confirmation password'
              }
              aria-controls="confirm-new-password"
            >
              {showConfirmPassword ? (
                <EyeOff size={22} aria-hidden="true" />
              ) : (
                <Eye size={22} aria-hidden="true" />
              )}
            </button>
          </div>
        </div>

        {error && (
          <div className="password-reset-error">
            <p
              className="password-feedback error"
              role="alert"
            >
              {error}
            </p>

            <p className="password-help">
              Need a new link?{' '}
              <Link to="/forgot-password">
                Request password reset
              </Link>
            </p>
          </div>
        )}

        <button
          type="submit"
          className="password-submit"
          disabled={loading}
        >
          {loading ? 'Resetting...' : 'Reset Password'}
        </button>
      </form>
    </>
  )
}

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')?.trim() || ''

  return (
    <div className="register-page password-page">
      <AuthHeader />

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
            {token ? (
              <ResetPasswordForm key={token} token={token} />
            ) : (
              <>
                <h1>Reset Password</h1>

                <p className="register-description">
                  Open the password reset link from your
                  email to choose a new password.
                </p>

                <div className="password-link-notice">
                  <p className="password-feedback error">
                    This page requires a reset link.
                    Request a new one to continue.
                  </p>

                  <Link
                    to="/forgot-password"
                    className="password-submit password-next-link"
                  >
                    Request Reset Link
                  </Link>
                </div>
              </>
            )}
          </div>
        </section>
      </main>
    </div>
  )
}