<template>
	<div style="z-index: 9998">
		<NuxtLayout>
			<NuxtPage />
		</NuxtLayout>
		<ToastContainer />
	</div>
</template>

<style lang="scss">
  .page-enter-active,
  .page-leave-active {
    transition: opacity 0.5s ease;
  }
  .page-enter-from,
  .page-leave-to {
    opacity: 0;
  }
</style>