
import { useState, type FormEvent } from 'react'
import {
  UserRound,
  Mail,
  LockKeyhole,
  Eye,
  EyeOff,
  ChevronDown,
  UsersRound,
  X, 
  Workflow,
  ShieldCheck,
  Check
} from 'lucide-react'
import { registerUser } from '../services/authApi'
import logo from '../assets/story2gherkin-logo.svg'
import registerIllustration from '../assets/register-illustration.png'
import './RegisterPage.css'
import { useNavigate } from 'react-router-dom'

function RegisterPage() {

  const navigate = useNavigate()
  // Password visibility
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)

  // Form values
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')

  // Validation messages
  const [passwordError, setPasswordError] = useState('')
  const [confirmPasswordError, setConfirmPasswordError] = useState('')
  const [formMessage, setFormMessage] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [messageType, setMessageType] = useState<'success' | 'error' | ''>('')

  // Terms & Privacy modal
  const [showTermsModal, setShowTermsModal] = useState(false)

  // Password validation
  const validatePassword = (value: string): string => {
    const hasUppercase = /[A-Z]/.test(value)
    const hasLowercase = /[a-z]/.test(value)
    const isValidLength = value.length >= 8

    if (!isValidLength || !hasUppercase || !hasLowercase) {
      return 'Password should be at least 8 characters long and use a mix of uppercase and lowercase letters.'
    }

    return ''
  }

  // Handle password changes
  const handlePasswordChange = (value: string) => {
    setPassword(value)

    if (passwordError) {
      setPasswordError(validatePassword(value))
    }

    if (confirmPasswordError) {
      setConfirmPasswordError(
        confirmPassword && confirmPassword !== value
          ? 'Passwords do not match.'
          : ''
      )
    }
  }

  // Handle confirm password changes
  const handleConfirmPasswordChange = (value: string) => {
    setConfirmPassword(value)

    if (confirmPasswordError) {
      setConfirmPasswordError(
        value !== password ? 'Passwords do not match.' : ''
      )
    }
  }
  const [selectedRole, setSelectedRole] = useState<
  'system_analyst' | 'qa' | ''
  >('')

  const [roleOpen, setRoleOpen] = useState(false)

  // Handle registration
  
  const handleRegister = async (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault()

      if (isSubmitting) return

      setFormMessage('')
      setMessageType('')

      // Ambil data dari form
      const formData = new FormData(event.currentTarget)
      
      const name = String(formData.get('name') ?? '').trim()
      const email = String(formData.get('email') ?? '').trim()
      const agreeTerms = formData.get('agreeTerms') === 'on'

      // Ambil role dari dropdown
      const role = formData.get('role')

      // Validasi role sebelum dikirim ke backend
      if (role !== 'system_analyst' && role !== 'qa') {
        setMessageType('error')
        setFormMessage('Please select your role.')
        return
      }

      // Validasi password
      const error = validatePassword(password)
      setPasswordError(error)

      if (error) {
          return
      }

      // Validasi konfirmasi password
      if (password !== confirmPassword) {
          setConfirmPasswordError('Passwords do not match.')
          return
      }

      setConfirmPasswordError('')

      // Validasi persetujuan terms
      if (!agreeTerms) {
          setFormMessage('Please agree to the Terms & Privacy Policies.')
          setMessageType('error')
          return
      }

      setIsSubmitting(true)

      try {
          // Kirim data ke backend Register
          await registerUser({
          name,
          email,
          password,
          role,
          })

          // Redirect langsung 
          navigate('/login', { replace: true })

      } catch (error) {
          setMessageType('error')

          setFormMessage(
          error instanceof Error
              ? error.message
              : 'Registration failed. Please try again.'
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

        {/* LEFT PANEL: ILLUSTRATION */}
        <aside className="register-left">
          <img
            src={registerIllustration}
            alt="Story2Gherkin illustration"
            className="register-illustration"
          />
        </aside>

        {/* RIGHT PANEL: REGISTER FORM */}
        <section className="register-content">
          <div className="register-form-container">

            <h1>Sign Up</h1>

            <p className="register-description">
              Create your new account to start generating
              and evaluating Gherkin scenarios.
            </p>

            <form
              className="register-form"
              onSubmit={handleRegister}
            >

              {/* FULL NAME */}
              <div className="form-group">
                <label htmlFor="fullName">Full Name</label>

                <div className="input-wrapper">
                  <UserRound
                    size={22}
                    className="input-icon"
                  />

                  <input
                    type="text"
                    id="fullName"
                    name="name"
                    placeholder="enter your name"
                    autoComplete="name"
                    required
                  />
                </div>
              </div>

              {/* EMAIL ADDRESS */}
              <div className="form-group">
                <label htmlFor="email">Email Address</label>

                <div className="input-wrapper">
                  <Mail
                    size={22}
                    className="input-icon"
                  />

                  <input
                    type="email"
                    id="email"
                    name="email"
                    placeholder="example@gmail.com"
                    autoComplete="email"
                    required
                  />
                </div>
              </div>

              
              
{/* ROLE */}
<div className="form-group">
  <label htmlFor="role-trigger">Role</label>

  <div
    className="role-picker"
    onBlur={(event) => {
      if (!event.currentTarget.contains(event.relatedTarget)) {
        setRoleOpen(false)
      }
    }}
    onKeyDown={(event) => {
      if (event.key === 'Escape') {
        setRoleOpen(false)
      }
    }}
  >
    {/* Tetap dikirim melalui FormData */}
    <input
      type="hidden"
      name="role"
      value={selectedRole}
    />

    {/* Dropdown Trigger */}
    <button
      type="button"
      id="role-trigger"
      className="role-trigger"
      aria-expanded={roleOpen}
      aria-haspopup="true"
      onClick={() => setRoleOpen(!roleOpen)}
    >
      <UsersRound size={22} className="input-icon" />

      <span className={selectedRole ? 'role-value' : 'role-placeholder'}>
        {selectedRole === 'system_analyst'
          ? 'System Analyst'
          : selectedRole === 'qa'
          ? 'Quality Assurance'
          : 'Select your role'}
      </span>

      <ChevronDown
        size={20}
        className={`role-chevron ${roleOpen ? 'open' : ''}`}
      />
    </button>

    {/* Dropdown Menu */}
          {roleOpen && (
            <div className="role-dropdown" aria-label="Role options">

              <button
                type="button"
                className={`role-option ${
                  selectedRole === 'system_analyst' ? 'selected' : ''
                }`}
                onClick={() => {
                  setSelectedRole('system_analyst')
                  setRoleOpen(false)
                }}
              >
                <Workflow size={20} className="role-option-icon" />
                <span className="role-code">SA</span>
                <span className="role-name">System Analyst</span>
                {selectedRole === 'system_analyst' && (
                  <Check size={18} className="role-check" />
                )}
              </button>

              <button
                type="button"
                className={`role-option ${
                  selectedRole === 'qa' ? 'selected' : ''
                }`}
                onClick={() => {
                  setSelectedRole('qa')
                  setRoleOpen(false)
                }}
              >
                <ShieldCheck size={20} className="role-option-icon" />
                <span className="role-code">QA</span>
                <span className="role-name">Quality Assurance</span>
                {selectedRole === 'qa' && (
                  <Check size={18} className="role-check" />
                )}
              </button>

            </div>
          )}
        </div>
      </div>



              {/* PASSWORD */}
              <div className="form-group">
                <label htmlFor="password">Password</label>

                <div
                  className={`password-field ${
                    passwordError ? 'has-error' : ''
                  }`}
                >
                  <div className="input-wrapper">
                    <LockKeyhole
                      size={24}
                      className="input-icon"
                    />

                    <input
                      type={showPassword ? 'text' : 'password'}
                      id="password"
                      name="password"
                      placeholder="Enter your password"
                      autoComplete="new-password"
                      value={password}
                      required
                      aria-invalid={Boolean(passwordError)}
                      aria-describedby={
                        passwordError ? 'password-error' : undefined
                      }
                      onChange={(event) =>
                        handlePasswordChange(event.target.value)
                      }
                      onBlur={() =>
                        setPasswordError(validatePassword(password))
                      }
                    />

                    <button
                      type="button"
                      className="toggle-password"
                      onClick={() =>
                        setShowPassword(!showPassword)
                      }
                      aria-label={
                        showPassword
                          ? 'Hide password'
                          : 'Show password'
                      }
                    >
                      {showPassword ? (
                        <Eye size={22} />
                      ) : (
                        <EyeOff size={22} />
                      )}
                    </button>
                  </div>

                  {passwordError && (
                    <p
                      className="password-error-message"
                      id="password-error"
                      role="alert"
                    >
                      {passwordError}
                    </p>
                  )}
                </div>
              </div>

              {/* CONFIRM PASSWORD */}
              <div className="form-group">
                <label htmlFor="confirmPassword">
                  Confirm Password
                </label>

                <div
                  className={`password-field ${
                    confirmPasswordError ? 'has-error' : ''
                  }`}
                >
                  <div className="input-wrapper">
                    <LockKeyhole
                      size={24}
                      className="input-icon"
                    />

                    <input
                      type={
                        showConfirmPassword ? 'text' : 'password'
                      }
                      id="confirmPassword"
                      name="confirmPassword"
                      placeholder="Confirm your password"
                      autoComplete="new-password"
                      value={confirmPassword}
                      required
                      aria-invalid={Boolean(confirmPasswordError)}
                      aria-describedby={
                        confirmPasswordError
                          ? 'confirm-password-error'
                          : undefined
                      }
                      onChange={(event) =>
                        handleConfirmPasswordChange(
                          event.target.value
                        )
                      }
                      onBlur={() =>
                        setConfirmPasswordError(
                          confirmPassword !== password
                            ? 'Passwords do not match.'
                            : ''
                        )
                      }
                    />

                    <button
                      type="button"
                      className="toggle-password"
                      onClick={() =>
                        setShowConfirmPassword(
                          !showConfirmPassword
                        )
                      }
                      aria-label={
                        showConfirmPassword
                          ? 'Hide confirm password'
                          : 'Show confirm password'
                      }
                    >
                      {showConfirmPassword ? (
                        <Eye size={22} />
                      ) : (
                        <EyeOff size={22} />
                      )}
                    </button>
                  </div>

                  {confirmPasswordError && (
                    <p
                      className="password-error-message"
                      id="confirm-password-error"
                      role="alert"
                    >
                      {confirmPasswordError}
                    </p>
                  )}
                </div>
              </div>

              {/* TERMS & PRIVACY */}
              <div className="register-terms">
                <input
                  type="checkbox"
                  id="agreeTerms"
                  name="agreeTerms"
                  required
                />

                <label htmlFor="agreeTerms">
                  I agree to the
                </label>

                <button
                  type="button"
                  className="terms-button"
                  onClick={() => setShowTermsModal(true)}
                >
                  Terms &amp; Privacy Policies
                </button>
              </div>

              {/* SIGN UP BUTTON */}
              <button
                type="submit"
                className="register-submit"
                disabled={isSubmitting}
                >
                {isSubmitting ? 'Creating Account...' : 'Sign Up'}
                </button>

              {/* FORM MESSAGE */}
              {formMessage && (
                <p
                    className={`register-message ${messageType}`}
                    role={messageType === 'error' ? 'alert' : 'status'}
                >
                    {formMessage}
                </p>
                )}

            </form>

            {/* DIVIDER */}
            <div className="register-divider">
              <span>or</span>
            </div>

            {/* LOGIN LINK */}
            <p className="register-login">
              Already have an account?{' '}
              <a href="/login">Login</a>
            </p>

          </div>
        </section>
      </main>

      {/* TERMS & PRIVACY MODAL */}
      {showTermsModal && (
        <div
          className="terms-modal-overlay"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) {
              setShowTermsModal(false)
            }
          }}
        >
          <div
            className="terms-modal"
            role="dialog"
            aria-modal="true"
            aria-labelledby="terms-modal-title"
            onKeyDown={(event) => {
              if (event.key === 'Escape') {
                setShowTermsModal(false)
              }
            }}
          >
            <div className="terms-modal-header">
              <h2 id="terms-modal-title">
                Terms &amp; Privacy Policies
              </h2>

              <button
                type="button"
                className="terms-modal-close"
                onClick={() => setShowTermsModal(false)}
                aria-label="Close terms and privacy policies"
                autoFocus
              >
                <X size={22} />
              </button>
            </div>

            <div className="terms-modal-body">

              <p>
                <strong>
                  Draft for project review — not the final policy.
                </strong>
              </p>

              <h3>1. Terms of Use</h3>
              <p>
                Story2Gherkin is designed to support the
                generation and evaluation of Gherkin
                scenarios. Users are expected to use the
                platform responsibly.
              </p>

              <h3>2. Account Information</h3>
              <p>
                Users are responsible for providing accurate
                account information and maintaining the
                confidentiality of their login credentials.
              </p>

              <h3>3. Privacy Policy</h3>
              <p>
                This section will explain how account
                information is collected, processed, stored,
                and protected. The final privacy policy
                will be provided by the project team.
              </p>

            </div>

            <button
              type="button"
              className="terms-modal-confirm"
              onClick={() => setShowTermsModal(false)}
            >
              Close
            </button>

          </div>
        </div>
      )}

    </div>
  )
}

export default RegisterPage
