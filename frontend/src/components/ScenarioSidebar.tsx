import { NavLink } from 'react-router-dom'
import { useAuthSession } from '../hooks/useAuthSession'

function sidebarLinkClassName({
  isActive,
}: {
  isActive: boolean
}) {
  return `generate-sidebar-link${isActive ? ' active' : ''}`
}

export default function ScenarioSidebar() {
  const session = useAuthSession()

  return (
    <aside className="generate-sidebar">
      <nav
        className="generate-sidebar-nav"
        aria-label="Workspace navigation"
      >
        {session?.role === 'system_analyst' && (
          <div className="generate-sidebar-section">
            <p className="generate-sidebar-heading">
              GENERATE
            </p>

            <NavLink
              to="/generate"
              className={sidebarLinkClassName}
            >
              Generate Scenario
            </NavLink>

            <NavLink
              to="/review"
              className={sidebarLinkClassName}
            >
              Review Scenario
            </NavLink>

            <NavLink
              to="/history"
              className={sidebarLinkClassName}
            >
              History
            </NavLink>
          </div>
        )}

        {session?.role === 'qa' && (
          <div className="generate-sidebar-section">
            <p className="generate-sidebar-heading">
              EVALUATION
            </p>

            <NavLink
              to="/evaluation/golden-dataset"
              className={sidebarLinkClassName}
            >
              Golden Dataset
            </NavLink>

            <span
              className="generate-sidebar-link disabled"
              aria-disabled="true"
              title="Coming soon"
            >
              Run Evaluation
            </span>

            <NavLink
              to="/evaluation/reports"
              className={sidebarLinkClassName}
            >
              Evaluation Report
            </NavLink>
          </div>
        )}
      </nav>
    </aside>
  )
}
