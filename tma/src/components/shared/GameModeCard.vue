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
	badgeColor: 'gray',
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
	<div
		:class="[
			'rounded-(--ui-radius) overflow-hidden bg-(--ui-bg-elevated) border border-(--ui-border) transition-all',
			disabled
				? 'opacity-50 cursor-not-allowed'
				: 'cursor-pointer hover:shadow-lg hover:scale-[1.01] active:scale-[0.99]',
		]"
		@click="handleClick"
	>
		<div class="flex items-center gap-3 px-4 py-4">
			<!-- Icon -->
			<div
				:class="[
					'flex-shrink-0 size-10 rounded-xl flex items-center justify-center text-xl',
					disabled ? 'bg-(--ui-bg-accented)' : 'bg-primary-500/15 dark:bg-primary-400/15',
				]"
			>
				<UIcon v-if="icon.startsWith('i-')" :name="icon" class="size-5 text-primary" />
				<span v-else>{{ icon }}</span>
			</div>

			<!-- Content -->
			<div class="flex-1 min-w-0">
				<div class="flex items-center gap-2">
					<h4 class="font-bold text-base text-(--ui-text-highlighted)">{{ title }}</h4>
					<UBadge v-if="badge" :color="badgeColor" size="sm" variant="subtle">
						{{ badge }}
					</UBadge>
				</div>
				<p class="text-xs text-(--ui-text-dimmed) mt-0.5">{{ description }}</p>
			</div>

			<!-- Arrow -->
			<UIcon
				v-if="!disabled"
				name="i-heroicons-chevron-right"
				class="flex-shrink-0 size-5 text-(--ui-text-dimmed)"
			/>
		</div>
	</div>
</template>
