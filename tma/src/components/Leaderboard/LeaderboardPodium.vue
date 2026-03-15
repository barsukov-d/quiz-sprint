<script setup lang="ts">
interface PodiumEntry {
	rank: number
	username: string
	avatarUrl?: string
	value: string | number
	label?: string
}

interface Props {
	entries: PodiumEntry[]
}

const props = defineProps<Props>()

const podiumOrder = [
	{
		rank: 2,
		order: 'order-1',
		height: 'h-10',
		bg: 'bg-gray-300 dark:bg-gray-600',
		ring: 'ring-gray-400',
		text: 'text-gray-600 dark:text-gray-300',
	},
	{
		rank: 1,
		order: 'order-2',
		height: 'h-14',
		bg: 'bg-yellow-400 dark:bg-yellow-500',
		ring: 'ring-yellow-500',
		text: 'text-yellow-600 dark:text-yellow-300',
	},
	{
		rank: 3,
		order: 'order-3',
		height: 'h-8',
		bg: 'bg-amber-600 dark:bg-amber-700',
		ring: 'ring-amber-600',
		text: 'text-amber-700 dark:text-amber-400',
	},
]

const getRankEmoji = (rank: number) => {
	if (rank === 1) return '🥇'
	if (rank === 2) return '🥈'
	if (rank === 3) return '🥉'
	return ''
}

const getEntry = (rank: number) => props.entries.find((e) => e.rank === rank)
</script>

<template>
	<div class="flex items-end justify-center gap-2 py-3 px-2">
		<div
			v-for="slot in podiumOrder"
			:key="slot.rank"
			class="flex flex-col items-center gap-1 flex-1"
			:class="slot.order"
		>
			<template v-if="getEntry(slot.rank)">
				<!-- Avatar -->
				<div class="relative">
					<UAvatar
						:src="getEntry(slot.rank)?.avatarUrl"
						:alt="getEntry(slot.rank)?.username"
						:size="slot.rank === 1 ? 'lg' : 'md'"
						:class="['ring-2', slot.ring]"
					/>
					<span class="absolute -bottom-1 -right-1 text-sm">{{
						getRankEmoji(slot.rank)
					}}</span>
				</div>

				<!-- Name -->
				<p
					class="text-xs font-semibold text-(--ui-text-highlighted) text-center truncate w-full max-w-[7rem]"
				>
					{{ getEntry(slot.rank)?.username }}
				</p>

				<!-- Value -->
				<div class="flex flex-col items-center">
					<p class="text-xs font-bold" :class="slot.text">
						{{ getEntry(slot.rank)?.value }}
					</p>
					<p
						v-if="getEntry(slot.rank)?.label"
						class="text-[10px] text-(--ui-text-dimmed)"
					>
						{{ getEntry(slot.rank)?.label }}
					</p>
				</div>
			</template>

			<!-- Pedestal -->
			<div
				class="w-full rounded-t-lg flex items-center justify-center"
				:class="[slot.height, slot.bg]"
			>
				<span class="text-white font-bold text-lg">{{ slot.rank }}</span>
			</div>
		</div>
	</div>
</template>
