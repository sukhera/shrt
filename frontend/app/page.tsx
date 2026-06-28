import { ShortenForm } from "@/components/app/shorten-form"

export default function Home() {
  return (
    <main className="mx-auto flex min-h-[calc(100vh-3.5rem)] max-w-4xl flex-col items-center justify-center px-4">
      <div className="mb-8 text-center">
        <h1 className="font-mono text-4xl font-bold text-primary">shrt</h1>
        <p className="mt-2 text-muted-foreground">
          Shorten any URL in one click. Free and open source.
        </p>
      </div>
      <ShortenForm />
    </main>
  )
}
