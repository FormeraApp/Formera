<script lang="ts" setup>
withDefaults(defineProps<{
	icon: string;
	extraClasses?: string;
}>(), {
	extraClasses: "",
});
</script>

<template>
	<ClientOnly>
		<span :key="icon" class="sys-icon-wrapper">
			<i :class="[icon, extraClasses]"></i>
		</span>
	</ClientOnly>
</template>

<style scoped>
.sys-icon-wrapper {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
</style>
