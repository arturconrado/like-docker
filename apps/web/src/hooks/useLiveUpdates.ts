import { useEffect, useRef, useState } from 'react'
import type { QueryClient } from '@tanstack/react-query'

import { API_BASE_URL } from '../api/client'

const EVENT_TYPES = [
  'ready',
  'workload_created',
  'workload_started',
  'workload_completed',
  'workload_failed',
  'workload_stopped',
  'workload_removed',
  'workload_flagged',
  'demo_seeded',
]

export type LiveMode = 'connecting' | 'live' | 'polling'

export function useLiveUpdates(queryClient: QueryClient) {
  const [mode, setMode] = useState<LiveMode>('connecting')
  const retryRef = useRef<number | null>(null)

  useEffect(() => {
    let closed = false
    let source: EventSource | null = null

    const refreshData = () => {
      queryClient.invalidateQueries({ queryKey: ['health'] })
      queryClient.invalidateQueries({ queryKey: ['workloads'] })
      queryClient.invalidateQueries({ queryKey: ['events'] })
      queryClient.invalidateQueries({ queryKey: ['summary'] })
    }

    const openConnection = () => {
      if (closed) return
      setMode((current) => (current === 'live' ? current : 'connecting'))

      source = new EventSource(`${API_BASE_URL}/api/stream`)
      source.onopen = () => {
        if (!closed) {
          setMode('live')
        }
      }

      EVENT_TYPES.forEach((eventType) => {
        source?.addEventListener(eventType, refreshData)
      })

      source.onerror = () => {
        source?.close()
        if (closed) return
        setMode('polling')

        if (retryRef.current !== null) {
          window.clearTimeout(retryRef.current)
        }
        retryRef.current = window.setTimeout(() => {
          openConnection()
        }, 5000)
      }
    }

    openConnection()

    return () => {
      closed = true
      source?.close()
      if (retryRef.current !== null) {
        window.clearTimeout(retryRef.current)
      }
    }
  }, [queryClient])

  return mode
}
