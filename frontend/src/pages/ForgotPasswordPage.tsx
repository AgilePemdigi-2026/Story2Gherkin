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
  const [emailInvalid, setEmailInvalid] = useState(false)

  const feedbackState = error
    ? 'has-error'
    : sent
      ? 'has-success'
      : ''

  const handleSubmit = async (
    event: FormEvent<HTMLFormElement>
  ) => {
    event.preventDefault()

    if (loading) return

    const form = event.currentTarget

    if (!form.checkValidity()) {
      setSent(false)
      setEmailInvalid(true)
      setError('Please enter a valid email address.')

      form
        .querySelector<HTMLInputElement>(
          'input[name="email"]'
        )
        ?.focus()

      return
    }

    setEmailInvalid(false)
    setError('')
    setLoading(true)

    try {
      await sendPasswordReset(email.trim())
      setSent(true)
    } catch (err) {
      setError(
        err instanceof TypeError
          ? 'Unable to connect to the server. Please try again.'
          : err instanceof Error
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
              noValidate
            >
              <div className="form-group">
                <label htmlFor="reset-email">
                  Email Address
                </label>

                <div
                  className={`password-field ${feedbackState}`}
                >
                  <div className="input-wrapper">
                    <Mail
                      size={22}
                      className="input-icon"
                      aria-hidden="true"
                    />

                    <input
                      id="reset-email"
                      name="email"
                      type="email"
                      placeholder="email@gmail.com"
                      autoComplete="email"
                      value={email}
                      disabled={loading}
                      aria-invalid={emailInvalid}
                      aria-describedby={
                        error || sent
                          ? 'reset-email-feedback'
                          : undefined
                      }
                      onChange={(event) => {
                        setEmail(event.target.value)
                        setSent(false)
                        setError('')
                        setEmailInvalid(false)
                      }}
                      required
                    />
                  </div>

                  {error ? (
                    <p
                      id="reset-email-feedback"
                      className="password-error-message"
                      role="alert"
                    >
                      {error}
                    </p>
                  ) : sent ? (
                    <p
                      id="reset-email-feedback"
                      className="password-success-message"
                      role="status"
                    >
                      Check your email! If the address is
                      registered, we’ve sent you a password
                      reset link.
                    </p>
                  ) : null}
                </div>
              </div>

              {sent ? (
                <div className="password-post-send">
                  <Link
                    to="/"
                    className="password-submit password-next-link"
                  >
                    Back to Home
                  </Link>

                  <p className="password-resend-note">
                    Didn’t receive the email?{' '}
                    <button
                      type="submit"
                      className="password-resend-link"
                      disabled={loading}
                    >
                      {loading
                        ? 'Resending...'
                        : 'Resend Reset Link'}
                    </button>
                  </p>
                </div>
              ) : (
                <button
                  type="submit"
                  className="password-submit"
                  disabled={loading}
                >
                  {loading
                    ? 'Sending...'
                    : 'Send Reset Link'}
                </button>
              )}
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