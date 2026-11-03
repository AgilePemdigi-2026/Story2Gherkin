
import { Link } from 'react-router-dom'
import { ArrowRight, ChevronRight } from 'lucide-react'

import PublicHeader from '../components/PublicHeader'
import homeIllustration from '../assets/homepage-illustration.png'

import './HomePage.css'
import { useAuthSession } from '../hooks/useAuthSession'
import { roleHome } from '../services/authSession'

function HomePage() {
  const session = useAuthSession()
  const actionLabel = !session ? 'Mulai Menggunakan' : session.role === 'system_analyst' ? 'Mulai Generate' : 'Buka Evaluation'

  return (
    <div className="home-page">

      {/* NAVBAR */}
      <PublicHeader />

      {/* HERO SECTION */}
      <main className="home-main">
        <section className="home-hero">

          {/* LEFT CONTENT */}
          <div className="home-hero-content">

            {/* BADGE */}
            <div className="home-hero-badge">
              Generate and evaluate your Gherkin scenarios
            </div>

            {/* HEADING */}
            <h1 className="home-hero-title">
              Build &amp; Evaluate
              <span>
                Gherkin Scenarios
                <br />
                with Confidence
              </span>
            </h1>

            {/* DESCRIPTION */}
            <p className="home-hero-description">
              Bantu System Analyst menyusun dan meninjau skenario Gherkin,
              serta QA mengevaluasi kualitasnya dalam satu platform.
              Masuk ke ruang kerja sesuai peran Anda.
            </p>

            {/* ACTION BUTTONS */}
            <div className="home-hero-actions">

              <Link
                to={session ? roleHome(session.role) : '/login'}
                className="home-primary-button"
              >
                {actionLabel}
                <ArrowRight size={20} />
              </Link>

              <Link
                to="/documentation"
                className="home-secondary-button"
              >
                Baca Panduan
                <ChevronRight size={20} />
              </Link>

            </div>
          </div>

          {/* RIGHT ILLUSTRATION */}
          <div className="home-hero-visual">
            <img
              src={homeIllustration}
              alt="Story2Gherkin AI scenario generation illustration"
              className="home-hero-image"
            />
          </div>

        </section>
      </main>

    </div>
  )
}

export default HomePage
