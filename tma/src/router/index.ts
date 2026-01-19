import { createRouter, createWebHistory } from 'vue-router'
import CategoriesView from '../views/CategoriesView.vue'
import QuizListView from '../views/QuizListView.vue'
import QuizDetailsView from '../views/QuizDetailsView.vue'
import QuizPlayView from '../views/QuizPlayView.vue'
import QuizResultsView from '../views/QuizResultsView.vue'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'categories',
      component: CategoriesView,
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
    // Legacy route (для обратной совместимости)
    {
      path: '/home',
      name: 'home',
      component: HomeView,
    },
  ],
})

export default router
