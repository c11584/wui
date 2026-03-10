import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface SystemState {
  mode: 'admin' | 'agent'
  setMode: (mode: 'admin' | 'agent') => void
}

export const useSystemStore = create<SystemState>()(
  persist(
    (set) => ({
      mode: 'admin',
      setMode: (mode) => set({ mode }),
    }),
    {
      name: 'system-storage',
    }
  )
)
