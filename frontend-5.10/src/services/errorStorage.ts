export interface SecureErrorDetails {
  id: string
  timestamp: number
  errorType: 'network' | 'server' | 'timeout' | 'unknown'
  errorStatus?: number
  url?: string
  method?: string
  message?: string
  userAgent: string
}

class ErrorStorageService {
  private readonly STORAGE_KEY = 'network_error_details'
  private readonly MAX_AGE = 5 * 60 * 1000
  private readonly MAX_ERRORS = 10

  storeError(
    errorType: string,
    errorStatus: number | string,
    url?: string,
    method?: string,
    message?: string
  ): string {
    try {
      const errorId = `ERR-${Date.now().toString(36).toUpperCase()}`
      const errorDetails: SecureErrorDetails = {
        id: errorId,
        timestamp: Date.now(),
        errorType: errorType as SecureErrorDetails['errorType'],
        errorStatus: typeof errorStatus === 'number' ? errorStatus : 0,
        url: this.sanitizeUrl(url || ''),
        method: method?.toUpperCase(),
        message: this.sanitizeMessage(message || ''),
        userAgent: navigator.userAgent,
      }

      const existingErrors = this.getStoredErrors()
      existingErrors.push(errorDetails)
      this.cleanExpiredErrors(existingErrors)
      this.limitErrors(existingErrors)
      sessionStorage.setItem(this.STORAGE_KEY, JSON.stringify(existingErrors))
      return errorId
    } catch (error) {
      console.warn('Failed to store error details:', error)
      return 'ERR-UNKNOWN'
    }
  }

  getError(errorId: string): SecureErrorDetails | null {
    try {
      const errors = this.getStoredErrors()
      const error = errors.find((e) => e.id === errorId)
      if (error) {
        this.removeError(errorId)
        return error
      }
      return null
    } catch (error) {
      console.warn('Failed to retrieve error details:', error)
      return null
    }
  }

  private sanitizeUrl(url: string): string {
    try {
      const urlObj = new URL(url, window.location.origin)
      const sensitivePaths = ['/auth/', '/api/', '/admin/', '/users/', '/login', '/token', '/password', '/reset']
      for (const sensitivePath of sensitivePaths) {
        if (urlObj.pathname.includes(sensitivePath)) return '[REDACTED]'
      }
      return urlObj.pathname
    } catch {
      return '[REDACTED]'
    }
  }

  private sanitizeMessage(message: string): string {
    if (!message) return ''
    let sanitized = message
      .replace(/https?:\/\/[^\s]+/g, '[REDACTED_URL]')
      .replace(/\/[a-zA-Z0-9_\-/.]+/g, '[REDACTED_PATH]')
      .replace(/[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/g, '[REDACTED_EMAIL]')
      .replace(/password[=:]\S+/gi, 'password=[REDACTED]')
      .replace(/token[=:]\S+/gi, 'token=[REDACTED]')
    if (sanitized.length > 200) sanitized = sanitized.substring(0, 197) + '...'
    return sanitized.trim()
  }

  private getStoredErrors(): SecureErrorDetails[] {
    try {
      const stored = sessionStorage.getItem(this.STORAGE_KEY)
      return stored ? JSON.parse(stored) : []
    } catch {
      return []
    }
  }

  private cleanExpiredErrors(errors: SecureErrorDetails[]): void {
    const now = Date.now()
    for (let i = errors.length - 1; i >= 0; i--) {
      if (now - errors[i].timestamp > this.MAX_AGE) errors.splice(i, 1)
    }
  }

  private limitErrors(errors: SecureErrorDetails[]): void {
    if (errors.length > this.MAX_ERRORS) {
      errors.splice(0, errors.length - this.MAX_ERRORS)
    }
  }

  private removeError(errorId: string): void {
    try {
      const errors = this.getStoredErrors()
      sessionStorage.setItem(this.STORAGE_KEY, JSON.stringify(errors.filter((e) => e.id !== errorId)))
    } catch {
      // ignore
    }
  }

  clearAll(): void {
    try {
      sessionStorage.removeItem(this.STORAGE_KEY)
    } catch {
      // ignore
    }
  }
}

export const errorStorage = new ErrorStorageService()

export class NetworkErrorRateLimit {
  private readonly STORAGE_KEY = 'network_error_rate_limit'
  private readonly MAX_ATTEMPTS = 3
  private readonly WINDOW_MS = 60 * 1000

  canRedirect(): boolean {
    try {
      const data = this.getData()
      const now = Date.now()
      data.attempts = data.attempts.filter((t: number) => now - t <= this.WINDOW_MS)
      if (data.attempts.length >= this.MAX_ATTEMPTS) return false
      data.attempts.push(now)
      data.lastAttempt = now
      sessionStorage.setItem(this.STORAGE_KEY, JSON.stringify(data))
      return true
    } catch {
      return true
    }
  }

  private getData() {
    try {
      const stored = sessionStorage.getItem(this.STORAGE_KEY)
      return stored ? JSON.parse(stored) : { attempts: [], lastAttempt: 0 }
    } catch {
      return { attempts: [], lastAttempt: 0 }
    }
  }
}

export const networkErrorRateLimit = new NetworkErrorRateLimit()
