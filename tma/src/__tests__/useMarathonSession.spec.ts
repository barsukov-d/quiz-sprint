import { describe, it, expect, beforeEach } from 'vitest'
import { useMarathonSession } from '@/composables/useMarathonSession'

describe('useMarathonSession', () => {
  beforeEach(() => {
    // Reset session between tests
    const session = useMarathonSession()
    session.resetSession()
  })

  it('starts with zero runs and zero best', () => {
    const session = useMarathonSession()
    expect(session.runCount.value).toBe(0)
    expect(session.sessionBest.value).toBe(0)
  })

  it('recordRunResult increments runCount', () => {
    const session = useMarathonSession()
    session.recordRunResult(47, 12)
    expect(session.runCount.value).toBe(1)
  })

  it('recordRunResult tracks session best', () => {
    const session = useMarathonSession()
    session.recordRunResult(47, 12)
    session.recordRunResult(23, 5)
    expect(session.sessionBest.value).toBe(47)
  })

  it('motivational prompt mentions deficit to record when behind', () => {
    const session = useMarathonSession()
    session.recordRunResult(75, 10)
    const prompt = session.getMotivationalPrompt(63, 87)
    expect(prompt).toContain('24')  // 87 - 63 = 24
  })

  it('resetSession zeroes all state', () => {
    const session = useMarathonSession()
    session.recordRunResult(47, 12)
    session.resetSession()
    expect(session.runCount.value).toBe(0)
    expect(session.sessionBest.value).toBe(0)
  })

  it('sessionLabel is null before any run', () => {
    const session = useMarathonSession()
    expect(session.sessionLabel.value).toBeNull()
  })

  it('sessionLabel shows run count and best after first run', () => {
    const session = useMarathonSession()
    session.recordRunResult(47, 12)
    expect(session.sessionLabel.value).toContain('47')
  })
})
