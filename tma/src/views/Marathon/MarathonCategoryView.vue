<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMarathon } from '@/composables/useMarathon'
import { useAuth } from '@/composables/useAuth'
import { useGetCategories } from '@/api/generated'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { currentUser } = useAuth()
const { t } = useI18n()
const playerId = currentUser.value?.id || 'guest'

const { startGame, isLoading } = useMarathon(playerId)

const { data: categoriesData, isLoading: isLoadingCategories } = useGetCategories()

const selectedCategory = ref<string | null>(null)

const handleSelectCategory = async (categoryId: string) => {
	selectedCategory.value = categoryId

	try {
		await startGame(categoryId)
	} catch (error) {
		console.error('Failed to start marathon:', error)
		selectedCategory.value = null
	}
}

const handleBack = () => {
	router.push({ name: 'home' })
}
</script>

<template>
	<div class="mx-auto max-w-[800px] pb-4">
		<!-- Header -->
		<div class="flex items-center gap-3 pt-4 pb-6">
			<UButton
				color="neutral"
				variant="ghost"
				icon="i-heroicons-arrow-left"
				size="sm"
				@click="handleBack"
			/>
			<h1 class="text-lg font-bold">{{ t('marathon.chooseCategory') }}</h1>
		</div>

		<!-- Loading -->
		<div
			v-if="isLoadingCategories"
			class="flex flex-col items-center justify-center min-h-[30vh]"
		>
			<UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary" />
			<p class="text-(--ui-text-muted) mt-4">
				{{ t('marathon.loadingCategories') }}
			</p>
		</div>

		<!-- Category Grid -->
		<div v-else class="grid grid-cols-2 gap-3">
			<!-- All Categories option -->
			<div
				class="rounded-(--ui-radius) bg-(--ui-bg-elevated) border border-(--ui-border) p-4 cursor-pointer transition-all active:scale-95"
				:class="
					selectedCategory === 'all'
						? 'ring-2 ring-primary'
						: 'hover:border-(--ui-border-accented)'
				"
				@click="handleSelectCategory('all')"
			>
				<div class="flex flex-col gap-2">
					<div class="flex items-center justify-between">
						<UIcon name="i-heroicons-squares-2x2" class="size-6 text-primary" />
						<UIcon
							v-if="isLoading && selectedCategory === 'all'"
							name="i-heroicons-arrow-path"
							class="size-4 animate-spin text-primary"
						/>
					</div>
					<p class="font-semibold text-sm leading-tight">
						{{ t('marathon.allCategories') }}
					</p>
					<p class="text-xs text-(--ui-text-muted)">
						{{ t('marathon.allCategoriesDesc') }}
					</p>
				</div>
			</div>

			<!-- Dynamic categories -->
			<div
				v-for="category in categoriesData?.data"
				:key="category.id"
				class="rounded-(--ui-radius) bg-(--ui-bg-elevated) border border-(--ui-border) p-4 cursor-pointer transition-all active:scale-95"
				:class="
					selectedCategory === category.id
						? 'ring-2 ring-primary'
						: 'hover:border-(--ui-border-accented)'
				"
				@click="() => handleSelectCategory(category.id)"
			>
				<div class="flex flex-col gap-2">
					<div class="flex items-center justify-between">
						<UIcon name="i-heroicons-tag" class="size-6 text-primary" />
						<UIcon
							v-if="isLoading && selectedCategory === category.id"
							name="i-heroicons-arrow-path"
							class="size-4 animate-spin text-primary"
						/>
					</div>
					<p class="font-semibold text-sm leading-tight">{{ category.name }}</p>
				</div>
			</div>
		</div>
	</div>
</template>
