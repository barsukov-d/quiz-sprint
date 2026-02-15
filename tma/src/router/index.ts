import { createRouter, createWebHistory } from 'vue-router'
import CategoriesView from '../views/CategoriesView.vue'
import QuizListView from '../views/QuizListView.vue'
import QuizDetailsView from '../views/QuizDetailsView.vue'
import QuizPlayView from '../views/QuizPlayView.vue'
import QuizResultsView from '../views/QuizResultsView.vue'
import LeaderboardView from '../views/LeaderboardView.vue'
import ProfileView from '../views/ProfileView.vue'
import HomeView from '../views/HomeView.vue'

// Daily Challenge Views
import DailyChallengePlayView from '../views/DailyChallenge/DailyChallengePlayView.vue'
import DailyChallengeResultsView from '../views/DailyChallenge/DailyChallengeResultsView.vue'

// Marathon Views
import MarathonCategoryView from '../views/Marathon/MarathonCategoryView.vue'
import MarathonPlayView from '../views/Marathon/MarathonPlayView.vue'
import MarathonGameOverView from '../views/Marathon/MarathonGameOverView.vue'

// PvP Duel Views
import DuelLobbyView from '../views/Duel/DuelLobbyView.vue'
import DuelPlayView from '../views/Duel/DuelPlayView.vue'
import DuelResultsView from '../views/Duel/DuelResultsView.vue'
// DailyChallengeReviewView removed — feedback is now shown inline during gameplay
// TODO: Import DailyChallengeIntroView when created
// import DailyChallengeIntroView from '../views/DailyChallenge/DailyChallengeIntroView.vue'

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
		// Daily Challenge Routes
		{
			path: '/daily-challenge',
			children: [
				// TODO: Uncomment when DailyChallengeIntroView is created
				// {
				//   path: '',
				//   name: 'daily-challenge-intro',
				//   component: DailyChallengeIntroView,
				// },
				{
					path: 'play',
					name: 'daily-challenge-play',
					component: DailyChallengePlayView,
				},
				{
					path: 'results',
					name: 'daily-challenge-results',
					component: DailyChallengeResultsView,
				},
			],
		},
		// Marathon Routes
		{
			path: '/marathon',
			children: [
				{
					path: 'category',
					name: 'marathon-category',
					component: MarathonCategoryView,
				},
				{
					path: 'play',
					name: 'marathon-play',
					component: MarathonPlayView,
				},
				{
					path: 'gameover',
					name: 'marathon-gameover',
					component: MarathonGameOverView,
				},
			],
		},
		// PvP Duel Routes
		{
			path: '/duel',
			children: [
				{
					path: '',
					name: 'duel-lobby',
					component: DuelLobbyView,
				},
				{
					path: 'play/:duelId',
					name: 'duel-play',
					component: DuelPlayView,
				},
				{
					path: 'results/:duelId',
					name: 'duel-results',
					component: DuelResultsView,
				},
			],
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
