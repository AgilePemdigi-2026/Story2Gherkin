
import { Link } from 'react-router-dom'
import logo from '../assets/story2gherkin-logo.svg'

function AuthHeader() {
  return (
    <header className="register-header">
      <Link
        to="/"
        className="register-logo-link"
        aria-label="Back to Home"
      >
        <img
          src={logo}
          alt="Story2Gherkin"
          className="register-logo"
        />
      </Link>
    </header>
  )
}

export default AuthHeader
