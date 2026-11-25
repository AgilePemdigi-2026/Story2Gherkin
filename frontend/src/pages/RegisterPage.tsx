
import { useState, type FormEvent } from 'react'
import {
  UserRound,
  Mail,
  Eye,
  EyeOff,
  ChevronDown,
  UsersRound,
  Workflow,
  ShieldCheck,
  Check
} from 'lucide-react'
import { registerUser } from '../services/authApi'
import AuthHeader from '../components/AuthHeader'
import AuthIllustration from '../components/AuthIllustration'
import LegalDocumentDialog from '../components/LegalDocumentDialog'
import './RegisterPage.css'
import './AuthPages.css'
import './AuthValidation.css'
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
  const [showLegalDocument, setShowLegalDocument] = useState(false)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  // Password validation
  const validatePassword = (value: string): string => {
    if (/[\p{White_Space}\uFEFF]/u.test(value)) {
      return 'Password must not contain spaces, tabs, or line breaks.'
    }

    const hasUppercase = /[A-Z]/.test(value)
    const hasLowercase = /[a-z]/.test(value)
    const isValidLength = Array.from(value).length >= 8

    if (!isValidLength || !hasUppercase || !hasLowercase) {
      return 'Password must contain at least 8 characters, including uppercase and lowercase letters.'
    }

    if (new TextEncoder().encode(value).length > 72) {
      return 'Password must not exceed 72 bytes.'
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

      const role = formData.get('role')
      const emailInput = event.currentTarget.elements.namedItem('email') as HTMLInputElement
      const errors: Record<string, string> = {}
      if (!name) errors.name = 'Full name is required.'
      if (!email) errors.email = 'Email address is required.'
      else if (emailInput.validity.typeMismatch) errors.email = 'Enter a valid email address.'
      if (role !== 'system_analyst' && role !== 'qa') errors.role = 'Please select your role.'
      const passwordIssue = !password ? 'Password is required.' : validatePassword(password)
      const confirmIssue = !confirmPassword ? 'Please confirm your password.' : password !== confirmPassword ? 'Passwords do not match.' : ''
      if (passwordIssue) errors.password = passwordIssue
      if (confirmIssue) errors.confirmPassword = confirmIssue
      if (!agreeTerms) errors.agreeTerms = 'Please agree to the Terms of Use and acknowledge the Privacy Policy.'
      setFieldErrors(errors)
      setPasswordError(passwordIssue)
      setConfirmPasswordError(confirmIssue)
      if (Object.keys(errors).length > 0) {
        const selectors: Record<string, string> = { name: '#fullName', email: '#email', role: '#role-trigger', password: '#password', confirmPassword: '#confirmPassword', agreeTerms: '#agreeTerms' }
        event.currentTarget.querySelector<HTMLElement>(selectors[Object.keys(errors)[0]])?.focus()
        return
      }
      // Keep role narrowed for the registration API.
      if (role !== 'system_analyst' && role !== 'qa') return

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
    <div className="register-page auth-redesign auth-signup-page">

      {/* HEADER */}
      <AuthHeader />

      {/* MAIN LAYOUT */}
      <main className="register-layout">

        {/* LEFT PANEL: ILLUSTRATION */}
        <aside className="register-left">
          <AuthIllustration />
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
              noValidate
              onSubmit={handleRegister}
              onChange={event => {
                const target = event.target
                if (target instanceof HTMLInputElement) {
                  setFieldErrors(current => ({ ...current, [target.name]: '' }))
                }
              }}
            >

              {/* FULL NAME */}
              <div className="form-group">
                <label htmlFor="fullName">Full Name</label>

                <div className={`auth-error-shell ${fieldErrors.name ? 'has-error' : ''}`}>
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
                    aria-invalid={Boolean(fieldErrors.name)}
                    aria-describedby={fieldErrors.name ? 'name-error' : undefined}
                  />
                </div>
                  {fieldErrors.name && <p id="name-error" className="password-error-message" role="alert">{fieldErrors.name}</p>}
                </div>
              </div>

              {/* EMAIL ADDRESS */}
              <div className="form-group">
                <label htmlFor="email">Email Address</label>

                <div className={`auth-error-shell ${fieldErrors.email ? 'has-error' : ''}`}>
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
                    aria-invalid={Boolean(fieldErrors.email)}
                    aria-describedby={fieldErrors.email ? 'email-error' : undefined}
                  />
                </div>
                  {fieldErrors.email && <p id="email-error" className="password-error-message" role="alert">{fieldErrors.email}</p>}
                </div>
              </div>

              
              
{/* ROLE */}
<div className="form-group">
  <label htmlFor="role-trigger">Role</label>

  <div className={`auth-error-shell ${fieldErrors.role ? 'has-error' : ''}`}>
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
      aria-invalid={Boolean(fieldErrors.role)}
      aria-describedby={fieldErrors.role ? 'role-error' : undefined}
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
                  setFieldErrors(current => ({ ...current, role: '' }))
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
                  setFieldErrors(current => ({ ...current, role: '' }))
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
        {fieldErrors.role && <p id="role-error" className="password-error-message" role="alert">{fieldErrors.role}</p>}
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
              <div className={`register-legal-agreement-wrap ${fieldErrors.agreeTerms ? 'has-error' : ''}`}>
                <div className="register-terms register-legal-agreement">
                  <input type="checkbox" id="agreeTerms" name="agreeTerms" required
                    aria-labelledby="agree-terms-copy"
                    aria-invalid={Boolean(fieldErrors.agreeTerms)}
                    aria-describedby={fieldErrors.agreeTerms ? 'agreement-error' : undefined} />
                  <p className="register-legal-copy" id="agree-terms-copy">
                    <label htmlFor="agreeTerms">I agree to the </label>
                    <button type="button" className="terms-button" aria-haspopup="dialog" onClick={() => setShowLegalDocument(true)}>Terms &amp; Privacy Policy</button>.
                  </p>
                </div>
                {fieldErrors.agreeTerms && <p id="agreement-error" className="password-error-message" role="alert">{fieldErrors.agreeTerms}</p>}
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

      {showLegalDocument && (
        <LegalDocumentDialog
          onClose={() => setShowLegalDocument(false)}
        />
      )}

    </div>
  )
}

export default RegisterPage
