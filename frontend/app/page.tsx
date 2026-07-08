import { ShortenForm } from "@/components/app/shorten-form"

export default function Home() {
  return (
    <main className="mx-auto flex min-h-[calc(100vh-3.5rem)] max-w-2xl flex-col items-center justify-center px-4">
      {/* Hero: wordmark + pitch + form */}
      <div className="mb-8 text-center">
        <h1 className="font-mono text-5xl font-bold tracking-tight">
          <span>s</span>
          <span className="text-primary">/</span>
          <span>hrt</span>
        </h1>
        <p className="mt-3 text-dim">
          Shorten any URL in one click.
        </p>
      </div>

      <ShortenForm />

      {/* Sub-fold: three terse feature lines */}
      <div className="mt-16 grid w-full max-w-md gap-4 text-center sm:grid-cols-3">
        <div>
          <span className="font-mono text-lg text-primary">&#x21C9;</span>
          <p className="mt-1 text-sm text-dim">Instant redirects</p>
        </div>
        <div>
          <span className="font-mono text-lg text-primary">&#x23F1;</span>
          <p className="mt-1 text-sm text-dim">Optional expiry</p>
        </div>
        <div>
          <span className="font-mono text-lg text-primary">&#x2318;</span>
          <p className="mt-1 text-sm text-dim">Custom aliases</p>
        </div>
      </div>

      {/* Sign-in nudge */}
      <p className="mt-8 text-sm text-faint">
        <a href="/login" className="text-primary hover:underline">Sign in</a>
        {" "}to manage and track your links.
      </p>
    </main>
  )
}
