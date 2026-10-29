
import {
  BrowserRouter,
  Routes,
  Route,
  Navigate
} from 'react-router-dom'

import RegisterPage from './pages/RegisterPage'
import LoginPage from './pages/LoginPage'
import ForgotPasswordPage from './pages/ForgotPasswordPage'
import ResetPasswordPage from './pages/ResetPasswordPage'
import HomePage from './pages/HomePage'
import AboutProjectPage from './pages/AboutProjectPage'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Default page */}
        <Route path="/" element={<HomePage />} />

        {/* Authentication pages */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />

        {/* Unknown pages */}
        <Route
          path="*"
          element={<Navigate to="/login" replace />}
        />
        <Route
          path="/forgot-password"
          element={<ForgotPasswordPage />}
        />

        <Route
          path="/reset-password"
          element={<ResetPasswordPage />}
        />

        <Route path="/" element={<HomePage />} />
        <Route
          path="/about"
          element={<AboutProjectPage />}
        />
      </Routes>
    </BrowserRouter>
  )
}

export default App
