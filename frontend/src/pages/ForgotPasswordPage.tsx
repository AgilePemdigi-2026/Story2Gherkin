import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Mail } from 'lucide-react'

import AuthHeader from '../components/AuthHeader'
import { sendPasswordReset } from '../services/authApi'
import illustration from '../assets/register-illustration.png'

import './RegisterPage.css'
import './PasswordPages.css'

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [sent, setSent] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (
    event: FormEvent<HTMLFormElement>
  ) => {
    event.preventDefault()

    if (loading) return

    setError('')
    setSent(false)
    setLoading(true)

    try {
      await sendPasswordReset(email.trim())
      setSent(true)
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'Unable to send the reset link. Please try again.'
      )
    } finally {
      setLoading(false)
    }
  }

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
            <h1>Reset Password</h1>

            <p className="register-description">
              Enter your email address and we’ll send you
              a password reset link.
            </p>

            <form
              className="password-form forgot-form"
              onSubmit={handleSubmit}
              aria-busy={loading}
            >
              <div className="form-group">
                <label htmlFor="reset-email">
                  Email Address
                </label>

                <div
                  className={`password-email-field${
                    sent ? ' has-success' : ''
                  }`}
                >
                  <div className="password-input-wrapper">
                    <Mail
                      size={22}
                      className="field-icon"
                      aria-hidden="true"
                    />

                    <input
                      id="reset-email"
                      type="email"
                      name="email"
                      placeholder="email@gmail.com"
                      autoComplete="email"
                      value={email}
                      disabled={loading}
                      aria-describedby={
                        sent
                          ? 'reset-email-success'
                          : undefined
                      }
                      onChange={(event) => {
                        setEmail(event.target.value)
                        setSent(false)
                        setError('')
                      }}
                      required
                    />
                  </div>

                  {sent && (
                    <p
                      id="reset-email-success"
                      className="password-email-success"
                      role="status"
                    >
                      Check your email! If the address is
                      registered, we’ve sent you a password
                      reset link. Open the link to choose
                      your new password.
                    </p>
                  )}
                </div>
              </div>

              {error && (
                <p
                  className="password-feedback error"
                  role="alert"
                >
                  {error}
                </p>
              )}

              <button
                type="submit"
                className="password-submit"
                disabled={loading}
              >
                {loading
                  ? 'Sending...'
                  : sent
                    ? 'Resend Reset Link'
                    : 'Send Reset Link'}
              </button>
            </form>

            <div className="password-divider">
              <span>or</span>
            </div>

            <p className="password-login">
              Remember your password?{' '}
              <Link to="/login">Login</Link>
            </p>
          </div>
        </section>
      </main>
    </div>
  )
}