import { NavLink } from 'react-router-dom'
import { useAuthSession } from '../hooks/useAuthSession'

export default function ScenarioSidebar() {
  const session = useAuthSession()
  return (
    <aside className="generate-sidebar">
      <nav className="generate-sidebar-nav" aria-label="Workspace navigation">
        {session?.role === 'system_analyst' && <div className="generate-sidebar-section">
          <p className="generate-sidebar-heading">GENERATE</p>
          <NavLink to="/generate" className={({ isActive }) => `generate-sidebar-link ${isActive ? 'active' : ''}`}>Generate Scenario</NavLink>
          <NavLink to="/review" className={({ isActive }) => `generate-sidebar-link ${isActive ? 'active' : ''}`}>Review Scenario</NavLink>
        </div>}
        {session?.role === 'qa' && <div className="generate-sidebar-section">
          <p className="generate-sidebar-heading">EVALUATION</p>
          <span className="generate-sidebar-link disabled" aria-disabled="true" title="Belum tersedia">Golden Dataset</span>
          <span className="generate-sidebar-link disabled" aria-disabled="true" title="Belum tersedia">Run Evaluation</span>
          <span className="generate-sidebar-link disabled" aria-disabled="true" title="Belum tersedia">Evaluation Report</span>
        </div>}
      </nav>
    </aside>
  )
}
