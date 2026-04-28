// Providers wrap JSON-RPC errors through several layers (the outer
// thrown Error, an `info.error` carrier, the underlying `error` object).
// extractCode / extractMessage drill through to find the original
// {code, message} regardless of where it landed so suites can assert
// against spec-reserved JSON-RPC error codes without coupling to the
// provider's wrapping shape.

export type CapturedError = {
  thrown: boolean
  code: number | undefined
  message: string
}

export function extractCode(err: any): number | undefined {
  if (!err) return undefined
  if (typeof err.code === "number") return err.code
  if (err.error && typeof err.error.code === "number") return err.error.code
  if (err.info && err.info.error && typeof err.info.error.code === "number") return err.info.error.code
  if (err.data && typeof err.data.code === "number") return err.data.code
  return undefined
}

export function extractMessage(err: any): string {
  if (!err) return ""
  if (err.error && typeof err.error.message === "string") return err.error.message
  if (err.info && err.info.error && typeof err.info.error.message === "string") return err.info.error.message
  if (typeof err.message === "string") return err.message
  return String(err)
}
