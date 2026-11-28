const ENCRYPTION_KEY = 'ArtacomFTTH_2025_Secure_Key_Hashing'

export function encrypt(text: string): string {
  if (!text) return ''
  try {
    let encrypted = ''
    for (let i = 0; i < text.length; i++) {
      encrypted += String.fromCharCode(
        text.charCodeAt(i) ^ ENCRYPTION_KEY.charCodeAt(i % ENCRYPTION_KEY.length)
      )
    }
    return btoa(encrypted)
  } catch (error) {
    console.error('Encryption failed:', error)
    return text
  }
}

export function decrypt(encryptedText: string): string {
  if (!encryptedText) return ''
  try {
    if (encryptedText.startsWith('eyJ')) return encryptedText
    if (encryptedText.startsWith('gAAAAA')) {
      console.warn('Fernet encryption detected in frontend - this should be handled by backend')
      return encryptedText
    }
    if (!isValidBase64(encryptedText)) return encryptedText

    const decoded = atob(encryptedText)
    let decrypted = ''
    for (let i = 0; i < decoded.length; i++) {
      decrypted += String.fromCharCode(
        decoded.charCodeAt(i) ^ ENCRYPTION_KEY.charCodeAt(i % ENCRYPTION_KEY.length)
      )
    }
    return decrypted
  } catch (error) {
    console.error('Decryption failed:', error)
    return encryptedText
  }
}

function isValidBase64(str: string): boolean {
  try {
    const base64Regex = /^[A-Za-z0-9+/]*={0,2}$/
    if (!base64Regex.test(str) || str.length % 4 !== 0) return false
    atob(str)
    return true
  } catch {
    return false
  }
}

export function setEncryptedToken(key: string, token: string): void {
  if (!token) {
    localStorage.removeItem(key)
    return
  }
  try {
    localStorage.setItem(key, encrypt(token))
  } catch (error) {
    console.error('Failed to encrypt and store token:', error)
    localStorage.setItem(key, token)
  }
}

export function getEncryptedToken(key: string): string | null {
  try {
    const encrypted = localStorage.getItem(key)
    if (!encrypted) return null
    if (encrypted.startsWith('eyJ')) return encrypted
    if (encrypted.startsWith('gAAAAA')) {
      console.warn('Fernet encrypted token found in localStorage - this should not happen')
      return null
    }
    if (isValidBase64(encrypted)) {
      try {
        const decrypted = decrypt(encrypted)
        if (decrypted.startsWith('eyJ')) return decrypted
        return encrypted
      } catch (decryptError) {
        console.warn('Failed to decrypt token, treating as plain:', decryptError)
        return encrypted
      }
    }
    return encrypted
  } catch (error) {
    console.error('Failed to retrieve and decrypt token:', error)
    return localStorage.getItem(key)
  }
}

export function removeEncryptedToken(key: string): void {
  localStorage.removeItem(key)
}
