import { createRouter, createWebHistory } from 'vue-router'
import CategoriesView from '../views/CategoriesView.vue'
import QuizListView from '../views/QuizListView.vue'
import QuizDetailsView from '../views/QuizDetailsView.vue'
import QuizPlayView from '../views/QuizPlayView.vue'
import QuizResultsView from '../views/QuizResultsView.vue'
import LeaderboardView from '../views/LeaderboardView.vue'
import ProfileView from '../views/ProfileView.vue'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/quizzes',
      name: 'quizzes',
      component: QuizListView,
    },
    {
      path: '/quiz/:id',
      name: 'quiz-details',
      component: QuizDetailsView,
    },
    {
      path: '/quiz/:id/play',
      name: 'quiz-play',
      component: QuizPlayView,
    },
    {
      path: '/quiz/results/:sessionId',
      name: 'quiz-results',
      component: QuizResultsView,
    },
    {
      path: '/leaderboard/:quizId?',
      name: 'leaderboard',
      component: LeaderboardView,
    },
    {
      path: '/profile',
      name: 'profile',
      component: ProfileView,
    },
    // Legacy route (для обратной совместимости)
    {
      path: '/categories',
      name: 'categories',
      component: CategoriesView,
    },
  ],
})

export default router
