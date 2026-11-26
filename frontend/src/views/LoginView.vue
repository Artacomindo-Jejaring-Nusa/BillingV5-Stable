<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import type { AxiosError } from 'axios';
import apiClient from '@/services/api';

interface ErrorResponse {
  detail?: string;
  message?: string;
}

const email = ref('');
const password = ref('');
const resetEmail = ref('');
const newPassword = ref('');
const resetToken = ref('');
const error = ref('');
const successMessage = ref('');
const loading = ref(false);
const showForgotPassword = ref(false);
const rememberMe = ref(false);
const showPassword = ref(false);

const router = useRouter();
const authStore = useAuthStore();

onMounted(() => {
  const rememberedEmail = localStorage.getItem('remember_email');
  if (rememberedEmail) {
    email.value = rememberedEmail;
    rememberMe.value = true;
  }
});

function togglePasswordVisibility() {
  showPassword.value = !showPassword.value;
}

function handleImageError(event: Event) {
  const target = event.target as HTMLImageElement;
  target.src = '/src/assets/Jelantik 1.svg';
}

async function handleLogin() {
  if (!email.value || !password.value) {
    error.value = 'Email dan kata sandi harus diisi';
    return;
  }

  error.value = '';
  loading.value = true;

  try {
    const success = await authStore.login(email.value, password.value);

    if (success) {
      if (rememberMe.value) {
        localStorage.setItem('remember_email', email.value);
      } else {
        localStorage.removeItem('remember_email');
      }
      router.push('/dashboard');
    } else {
      error.value = 'Email atau kata sandi salah';
    }
  } catch (err) {
    const errorResponse = err as AxiosError<ErrorResponse>;
    console.error('Login error:', errorResponse.response?.data || errorResponse.message);
    error.value =
      errorResponse.response?.data?.detail ||
      errorResponse.response?.data?.message ||
      'Terjadi kesalahan saat login';
  } finally {
    loading.value = false;
  }
}

async function handleResetPassword() {
  if (!resetEmail.value || !newPassword.value || !resetToken.value) {
    error.value = 'Email, kata sandi baru, dan token harus diisi';
    return;
  }

  error.value = '';
  successMessage.value = '';
  loading.value = true;

  try {
    const response = await apiClient.post<{ message: string }>(
      '/users/reset-password',
      {
        email: resetEmail.value,
        new_password: newPassword.value,
        token: resetToken.value,
      },
      { headers: { 'Content-Type': 'application/json' } }
    );
    successMessage.value = response.data.message;
    showForgotPassword.value = false;
    setTimeout(() => router.push('/login'), 2000);
  } catch (err) {
    const errorResponse = err as AxiosError<ErrorResponse>;
    console.error('Reset password error:', errorResponse.response?.data || errorResponse.message);
    error.value =
      errorResponse.response?.data?.detail ||
      errorResponse.response?.data?.message ||
      'Terjadi kesalahan saat mengatur ulang kata sandi';
  } finally {
    loading.value = false;
  }
}

function openForgotPassword() {
  error.value = '';
  successMessage.value = '';
  showForgotPassword.value = true;
}

function backToLogin() {
  showForgotPassword.value = false;
  error.value = '';
  successMessage.value = '';
}
</script>

