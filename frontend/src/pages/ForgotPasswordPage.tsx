import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Mail } from 'lucide-react'
import AuthHeader from '../components/AuthHeader'
import AuthIllustration from '../components/AuthIllustration'
import { sendPasswordReset } from '../services/authApi'
import './RegisterPage.css'
import './AuthPages.css'
import './ForgotPasswordPage.css'

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [sent, setSent] = useState(false)
  const [error, setError] = useState('')
  const [emailInvalid, setEmailInvalid] = useState(false)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (loading) return
    const input = event.currentTarget.elements.namedItem('email') as HTMLInputElement
    const address = email.trim()
    const issue = !address
      ? 'Email address is required.'
      : input.validity.typeMismatch
        ? 'Enter a valid email address.'
        : ''

    if (issue) {
      setSent(false)
      setEmailInvalid(true)
      setError(issue)
      input.focus()
      return
    }

    setEmailInvalid(false)
    setError('')
    setLoading(true)
    try {
      await sendPasswordReset(address)
      setSent(true)
    } catch (err) {
      setSent(false)
      setError(
        err instanceof TypeError
          ? 'Unable to connect to the server. Please try again.'
          : err instanceof Error
            ? err.message
            : 'Unable to send the reset link. Please try again.',
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="register-page auth-redesign auth-forgot-page">
      <AuthHeader />
      <main className="register-layout">
        <aside className="register-left"><AuthIllustration /></aside>
        <section className="register-content">
          <div className="register-form-container">
            <h1>Forgot Password?</h1>
            <p className="register-description">
              Enter your email address to request a password reset link.
            </p>

            <form className="register-form forgot-password-form" onSubmit={handleSubmit} aria-busy={loading} noValidate>
              <div className="form-group">
                <label htmlFor="reset-email">Email Address</label>
                <div className={`password-field${error ? ' has-error' : sent ? ' has-success' : ''}`}>
                  <div className="input-wrapper">
                    <Mail size={19} className="input-icon" aria-hidden="true" />
                    <input
                      id="reset-email"
                      name="email"
                      type="email"
                      placeholder="example@gmail.com"
                      autoComplete="email"
                      value={email}
                      disabled={loading}
                      required
                      aria-invalid={emailInvalid}
                      aria-describedby={error || sent ? 'reset-email-feedback' : undefined}
                      onChange={event => {
                        setEmail(event.target.value)
                        setSent(false)
                        setError('')
                        setEmailInvalid(false)
                      }}
                    />
                  </div>
                  {error ? (
                    <p id="reset-email-feedback" className="password-error-message" role="alert">{error}</p>
                  ) : sent ? (
                    <p id="reset-email-feedback" className="forgot-password-success" role="status">
                      If this email address is registered, a password reset link has been sent. Check your inbox and spam folder.
                    </p>
                  ) : null}
                </div>
              </div>

              {sent ? (
                <div className="forgot-password-post-send">
                  <Link to="/" className="register-submit forgot-password-home">Back to Home</Link>
                  <p className="forgot-password-resend-note">
                    Didn’t receive the email?{' '}
                    <button type="submit" className="forgot-password-resend" disabled={loading}>
                      {loading ? 'Resending...' : 'Resend Reset Link'}
                    </button>
                  </p>
                </div>
              ) : (
                <button type="submit" className="register-submit" disabled={loading}>
                  {loading ? 'Sending...' : 'Send Reset Link'}
                </button>
              )}
            </form>

            <div className="register-divider"><span>or</span></div>
            <p className="register-login">
              Remember your password? <Link to="/login">Login</Link>
            </p>
          </div>
        </section>
      </main>
    </div>
  )
}
