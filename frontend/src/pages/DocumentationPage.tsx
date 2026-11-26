import { useEffect, useLayoutEffect, useRef, useState } from 'react'
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
  X,
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
  const location = useLocation()

  return (
    <ul>
      {article.sections.map((section) => (
        <li key={section.id}>
          <Link
            to={`/documentation?topic=${article.slug}#${section.id}`}
            onClick={(event) => {
              if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
              if (location.hash === `#${section.id}`) {
                event.preventDefault()
                document.getElementById(section.id)?.scrollIntoView({
                  behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
                  block: 'start',
                })
              }
            }}
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
  const [visibleSection, setVisibleSection] = useState('')
  const navRef = useRef<HTMLElement>(null)
  const indicatorRef = useRef<HTMLSpanElement>(null)
  const mainRef = useRef<HTMLElement>(null)
  const previousArticleRef = useRef<string | null>(null)

  const term = query.trim().toLocaleLowerCase()

  const matches = documentationArticles.filter((item) =>
    `${item.title} ${item.description} ${item.keywords}`
      .toLocaleLowerCase()
      .includes(term)
  )

  const index = documentationArticles.indexOf(article)
  const previous = documentationArticles[index - 1]
  const next = documentationArticles[index + 1]

  useLayoutEffect(() => {
    const nav = navRef.current
    const indicator = indicatorRef.current
    if (!nav || !indicator) return
    const update = () => {
      const link = nav.querySelector<HTMLAnchorElement>('[aria-current="page"]')
      if (!link || link.getClientRects().length === 0) {
        indicator.style.opacity = '0'
        return
      }
      const parent = nav.getBoundingClientRect()
      const rect = link.getBoundingClientRect()
      Object.assign(indicator.style, {
        width: `${rect.width}px`, height: `${rect.height}px`,
        transform: `translate3d(${rect.left - parent.left}px, ${rect.top - parent.top}px, 0)`,
        opacity: '1',
      })
    }
    update()
    const observer = new ResizeObserver(update)
    observer.observe(nav)
    nav.addEventListener('toggle', update, true)
    return () => {
      observer.disconnect()
      nav.removeEventListener('toggle', update, true)
    }
  }, [article.slug, term, menuOpen])

  useEffect(() => {
    let frame = 0
    const update = () => {
      const sections = article.sections.map((section) => document.getElementById(section.id))
      const passed = sections.filter((section) => section && section.getBoundingClientRect().top <= 180)
      setVisibleSection(passed.at(-1)?.id ?? article.sections[0]?.id ?? '')
    }
    const onScroll = () => {
      cancelAnimationFrame(frame)
      frame = requestAnimationFrame(update)
    }
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => {
      cancelAnimationFrame(frame)
      window.removeEventListener('scroll', onScroll)
    }
  }, [article])

  useEffect(() => {
    const sameArticle = previousArticleRef.current === article.slug
    previousArticleRef.current = article.slug
    const target = document.getElementById(location.hash.slice(1))
    if (target) {
      target.scrollIntoView({
        behavior: sameArticle && !window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'smooth' : 'auto',
        block: 'start',
      })
    } else {
      window.scrollTo({ top: 0, behavior: 'instant' })
    }
  }, [article.slug, location.hash])

  useEffect(() => {
    const cleanups: (() => void)[] = []
    mainRef.current?.querySelectorAll<HTMLDetailsElement>('.docs-faq details').forEach((details) => {
      const summary = details.querySelector('summary')
      if (!summary) return
      let animation: Animation | null = null
      let expanded = details.open
      const reset = () => {
        details.style.removeProperty('height')
        details.style.removeProperty('overflow')
        details.style.removeProperty('box-sizing')
      }
      const toggle = (event: MouseEvent) => {
        event.preventDefault()
        expanded = !expanded
        const startHeight = details.getBoundingClientRect().height
        animation?.cancel()
        if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
          details.open = expanded
          reset()
          return
        }
        reset()
        details.open = true
        const endHeight = expanded
          ? details.getBoundingClientRect().height
          : summary.getBoundingClientRect().height + 2
        details.style.boxSizing = 'border-box'
        details.style.overflow = 'hidden'
        details.style.height = `${endHeight}px`
        animation = details.animate(
          [{ height: `${startHeight}px` }, { height: `${endHeight}px` }],
          { duration: expanded ? 300 : 190, easing: 'cubic-bezier(0.22, 1, 0.36, 1)' },
        )
        animation.onfinish = () => {
          details.open = expanded
          reset()
          animation = null
        }
      }
      summary.addEventListener('click', toggle)
      cleanups.push(() => {
        animation?.cancel()
        summary.removeEventListener('click', toggle)
        reset()
      })
    })
    return () => cleanups.forEach((cleanup) => cleanup())
  }, [article.slug])

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
            {query && (
              <button type="button" className="docs-search-clear" aria-label="Clear topic search" onClick={() => setQuery('')}>
                <X size={16} aria-hidden="true" />
              </button>
            )}
          </div>

          <nav ref={navRef} className="docs-topic-nav" aria-label="Documentation navigation">
            <span ref={indicatorRef} className="docs-nav-indicator" aria-hidden="true" />
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
              <div className="docs-no-results" role="status">
                <Search size={24} aria-hidden="true" />
                <strong>No topics found</strong>
                <p>Try a different keyword for “{query}”.</p>
                <button type="button" onClick={() => setQuery('')}>Clear search</button>
              </div>
            )}
          </nav>
        </aside>

        <main ref={mainRef} className="docs-main" id="docs-main">
          <article
            key={article.slug}
            className="docs-article"
            aria-labelledby="docs-article-title"
          >
            <header className="docs-article-heading">
              <p>STORY2GHERKIN GUIDE <span>{String(index + 1).padStart(2, '0')} / {String(documentationArticles.length).padStart(2, '0')}</span></p>

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
                  hash={`#${visibleSection}`}
                />
              </nav>
            </details>

            {article.sections.map((section, sectionIndex) => (
              <section
                className="docs-section"
                id={section.id}
                key={section.id}
                aria-labelledby={`heading-${section.id}`}
              >
                <h2 id={`heading-${section.id}`}>
                  <span className="docs-section-number" aria-hidden="true">{String(sectionIndex + 1).padStart(2, '0')}. </span>
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
              hash={`#${visibleSection}`}
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

      <DocumentationWorkspace article={article} />
    </div>
  )
}