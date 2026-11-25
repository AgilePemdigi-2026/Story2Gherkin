import { useAuthSession } from '../hooks/useAuthSession'
import AnimatedSANav from './AnimatedSANav'
import AnimatedQANav from './AnimatedQANav'

export default function ScenarioSidebar() {
  const session = useAuthSession()

  const sidebarClass =
    session?.role === 'system_analyst'
      ? 'generate-sidebar sa-workspace-sidebar'
      : session?.role === 'qa'
        ? 'generate-sidebar qa-workspace-sidebar'
        : 'generate-sidebar'

  return (
    <aside className={sidebarClass}>
      <nav
        className="generate-sidebar-nav"
        aria-label="Workspace navigation"
      >
        {session?.role === 'system_analyst' && (
          <AnimatedSANav />
        )}

        {session?.role === 'qa' && (
          <AnimatedQANav />
        )}
      </nav>
    </aside>
  )
}