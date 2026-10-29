import {
  Target,
  Settings2,
  Lightbulb,
  CircleCheck,
  Code2,
  Bot,
  FileCode2,
  FileCheck2,
  Terminal,
  Globe2,
} from 'lucide-react'

import PublicHeader from '../components/PublicHeader'
import aboutIllustration from '../assets/about-project-illustration.png'

import './AboutProjectPage.css'

const objectives = [
  'Mengotomatisasi pembuatan scenario Gherkin.',
  'Mengurangi pekerjaan manual dalam penulisan test scenario.',
  'Memvalidasi struktur Gherkin secara otomatis.',
  'Memeriksa cakupan elemen GUI.',
  'Membandingkan hasil LLM dengan pendekatan rule-based.',
]

const benefits = [
  'Menghemat waktu penyusunan test scenario.',
  'Membantu menjaga konsistensi format Gherkin.',
  'Mengurangi human error.',
  'Membantu analis menemukan elemen GUI yang belum tercakup.',
  'Menyediakan hasil evaluasi yang lebih objektif.',
]

const technologies = [
  { name: 'Python', Icon: Code2, color: '#f59e0b' },
  { name: 'Open-source Large Language Model', Icon: Bot, color: '#059669' },
  { name: 'Gherkin Parser', Icon: FileCode2, color: '#0ea5e9' },
  { name: 'Behave', Icon: FileCheck2, color: '#6366f1' },
  { name: 'Rule-Based BDD Generator', Icon: Terminal, color: '#0891b2' },
  { name: 'Web Application', Icon: Globe2, color: '#2563eb' },
]

function AboutProjectPage() {
  return (
    <div className="about-page">
      {/* NAVBAR */}
      <PublicHeader />

      <main className="about-main">

        {/* ABOUT HERO */}
        <section className="about-hero">
          <div className="about-hero-content">
            <h1 className="about-title">
              Tentang
              <span>Project Story2Gherkin</span>
            </h1>

            <p className="about-description">
              Sistem Generator dan Evaluator Gherkin dikembangkan
              untuk membantu System Analyst dan QA dalam
              menghasilkan serta mengevaluasi skenario pengujian
              secara lebih cepat, konsisten, dan terstruktur
              dengan memanfaatkan Large Language Model (LLM).
            </p>
          </div>

          <div className="about-hero-visual">
            <img
              src={aboutIllustration}
              alt="Ilustrasi proses User Story menjadi Gherkin Scenario dan evaluasi"
              className="about-illustration"
            />
          </div>
        </section>

        {/* INFORMATION CARDS */}
        <section
          className="about-info-section"
          aria-label="Informasi tentang proyek"
        >
          <div className="about-cards">

            {/* TUJUAN */}
            <article className="about-card">
              <div className="about-card-header">
                <div className="about-card-icon objective-icon">
                  <Target size={28} />
                </div>
                <h2>Tujuan</h2>
              </div>

              <ul className="about-card-list">
                {objectives.map((item) => (
                  <li key={item}>
                    <CircleCheck
                      size={18}
                      className="about-check objective-check"
                    />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </article>

            {/* TEKNOLOGI */}
            <article className="about-card">
              <div className="about-card-header">
                <div className="about-card-icon technology-icon">
                  <Settings2 size={28} />
                </div>
                <h2>Teknologi</h2>
              </div>

              <ul className="about-card-list technology-list">
                {technologies.map(({ name, Icon, color }) => (
                  <li key={name}>
                    <Icon
                      size={20}
                      color={color}
                      className="about-tech-icon"
                    />
                    <span>{name}</span>
                  </li>
                ))}
              </ul>
            </article>

            {/* MANFAAT */}
            <article className="about-card">
              <div className="about-card-header">
                <div className="about-card-icon benefit-icon">
                  <Lightbulb size={28} />
                </div>
                <h2>Manfaat</h2>
              </div>

              <ul className="about-card-list">
                {benefits.map((item) => (
                  <li key={item}>
                    <CircleCheck
                      size={18}
                      className="about-check benefit-check"
                    />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </article>

          </div>
        </section>

      </main>
    </div>
  )
}

export default AboutProjectPage
