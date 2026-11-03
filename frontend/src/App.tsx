
import { createBrowserRouter, Navigate, RouterProvider } from 'react-router-dom'

import RegisterPage from './pages/RegisterPage'
import LoginPage from './pages/LoginPage'
import ForgotPasswordPage from './pages/ForgotPasswordPage'
import ResetPasswordPage from './pages/ResetPasswordPage'
import HomePage from './pages/HomePage'
import AboutProjectPage from './pages/AboutProjectPage'
import GenerateScenarioPage from './pages/GenerateScenarioPage'
import ReviewScenarioPage from './pages/ReviewScenarioPage'

// A data router enables useBlocker to protect unsaved review corrections.
const router = createBrowserRouter([
  { path: '/', element: <HomePage /> },
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
  { path: '/forgot-password', element: <ForgotPasswordPage /> },
  { path: '/reset-password', element: <ResetPasswordPage /> },
  { path: '/about', element: <AboutProjectPage /> },
  { path: '/generate', element: <GenerateScenarioPage /> },
  { path: '/review', element: <ReviewScenarioPage /> },
  { path: '*', element: <Navigate to="/login" replace /> },
])

function App() {
  return <RouterProvider router={router} />
}

export default App
