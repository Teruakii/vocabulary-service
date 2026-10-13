import { createRouter, createWebHistory } from 'vue-router'
import { auth } from '@/api/client'
import AppLayout from '@/layouts/AppLayout.vue'
import AuthLayout from '@/layouts/AuthLayout.vue'
import FocusLayout from '@/layouts/FocusLayout.vue'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    guest?: boolean
  }
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // All three layouts share the '/' prefix. A layout without a '' child would
    // itself match '/' and render empty, so the one owning '' must come first.
    {
      path: '/',
      component: AppLayout,
      meta: { requiresAuth: true },
      children: [
        { path: '', name: 'dashboard', component: () => import('@/views/DashboardView.vue') },
        { path: 'my-words', name: 'my-words', component: () => import('@/views/MyWordsView.vue') },
        { path: 'library', name: 'library', component: () => import('@/views/LibraryView.vue') },
        { path: 'quizzes', name: 'quizzes', component: () => import('@/views/QuizListView.vue') },
        { path: 'categories', name: 'categories', component: () => import('@/views/CategoriesView.vue') },
        { path: 'profile', name: 'profile', component: () => import('@/views/ProfileView.vue') },
      ],
    },
    {
      path: '/',
      component: AuthLayout,
      meta: { guest: true },
      children: [
        { path: 'login', name: 'login', component: () => import('@/views/auth/LoginView.vue') },
        { path: 'register', name: 'register', component: () => import('@/views/auth/RegisterView.vue') },
        {
          path: 'forgot-password',
          name: 'forgot-password',
          component: () => import('@/views/auth/ForgotPasswordView.vue'),
        },
        {
          path: 'reset-password',
          name: 'reset-password',
          component: () => import('@/views/auth/ResetPasswordView.vue'),
        },
      ],
    },
    {
      path: '/',
      component: FocusLayout,
      meta: { requiresAuth: true },
      children: [
        { path: 'review', name: 'review', component: () => import('@/views/ReviewView.vue') },
        { path: 'quizzes/:id(\\d+)', name: 'quiz-play', component: () => import('@/views/QuizPlayView.vue') },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: { name: 'dashboard' } },
  ],
})

router.beforeEach(async (to) => {
  const signedIn = await auth.hasSession()
  if (to.meta.requiresAuth && !signedIn) return { name: 'login', query: { redirect: to.fullPath } }
  if (to.meta.guest && signedIn) return { name: 'dashboard' }
})

export default router
