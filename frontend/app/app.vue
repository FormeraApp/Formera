<script lang="ts" setup>
// Stores are initialized in the Pinia plugin (00.pina.client.ts)
// No need to initialize here - plugin handles it
</script>

<template>
	<div>
		<NuxtPwaManifest />
		<NuxtLayout>
			<NuxtPage />
		</NuxtLayout>
		<ToastContainer />
	</div>
</template>
