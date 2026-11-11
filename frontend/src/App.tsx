
import { createBrowserRouter, Navigate, RouterProvider } from 'react-router-dom'

import RegisterPage from './pages/RegisterPage'
import LoginPage from './pages/LoginPage'
import ForgotPasswordPage from './pages/ForgotPasswordPage'
import ResetPasswordPage from './pages/ResetPasswordPage'
import HomePage from './pages/HomePage'
import AboutProjectPage from './pages/AboutProjectPage'
import GenerateScenarioPage from './pages/GenerateScenarioPage'
import ReviewScenarioPage from './pages/ReviewScenarioPage'
import DocumentationPage from './pages/DocumentationPage'
import EvaluationPage from './pages/EvaluationPage'
import RequireRole from './components/RequireRole'
import HistoryPage from './pages/HistoryPage'
import HistoryDetailPage from './pages/HistoryDetailPage'

// A data router enables useBlocker to protect unsaved review corrections.
const router = createBrowserRouter([
  { path: '/', element: <HomePage /> },
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
  { path: '/forgot-password', element: <ForgotPasswordPage /> },
  { path: '/reset-password', element: <ResetPasswordPage /> },
  { path: '/about', element: <AboutProjectPage /> },
  { path: '/documentation', element: <DocumentationPage /> },
  { path: '/generate', element: <RequireRole role="system_analyst"><GenerateScenarioPage /></RequireRole> },
  { path: '/review', element: <RequireRole role="system_analyst"><ReviewScenarioPage /></RequireRole> },
  {  path: '/history',  element: (    <RequireRole role="system_analyst">      <HistoryPage />    </RequireRole>),},
  {  path: '/history/:id',  element: (<RequireRole role="system_analyst"><HistoryDetailPage /></RequireRole>)},
  { path: '/evaluation', element: <RequireRole role="qa"><EvaluationPage /></RequireRole> },
  { path: '*', element: <Navigate to="/" replace /> },
])

function App() {
  return <RouterProvider router={router} />
}

export default App
