import { Link } from 'react-router-dom'
import { ArrowRight, FileCheck2 } from 'lucide-react'
import PublicHeader from '../components/PublicHeader'
import ScenarioSidebar from '../components/ScenarioSidebar'
import './GenerateScenarioPage.css'
import './DocumentationPage.css'

export default function EvaluationPage() {
  return (
    <div className="generate-page">
      <PublicHeader />
      <div className="generate-layout">
        <ScenarioSidebar />
        <main className="generate-main">
          <div className="generate-page-heading"><h1>Evaluation</h1><p>A Quality Assurance workspace for evaluating the quality of Gherkin scenarios.</p></div>
          <section className="documentation-card" aria-labelledby="evaluation-status-title">
            <FileCheck2 className="documentation-icon" size={32} />
            <h2 id="evaluation-status-title">Evaluation Features Are Under Development</h2>
            <span className="documentation-status">Not Yet Available</span>
            <p>
              Golden Dataset, Run Evaluation, and Evaluation Report will be available
              once development is complete. For now, you can read the Story2Gherkin
              user guide.
            </p>
            <Link to="/documentation#qa" className="documentation-action">Read the QA Documentation <ArrowRight size={18} /></Link>
          </section>
        </main>
      </div>
    </div>
  )
}
