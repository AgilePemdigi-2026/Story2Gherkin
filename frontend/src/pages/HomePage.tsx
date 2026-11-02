
import { Link } from 'react-router-dom'
import { ArrowRight, ChevronRight } from 'lucide-react'

import PublicHeader from '../components/PublicHeader'
import homeIllustration from '../assets/homepage-illustration.png'

import './HomePage.css'

function HomePage() {
  // Cek apakah user sudah login
  const isLoggedIn = Boolean(
    localStorage.getItem('access_token') ||
    sessionStorage.getItem('access_token')
  )

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
              Turn your requirements into test scenarios
            </div>

            {/* HEADING */}
            <h1 className="home-hero-title">
              Generate Better
              <span>
                Gherkin Scenarios
                <br />
                with AI
              </span>
            </h1>

            {/* DESCRIPTION */}
            <p className="home-hero-description">
              Ubah user story, elemen GUI, dan deskripsi
              skenario menjadi skenario pengujian Gherkin
              yang siap digunakan. Lebih cepat, konsisten,
              dan tervalidasi.
            </p>

            {/* ACTION BUTTONS */}
            <div className="home-hero-actions">

              <Link
                to={isLoggedIn ? '/generate' : '/login'}
                className="home-primary-button"
              >
                Mulai Generate
                <ArrowRight size={20} />
              </Link>

              <Link
                to="/about"
                className="home-secondary-button"
              >
                Pelajari Lebih Lanjut
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
