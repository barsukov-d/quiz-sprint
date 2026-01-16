import axios, { type AxiosRequestConfig } from 'axios'

const getBaseURL = () => {
  // Runtime detection: if accessed via tunnel domain, use that domain for API
  const hostname = typeof window !== 'undefined' ? window.location.hostname : 'localhost'

  if (hostname === 'dev.quiz-sprint-tma.online') {
    return 'https://dev.quiz-sprint-tma.online/api/v1'
  }
  if (hostname === 'staging.quiz-sprint-tma.online') {
    return 'https://staging.quiz-sprint-tma.online/api/v1'
  }
  if (hostname === 'quiz-sprint-tma.online') {
    return 'https://quiz-sprint-tma.online/api/v1'
  }

  // Fallback to env variable or localhost
  return import.meta.env.VITE_API_BASE_URL || 'http://localhost:3000/api/v1'
}

export const apiClient = axios.create({
  baseURL: getBaseURL(),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Custom fetch function for generated hooks
export default async function fetch<TData, TError = unknown, TVariables = unknown>(
  config: AxiosRequestConfig<TVariables>
): Promise<{ data: TData }> {
  const response = await apiClient.request<TData>(config)
  return { data: response.data }
}

export type RequestConfig<TVariables = unknown> = AxiosRequestConfig<TVariables>
export type ResponseErrorConfig<TError = unknown> = TError

// Request interceptor (для Telegram auth в будущем)
apiClient.interceptors.request.use(
  (config) => {
    // TODO: Добавить Telegram WebApp initData
    // const initData = window.Telegram?.WebApp?.initData
    // if (initData) {
    //   config.headers['X-Telegram-Init-Data'] = initData
    // }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor (обработка ошибок)
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // Log detailed error info for debugging
    if (error.response) {
      // Server responded with error status
      console.error('API Error Response:', {
        status: error.response.status,
        data: error.response.data,
        url: error.config?.url,
      })
    } else if (error.request) {
      // Request made but no response received
      console.error('API Network Error:', {
        message: error.message,
        url: error.config?.url,
      })
    } else {
      // Something else happened
      console.error('API Error:', error.message)
    }
    return Promise.reject(error)
  }
)
