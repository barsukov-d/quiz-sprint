<script setup lang="ts">
interface Props {
  title: string
  icon: string
  description: string
  disabled?: boolean
  badge?: string
  badgeColor?: 'gray' | 'yellow' | 'blue' | 'green' | 'red'
  lives?: number
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  badgeColor: 'gray'
})

const emit = defineEmits<{
  click: []
}>()

const handleClick = () => {
  if (!props.disabled) {
    emit('click')
  }
}
</script>

<template>
  <UCard
    :class="[
      'transition-all duration-200',
      disabled
        ? 'opacity-50 cursor-not-allowed'
        : 'cursor-pointer hover:shadow-lg hover:scale-[1.02]'
    ]"
    @click="handleClick"
  >
    <div class="flex items-center gap-4">
      <!-- Icon -->
      <div
        :class="[
          'flex-shrink-0 w-12 h-12 rounded-lg flex items-center justify-center text-2xl',
          disabled
            ? 'bg-gray-100 dark:bg-gray-800'
            : 'bg-primary-100 dark:bg-primary-900/30'
        ]"
      >
        <UIcon v-if="icon.startsWith('i-')" :name="icon" class="size-6" />
        <span v-else>{{ icon }}</span>
      </div>

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-1">
          <h4 class="font-semibold text-base">{{ title }}</h4>
          <UBadge v-if="badge" :color="badgeColor" size="xs" variant="subtle">
            {{ badge }}
          </UBadge>
        </div>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ description }}
        </p>

        <!-- Lives indicator (для Marathon) -->
        <div v-if="lives !== undefined" class="flex items-center gap-1 mt-2">
          <UIcon
            v-for="i in 3"
            :key="i"
            :name="i <= lives ? 'i-heroicons-heart-solid' : 'i-heroicons-heart'"
            :class="i <= lives ? 'text-red-500' : 'text-gray-300 dark:text-gray-600'"
            class="size-4"
          />
          <span class="text-xs text-gray-500 ml-1">{{ lives }} lives</span>
        </div>
      </div>

      <!-- Arrow -->
      <UIcon
        v-if="!disabled"
        name="i-heroicons-chevron-right"
        class="flex-shrink-0 size-5 text-gray-400"
      />
    </div>
  </UCard>
</template>
