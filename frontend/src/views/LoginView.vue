<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import type { AxiosError } from 'axios';
import apiClient from '@/services/api';

// Definisikan tipe untuk respons error dari backend
interface ErrorResponse {
  detail?: string;
  message?: string;
}

const email = ref('');
const password = ref('');
const resetEmail = ref('');
const newPassword = ref('');
const error = ref('');
const successMessage = ref('');
const loading = ref(false);
const showForgotPassword = ref(false);
const router = useRouter();
const authStore = useAuthStore();
const resetToken = ref(''); // Untuk menyimpan token dari forgot password
const rememberMe = ref(false); // Untuk checkbox remember me
const showPassword = ref(false); // Untuk toggle password visibility

// Load remembered email on component mount
onMounted(() => {
  const rememberedEmail = localStorage.getItem('remember_email');
  if (rememberedEmail) {
    email.value = rememberedEmail;
    rememberMe.value = true;
  }
});

// Toggle password visibility
function togglePasswordVisibility() {
  showPassword.value = !showPassword.value;
}

// Handle image error fallback
function handleImageError(event: Event) {
  const target = event.target as HTMLImageElement;
  target.src = '/src/assets/Jelantik 1.svg';
}

async function handleLogin() {
  if (!email.value || !password.value) {
    error.value = 'Email dan password harus diisi';
    return;
  }

  error.value = '';
  loading.value = true;

  try {
    const success = await authStore.login(email.value, password.value);
    if (success) {
      // Store remember me preference if checked
      if (rememberMe.value) {
        localStorage.setItem('remember_email', email.value);
      } else {
        localStorage.removeItem('remember_email');
      }
      router.push('/dashboard');
    } else {
      error.value = 'Email atau password salah!';
    }
  } catch (err) {
    const errorResponse = err as AxiosError<ErrorResponse>;
    console.error('Login error:', errorResponse.response?.data || errorResponse.message);
    error.value = errorResponse.response?.data?.detail || errorResponse.response?.data?.message || 'Terjadi kesalahan saat login';
  } finally {
    loading.value = false;
  }
}

async function handleResetPassword() {
  if (!resetEmail.value || !newPassword.value || !resetToken.value) {
    error.value = 'Email, password baru, dan token harus diisi';
    return;
  }

  error.value = '';
  successMessage.value = '';
  loading.value = true;

  try {
    // Use the configured API client instead of hardcoded URL
    const response = await apiClient.post<{ message: string }>('/users/reset-password', {
      email: resetEmail.value,
      new_password: newPassword.value,
      token: resetToken.value,
    }, {
      headers: { 'Content-Type': 'application/json' },
    });
    successMessage.value = response.data.message;
    showForgotPassword.value = false;
    setTimeout(() => router.push('/login'), 2000);
  } catch (err) {
    const errorResponse = err as AxiosError<ErrorResponse>;
    console.error('Reset password error:', errorResponse.response?.data || errorResponse.message);
    error.value = errorResponse.response?.data?.detail || errorResponse.response?.data?.message || 'Terjadi kesalahan saat reset password';
  } finally {
    loading.value = false;
  }
}

function backToLogin() {
  showForgotPassword.value = false;
  error.value = '';
  successMessage.value = '';
}
</script>

