import { useState } from 'react'
import { Check, Copy } from 'lucide-react'

export default function DocumentationCode({ code, filename = 'scenario.feature' }: { code: string; filename?: string }) {
  const [copyState, setCopyState] = useState<'idle' | 'copied' | 'failed'>('idle')
  async function copyCode() {
    try { await navigator.clipboard.writeText(code); setCopyState('copied') }
    catch { setCopyState('failed') }
  }
  return <figure className="docs-code">
    <figcaption><span>{filename}</span><button type="button" onClick={copyCode} aria-label={`Salin ${filename}`}>{copyState === 'copied' ? <Check size={15} /> : <Copy size={15} />}{copyState === 'copied' ? 'Disalin' : 'Salin'}</button></figcaption>
    <pre tabIndex={0} aria-label={`Contoh ${filename}`}><code>{code.split('\n').map((line, index, lines) => {
      const match = line.match(/^(\s*)(Feature:|Scenario:|Given\b|When\b|Then\b|And\b)(.*)$/)
      return <span key={index}>{match ? <>{match[1]}<span className={match[2].endsWith(':') ? 'docs-code-heading' : 'docs-code-step'}>{match[2]}</span>{match[3]}</> : line}{index < lines.length - 1 ? '\n' : ''}</span>
    })}</code></pre>
    {copyState !== 'idle' && <span className="docs-copy-status" role="status">{copyState === 'copied' ? 'Contoh berhasil disalin.' : 'Tidak dapat menyalin otomatis. Pilih dan salin teks contoh secara manual.'}</span>}
  </figure>
}
