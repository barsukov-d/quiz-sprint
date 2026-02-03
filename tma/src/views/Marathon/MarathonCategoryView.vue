<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMarathon } from '@/composables/useMarathon'
import { useAuth } from '@/composables/useAuth'
import { useGetCategories } from '@/api/generated'

const router = useRouter()
const { currentUser } = useAuth()
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
	<div class="min-h-screen mx-auto max-w-[800px] px-4 pt-14 pb-8 sm:px-3 sm:pt-12">
		<!-- Header -->
		<div class="flex items-center gap-3 mb-6">
			<UButton
				color="gray"
				variant="ghost"
				icon="i-heroicons-arrow-left"
				size="sm"
				@click="handleBack"
			/>
			<div>
				<h1 class="text-xl font-bold">Choose Category</h1>
				<p class="text-sm text-gray-500 dark:text-gray-400">
					Select a category for your marathon
				</p>
			</div>
		</div>

		<!-- Loading -->
		<div
			v-if="isLoadingCategories"
			class="flex flex-col items-center justify-center min-h-[30vh]"
		>
			<UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary" />
			<p class="text-gray-500 dark:text-gray-400 mt-4">Loading categories...</p>
		</div>

		<!-- Category List -->
		<div v-else class="flex flex-col gap-3">
			<!-- All Categories option -->
			<UCard
				:class="[
					'cursor-pointer transition-all',
					selectedCategory === 'all' ? 'ring-2 ring-primary' : 'hover:ring-1 hover:ring-gray-300',
				]"
				@click="handleSelectCategory('all')"
			>
				<div class="flex items-center gap-3">
					<UIcon name="i-heroicons-squares-2x2" class="size-6 text-primary" />
					<div class="flex-1">
						<p class="font-semibold">All Categories</p>
						<p class="text-sm text-gray-500 dark:text-gray-400">
							Questions from all topics
						</p>
					</div>
					<UIcon
						v-if="isLoading && selectedCategory === 'all'"
						name="i-heroicons-arrow-path"
						class="size-5 animate-spin text-primary"
					/>
					<UIcon v-else name="i-heroicons-chevron-right" class="size-5 text-gray-400" />
				</div>
			</UCard>

			<!-- Dynamic categories -->
			<UCard
				v-for="category in categoriesData?.data"
				:key="category.id"
				:class="[
					'cursor-pointer transition-all',
					selectedCategory === category.id ? 'ring-2 ring-primary' : 'hover:ring-1 hover:ring-gray-300',
				]"
				@click="handleSelectCategory(category.id)"
			>
				<div class="flex items-center gap-3">
					<UIcon name="i-heroicons-tag" class="size-6 text-gray-500" />
					<div class="flex-1">
						<p class="font-semibold">{{ category.name }}</p>
					</div>
					<UIcon
						v-if="isLoading && selectedCategory === category.id"
						name="i-heroicons-arrow-path"
						class="size-5 animate-spin text-primary"
					/>
					<UIcon v-else name="i-heroicons-chevron-right" class="size-5 text-gray-400" />
				</div>
			</UCard>
		</div>
	</div>
</template>