<template>
  <div class="min-h-screen gradient-bg flex flex-col items-center justify-center p-6 md:p-12 relative pb-20">
    <!-- Main Login Card -->
    <div class="w-full max-w-7xl flex overflow-hidden login-card-enhanced relative z-10 mx-auto" style="width: 95vw; max-width: 1200px; border-radius: 1.5rem 1.5rem 0 0; position: relative; overflow: visible;">

      <!-- Login Form - Left Side (LoginForm.js) -->
      <div class="w-full lg:w-[60%] glass-effect px-4 sm:px-6 md:px-8 lg:px-12 py-6 sm:py-8 flex flex-col">
        <div class="w-full max-w-md mx-auto">
          <div class="logo-header mb-4">
            <h2 class="text-lg font-semibold text-black pb-1 border-b-2 border-gray-300 inline-block">
              {{ showForgotPassword ? 'Atur Ulang Kata Sandi' : 'Silakan Login' }}
            </h2>
            <div v-if="!showForgotPassword" class="logo-container">
              <img
                src="/src/assets/icon_dark.ico"
                alt="Jelantik Logo"
                class="h-20 w-auto object-contain"
                fetchpriority="high"
                width="80"
                height="80"
                @error="handleImageError"
              />
            </div>
          </div>

          <!-- Login Form -->
          <form v-if="!showForgotPassword" @submit.prevent="handleLogin" class="mt-10">
            <!-- Email Input -->
            <div class="mb-6">
              <label class="form-label">
                Alamat Email
              </label>
              <div class="input-box-wrapper">
                <div class="input-prefix-icon">
                  <v-icon size="20" color="grey-darken-1">mdi-email-outline</v-icon>
                </div>
                <input
                  type="email"
                  v-model="email"
                  placeholder="Masukkan alamat email Anda"
                  class="custom-text-input"
                />
                <div class="input-suffix-icon pr-1">
                  <div
                    class="w-2 h-2 rounded-full bg-green-500 transition-all duration-200"
                    :class="email.length > 0 ? 'opacity-100' : 'opacity-0'"
                    :style="email.length > 0 ? 'box-shadow: 0 0 8px rgba(34, 197, 94, 0.5);' : ''"
                  ></div>
                </div>
              </div>
              <div v-if="email && !email.includes('@')" class="mt-1 text-xs" style="color: #dc2626 !important; font-weight: 500;">
                Mohon masukkan alamat email yang valid
              </div>
            </div>

            <!-- Password Input -->
            <div class="mb-4">
              <label class="form-label">
                Kata Sandi
              </label>
              <div class="input-box-wrapper">
                <div class="input-prefix-icon">
                  <v-icon size="20" color="grey-darken-1">mdi-lock-outline</v-icon>
                </div>
                <input
                  :type="showPassword ? 'text' : 'password'"
                  v-model="password"
                  placeholder="Masukkan kata sandi Anda"
                  class="custom-text-input"
                />
                <div class="input-suffix-icon">
                  <button
                    type="button"
                    @click="togglePasswordVisibility"
                    class="toggle-password-btn"
                    tabindex="-1"
                    title="Lihat Kata Sandi"
                  >
                    <v-icon size="22" color="grey-darken-2">{{ showPassword ? 'mdi-eye-off' : 'mdi-eye' }}</v-icon>
                  </button>
                </div>
              </div>
              <!-- Password Strength Indicator -->
              <div class="password-strength">
                <div
                  class="password-strength-bar"
                  :style="{ width: password.length > 0 ? Math.min((password.length / 8) * 100, 100) + '%' : '0%' }"
                ></div>
              </div>
              <div v-if="password.length > 0 && password.length < 6" class="mt-1 text-xs" style="color: #dc2626 !important; font-weight: 500;">
                Kata sandi minimal 6 karakter
              </div>
            </div>

            <div class="flex items-center mb-6">
              <input
                type="checkbox"
                v-model="rememberMe"
                :disabled="loading"
                class="w-4 h-4 accent-[var(--primary-color)] cursor-pointer"
                id="remember"
              />
              <label for="remember" class="ml-2 text-sm text-black cursor-pointer">
                Ingat Saya
              </label>
            </div>

            <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
              <div class="flex items-center">
                <v-icon size="20" color="error" class="mr-2">mdi-alert-circle</v-icon>
                {{ error }}
              </div>
            </div>

            <button
              type="submit"
              :disabled="loading"
              class="w-full rounded font-medium flex items-center justify-center hover:opacity-90 transition-all duration-200 disabled:opacity-70 disabled:cursor-not-allowed"
              style="
                background: #0d2691;
                color: white;
                padding: 12px 24px;
                font-size: 16px;
                font-weight: 600;
                border: none;
                min-height: 52px;
                box-shadow: 0 4px 14px rgba(13, 38, 145, 0.3);
              "
            >
              <v-progress-circular v-if="loading" indeterminate size="20" width="2" color="white" class="mr-2"></v-progress-circular>
              {{ loading ? 'SABAR LAGI LOGIN...' : 'MASUK' }}
            </button>
          </form>

          <!-- Reset Password Form -->
          <form v-else @submit.prevent="handleResetPassword" class="mt-10">
            <div class="mb-4">
              <button
                type="button"
                @click="backToLogin"
                class="flex items-center text-[var(--text-secondary)] hover:text-[var(--primary-color)] transition-colors text-sm"
              >
                <v-icon size="18" class="mr-1">mdi-chevron-left</v-icon>
                Kembali ke login
              </button>
            </div>

            <div class="mb-5">
              <div class="flex items-center input-border bg-white rounded px-3 py-2.5 shadow-sm">
                <v-icon size="20" color="grey-darken-1" class="mr-3">mdi-email-outline</v-icon>
                <input
                  type="email"
                  v-model="resetEmail"
                  placeholder="Alamat email Anda"
                  :disabled="loading"
                  class="flex-1 outline-none text-gray-700 bg-transparent placeholder-gray-400 text-sm"
                  required
                />
              </div>
            </div>

            <div class="mb-5">
              <div class="flex items-center input-border bg-white rounded px-3 py-2.5 shadow-sm">
                <v-icon size="20" color="grey-darken-1" class="mr-3">mdi-lock-outline</v-icon>
                <input
                  type="password"
                  v-model="newPassword"
                  placeholder="Kata sandi baru"
                  :disabled="loading"
                  class="flex-1 outline-none text-gray-700 bg-transparent placeholder-gray-400 text-sm"
                  required
                />
              </div>
            </div>

            <div class="mb-5">
              <div class="flex items-center input-border bg-white rounded px-3 py-2.5 shadow-sm">
                <v-icon size="20" color="grey-darken-1" class="mr-3">mdi-shield-key-outline</v-icon>
                <input
                  type="text"
                  v-model="resetToken"
                  placeholder="Token reset"
                  :disabled="loading"
                  class="flex-1 outline-none text-gray-700 bg-transparent placeholder-gray-400 text-sm"
                  required
                />
              </div>
            </div>

            <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
              <div class="flex items-center">
                <v-icon size="20" color="error" class="mr-2">mdi-alert-circle</v-icon>
                {{ error }}
              </div>
            </div>

            <div v-if="successMessage" class="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg text-green-700 text-sm">
              <div class="flex items-center">
                <v-icon size="20" color="success" class="mr-2">mdi-check-circle</v-icon>
                {{ successMessage }}
              </div>
            </div>

            <button
              type="submit"
              :disabled="loading"
              class="w-full bg-green-600 hover:bg-green-700 text-white py-3 rounded font-medium transition-all flex items-center justify-center disabled:opacity-70"
            >
              <v-progress-circular v-if="loading" indeterminate size="20" width="2" color="white" class="mr-2"></v-progress-circular>
              <v-icon v-else size="20" color="white" class="mr-2">mdi-refresh</v-icon>
              <span>{{ loading ? 'Memproses...' : 'Atur Ulang Kata Sandi' }}</span>
            </button>
          </form>

          <!-- Security Notice for Reset Password -->
          <div v-if="showForgotPassword" class="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <div class="flex">
              <v-icon size="20" color="primary" class="mr-2 flex-shrink-0 mt-0.5">mdi-information</v-icon>
              <div class="text-sm text-blue-800">
                <p class="font-medium mb-1">Pemberitahuan Keamanan</p>
                <p class="text-xs">Ingat kata sandi baru Anda dengan baik, karena tidak ada OTP saat login kembali.</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      
      <!-- Welcome Panel - Right Side (WelcomePanel.js) -->
      <div class="hidden lg:flex lg:w-[40%] blue-gradient-bg relative overflow-hidden items-center justify-center">
        <div class="absolute inset-0">
          <div class="absolute -top-32 -right-32 w-[400px] h-[400px] bg-blue-400 rounded-full opacity-30 blur-3xl"></div>
          <div class="absolute top-20 right-20 w-[350px] h-[350px] bg-blue-500 rounded-full opacity-25 blur-3xl"></div>
          <div class="absolute -bottom-40 -right-40 w-[500px] h-[500px] bg-blue-600 rounded-full opacity-30 blur-3xl"></div>
          <div class="absolute bottom-32 right-10 w-[380px] h-[380px] bg-blue-400 rounded-full opacity-20 blur-3xl"></div>
        </div>

        <!-- Floating Logo in Welcome Panel -->
        <div class="welcome-panel-logo">
          <img
            src="/src/assets/icon_light.ico"
            alt="Jelantik Logo"
            @error="handleImageError"
          />
        </div>

        <div class="relative z-10 text-center text-white px-12">
          <h1 class="text-5xl font-bold mb-2 tracking-wide">WELCOME!</h1>
          <p class="text-base opacity-90 max-w-sm mx-auto leading-relaxed">
            PORTAL FTTH & BILLING AJNUSA V5.0
          </p>
        </div>
      </div>

    </div>

    <!-- Credit dan Copyright Section - Menyatu dengan Login Card -->
    <div class="w-full max-w-7xl login-card-enhanced relative z-10 mx-auto" style="width: 95vw; max-width: 1200px; margin-top: -1px; position: relative;">
      <div class="copyright-section-merged bg-white/95 backdrop-blur-xl border border-gray-200/50 shadow-2xl relative" style="border-radius: 0 0 1.5rem 1.5rem; box-shadow: 0 -8px 32px rgba(0, 0, 0, 0.12), 0 -2px 8px rgba(0, 0, 0, 0.08), inset 0 1px 0 rgba(255, 255, 255, 0.9);">

        <!-- Corner fixes to remove sharp edges -->
        <div class="absolute inset-x-0 -top-4 h-12 overflow-hidden pointer-events-none z-30">
          <!-- Left corner cover -->
          <div class="absolute left-0 top-0 w-40 h-40 bg-white rounded-full transform -translate-x-24 -translate-y-20 opacity-100 shadow-2xl" style="box-shadow: 0 -8px 32px rgba(0, 0, 0, 0.15), 0 -4px 16px rgba(0, 0, 0, 0.1);"></div>
          <!-- Right corner cover -->
          <div class="absolute right-0 top-0 w-40 h-40 bg-white rounded-full transform translate-x-24 -translate-y-20 opacity-100 shadow-2xl" style="box-shadow: 0 -8px 32px rgba(0, 0, 0, 0.15), 0 -4px 16px rgba(0, 0, 0, 0.1);"></div>
          <!-- Enhanced gradient overlay -->
          <div class="absolute inset-0 bg-gradient-to-b from-white/95 via-white/80 to-transparent"></div>
        </div>

        <!-- Additional corner smoothing -->
        <div class="absolute -top-2 left-0 right-0 h-4 z-25">
          <div class="absolute left-0 top-0 w-8 h-8 bg-white rounded-full transform -translate-x-4 -translate-y-4"></div>
          <div class="absolute right-0 top-0 w-8 h-8 bg-white rounded-full translate-x-4 -translate-y-4"></div>
        </div>

        <!-- Gradient accent strip -->
        <div class="h-1 bg-gradient-to-r from-[var(--primary-color)] via-blue-500 to-[var(--primary-color)] opacity-80 relative overflow-hidden">
          <!-- Animated shimmer effect -->
          <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 to-transparent shimmer-animation"></div>
        </div>

        <div class="flex items-center justify-center px-6 sm:px-8 md:px-12 lg:px-16 py-5 sm:py-6 space-x-2 sm:space-x-4 md:space-x-6">
          <a
            href="https://www.ajnusa.com"
            target="_blank"
            rel="noopener noreferrer"
            class="text-gray-600 font-bold text-sm sm:text-base py-2 px-3 rounded-lg"
          >
            www.ajnusa.com
          </a>
          <span class="text-gray-600 text-sm sm:text-base">|</span>
          <a
            href="https://www.jelantik.com"
            target="_blank"
            rel="noopener noreferrer"
            class="text-gray-600 font-bold text-sm sm:text-base py-2 px-3 rounded-lg"
          >
            www.jelantik.com
          </a>
              <span class="hidden sm:inline text-gray-600 font-bold text-sm sm:text-base">
               |  © Copyright by PT. Artacomindo Jejaring Nusa
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* CSS Variables from template */
:root {
  --primary-color: #0d2691;
  --secondary-color: #1e40af;
  --accent-color: #3b82f6;
  --text-primary: #374151;
  --text-secondary: #6b7280;
  --bg-light: #f3f4f6;
  --card-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

/* Base Typography */
.min-h-screen {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Template Base Styles */
.gradient-bg {
  background: #f0f2f5;
}

.glass-effect {
  background: #ffffff;
}

.input-border {
  border-left: 3px solid var(--primary-color);
}

.btn-primary {
  background: var(--primary-color) !important;
  color: #ffffff !important;
  height: auto;
  min-height: 48px;
  font-weight: 500;
  font-size: 16px;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  box-shadow: 0 4px 12px rgba(13, 38, 145, 0.3);
  transition: all 0.2s ease;
}

.btn-primary:hover {
  opacity: 0.9 !important;
  box-shadow: 0 6px 16px rgba(13, 38, 145, 0.4);
}

.blue-gradient-bg {
  background: linear-gradient(135deg, #0d47a1 0%, #2196f3 50%, #64b5f6 100%);
}

.card-shadow {
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

/* Enhanced base styles */
body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: #e5e7eb;
}

/* Button visibility fix */
button[type="submit"] {
  position: relative !important;
  z-index: 10 !important;
  opacity: 1 !important;
  visibility: visible !important;
}

/* Button text visibility fix */
button[type="submit"] {
  color: white !important;
}

button[type="submit"] * {
  color: white !important;
}

/* Enhanced Form Styling */
.form-input {
  position: relative;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 12px;
  background: linear-gradient(145deg, #ffffff, #f8fafc);
}

.form-input:hover {
  transform: translateY(-2px);
  box-shadow:
    0 8px 16px rgba(0, 0, 0, 0.12),
    0 4px 8px rgba(0, 0, 0, 0.08),
    inset 0 1px 2px rgba(255, 255, 255, 0.8);
}

.form-input:focus-within {
  transform: translateY(-3px);
  box-shadow:
    0 12px 24px rgba(13, 38, 145, 0.15),
    0 6px 12px rgba(13, 38, 145, 0.1),
    inset 0 1px 2px rgba(255, 255, 255, 0.9);
}

.input-box-wrapper {
  position: relative;
  background: #ffffff;
  border-radius: 12px;
  border: 2px solid #e5e7eb;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex !important;
  align-items: center !important;
  min-height: 52px;
  padding: 0 14px;
  gap: 12px;
  width: 100%;
}

.input-box-wrapper:hover {
  border-color: #cbd5e1;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.input-box-wrapper:focus-within {
  border-color: var(--primary-color, #0d2691) !important;
  box-shadow: 0 0 0 4px rgba(13, 38, 145, 0.12) !important;
  background: #ffffff !important;
}

.input-prefix-icon {
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  flex-shrink: 0;
  width: 24px;
  height: 24px;
}

.custom-text-input {
  flex: 1 1 auto !important;
  width: 100% !important;
  min-width: 0 !important;
  background: transparent !important;
  border: none !important;
  outline: none !important;
  padding: 14px 0 !important;
  font-size: 14px !important;
  color: #1f2937 !important;
  line-height: 1.5 !important;
}

.custom-text-input:focus {
  outline: none !important;
  border: none !important;
}

.custom-text-input::placeholder {
  color: #9ca3af !important;
  font-size: 14px !important;
  opacity: 1 !important;
}

.input-suffix-icon {
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  flex-shrink: 0;
}

.input-icon {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
  background: linear-gradient(145deg, #9ca3af, #6b7280);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.input-group:hover .input-icon {
  background: linear-gradient(145deg, var(--primary-color), #1e40af);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  filter: drop-shadow(0 4px 8px rgba(13, 38, 145, 0.3)) drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
  transform: scale(1.1);
}

.input-group:focus-within .input-icon {
  background: linear-gradient(145deg, var(--primary-color), #1e40af);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  filter: drop-shadow(0 6px 12px rgba(13, 38, 145, 0.4)) drop-shadow(0 3px 6px rgba(0, 0, 0, 0.3));
  transform: scale(1.15);
}

/* Remove background from icons in buttons */
.toggle-password-btn {
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  background: transparent !important;
  border: none !important;
  cursor: pointer !important;
  padding: 6px !important;
  border-radius: 8px !important;
  transition: all 0.2s ease !important;
  color: #6b7280 !important;
}

.toggle-password-btn:hover {
  background-color: rgba(13, 38, 145, 0.08) !important;
  color: var(--primary-color) !important;
  transform: scale(1.08);
}

.toggle-password-btn:active {
  transform: scale(0.95);
}

button .input-icon,
button .lucide {
  background: none !important;
  -webkit-text-fill-color: unset !important;
  color: inherit !important;
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.2)) !important;
}

button:hover .input-icon,
button:hover .lucide {
  filter: drop-shadow(0 2px 4px rgba(255, 255, 255, 0.3)) !important;
}

/* Elegant Gradient Background - Logo Theme */
.gradient-bg {
  background: linear-gradient(
    135deg,
    #0d2691 0%,
    #1e40af 30%,
    #2563eb 60%,
    #3b82f6 100%
  );
  background-size: 100% 100%;
  position: relative;
}

.gradient-bg::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background:
    radial-gradient(
      circle at 20% 30%,
      rgba(13, 38, 145, 0.4),
      transparent 50%
    ),
    radial-gradient(
      circle at 80% 70%,
      rgba(59, 130, 246, 0.3),
      transparent 50%
    ),
    radial-gradient(
      circle at 50% 90%,
      rgba(37, 99, 235, 0.2),
      transparent 40%
    );
  z-index: 1;
  opacity: 0.8;
}

/* Enhanced Login Card with Glassmorphism */
.login-card-enhanced {
  background: rgba(255, 255, 255, 0.95) !important;
  backdrop-filter: blur(20px) !important;
  -webkit-backdrop-filter: blur(20px) !important;
  border: 1px solid rgba(255, 255, 255, 0.3) !important;
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.25),
    0 0 100px rgba(120, 119, 198, 0.1),
    inset 0 1px 3px rgba(255, 255, 255, 0.6) !important;
}

.glass-effect {
  background: rgba(255, 255, 255, 0.98) !important;
  backdrop-filter: blur(10px) !important;
  -webkit-backdrop-filter: blur(10px) !important;
  border-right: 1px solid rgba(255, 255, 255, 0.2) !important;
}


/* Password strength indicator */
.password-strength {
  height: 4px;
  border-radius: 2px;
  overflow: hidden;
  background: #e5e7eb;
  margin-top: 8px;
}

.password-strength-bar {
  height: 100%;
  background: linear-gradient(90deg, #ef4444 0%, #f59e0b 50%, #10b981 100%);
  transition: width 0.3s ease;
  border-radius: 2px;
}

/* Floating label effect */
.form-label {
  color: #374151 !important;
  font-weight: 500 !important;
  font-size: 14px !important;
  margin-bottom: 8px !important;
  display: block !important;
}



/* Tailwind-like utilities for compatibility */
.min-h-screen {
  min-height: 100vh;
}

.flex {
  display: flex;
}

.flex-col {
  flex-direction: column;
}

.items-center {
  align-items: center;
}

.justify-center {
  justify-content: center;
}

.w-full {
  width: 100%;
}

.max-w-4xl {
  max-width: 56rem;
}

.max-w-6xl {
  max-width: 72rem;
}

.max-w-7xl {
  max-width: 80rem;
}

.rounded-3xl {
  border-radius: 1.5rem;
}

.rounded-t-2xl {
  border-top-left-radius: 1rem;
  border-top-right-radius: 1rem;
}

.rounded-b-3xl {
  border-bottom-left-radius: 1.5rem;
  border-bottom-right-radius: 1.5rem;
}

.overflow-hidden {
  overflow: hidden;
}

.bg-white {
  background-color: #ffffff;
}

.hidden {
  display: none;
}

.lg\:flex {
  display: flex;
}

.lg\:w-\[60\%\] {
  width: 60%;
}

.lg\:w-\[55\%\] {
  width: 55%;
}

.lg\:w-\[45\%\] {
  width: 45%;
}

.lg\:w-\[40\%\] {
  width: 40%;
}

.relative {
  position: relative;
}

.absolute {
  position: absolute;
}

.top-4 {
  top: 1rem;
}

.right-4 {
  right: 1rem;
}

.z-10 {
  z-index: 10;
}

.absolute {
  position: absolute;
}

.inset-0 {
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
}

.z-10 {
  z-index: 10;
}

.text-center {
  text-align: center;
}

.text-white {
  color: #ffffff;
}

.px-12 {
  padding-left: 3rem;
  padding-right: 3rem;
}

.px-8 {
  padding-left: 2rem;
  padding-right: 2rem;
}

.px-3 {
  padding-left: 0.75rem;
  padding-right: 0.75rem;
}

.pr-4 {
  padding-right: 1rem;
}

.pr-10 {
  padding-right: 2.5rem;
}


.py-2\.5 {
  padding-top: 0.625rem;
  padding-bottom: 0.625rem;
}

.py-16 {
  padding-top: 4rem;
  padding-bottom: 4rem;
}

.py-3 {
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
}

.mb-5 {
  margin-bottom: 1.25rem;
}

.mb-8 {
  margin-bottom: 2rem;
}

.mb-3 {
  margin-bottom: 0.75rem;
}

.mb-1 {
  margin-bottom: 0.25rem;
}

.mb-4 {
  margin-bottom: 1rem;
}

.mb-6 {
  margin-bottom: 1.5rem;
}

.mr-2 {
  margin-right: 0.5rem;
}

.mr-3 {
  margin-right: 0.75rem;
}

.ml-2 {
  margin-left: 0.5rem;
}

.mt-10 {
  margin-top: 2.5rem;
}

.mt-6 {
  margin-top: 1.5rem;
}

.max-w-sm {
  max-width: 24rem;
}

.max-w-md {
  max-width: 28rem;
}

.mx-auto {
  margin-left: auto;
  margin-right: auto;
}

.text-xl {
  font-size: 1.25rem;
  line-height: 1.75rem;
}

.text-base {
  font-size: 1rem;
  line-height: 1.5rem;
}

.text-5xl {
  font-size: 3rem;
  line-height: 2;
}

.text-sm {
  font-size: 0.875rem;
  line-height: 1.25rem;
}

.text-xs {
  font-size: 0.75rem;
  line-height: 1rem;
}

.font-semibold {
  font-weight: 600;
}

.font-bold {
  font-weight: 700;
}

.font-medium {
  font-weight: 500;
}

.tracking-wide {
  letter-spacing: 0.05em;
}

.opacity-90 {
  opacity: 0.9;
}

.rounded {
  border-radius: 0.25rem;
}

.shadow-sm {
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
}

.shadow-lg {
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

.text-gray-700 {
  color: #374151;
}

.text-gray-400 {
  color: #9ca3af;
}

.bg-red-50 {
  background-color: #fef2f2;
}

.bg-green-50 {
  background-color: #f0fdf4;
}

.bg-blue-50 {
  background-color: #eff6ff;
}

.text-red-700 {
  color: #b91c1c;
}

.text-green-700 {
  color: #15803d;
}

.text-blue-700 {
  color: #1d4ed8;
}

.text-blue-800 {
  color: #1e40af;
}

.text-blue-600 {
  color: #2563eb;
}

.border-b-2 {
  border-bottom-width: 2px;
}

.border-gray-300 {
  border-color: #d1d5db;
}

.border-red-200 {
  border-color: #fecaca;
}

.border-green-200 {
  border-color: #bbf7d0;
}

.border-blue-200 {
  border-color: #bfdbfe;
}

.rounded-lg {
  border-radius: 0.5rem;
}

.inline-block {
  display: inline-block;
}

.transition-colors {
  transition-property: color, background-color, border-color, text-decoration-color, fill, stroke;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.transition-all {
  transition-property: all;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.hover\:opacity-90:hover {
  opacity: 0.9;
}

.hover\:bg-green-700:hover {
  background-color: #15803d;
}

.hover\:text-\[var\(--primary-color\)\]:hover {
  color: var(--primary-color);
}

.focus\:outline-none:focus {
  outline: 2px solid transparent;
  outline-offset: 2px;
}

.disabled\:cursor-not-allowed:disabled {
  cursor: not-allowed;
}

.cursor-pointer {
  cursor: pointer;
}

.flex-1 {
  flex: 1 1 0%;
}

.flex-shrink-0 {
  flex-shrink: 0;
}

.mt-0\.5 {
  margin-top: 0.125rem;
}

.font-medium {
  font-weight: 500;
}

.inline-flex {
  display: inline-flex;
}

.-ml-1 {
  margin-left: -0.25rem;
}

.h-4 {
  height: 1rem;
}

.h-10 {
  height: 4.5rem;
}

.w-4 {
  width: 1rem;
}

.w-14 {
  width: 3.5rem;
}

.h-14 {
  height: 3.5rem;
}

.rounded-xl {
  border-radius: 0.75rem;
}

.flex-shrink-0 {
  flex-shrink: 0;
}

/* Additional utility classes for responsive design */
.sm\:px-6 {
  padding-left: 1.5rem;
  padding-right: 1.5rem;
}

.md\:px-8 {
  padding-left: 2rem;
  padding-right: 2rem;
}

.py-6 {
  padding-top: 1.5rem;
  padding-bottom: 1.5rem;
}

/* Logo header styles */
.logo-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: 1rem;
}

.logo-header h2 {
  flex: 1;
  margin: 0;
}

.logo-container {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.logo-container img {
  transition: all 0.3s ease;
  max-height: 5rem;
  width: auto;
}

.logo-container:hover img {
  transform: scale(1.05);
}

/* Hide logo on desktop (min-width: 1024px) */
@media (min-width: 1024px) {
  .logo-container {
    display: none;
  }

  .logo-header {
    justify-content: flex-start;
  }
}

/* Desktop enhancements (min-width: 1024px) */
@media (min-width: 1024px) {
  .logo-container {
    display: none;
  }

  .logo-header {
    justify-content: flex-start;
    margin-bottom: 16px;
  }

  /* Floating logo animation for welcome panel */
  .welcome-panel-logo {
    position: absolute;
    top: 25%;
    left: 50%;
    transform: translate(-50%, -50%);
    animation: floating 3s ease-in-out infinite;
    z-index: 5;
  }

  /* Floating animation for logo */
  @keyframes floating {
    0%, 100% {
      transform: translate(-50%, -50%) translateY(0px);
    }
    50% {
      transform: translate(-50%, -50%) translateY(-15px);
    }
  }

  /* Logo glow effect on desktop */
  .welcome-panel-logo img {
    filter: drop-shadow(0 15px 30px rgba(255, 255, 255, 0.5));
    animation: glow 2s ease-in-out infinite alternate;
    height: 150px !important;
    width: auto;
    transition: all 0.3s ease;
  }

  .welcome-panel-logo:hover img {
    transform: scale(1.1) !important;
    filter: drop-shadow(0 25px 50px rgba(255, 255, 255, 0.8));
  }

  @keyframes glow {
    0% {
      filter: drop-shadow(0 15px 30px rgba(255, 255, 255, 0.5));
    }
    100% {
      filter: drop-shadow(0 25px 40px rgba(255, 255, 255, 0.7));
    }
  }

  /* Enlarge form container on desktop */
  .max-w-md {
    max-width: 700px !important;
  }

  /* Enlarge form inputs on desktop */
  .input-field {
    padding: 18px 20px !important;
    font-size: 18px !important;
    min-height: 56px !important;
  }

  .input-group {
    padding: 8px !important;
  }

  /* Enlarge form labels on desktop */
  .form-label {
    font-size: 16px !important;
    font-weight: 600 !important;
    margin-bottom: 12px !important;
  }

  /* Enlarge form heading on desktop */
  .logo-header h2 {
    font-size: 24px !important;
    margin-bottom: 8px !important;
  }

  /* Enlarge button on desktop */
  button[type="submit"] {
    padding: 16px 32px !important;
    font-size: 18px !important;
    min-height: 56px !important;
    font-weight: 600 !important;
  }

  /* Increase spacing on desktop */
  .mb-4 {
    margin-bottom: 24px !important;
  }

  .mb-6 {
    margin-bottom: 32px !important;
  }

  .mt-10 {
    margin-top: 40px !important;
  }

  /* Enlarge icons on desktop */
  .input-icon {
    font-size: 20px !important;
  }

  /* Increase container padding on desktop */
  .glass-effect {
    padding: 48px !important;
  }
}

/* Show logo on mobile and tablets only */
@media (max-width: 1023px) {
  .w-full {
    width: 100% !important;
  }

  .logo-container {
    display: flex;
    position: relative;
    top: auto;
    left: auto;
    transform: none;
    animation: none;
  }

  .logo-header {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin-bottom: 16px;
  }

  .logo-container img {
    height: 4.5rem !important;
    filter: none;
    animation: none;
    transition: all 0.3s ease;
  }

  .logo-container:hover img {
    transform: scale(1.05);
    filter: none;
  }
}

@media (max-width: 768px) {
  .logo-header {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .logo-container img {
    height: 4rem !important;
  }
}

@media (max-width: 640px) {
  .logo-header {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .logo-container img {
    height: 3.5rem !important;
  }

  .logo-header h2 {
    flex: 1;
    margin: 0;
  }
}

.bg-\[var\(--secondary-color\)\] {
  background-color: var(--secondary-color);
}

.text-\[var\(--primary-color\)\] {
  color: var(--primary-color);
}

.text-2xl {
  font-size: 1.5rem;
  line-height: 2rem;
}

/* Loading spinner */
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}

/* Responsive styles */
@media (min-width: 1024px) {
  .hidden {
    display: block;
  }

  .lg\:flex {
    display: flex;
  }

  .lg\:w-\[55\%\] {
    width: 55%;
  }

  .lg\:w-\[45\%\] {
    width: 45%;
  }
}

@media (max-width: 1023px) {
  /* Hide welcome panel on tablets and smaller */
  .lg\:flex {
    display: none;
  }

  /* Make form container more compact on smaller screens */
  .login-card-enhanced {
    width: 95vw !important;
    max-width: 95vw !important;
    border-radius: 1.5rem !important;
    min-height: auto !important;
    height: auto;
    margin: 0 auto;
  }

  /* Adjust form section to be full width on mobile */
  .glass-effect {
    width: 100% !important;
    border-right: none !important;
    padding: 2rem 1.5rem !important;
  }

  /* Center form better on mobile */
  .max-w-md {
    max-width: 90%;
  }
}

@media (max-width: 768px) {
  .px-12 {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .py-16 {
    padding-top: 1rem;
    padding-bottom: 1rem;
  }

  .px-8 {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  /* Additional mobile optimizations */
  .gradient-bg {
    padding: 1rem;
  }

  .login-card-enhanced {
    width: 95vw !important;
    max-width: 95vw !important;
    border-radius: 1.5rem !important;
    height: auto !important;
    min-height: auto !important;
  }

  .glass-effect {
    padding: 1.5rem 1rem !important;
  }
}

@media (max-width: 640px) {
  .gradient-bg {
    padding: 0.5rem;
  }

  .glass-effect {
    padding: 1rem 0.75rem !important;
  }

  .max-w-md {
    max-width: 95%;
  }

  .login-card-enhanced {
    width: 98vw !important;
    max-width: 98vw !important;
    border-radius: 1rem !important;
    height: auto !important;
    min-height: auto !important;
  }
}

/* Credit Section Styles */
.pb-20 {
  padding-bottom: 1rem;
}

.text-gray-600 {
  color: #6b7280;
}

.text-gray-400 {
  color: #9ca3af;
}

.hover\:text-\[var\(--primary-color\)\]:hover {
  color: var(--primary-color) !important;
}

.space-x-6 > :not([hidden]) ~ :not([hidden]) {
  --space-x-reverse: 0;
  margin-right: calc(1.5rem * var(--space-x-reverse));
  margin-left: calc(1.5rem * calc(1 - var(--space-x-reverse)));
}

.space-x-4 > :not([hidden]) ~ :not([hidden]) {
  --space-x-reverse: 0;
  margin-right: calc(1rem * var(--space-x-reverse));
  margin-left: calc(1rem * calc(1 - var(--space-x-reverse)));
}

.space-y-2 > :not([hidden]) ~ :not([hidden]) {
  --space-y-reverse: 0;
  margin-top: calc(0.5rem * calc(1 - var(--space-y-reverse)));
  margin-bottom: calc(0.5rem * var(--space-y-reverse));
}

.flex-shrink-0 {
  flex-shrink: 0;
}

/* Bottom credit bar styles */
.bottom-0 {
  bottom: 0;
}

.left-8 {
  left: 2rem;
}

.right-8 {
  right: 2rem;
}

.px-8 {
  padding-left: 2rem;
  padding-right: 2rem;
}

.py-4 {
  padding-top: 1rem;
  padding-bottom: 1rem;
}

.bg-white\/90 {
  background-color: rgba(255, 255, 255, 0.9);
}

.backdrop-blur-sm {
  backdrop-filter: blur(4px);
}

.border-t {
  border-top-width: 1px;
}

.border-gray-200 {
  border-color: #e5e7eb;
}

.rounded-t-2xl {
  border-top-left-radius: 1rem;
  border-top-right-radius: 1rem;
}

.z-50 {
  z-index: 50;
}

/* Enhanced Copyright Section Styles */
.copyright-section-merged {
  position: relative;
  background: rgba(255, 255, 255, 0.98) !important;
  backdrop-filter: blur(24px) saturate(180%) !important;
  -webkit-backdrop-filter: blur(24px) saturate(180%) !important;
  border: 1px solid rgba(255, 255, 255, 0.3) !important;
  border-top: none !important;
  box-shadow:
    0 -8px 32px rgba(0, 0, 0, 0.12),
    0 -2px 8px rgba(0, 0, 0, 0.08),
    0 0 0 1px rgba(255, 255, 255, 0.5) inset,
    0 1px 0 rgba(255, 255, 255, 0.9) inset !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: visible;
}

.copyright-section-merged:hover {
  transform: translateY(-2px);
  box-shadow:
    0 -12px 40px rgba(0, 0, 0, 0.15),
    0 -4px 12px rgba(0, 0, 0, 0.1),
    0 0 0 1px rgba(255, 255, 255, 0.6) inset,
    0 2px 0 rgba(255, 255, 255, 0.95) inset,
    0 0 0 1px rgba(13, 38, 145, 0.1);
}

.copyright-section-merged::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    180deg,
    rgba(255, 255, 255, 0.1) 0%,
    transparent 100%
  );
  pointer-events: none;
  z-index: 1;
}

/* Enhanced curved edges */
.copyright-section-merged > div:first-child {
  z-index: 2;
}

/* Better curve styling for left and right sides */
.copyright-section-merged > div:first-child > div:first-child,
.copyright-section-merged > div:first-child > div:nth-child(2) {
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.1);
  background: linear-gradient(135deg, #ffffff 0%, #f8fafc 100%);
}

/* Gradient overlay improvements */
.copyright-section-merged > div:first-child > div:nth-child(3) {
  background: linear-gradient(180deg,
    rgba(255, 255, 255, 0.9) 0%,
    rgba(255, 255, 255, 0.3) 30%,
    transparent 100%
  );
}

.website-links {
  position: relative;
  z-index: 2;
}

.website-link {
  position: relative;
  overflow: hidden;
  color: #6b7280 !important;
  text-decoration: none !important;
  background: transparent;
  border-radius: 8px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.website-link::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(13, 38, 145, 0.1), transparent);
  transition: left 0.6s ease;
  z-index: -1;
}

.website-link:hover::before {
  left: 100%;
}

.website-link:hover {
  color: var(--primary-color) !important;
  background: rgba(13, 38, 145, 0.05) !important;
  border-radius: 8px;
  transform: translateY(-1px) scale(1.02);
  box-shadow: 0 2px 8px rgba(13, 38, 145, 0.15);
  text-decoration: none !important;
}

.instagram-credit {
  position: relative;
  z-index: 2;
}

.instagram-container {
  position: relative;
}

.instagram-link {
  position: relative;
  overflow: hidden;
  color: #6b7280 !important;
  text-decoration: none !important;
  background: transparent;
  border-radius: 8px;
  border: 1px solid transparent;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.instagram-link::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg,
    transparent,
    rgba(219, 39, 119, 0.1),
    rgba(147, 51, 234, 0.1),
    transparent
  );
  transition: left 0.6s ease;
  z-index: -1;
}

.instagram-link:hover::before {
  left: 100%;
}

.instagram-link:hover {
  color: var(--primary-color) !important;
  background: linear-gradient(135deg,
    rgba(219, 39, 119, 0.05),
    rgba(147, 51, 234, 0.05)
  ) !important;
  border-color: rgba(219, 39, 119, 0.2) !important;
  border-radius: 8px;
  transform: translateY(-1px) scale(1.02);
  box-shadow:
    0 2px 8px rgba(219, 39, 119, 0.15),
    0 1px 4px rgba(147, 51, 234, 0.1);
  text-decoration: none !important;
}

/* Enhanced hover effects for Instagram icon */
.instagram-container svg {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.1));
}

.instagram-container:hover svg {
  color: #ec4899 !important;
  transform: scale(1.1) rotate(5deg);
  filter: drop-shadow(0 2px 8px rgba(236, 72, 153, 0.4));
}

/* Gradient accent strip animation */
.copyright-container > div:first-child {
  position: relative;
  overflow: hidden;
}

.copyright-container > div:first-child::after {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.3),
    transparent
  );
  animation: shimmer 3s infinite;
}

@keyframes shimmer {
  0% {
    left: -100%;
  }
  100% {
    left: 100%;
  }
}

.shimmer-animation {
  animation: shimmer 4s ease-in-out infinite;
}

/* Add pulse effect to gradient accent strip */
.copyright-container > div:first-child {
  animation: pulse 3s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 0.8;
  }
  50% {
    opacity: 1;
  }
}

/* Enhanced responsive design */
@media (max-width: 768px) {
  .copyright-section-merged {
    margin-left: 0.75rem !important;
    margin-right: 0.75rem !important;
  }

  .copyright-section-merged {
    border-radius: 0 0 1.25rem 1.25rem !important;
  }

  .website-link,
  .instagram-link {
    font-size: 0.75rem !important;
    padding: 0.5rem 0.75rem !important;
  }
}

/* Hide copyright section on mobile devices for cleaner look */
@media (max-width: 640px) {
  .copyright-section-merged {
    display: none !important;
  }
}
</style>

