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
  'Automate Gherkin scenario generation.',
  'Reduce manual effort in writing test scenarios.',
  'Automatically validate Gherkin structure.',
  'Check GUI element coverage.',
  'Compare LLM-generated results with a rule-based approach.',
]

const benefits = [
  'Save time when creating test scenarios.',
  'Maintain consistent Gherkin formatting.',
  'Reduce human error.',
  'Help analysts identify GUI elements that are not covered.',
  'Provide more objective evaluation results.',
]

const technologies = [
  { name: 'Python', Icon: Code2, color: '#f59e0b' },
  {
    name: 'Open-Source Large Language Model',
    Icon: Bot,
    color: '#059669',
  },
  { name: 'Gherkin Parser', Icon: FileCode2, color: '#0ea5e9' },
  { name: 'Behave', Icon: FileCheck2, color: '#6366f1' },
  {
    name: 'Rule-Based BDD Generator',
    Icon: Terminal,
    color: '#0891b2',
  },
  { name: 'Web Application', Icon: Globe2, color: '#2563eb' },
]

function AboutProjectPage() {
  return (
    <div className="about-page">
      {/* NAVIGATION */}
      <PublicHeader />

      <main className="about-main">
        {/* ABOUT HERO */}
        <section className="about-hero">
          <div className="about-hero-content">
            <h1 className="about-title">
              About
              <span>the Story2Gherkin Project</span>
            </h1>

            <p className="about-description">
              Story2Gherkin is a Gherkin generation and evaluation
              system designed to help System Analysts and QA
              professionals create and evaluate test scenarios
              more quickly, consistently, and systematically
              using a Large Language Model (LLM).
            </p>
          </div>

          <div className="about-hero-visual">
            <img
              src={aboutIllustration}
              alt="Illustration of user stories being converted into Gherkin scenarios and evaluated"
              className="about-illustration"
            />
          </div>
        </section>

        {/* INFORMATION CARDS */}
        <section
          className="about-info-section"
          aria-label="Project information"
        >
          <div className="about-cards">
            {/* OBJECTIVES */}
            <article className="about-card">
              <div className="about-card-header">
                <div className="about-card-icon objective-icon">
                  <Target size={28} />
                </div>

                <h2>Objectives</h2>
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

            {/* TECHNOLOGIES */}
            <article className="about-card">
              <div className="about-card-header">
                <div className="about-card-icon technology-icon">
                  <Settings2 size={28} />
                </div>

                <h2>Technologies</h2>
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

            {/* BENEFITS */}
            <article className="about-card">
              <div className="about-card-header">
                <div className="about-card-icon benefit-icon">
                  <Lightbulb size={28} />
                </div>

                <h2>Benefits</h2>
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