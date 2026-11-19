import { useEffect, useState } from 'react'
import {
  Link,
  useLocation,
  useSearchParams,
} from 'react-router-dom'
import {
  ArrowLeft,
  ArrowRight,
  ChevronDown,
  ChevronRight,
  Menu,
  Search,
} from 'lucide-react'

import PublicHeader from '../components/PublicHeader'
import {
  documentationArticles,
  documentationGroups,
  resolveDocumentationTopic,
  type DocumentationArticle,
} from '../content/documentation'

import './DocumentationPage.css'

function TableOfContents({
  article,
  hash,
}: {
  article: DocumentationArticle
  hash: string
}) {
  return (
    <ul>
      {article.sections.map((section) => (
        <li key={section.id}>
          <Link
            to={`/documentation?topic=${article.slug}#${section.id}`}
            aria-current={
              hash === `#${section.id}`
                ? 'location'
                : undefined
            }
          >
            {section.title}
          </Link>
        </li>
      ))}
    </ul>
  )
}

function DocumentationWorkspace({
  article,
}: {
  article: DocumentationArticle
}) {
  const location = useLocation()

  const [query, setQuery] = useState('')
  const [menuOpen, setMenuOpen] = useState(false)

  const term = query.trim().toLocaleLowerCase()

  const matches = documentationArticles.filter((item) =>
    `${item.title} ${item.description} ${item.keywords}`
      .toLocaleLowerCase()
      .includes(term)
  )

  const index = documentationArticles.indexOf(article)
  const previous = documentationArticles[index - 1]
  const next = documentationArticles[index + 1]

  useEffect(() => {
    const id = location.hash.slice(1)
    const target = id
      ? document.getElementById(id)
      : null

    if (target) {
      target.scrollIntoView()
    } else {
      window.scrollTo(0, 0)
    }
  }, [article.slug, location.hash])

  return (
    <>
      <div className="docs-mobile-toolbar">
        <button
          type="button"
          aria-expanded={menuOpen}
          aria-controls="docs-sidebar"
          onClick={() => setMenuOpen(!menuOpen)}
        >
          <Menu size={18} aria-hidden="true" />
          Documentation Topics
          <ChevronDown
            size={16}
            className={menuOpen ? 'is-open' : ''}
            aria-hidden="true"
          />
        </button>
      </div>

      <div className="docs-layout">
        <aside
          id="docs-sidebar"
          className={`docs-sidebar ${
            menuOpen ? 'is-open' : ''
          }`}
          aria-label="Documentation topics"
        >
          <Link
            to="/documentation"
            className="docs-sidebar-title"
          >
            <span>
              Documentation
              <small>STORY2GHERKIN</small>
            </span>
          </Link>

          <div className="docs-search">
            <Search size={17} aria-hidden="true" />

            <input
              type="search"
              aria-label="Search documentation topics"
              placeholder="Search topics…"
              value={query}
              onChange={(event) =>
                setQuery(event.target.value)
              }
            />
          </div>

          <nav aria-label="Documentation navigation">
            {documentationGroups.map((group) => {
              const items = matches.filter(
                (item) => item.group === group.id
              )

              if (!items.length) return null

              return (
                <details
                  className="docs-nav-group"
                  key={`${group.id}-${term}`}
                  open
                >
                  <summary>
                    {group.label}
                    <ChevronRight
                      size={15}
                      aria-hidden="true"
                    />
                  </summary>

                  <ul>
                    {items.map((item) => (
                      <li key={item.slug}>
                        <Link
                          to={`/documentation?topic=${item.slug}`}
                          aria-current={
                            article.slug === item.slug
                              ? 'page'
                              : undefined
                          }
                          onClick={() =>
                            setMenuOpen(false)
                          }
                        >
                          {item.title}
                        </Link>
                      </li>
                    ))}
                  </ul>
                </details>
              )
            })}

            {matches.length === 0 && (
              <p
                className="docs-no-results"
                role="status"
              >
                No topics found for “{query}”. Try another
                keyword.
              </p>
            )}
          </nav>
        </aside>

        <main className="docs-main" id="docs-main">
          <article
            className="docs-article"
            aria-labelledby="docs-article-title"
          >
            <header className="docs-article-heading">
              <p>STORY2GHERKIN GUIDE</p>

              <h1 id="docs-article-title">
                {article.title}
              </h1>

              <div className="docs-description">
                {article.description}
              </div>
            </header>

            <details className="docs-inline-toc">
              <summary>
                On This Page
                <ChevronDown
                  size={16}
                  aria-hidden="true"
                />
              </summary>

              <nav aria-label="Article contents">
                <TableOfContents
                  article={article}
                  hash={location.hash}
                />
              </nav>
            </details>

            {article.sections.map((section) => (
              <section
                className="docs-section"
                id={section.id}
                key={section.id}
                aria-labelledby={`heading-${section.id}`}
              >
                <h2 id={`heading-${section.id}`}>
                  {section.title}
                </h2>

                {section.content}
              </section>
            ))}
          </article>

          <nav
            className="docs-pagination"
            aria-label="Previous and next articles"
          >
            {previous ? (
              <Link
                to={`/documentation?topic=${previous.slug}`}
              >
                <ArrowLeft
                  size={17}
                  aria-hidden="true"
                />

                <span>
                  <small>Previous</small>
                  {previous.title}
                </span>
              </Link>
            ) : (
              <span />
            )}

            {next && (
              <Link
                to={`/documentation?topic=${next.slug}`}
                className="docs-next"
              >
                <span>
                  <small>Next</small>
                  {next.title}
                </span>

                <ArrowRight
                  size={17}
                  aria-hidden="true"
                />
              </Link>
            )}
          </nav>
        </main>

        <aside
          className="docs-toc"
          aria-label="On this page"
        >
          <p>On This Page</p>

          <nav aria-label="Article contents">
            <TableOfContents
              article={article}
              hash={location.hash}
            />
          </nav>
        </aside>
      </div>
    </>
  )
}

export default function DocumentationPage() {
  const [params] = useSearchParams()
  const location = useLocation()

  const article = resolveDocumentationTopic(
    params.get('topic'),
    location.hash
  )

  return (
    <div className="documentation-page">
      <PublicHeader />

      <DocumentationWorkspace
        key={article.slug}
        article={article}
      />
    </div>
  )
}