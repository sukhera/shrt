"use client"

import {
  createContext,
  useContext,
  useEffect,
  useReducer,
  useCallback,
  type ReactNode,
} from "react"

import { auth } from "@/lib/api"
import { setAccessToken } from "@/lib/auth"
import type { Credentials, User } from "@/types/api"

type Status = "loading" | "authenticated" | "anonymous"

interface AuthState {
  status: Status
  user: User | null
}

type Action =
  | { type: "authenticated"; user: User | null }
  | { type: "anonymous" }

function reducer(state: AuthState, action: Action): AuthState {
  switch (action.type) {
    case "authenticated":
      return { status: "authenticated", user: action.user }
    case "anonymous":
      return { status: "anonymous", user: null }
    default:
      return state
  }
}

interface AuthContextValue extends AuthState {
  login: (creds: Credentials) => Promise<void>
  register: (creds: Credentials) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, { status: "loading", user: null })

  // On mount, attempt a silent refresh to restore a session from the httpOnly
  // refresh cookie. Success populates the in-memory access token via the api
  // layer; failure leaves the user anonymous.
  useEffect(() => {
    let active = true
    auth
      .refresh()
      .then((token) => {
        if (!active) return
        dispatch(token ? { type: "authenticated", user: null } : { type: "anonymous" })
      })
      .catch(() => {
        if (active) dispatch({ type: "anonymous" })
      })
    return () => {
      active = false
    }
  }, [])

  const login = useCallback(async (creds: Credentials) => {
    const { access_token } = await auth.login(creds)
    setAccessToken(access_token)
    dispatch({ type: "authenticated", user: { id: "", email: creds.email } })
  }, [])

  const register = useCallback(async (creds: Credentials) => {
    const { user, access_token } = await auth.register(creds)
    setAccessToken(access_token)
    dispatch({ type: "authenticated", user })
  }, [])

  const logout = useCallback(async () => {
    await auth.logout()
    dispatch({ type: "anonymous" })
  }, [])

  const value: AuthContextValue = { ...state, login, register, logout }
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

/** Access auth state and actions. Throws if used outside AuthProvider. */
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return ctx
}
