import { createRouter, createWebHistory } from 'vue-router'
import CategoriesView from '../views/CategoriesView.vue'
import QuizListView from '../views/QuizListView.vue'
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
    // Legacy route (для обратной совместимости)
    {
      path: '/home',
      name: 'home',
      component: HomeView,
    },
  ],
})

export default router
