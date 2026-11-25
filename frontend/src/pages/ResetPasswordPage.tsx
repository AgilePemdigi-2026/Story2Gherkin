import { useState, type FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Eye, EyeOff, LockKeyhole } from 'lucide-react'
import AuthHeader from '../components/AuthHeader'
import AuthIllustration from '../components/AuthIllustration'
import { resetPassword } from '../services/authApi'
import './RegisterPage.css'
import './AuthPages.css'
import './ResetPasswordPage.css'

function validatePassword(value: string): string {
  if (!value) return 'New password is required.'
  if (/[\p{White_Space}\uFEFF]/u.test(value)) return 'Password must not contain spaces, tabs, or line breaks.'
  if (Array.from(value).length < 8 || !/[A-Z]/.test(value) || !/[a-z]/.test(value)) {
    return 'Password must contain at least 8 characters, including uppercase and lowercase letters.'
  }
  if (new TextEncoder().encode(value).length > 72) return 'Password must not exceed 72 bytes.'
  return ''
}

function ResetPasswordForm({ token }: { token: string }) {
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [passwordError, setPasswordError] = useState('')
  const [confirmError, setConfirmError] = useState('')
  const [requestError, setRequestError] = useState('')
  const [loading, setLoading] = useState(false)
  const [completed, setCompleted] = useState(false)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (loading || completed) return
    setRequestError('')
    const passwordIssue = validatePassword(password)
    const confirmIssue = !confirmPassword
      ? 'Please confirm your new password.'
      : password !== confirmPassword ? 'Passwords do not match.' : ''
    setPasswordError(passwordIssue)
    setConfirmError(confirmIssue)
    if (passwordIssue || confirmIssue) {
      event.currentTarget.querySelector<HTMLInputElement>(passwordIssue ? '#new-password' : '#confirm-new-password')?.focus()
      return
    }

    setLoading(true)
    try {
      await resetPassword(token, password)
      setPassword('')
      setConfirmPassword('')
      setCompleted(true)
    } catch (err) {
      setRequestError(
        err instanceof TypeError
          ? 'Unable to connect to the server. Please try again.'
          : err instanceof Error ? err.message : 'Unable to reset your password. Please try again.',
      )
    } finally {
      setLoading(false)
    }
  }

  if (completed) {
    return (
      <>
        <h1>Password Updated</h1>
        <p className="register-description">You can now log in with your new password.</p>
        <div className="reset-password-success" role="status">Your password has been reset successfully.</div>
        <Link to="/login" replace className="register-submit reset-password-link">Back to Login</Link>
      </>
    )
  }

  return (
    <>
      <h1>Reset Password</h1>
      <p className="register-description">Choose a new password to secure your account.</p>
      <form className="register-form reset-password-form" onSubmit={handleSubmit} aria-busy={loading} noValidate>
        <div className="form-group">
          <label htmlFor="new-password">New Password</label>
          <div className={`password-field${passwordError ? ' has-error' : ''}`}>
            <div className="input-wrapper">
              <LockKeyhole size={19} className="input-icon" aria-hidden="true" />
              <input
                id="new-password" name="new_password" type={showPassword ? 'text' : 'password'}
                placeholder="Enter your new password" autoComplete="new-password"
                value={password} disabled={loading} required
                aria-invalid={Boolean(passwordError)}
                aria-describedby={passwordError ? 'reset-password-guidance new-password-error' : 'reset-password-guidance'}
                onChange={event => {
                  const value = event.target.value
                  setPassword(value)
                  setRequestError('')
                  if (passwordError) setPasswordError(validatePassword(value))
                  if (confirmError) setConfirmError(!confirmPassword ? 'Please confirm your new password.' : value !== confirmPassword ? 'Passwords do not match.' : '')
                }}
              />
              <button type="button" className="toggle-password" disabled={loading} onClick={() => setShowPassword(value => !value)} aria-label={showPassword ? 'Hide password' : 'Show password'} aria-controls="new-password">
                {showPassword ? <EyeOff size={19} aria-hidden="true" /> : <Eye size={19} aria-hidden="true" />}
              </button>
            </div>
            {passwordError && <p id="new-password-error" className="password-error-message" role="alert">{passwordError}</p>}
          </div>
          <p className="reset-password-guidance" id="reset-password-guidance">Use at least 8 characters with uppercase and lowercase letters. Maximum 72 bytes. No spaces, tabs, or line breaks.</p>
        </div>

        <div className="form-group">
          <label htmlFor="confirm-new-password">Confirm Password</label>
          <div className={`password-field${confirmError ? ' has-error' : ''}`}>
            <div className="input-wrapper">
              <LockKeyhole size={19} className="input-icon" aria-hidden="true" />
              <input
                id="confirm-new-password" name="confirm_password" type={showConfirmPassword ? 'text' : 'password'}
                placeholder="Confirm your new password" autoComplete="new-password"
                value={confirmPassword} disabled={loading} required
                aria-invalid={Boolean(confirmError)} aria-describedby={confirmError ? 'confirm-new-password-error' : undefined}
                onChange={event => {
                  const value = event.target.value
                  setConfirmPassword(value)
                  setRequestError('')
                  if (confirmError) setConfirmError(!value ? 'Please confirm your new password.' : value !== password ? 'Passwords do not match.' : '')
                }}
              />
              <button type="button" className="toggle-password" disabled={loading} onClick={() => setShowConfirmPassword(value => !value)} aria-label={showConfirmPassword ? 'Hide confirmation password' : 'Show confirmation password'} aria-controls="confirm-new-password">
                {showConfirmPassword ? <EyeOff size={19} aria-hidden="true" /> : <Eye size={19} aria-hidden="true" />}
              </button>
            </div>
            {confirmError && <p id="confirm-new-password-error" className="password-error-message" role="alert">{confirmError}</p>}
          </div>
        </div>

        {requestError && (
          <div className="reset-password-request-error">
            <p className="register-message error" role="alert">{requestError}</p>
            <p className="reset-password-help">Need a new link? <Link to="/forgot-password">Request password reset</Link></p>
          </div>
        )}
        <button type="submit" className="register-submit" disabled={loading}>{loading ? 'Resetting...' : 'Reset Password'}</button>
      </form>
      <div className="register-divider"><span>or</span></div>
      <p className="register-login">Remember your password? <Link to="/login">Login</Link></p>
    </>
  )
}

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')?.trim() || ''
  return (
    <div className="register-page auth-redesign auth-reset-page">
      <AuthHeader />
      <main className="register-layout">
        <aside className="register-left"><AuthIllustration /></aside>
        <section className="register-content">
          <div className="register-form-container">
            {token ? <ResetPasswordForm key={token} token={token} /> : (
              <>
                <h1>Reset Password</h1>
                <p className="register-description">Open the reset link from your email to choose a new password.</p>
                <div className="reset-password-missing">
                  <p className="register-message error" role="alert">A password reset link is required. Request a new link to continue.</p>
                  <Link to="/forgot-password" className="register-submit reset-password-link">Request Reset Link</Link>
                </div>
                <p className="register-login reset-password-login"><Link to="/login">Back to Login</Link></p>
              </>
            )}
          </div>
        </section>
      </main>
    </div>
  )
}