<template>
  <div class="login-page">
    <!-- Tablet Top Bar (640px - 1023px) -->
    <div class="tablet-topbar">
      <span>JELANTIK / PORTAL FTTH &amp; BILLING</span>
      <span>AJNUSA V5.0</span>
    </div>

    <!-- Mobile Top Brand Header (< 640px) -->
    <div class="mobile-brand-header">
      <div class="mobile-header-top">
        <img
          src="/src/assets/icon_light.ico"
          alt="Jelantik Logo"
          class="mobile-logo-mark"
          width="48"
          height="48"
          @error="handleImageError"
        />
        <span class="version-badge">AJNUSA V5.0</span>
      </div>
      <h2 class="mobile-welcome-title">WELCOME!</h2>
      <p class="mobile-welcome-subtitle">Portal FTTH &amp; Billing Ajnusa V5.0</p>
    </div>

    <!-- Left Panel: Sign-in Form -->
    <div class="login-panel">
      <div class="login-panel-inner">
        <div class="sign-in-content">
          <!-- Brand Logo -->
          <div class="jelantik-logo">
            <img
              src="/src/assets/icon_dark.ico"
              alt="Jelantik Logo"
              class="logo-mark"
              fetchpriority="high"
              width="72"
              height="72"
              @error="handleImageError"
            />
          </div>

          <!-- Login Form -->
          <form v-if="!showForgotPassword" @submit.prevent="handleLogin" class="login-form" novalidate>
            <div class="welcome-message">
              <h1 class="welcome-title">Selamat Datang Kembali</h1>
              <p class="welcome-subtitle">Silakan login untuk mengakses portal Anda</p>
            </div>

            <div class="credentials-group">
              <!-- Alamat Email -->
              <div class="field-group">
                <label for="login-email" class="field-label">Alamat Email</label>
                <div class="input-control" :class="{ 'is-error': (email && !email.includes('@')) || Boolean(error), 'is-disabled': loading }">
                  <v-icon size="18" class="input-icon">mdi-email-outline</v-icon>
                  <input
                    id="login-email"
                    type="email"
                    v-model="email"
                    placeholder="Masukkan alamat email Anda"
                    :disabled="loading"
                    autocomplete="email"
                    class="input-element"
                  />
                </div>
                <p v-if="email && !email.includes('@')" class="field-error-text">
                  Mohon masukkan alamat email yang valid
                </p>
              </div>

              <!-- Kata Sandi -->
              <div class="field-group">
                <label for="login-password" class="field-label">Kata Sandi</label>
                <div class="input-control" :class="{ 'is-error': (password.length > 0 && password.length < 6) || Boolean(error), 'is-disabled': loading }">
                  <v-icon size="18" class="input-icon">mdi-lock-outline</v-icon>
                  <input
                    id="login-password"
                    :type="showPassword ? 'text' : 'password'"
                    v-model="password"
                    placeholder="Masukkan kata sandi Anda"
                    :disabled="loading"
                    autocomplete="current-password"
                    class="input-element"
                  />
                  <button
                    type="button"
                    @click="togglePasswordVisibility"
                    class="toggle-password-btn"
                    :aria-label="showPassword ? 'Sembunyikan kata sandi' : 'Lihat kata sandi'"
                  >
                    <v-icon size="18">{{ showPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline' }}</v-icon>
                  </button>
                </div>
                <p v-if="password.length > 0 && password.length < 6" class="field-error-text">
                  Kata sandi minimal 6 karakter
                </p>
              </div>
            </div>

            <!-- Persistent Error Banner -->
            <div v-if="error" class="error-banner" role="alert">
              <v-icon size="18" color="#dc2626" class="error-banner-icon">mdi-alert-circle-outline</v-icon>
              <div>
                <div class="error-banner-title">{{ error }}</div>
                <div class="error-banner-subtitle">Periksa kembali data Anda dan coba lagi.</div>
              </div>
            </div>

            <!-- Actions -->
            <div class="sign-in-actions">
              <div class="remember-recovery">
                <label class="remember-label">
                  <input
                    type="checkbox"
                    v-model="rememberMe"
                    :disabled="loading"
                    class="remember-checkbox"
                  />
                  <span>Ingat Saya</span>
                </label>
                <button type="button" @click="openForgotPassword" class="forgot-link">
                  Lupa kata sandi?
                </button>
              </div>

              <button type="submit" :disabled="loading" class="primary-btn">
                <v-progress-circular
                  v-if="loading"
                  indeterminate
                  size="18"
                  width="2"
                  color="white"
                  class="mr-2"
                />
                <span>{{ loading ? 'MEMPROSES...' : 'MASUK' }}</span>
              </button>
            </div>

            <div class="secure-sign-in">
              <v-icon size="16" color="#64748b">mdi-shield-check-outline</v-icon>
              <span>Akses aman ke portal Anda</span>
            </div>
          </form>

          <!-- Reset Password Form -->
          <form v-else @submit.prevent="handleResetPassword" class="login-form" novalidate>
            <button type="button" @click="backToLogin" class="back-btn">
              <v-icon size="18">mdi-chevron-left</v-icon>
              <span>Kembali ke login</span>
            </button>

            <div class="welcome-message">
              <h1 class="welcome-title">Atur Ulang Kata Sandi</h1>
              <p class="welcome-subtitle">Masukkan email, token reset, dan kata sandi baru Anda</p>
            </div>

            <div class="credentials-group">
              <div class="field-group">
                <label for="reset-email" class="field-label">Alamat Email</label>
                <div class="input-control">
                  <v-icon size="18" class="input-icon">mdi-email-outline</v-icon>
                  <input
                    id="reset-email"
                    type="email"
                    v-model="resetEmail"
                    placeholder="Masukkan alamat email Anda"
                    :disabled="loading"
                    class="input-element"
                    required
                  />
                </div>
              </div>

              <div class="field-group">
                <label for="reset-password" class="field-label">Kata Sandi Baru</label>
                <div class="input-control">
                  <v-icon size="18" class="input-icon">mdi-lock-outline</v-icon>
                  <input
                    id="reset-password"
                    type="password"
                    v-model="newPassword"
                    placeholder="Masukkan kata sandi baru"
                    :disabled="loading"
                    class="input-element"
                    required
                  />
                </div>
              </div>

              <div class="field-group">
                <label for="reset-token" class="field-label">Token Reset</label>
                <div class="input-control">
                  <v-icon size="18" class="input-icon">mdi-shield-key-outline</v-icon>
                  <input
                    id="reset-token"
                    type="text"
                    v-model="resetToken"
                    placeholder="Masukkan token reset"
                    :disabled="loading"
                    class="input-element"
                    required
                  />
                </div>
              </div>
            </div>

            <div v-if="error" class="error-banner" role="alert">
              <v-icon size="18" color="#dc2626" class="error-banner-icon">mdi-alert-circle-outline</v-icon>
              <div class="error-banner-title">{{ error }}</div>
            </div>

            <div v-if="successMessage" class="success-banner" role="status">
              <v-icon size="18" color="#16a34a" class="error-banner-icon">mdi-check-circle-outline</v-icon>
              <div class="success-banner-title">{{ successMessage }}</div>
            </div>

            <button type="submit" :disabled="loading" class="primary-btn">
              <v-progress-circular
                v-if="loading"
                indeterminate
                size="18"
                width="2"
                color="white"
                class="mr-2"
              />
              <span>{{ loading ? 'MEMPROSES...' : 'ATUR ULANG KATA SANDI' }}</span>
            </button>
          </form>
        </div>

        <!-- Legal Footer -->
        <footer class="legal-footer">
          <div class="legal-links">
            <a href="https://www.ajnusa.com" target="_blank" rel="noopener noreferrer">www.ajnusa.com</a>
            <span>|</span>
            <a href="https://www.jelantik.com" target="_blank" rel="noopener noreferrer">www.jelantik.com</a>
            <span>|</span>
            <router-link to="/privacy-policy">Kebijakan Privasi</router-link>
          </div>
          <!-- <p class="legal-copyright">© PT. Artacomindo Jejaring Nusa</p> -->
        </footer>
      </div>
    </div>

    <!-- Tablet Bottom Tagline (640px - 1023px) -->
    <p class="tablet-tagline">Koneksi yang menyatukan.</p>

    <!-- Right Panel: Brand & FTTH Network Illustration -->
    <div class="brand-panel">
      <div class="brand-panel-content">
        <!-- Masthead -->
        <div class="brand-masthead">
          <span class="masthead-label">AJNUSA / CONNECTED SERVICES</span>
          <span class="version-badge">V5.0</span>
        </div>

        <!-- Welcome Heading -->
        <div class="brand-welcome">
          <h2 class="brand-title">WELCOME!</h2>
          <p class="brand-subtitle">Portal FTTH &amp; Billing Ajnusa V5.0</p>
        </div>

        <!-- Capabilities -->
        <div class="portal-capabilities">
          <div class="capability-chip">
            <v-icon size="15" color="#2bb5f0">mdi-chart-timeline-variant</v-icon>
            <span>Monitoring Jaringan</span>
          </div>
          <div class="capability-chip">
            <v-icon size="15" color="#2bb5f0">mdi-receipt-text-outline</v-icon>
            <span>Billing &amp; Tagihan</span>
          </div>
          <div class="capability-chip">
            <v-icon size="15" color="#2bb5f0">mdi-account-group-outline</v-icon>
            <span>Manajemen Pelanggan</span>
          </div>
        </div>

        <!-- FTTH Network Illustration -->
        <div class="network-illustration" aria-hidden="true">
          <svg viewBox="0 0 520 260" fill="none" class="network-svg">
            <!-- Fiber Connectors -->
            <path d="M260 112 L120 185" stroke="rgba(232, 245, 255, 0.45)" stroke-width="2" stroke-dasharray="5 5" />
            <path d="M260 112 L400 185" stroke="rgba(232, 245, 255, 0.45)" stroke-width="2" stroke-dasharray="5 5" />
            <path d="M260 112 L260 212" stroke="rgba(232, 245, 255, 0.45)" stroke-width="2" stroke-dasharray="5 5" />
            <circle cx="190" cy="148" r="4" fill="#ff6b3d" />
            <circle cx="330" cy="148" r="4" fill="#ff6b3d" />
            <circle cx="260" cy="166" r="4" fill="#ff6b3d" />

            <!-- Central OLT / Hub -->
            <g transform="translate(204, 24)">
              <path d="M36 16 Q56 0 76 16" stroke="#2bb5f0" stroke-width="2.5" stroke-linecap="round" />
              <path d="M44 24 Q56 14 68 24" stroke="#ffffff" stroke-width="2.5" stroke-linecap="round" />
              <rect x="4" y="32" width="104" height="64" rx="10" fill="#ffffff" />
              <rect x="14" y="44" width="84" height="12" rx="4" fill="#e8f5ff" />
              <rect x="14" y="64" width="52" height="10" rx="4" fill="#e8f5ff" />
              <circle cx="88" cy="69" r="4" fill="#ff6b3d" />
              <rect x="18" y="82" width="12" height="6" rx="2" fill="#0b1f6b" />
              <rect x="36" y="82" width="12" height="6" rx="2" fill="#0b1f6b" />
            </g>

            <!-- Left Home Node -->
            <g transform="translate(83, 168)">
              <path d="M37 4 L4 26 L10 26 L10 66 L64 66 L64 26 L70 26 Z" fill="#e8f5ff" />
              <path d="M37 0 L0 25 L74 25 Z" fill="#ffffff" />
              <rect x="28" y="40" width="18" height="26" rx="2" fill="#0f2a8f" />
              <circle cx="37" cy="31" r="5" fill="#2bb5f0" />
            </g>

            <!-- Bottom Center Home Node -->
            <g transform="translate(231, 196)">
              <path d="M29 4 L4 20 L8 20 L8 52 L50 52 L50 20 L54 20 Z" fill="#e8f5ff" />
              <path d="M29 0 L0 19 L58 19 Z" fill="#ffffff" />
              <rect x="22" y="32" width="14" height="20" rx="2" fill="#0f2a8f" />
              <circle cx="29" cy="24" r="4" fill="#2bb5f0" />
            </g>

            <!-- Right Home Node -->
            <g transform="translate(363, 168)">
              <path d="M37 4 L4 26 L10 26 L10 66 L64 66 L64 26 L70 26 Z" fill="#e8f5ff" />
              <path d="M37 0 L0 25 L74 25 Z" fill="#ffffff" />
              <rect x="28" y="40" width="18" height="26" rx="2" fill="#0f2a8f" />
              <circle cx="37" cy="31" r="5" fill="#2bb5f0" />
            </g>
          </svg>
        </div>

        <!-- Brand Signature -->
        <div class="brand-signature">
          <p class="signature-tagline">Koneksi yang menyatukan.</p>
          <p class="signature-company">PT. Artacomindo Jejaring Nusa</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

@media (min-width: 1024px) {
  .login-page {
    flex-direction: row;
  }
}

/* Mobile Header (< 640px) */
.mobile-brand-header {
  display: block;
  background: linear-gradient(145deg, #0b1f6b 0%, #0f2a8f 55%, #1e4fd8 100%);
  padding: 24px 20px 28px;
  color: #ffffff;
}

.mobile-header-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.mobile-logo-mark {
  width: 48px;
  height: 48px;
  object-fit: contain;
}

.mobile-welcome-title {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.15;
  margin: 0 0 4px;
}

.mobile-welcome-subtitle {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.9);
  margin: 0;
}

.tablet-topbar,
.tablet-tagline {
  display: none;
}

/* Left Panel */
.login-panel {
  width: 100%;
  flex: 1;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  padding: 24px 20px;
}

@media (max-width: 639px) {
  .jelantik-logo {
    display: none;
  }
}

@media (min-width: 640px) and (max-width: 1023px) {
  .mobile-brand-header {
    display: none;
  }

  .login-page {
    background: linear-gradient(145deg, #0b1f6b 0%, #0f2a8f 48%, #1e4fd8 100%);
    align-items: center;
    justify-content: space-between;
    padding: 28px 32px;
  }

  .tablet-topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.08em;
    color: rgba(255, 255, 255, 0.85);
  }

  .tablet-tagline {
    display: block;
    font-size: 13px;
    color: rgba(255, 255, 255, 0.85);
    margin: 0;
  }

  .login-panel {
    flex: 0 0 auto;
    width: 100%;
    max-width: 480px;
    border-radius: 20px;
    box-shadow: 0 24px 48px rgba(0, 0, 0, 0.22);
    padding: 36px 36px 28px;
    margin: 24px 0;
  }
}

@media (min-width: 1024px) {
  .mobile-brand-header {
    display: none;
  }

  .login-panel {
    width: 50%;
    padding: 40px 48px;
  }
}

.login-panel-inner {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.sign-in-content {
  width: 100%;
  max-width: 400px;
  margin: auto 0;
  padding: 16px 0;
}

.jelantik-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 32px;
}

.logo-mark {
  width: 72px;
  height: 72px;
  object-fit: contain;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.welcome-message {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.welcome-title {
  font-size: 28px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.25;
  margin: 0;
}

.welcome-subtitle {
  font-size: 14px;
  font-weight: 400;
  color: #64748b;
  margin: 0;
}

.credentials-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 14px;
  font-weight: 500;
  color: #0f172a;
}

.input-control {
  display: flex;
  align-items: center;
  height: 48px;
  padding: 0 14px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.input-control:focus-within {
  border-color: #1e4fd8;
  box-shadow: 0 0 0 3px rgba(30, 79, 216, 0.12);
}

.input-control.is-error {
  border-color: #ef4444;
}

.input-control.is-error:focus-within {
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.12);
}

.input-control.is-disabled {
  background: #f8fafc;
  color: #94a3b8;
}

.input-icon {
  color: #64748b;
  margin-right: 10px;
  flex-shrink: 0;
}

.input-control:focus-within .input-icon {
  color: #1e4fd8;
}

.input-element {
  flex: 1;
  width: 100%;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
  color: #0f172a;
}

.input-element::placeholder {
  color: #94a3b8;
}

.toggle-password-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  min-height: 32px;
  color: #64748b;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0;
}

.toggle-password-btn:hover {
  color: #0f172a;
}

.field-error-text {
  font-size: 12px;
  color: #dc2626;
  margin: 0;
}

.error-banner {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 14px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
}

.error-banner-icon {
  flex-shrink: 0;
  margin-top: 1px;
}

.error-banner-title {
  font-size: 13px;
  font-weight: 600;
  color: #dc2626;
}

.error-banner-subtitle {
  font-size: 12px;
  color: #b91c1c;
  margin-top: 2px;
}

.success-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
}

.success-banner-title {
  font-size: 13px;
  font-weight: 600;
  color: #15803d;
}

.sign-in-actions {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.remember-recovery {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.remember-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #0f172a;
  cursor: pointer;
}

.remember-checkbox {
  width: 16px;
  height: 16px;
  accent-color: #1e4fd8;
  cursor: pointer;
}

.forgot-link {
  font-size: 14px;
  font-weight: 600;
  color: #1e4fd8;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 4px 0;
}

.forgot-link:hover {
  color: #0f2a8f;
  text-decoration: underline;
}

.primary-btn {
  width: 100%;
  height: 48px;
  border-radius: 8px;
  background: #1e4fd8;
  color: #ffffff;
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.04em;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.15s ease;
}

.primary-btn:hover:not(:disabled) {
  background: #0f2a8f;
}

.primary-btn:disabled {
  background: #94a3b8;
  cursor: not-allowed;
}

.secure-sign-in {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #64748b;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  font-weight: 500;
  color: #64748b;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0;
  align-self: flex-start;
}

.back-btn:hover {
  color: #0f172a;
}

/* Legal Footer */
.legal-footer {
  width: 100%;
  max-width: 640px;
  text-align: center;
  padding-top: 24px;
}

.legal-links {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #64748b;
}

.legal-links a {
  color: #64748b;
  text-decoration: none;
}

.legal-links a:hover {
  color: #1e4fd8;
  text-decoration: underline;
}

.legal-copyright {
  font-size: 12px;
  color: #64748b;
  margin: 4px 0 0;
}

/* Right Panel: Brand */
.brand-panel {
  display: none;
}

@media (min-width: 1024px) {
  .brand-panel {
    display: flex;
    width: 50%;
    position: relative;
    overflow: hidden;
    background: linear-gradient(145deg, #0b1f6b 0%, #0f2a8f 48%, #1e4fd8 100%);
    padding: 40px 56px;
    color: #ffffff;
  }
}

.brand-panel-content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 100%;
}

.brand-masthead {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.masthead-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: rgba(255, 255, 255, 0.85);
}

.version-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 28px;
  padding: 0 14px;
  background: rgba(15, 42, 143, 0.85);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  font-size: 11px;
  font-weight: 600;
  color: #ffffff;
}

.brand-welcome {
  margin-top: 24px;
}

.brand-eyebrow {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.1em;
  color: #2bb5f0;
  margin: 0 0 8px;
}

.brand-title {
  font-size: 56px;
  font-weight: 700;
  line-height: 1.1;
  letter-spacing: -0.02em;
  margin: 0 0 8px;
}

.brand-subtitle {
  font-size: 21px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
  margin: 0;
}

.portal-capabilities {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.capability-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 14px;
  background: rgba(15, 42, 143, 0.72);
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 8px;
  font-size: 12px;
  font-weight: 500;
  color: #ffffff;
}

.network-illustration {
  width: 100%;
  max-width: 520px;
  margin: 8px auto;
}

.network-svg {
  width: 100%;
  height: auto;
  display: block;
}

.brand-signature {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.signature-tagline {
  font-size: 22px;
  font-weight: 600;
  margin: 0;
}

.signature-company {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.75);
  margin: 0;
}
</style>
