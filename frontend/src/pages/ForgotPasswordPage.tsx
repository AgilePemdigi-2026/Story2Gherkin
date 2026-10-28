
import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Mail } from 'lucide-react'
import { sendPasswordReset } from '../services/authApi'

import logo from '../assets/story2gherkin-logo.svg'
import illustration from '../assets/register-illustration.png'

import './RegisterPage.css'
import './PasswordPages.css'

    function ForgotPasswordPage() {
    const [email, setEmail] = useState('')
    const [loading, setLoading] = useState(false)
    const [sent, setSent] = useState(false)
    const [error, setError] = useState('')

    
    const requestReset = async () => {
    if (loading) return

    setError('')
    setLoading(true)

    try {
        await sendPasswordReset(email.trim())
        setSent(true)
    } catch (err) {
        setError(
        err instanceof Error
            ? err.message
            : 'Failed to send password reset email.'
        )
    } finally {
        setLoading(false)
    }
    }

    const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    void requestReset()
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
              Enter your email and we’ll send you a reset
              link to your account.
            </p>

            <form
              className="password-form forgot-form"
              onSubmit={handleSubmit}
            >
              <div className="form-group">
                <label htmlFor="reset-email">
                  Email Address
                </label>

                <div className="password-input-wrapper">
                  <Mail size={22} className="field-icon" />

                  <input
                    id="reset-email"
                    type="email"
                    name="email"
                    placeholder="example@gmail.com"
                    autoComplete="email"
                    value={email}
                    onChange={(event) => {
                      setEmail(event.target.value)
                      setSent(false)
                    }}
                    required
                  />
                </div>
              </div>


                {/* ERROR MESSAGE */}
                {error && (
                <p className="password-feedback error" role="alert">
                    {error}
                </p>
                )}

                {/* SUCCESS MESSAGE */}
                {sent && (
                <p className="password-feedback success" role="status">
                    Check your email! If the address is registered,
                    we've sent you a password reset token.
                </p>
                )}

                {/* MAIN ACTION */}
                {!sent ? (
                <button
                    type="submit"
                    className="password-submit"
                    disabled={loading}
                >
                    {loading ? 'Sending...' : 'Send Reset Link'}
                </button>
                ) : (
                <>
                    {/* NEXT PAGE BUTTON */}
                    <Link
                    to="/reset-password"
                    className="password-submit password-next-link"
                    >
                    Continue to Reset Password
                    </Link>

                    {/* RESEND TOKEN LINK */}
                    <p className="password-resend">
                    Didn't receive the email?{' '}
                    <button
                        type="button"
                        className="password-resend-link"
                        onClick={() => void requestReset()}
                        disabled={loading}
                    >
                        {loading ? 'Resending...' : 'Resend token'}
                    </button>
                    </p>
                </>
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

export default ForgotPasswordPage
