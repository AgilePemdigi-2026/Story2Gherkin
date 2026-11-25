
import { useState, type FormEvent } from 'react'
import { Mail, LockKeyhole, Eye, EyeOff } from 'lucide-react'
import { loginUser } from '../services/authApi'
import { loginDestination, saveAuthSession } from '../services/authSession'
import AuthHeader from '../components/AuthHeader'
import AuthIllustration from '../components/AuthIllustration'
import { useLocation, useNavigate } from 'react-router-dom'
import './RegisterPage.css'
import './LoginPage.css'
import './AuthPages.css'
import './AuthValidation.css'

function LoginPage() {
    const [password, setPassword] = useState('')
    const [showPassword, setShowPassword] = useState(false)
    const [rememberMe, setRememberMe] = useState(false)
    
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [loginError, setLoginError] = useState('')
    const [emailError, setEmailError] = useState('')

    const navigate = useNavigate()
    const location = useLocation()

    // Halaman tempat pengguna menekan tombol Login
    const from = (location.state as { from?: string } | null)?.from

    const handleLogin = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    if (isSubmitting) return

    const formData = new FormData(event.currentTarget)
    const email = String(formData.get('email') ?? '').trim()

    const emailInput = event.currentTarget.elements.namedItem('email') as HTMLInputElement
    const emailIssue = !email ? 'Email address is required.' : emailInput.validity.typeMismatch ? 'Enter a valid email address.' : ''
    const passwordIssue = !password ? 'Password is required.' : ''
    setEmailError(emailIssue)
    setLoginError(passwordIssue)
    if (emailIssue || passwordIssue) {
      event.currentTarget.querySelector<HTMLElement>(emailIssue ? '#loginEmail' : '#loginPassword')?.focus()
      return
    }
    setIsSubmitting(true)

    try {
        const result = await loginUser({
        email,
        password,
        })

        saveAuthSession(result, rememberMe)
        navigate(loginDestination(from, result.role), { replace: true })

    } catch (error) {
        // Kosongkan password, tetapi pertahankan email
        setPassword('')
        setShowPassword(false)

        setLoginError(
        error instanceof Error
            ? error.message
            : 'Email or password is incorrect.'
        )

    } finally {
        setIsSubmitting(false)
    }
    }


  return (
    <div className="register-page auth-redesign auth-login-page">

      {/* HEADER */}
      <AuthHeader />

      {/* MAIN LAYOUT */}
      <main className="register-layout">

        {/* LEFT PANEL */}
        <aside className="register-left">
          <AuthIllustration />
        </aside>

        {/* RIGHT PANEL */}
        <section className="register-content">
          <div className="register-form-container login-container">

            <h1>Login</h1>

            <p className="register-description">
              Login to your account to start and continue
              generating and evaluating Gherkin scenarios.
            </p>

            {/* LOGIN FORM */}
            <form
            className="register-form login-form"
            noValidate
            onSubmit={handleLogin}
            >

              {/* EMAIL ADDRESS */}
              <div className="form-group">
                <label htmlFor="loginEmail">
                  Email Address
                </label>

                <div className={`auth-error-shell ${emailError ? 'has-error' : ''}`}>
                <div className="input-wrapper">
                  <Mail
                    size={22}
                    className="input-icon"
                  />

                  <input
                    type="email"
                    id="loginEmail"
                    name="email"
                    placeholder="example@gmail.com"
                    autoComplete="email"
                    required
                    onChange={() => { setEmailError(''); setLoginError('') }}
                    aria-invalid={Boolean(emailError)}
                    aria-describedby={emailError ? 'login-email-error' : undefined}
                  />
                </div>
                  {emailError && <p id="login-email-error" className="password-error-message" role="alert">{emailError}</p>}
                </div>
              </div>

              
        {/* PASSWORD */}
        <div className="form-group">
        <label htmlFor="loginPassword">Password</label>

        <div className={`password-field ${loginError ? 'has-error' : ''}`}>

            <div className="input-wrapper">
            <LockKeyhole
                size={24}
                className="input-icon"
            />

            <input
                type={showPassword ? 'text' : 'password'}
                id="loginPassword"
                name="password"
                placeholder="Enter your password"
                autoComplete="current-password"
                value={password}
                onChange={(event) => {
                setPassword(event.target.value)

                if (loginError) {
                    setLoginError('')
                }
                }}
                aria-invalid={Boolean(loginError)}
                aria-describedby={
                loginError ? 'login-password-error' : undefined
                }
                required
            />

            <button
                type="button"
                className="toggle-password"
                onClick={() => setShowPassword(!showPassword)}
                aria-label={
                showPassword ? 'Hide password' : 'Show password'
                }
            >
                {showPassword ? (
                <Eye size={22} />
                ) : (
                <EyeOff size={22} />
                )}
            </button>
            </div>

            {/* ERROR MESSAGE BELOW PASSWORD */}
            {loginError && (
            <p
                className="password-error-message"
                id="login-password-error"
                role="alert"
            >
                {loginError}
            </p>
            )}

        </div>
        </div>


              {/* REMEMBER ME & FORGOT PASSWORD */}
              <div className="login-options">

                <label className="login-remember">
                  <input
                    type="checkbox"
                    checked={rememberMe}
                    onChange={(event) =>
                      setRememberMe(event.target.checked)
                    }
                  />

                  <span>Remember me</span>
                </label>

                <a
                  href="/forgot-password"
                  className="login-forgot"
                >
                  Forgot password?
                </a>
              </div>
              
                {/* LOGIN BUTTON */}
                <button
                type="submit"
                className="register-submit login-submit"
                disabled={isSubmitting}
                >
                {isSubmitting ? 'Logging in...' : 'Login'}
                </button>

            </form>

            {/* DIVIDER */}
            <div className="register-divider login-divider">
              <span>or</span>
            </div>

            {/* SIGN UP LINK */}
            <p className="register-login login-signup">
              Don&apos;t have an account?{' '}
              <a href="/register">
                Sign Up
              </a>
            </p>

          </div>
        </section>

      </main>
    </div>
  )
}

export default LoginPage
