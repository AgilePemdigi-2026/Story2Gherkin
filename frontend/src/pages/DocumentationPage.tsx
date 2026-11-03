import { useEffect, useState } from 'react'
import { Link, useLocation, useSearchParams } from 'react-router-dom'
import { ArrowLeft, ArrowRight, BookOpen, ChevronDown, ChevronRight, House, Menu, Search } from 'lucide-react'
import PublicHeader from '../components/PublicHeader'
import { documentationArticles, documentationGroups, resolveDocumentationTopic, type DocumentationArticle } from '../content/documentation'
import './DocumentationPage.css'

function TableOfContents({ article, hash }: { article: DocumentationArticle; hash: string }) {
  return <ul>{article.sections.map(section => <li key={section.id}><Link to={`/documentation?topic=${article.slug}#${section.id}`} aria-current={hash === `#${section.id}` ? 'location' : undefined}>{section.title}</Link></li>)}</ul>
}

function DocumentationWorkspace({ article }: { article: DocumentationArticle }) {
  const location = useLocation()
  const [query, setQuery] = useState('')
  const [menuOpen, setMenuOpen] = useState(false)
  const term = query.trim().toLocaleLowerCase()
  const matches = documentationArticles.filter(item => `${item.title} ${item.description} ${item.keywords}`.toLocaleLowerCase().includes(term))
  const index = documentationArticles.indexOf(article)
  const previous = documentationArticles[index - 1]
  const next = documentationArticles[index + 1]

  useEffect(() => {
    const id = location.hash.slice(1)
    const target = id ? document.getElementById(id) : null
    if (target) target.scrollIntoView()
    else window.scrollTo(0, 0)
  }, [article.slug, location.hash])

  return <>
    <div className="docs-mobile-toolbar"><button type="button" aria-expanded={menuOpen} aria-controls="docs-sidebar" onClick={() => setMenuOpen(!menuOpen)}><Menu size={18} />Topik dokumentasi<ChevronDown size={16} className={menuOpen ? 'is-open' : ''} /></button></div>
    <div className="docs-layout">
      <aside id="docs-sidebar" className={`docs-sidebar ${menuOpen ? 'is-open' : ''}`} aria-label="Topik dokumentasi">
        <Link to="/documentation" className="docs-sidebar-title"><BookOpen size={21} /><span>Documentation<small>STORY2GHERKIN</small></span></Link>
        <div className="docs-search"><Search size={17} /><input type="search" aria-label="Cari topik dokumentasi" placeholder="Cari topik…" value={query} onChange={event => setQuery(event.target.value)} /></div>
        <nav aria-label="Navigasi dokumentasi">
          {documentationGroups.map(group => {
            const items = matches.filter(item => item.group === group.id)
            if (!items.length) return null
            return <details className="docs-nav-group" key={`${group.id}-${term}`} open>
              <summary>{group.label}<ChevronRight size={15} /></summary>
              <ul>{items.map(item => <li key={item.slug}><Link to={`/documentation?topic=${item.slug}`} aria-current={article.slug === item.slug ? 'page' : undefined} onClick={() => setMenuOpen(false)}>{item.title}</Link></li>)}</ul>
            </details>
          })}
          {matches.length === 0 && <p className="docs-no-results" role="status">Tidak ada topik untuk “{query}”. Coba kata lain.</p>}
        </nav>
      </aside>

      <main className="docs-main" id="docs-main">
        <nav className="docs-breadcrumb" aria-label="Breadcrumb"><Link to="/" aria-label="Home"><House size={16} /></Link><ChevronRight size={14} /><Link to="/documentation">Documentation</Link><ChevronRight size={14} /><span aria-current="page">{article.title}</span></nav>
        <article className="docs-article" aria-labelledby="docs-article-title">
          <header className="docs-article-heading"><p>PANDUAN STORY2GHERKIN</p><h1 id="docs-article-title">{article.title}</h1><div className="docs-description">{article.description}</div></header>
          <details className="docs-inline-toc"><summary>Di halaman ini<ChevronDown size={16} /></summary><nav aria-label="Daftar isi artikel"><TableOfContents article={article} hash={location.hash} /></nav></details>
          {article.sections.map(section => <section className="docs-section" id={section.id} key={section.id} aria-labelledby={`heading-${section.id}`}><h2 id={`heading-${section.id}`}>{section.title}</h2>{section.content}</section>)}
        </article>
        <nav className="docs-pagination" aria-label="Artikel sebelum dan selanjutnya">
          {previous ? <Link to={`/documentation?topic=${previous.slug}`}><ArrowLeft size={17} /><span><small>Sebelumnya</small>{previous.title}</span></Link> : <span />}
          {next && <Link to={`/documentation?topic=${next.slug}`} className="docs-next"><span><small>Selanjutnya</small>{next.title}</span><ArrowRight size={17} /></Link>}
        </nav>
      </main>

      <aside className="docs-toc" aria-label="Di halaman ini"><p>Di halaman ini</p><nav aria-label="Daftar isi artikel"><TableOfContents article={article} hash={location.hash} /></nav></aside>
    </div>
  </>
}

export default function DocumentationPage() {
  const [params] = useSearchParams()
  const location = useLocation()
  const article = resolveDocumentationTopic(params.get('topic'), location.hash)
  return <div className="documentation-page"><PublicHeader /><DocumentationWorkspace key={article.slug} article={article} /></div>
}
