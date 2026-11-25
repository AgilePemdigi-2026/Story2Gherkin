import { useLayoutEffect, useRef } from 'react'

function resizeTextarea(element: HTMLTextAreaElement | null) {
  if (!element) return

  // Reset before measuring so the field can also shrink.
  element.style.height = 'auto'

  const styles = window.getComputedStyle(element)

  const borderHeight =
    (Number.parseFloat(styles.borderTopWidth) || 0) +
    (Number.parseFloat(styles.borderBottomWidth) || 0)

  const minimumHeight =
    Number.parseFloat(styles.minHeight) || 0

  const nextHeight = Math.max(
    minimumHeight,
    element.scrollHeight + borderHeight,
  )

  element.style.height = `${Math.ceil(nextHeight)}px`
}

export function useAutoGrowTextarea(value: string) {
  const ref = useRef<HTMLTextAreaElement>(null)

  useLayoutEffect(() => {
    resizeTextarea(ref.current)
  }, [value])

  useLayoutEffect(() => {
    const element = ref.current

    if (!element) return

    let disposed = false
    let previousWidth = element.getBoundingClientRect().width

    const updateSize = () => {
      if (!disposed) {
        resizeTextarea(element)
      }
    }

    const observer =
      typeof ResizeObserver !== 'undefined'
        ? new ResizeObserver(() => {
            const nextWidth = element.getBoundingClientRect().width

            // Height changes should not trigger another resize cycle.
            if (Math.abs(nextWidth - previousWidth) < 0.5) return

            previousWidth = nextWidth
            updateSize()
          })
        : null

    observer?.observe(element)

    window.addEventListener('resize', updateSize)

    // Recalculate after web fonts finish loading.
    void document.fonts.ready.then(updateSize)

    return () => {
      disposed = true
      observer?.disconnect()
      window.removeEventListener('resize', updateSize)
    }
  }, [])

  return ref
}