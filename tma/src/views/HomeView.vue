<script setup lang="ts">
import { ref } from 'vue'
const value = ref(false)
</script>

<template>
  <div>
    <h1>Welcome to Quiz Sprint!</h1>
    <p>Start your journey by creating a new quiz.</p>

    <UCheckbox v-model="value" />
    <UCheckbox v-model="value" />
    <UCheckbox v-model="value" />
    <UCheckbox v-model="value" />
    <UCheckbox v-model="value" />
    <UCheckbox v-model="value" />


    <p>Checkbox value: {{ value }}</p>
  </div>
</template>

<style scoped>
/* Additional styles if needed */
</style>
