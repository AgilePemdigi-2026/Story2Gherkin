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
          <div className="generate-page-heading"><h1>Evaluation</h1><p>Ruang kerja Quality Assurance untuk mengevaluasi kualitas skenario Gherkin.</p></div>
          <section className="documentation-card" aria-labelledby="evaluation-status-title">
            <FileCheck2 className="documentation-icon" size={32} />
            <h2 id="evaluation-status-title">Fitur evaluasi sedang dikembangkan</h2>
            <span className="documentation-status">Belum tersedia</span>
            <p>Golden Dataset, Run Evaluation, dan Evaluation Report akan tersedia setelah pengembangannya selesai. Saat ini Anda dapat membaca panduan penggunaan Story2Gherkin.</p>
            <Link to="/documentation#qa" className="documentation-action">Baca Panduan QA <ArrowRight size={18} /></Link>
          </section>
        </main>
      </div>
    </div>
  )
}
