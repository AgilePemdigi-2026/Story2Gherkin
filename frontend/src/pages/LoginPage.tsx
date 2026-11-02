
import { useState, type FormEvent } from 'react'
import { Mail, LockKeyhole, Eye, EyeOff } from 'lucide-react'
import { loginUser } from '../services/authApi'
import logo from '../assets/story2gherkin-logo.svg'
import authIllustration from '../assets/register-illustration.png'
import { useLocation, useNavigate } from 'react-router-dom'
import './RegisterPage.css'
import './LoginPage.css'

function LoginPage() {
    const [password, setPassword] = useState('')
    const [showPassword, setShowPassword] = useState(false)
    const [rememberMe, setRememberMe] = useState(false)
    
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [loginError, setLoginError] = useState('')

    const navigate = useNavigate()
    const location = useLocation()

    // Halaman tempat pengguna menekan tombol Login
    const from = (location.state as { from?: string } | null)?.from

    // Pastikan redirect hanya ke path internal aplikasi
    const redirectTo =
      typeof from === 'string' &&
      from.startsWith('/') &&
      !from.startsWith('//') &&
      !from.includes('\\') &&
      !['/login', '/register'].includes(from.split(/[?#]/)[0])
        ? from
        : '/'
    const handleLogin = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    if (isSubmitting) return

    const formData = new FormData(event.currentTarget)
    const email = String(formData.get('email') ?? '').trim()

    setLoginError('')
    setIsSubmitting(true)

    try {
        const result = await loginUser({
        email,
        password,
        })

        // Hapus token sesi sebelumnya
        sessionStorage.removeItem('access_token')
        localStorage.removeItem('access_token')

        // Simpan token sesuai Remember Me
        const storage = rememberMe
        ? localStorage
        : sessionStorage

        storage.setItem('access_token', result.access_token)

        // Kembali ke halaman asal setelah login berhasil
        navigate(redirectTo, { replace: true }) 

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
    <div className="register-page">

      {/* HEADER */}
      <header className="register-header">
        <img
          src={logo}
          alt="Story2Gherkin"
          className="register-logo"
        />
      </header>

      {/* MAIN LAYOUT */}
      <main className="register-layout">

        {/* LEFT PANEL */}
        <aside className="register-left">
          <img
            src={authIllustration}
            alt="Story2Gherkin illustration"
            className="register-illustration"
          />
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
            onSubmit={handleLogin}
            >

              {/* EMAIL ADDRESS */}
              <div className="form-group">
                <label htmlFor="loginEmail">
                  Email Address
                </label>

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
                  />
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
