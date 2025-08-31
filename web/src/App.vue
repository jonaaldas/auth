<script setup lang="ts">
import { ref } from "vue";
import { ofetch } from "ofetch";
import { onMounted } from "vue";
const isAuthenticated = ref(false);
const user = ref(null);
const userId = ref(null);
const authError = ref("");
async function handleLogin() {
  isAuthenticated.value = false;

  const res = await ofetch("/api/login", {
    method: "POST",
    body: {
      email: "jonaaldas@gmail.com",
      password: "123",
    },
  });

  if (res.success) {
    isAuthenticated.value = true;
    userId.value = res.user.id;
  }
  if (res.error) {
    authError.value = res.error;
  }
}

async function handleLogout() {
  const res = await ofetch("/api/logout", {
    method: "POST",
  });
  if (res.success) {
    isAuthenticated.value = false;
    user.value = null;
    userId.value = null;
    authError.value = "";
  }
}

async function callProtectedRoute() {
  try {
    const res = await ofetch("/api/protected/profile", {
      method: "GET",
    });

    if (!res.success) {
      isAuthenticated.value = false;
      user.value = null;
      authError.value = "Unauthorized - No session cookie";
      return;
    }

    isAuthenticated.value = true;
    user.value = res.user;
    authError.value = "";
  } catch (error) {
    isAuthenticated.value = false;
    user.value = null;
    authError.value = "Failed to call protected route";
  }
}

onMounted(() => {
  const cookie = document.cookie;
  const sessionCookie = cookie
    .split(";")
    .find((c: string) => c.trim().startsWith("session="));
  if (sessionCookie) {
    isAuthenticated.value = true;
    ofetch("/api/protected/profile", {
      method: "GET",
    }).then((res) => {
      if (res.success) {
        user.value = res.user;
      }
      if (res.error) {
        authError.value = res.error;
      }
    });
  }
});
</script>

<template>
  <div class="min-h-screen bg-black flex items-center justify-center p-4">
    <!-- Aura gradient background -->
    <div class="absolute inset-0 overflow-hidden">
      <div
        class="absolute -top-40 -right-40 w-80 h-80 bg-purple-500 rounded-full mix-blend-multiply filter blur-3xl opacity-30 animate-blob"
      ></div>
      <div
        class="absolute -bottom-40 -left-40 w-80 h-80 bg-pink-500 rounded-full mix-blend-multiply filter blur-3xl opacity-30 animate-blob animation-delay-2000"
      ></div>
      <div
        class="absolute top-40 left-40 w-80 h-80 bg-blue-500 rounded-full mix-blend-multiply filter blur-3xl opacity-30 animate-blob animation-delay-4000"
      ></div>
    </div>

    <!-- Auth Card -->
    <div class="relative z-10 w-full max-w-md">
      <div
        class="backdrop-blur-xl bg-white/10 rounded-3xl shadow-2xl border border-white/20 p-8"
      >
        <!-- Card Header -->
        <div class="text-center mb-8">
          <div
            class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-violet-500 to-pink-500 mb-4"
          >
            <svg
              class="w-8 h-8 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
              ></path>
            </svg>
          </div>
          <h2 class="text-3xl font-bold text-white mb-2">Authentication</h2>
          <p class="text-gray-300 text-sm">Secure access control system</p>
        </div>

        <!-- Status Indicator -->
        <div class="mb-6 p-3 rounded-xl bg-white/5 border border-white/10">
          <div class="flex items-center justify-between">
            <span class="text-gray-300 text-sm">Status:</span>
            <div class="flex items-center gap-2">
              <div v-if="userId" class="flex items-center gap-2">
                <span class="text-gray-300 text-sm">User ID:</span>
                <span class="text-white text-sm font-medium">{{ userId }}</span>
              </div>
              <div
                :class="isAuthenticated ? 'bg-green-500' : 'bg-red-500'"
                class="w-2 h-2 rounded-full animate-pulse"
              ></div>
              <span class="text-white text-sm font-medium">{{
                isAuthenticated ? "Authenticated" : "Not Authenticated"
              }}</span>
            </div>
          </div>
        </div>

        <div
          v-if="user"
          class="mb-6 p-3 rounded-xl bg-white/5 border border-white/10"
        >
          <pre class="text-white overflow-x-auto"><code>{{ user }}</code></pre>
        </div>

        <div
          v-if="authError"
          class="mb-6 p-3 rounded-xl bg-red-500/20 border border-red-500/30"
        >
          <div class="flex items-center gap-2 mb-2">
            <svg class="w-5 h-5 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.732 16.5c-.77.833.192 2.5 1.732 2.5z"></path>
            </svg>
            <span class="text-red-400 text-sm font-medium">Error</span>
          </div>
          <p class="text-red-200 text-sm">{{ authError }}</p>
        </div>

        <!-- Buttons -->
        <div class="space-y-3">
          <!-- Login Button -->
          <button
            @click="handleLogin"
            :disabled="isAuthenticated"
            :class="isAuthenticated ? 'opacity-50 cursor-not-allowed' : 'hover:from-violet-700 hover:to-pink-700 hover:scale-[1.02] active:scale-[0.98] hover:shadow-xl'"
            class="w-full py-3 px-4 bg-gradient-to-r from-violet-600 to-pink-600 text-white font-semibold rounded-xl transition-all duration-200 transform shadow-lg"
          >
            <div class="flex items-center justify-center gap-2">
              <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1"
                ></path>
              </svg>
              <span>Login</span>
            </div>
          </button>

          <!-- Logout Button -->
          <button
            @click="handleLogout"
            :disabled="!isAuthenticated"
            :class="!isAuthenticated ? 'opacity-50 cursor-not-allowed' : 'hover:bg-white/20 hover:scale-[1.02] active:scale-[0.98]'"
            class="w-full py-3 px-4 bg-white/10 backdrop-blur-sm text-white font-semibold rounded-xl transition-all duration-200 transform border border-white/20"
          >
            <div class="flex items-center justify-center gap-2">
              <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
                ></path>
              </svg>
              <span>Logout</span>
            </div>
          </button>

          <!-- Protected Route Button -->
          <button
            @click="callProtectedRoute"
            class="w-full py-3 px-4 bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-600 hover:to-blue-700 text-white font-semibold rounded-xl transition-all duration-200 transform hover:scale-[1.02] active:scale-[0.98] shadow-lg hover:shadow-xl"
          >
            <div class="flex items-center justify-center gap-2">
              <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
                ></path>
              </svg>
              <span>Get Profile Data</span>
            </div>
          </button>
        </div>

        <!-- Footer -->
        <div class="mt-8 pt-6 border-t border-white/10">
          <p class="text-center text-gray-400 text-xs">
            Secure authentication with modern UI
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@keyframes blob {
  0% {
    transform: translate(0px, 0px) scale(1);
  }
  33% {
    transform: translate(30px, -50px) scale(1.1);
  }
  66% {
    transform: translate(-20px, 20px) scale(0.9);
  }
  100% {
    transform: translate(0px, 0px) scale(1);
  }
}

.animate-blob {
  animation: blob 7s infinite;
}

.animation-delay-2000 {
  animation-delay: 2s;
}

.animation-delay-4000 {
  animation-delay: 4s;
}
</style>
