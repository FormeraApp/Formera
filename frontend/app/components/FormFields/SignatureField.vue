<script lang="ts" setup>
withDefaults(
	defineProps<{
		modelValue?: string;
	}>(),
	{
		modelValue: "",
	}
);

const emit = defineEmits<{
	"update:modelValue": [value: string];
	change: [];
}>();

const { t } = useI18n();
const canvasRef = ref<HTMLCanvasElement | null>(null);
const ctx = ref<CanvasRenderingContext2D | null>(null);
const isDrawing = ref(false);

const initCanvas = () => {
	if (!canvasRef.value) return;
	ctx.value = canvasRef.value.getContext("2d");
	if (ctx.value) {
		ctx.value.strokeStyle = "#000";
		ctx.value.lineWidth = 2;
		ctx.value.lineCap = "round";
	}
};

const getEventCoords = (event: MouseEvent | TouchEvent) => {
	if (!canvasRef.value) return { x: 0, y: 0 };
	const rect = canvasRef.value.getBoundingClientRect();
	if ("touches" in event && event.touches[0]) {
		return {
			x: event.touches[0].clientX - rect.left,
			y: event.touches[0].clientY - rect.top,
		};
	}
	const mouseEvent = event as MouseEvent;
	return {
		x: mouseEvent.clientX - rect.left,
		y: mouseEvent.clientY - rect.top,
	};
};

const startDrawing = (event: MouseEvent | TouchEvent) => {
	if (!ctx.value) return;
	isDrawing.value = true;
	const { x, y } = getEventCoords(event);
	ctx.value.beginPath();
	ctx.value.moveTo(x, y);
};

const draw = (event: MouseEvent | TouchEvent) => {
	if (!isDrawing.value || !ctx.value) return;
	const { x, y } = getEventCoords(event);
	ctx.value.lineTo(x, y);
	ctx.value.stroke();
};

const stopDrawing = () => {
	if (isDrawing.value && canvasRef.value) {
		isDrawing.value = false;
		emit("update:modelValue", canvasRef.value.toDataURL());
		emit("change");
	}
};

const clearSignature = () => {
	if (canvasRef.value && ctx.value) {
		ctx.value.clearRect(0, 0, canvasRef.value.width, canvasRef.value.height);
		emit("update:modelValue", "");
		emit("change");
	}
};

onMounted(() => {
	initCanvas();
});
</script>

<template>
	<div class="signature-field">
		<canvas
			ref="canvasRef"
			class="signature-canvas"
			width="400"
			height="150"
			@mousedown="startDrawing"
			@mousemove="draw"
			@mouseup="stopDrawing"
			@mouseleave="stopDrawing"
			@touchstart.prevent="startDrawing"
			@touchmove.prevent="draw"
			@touchend="stopDrawing"
		/>
		<button type="button" class="btn btn-sm btn-secondary" @click="clearSignature">
			<UISysIcon icon="fa-solid fa-eraser" /> {{ t("signature.clear") }}
		</button>
	</div>
</template>

<style scoped>
.signature-field {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.signature-canvas {
	width: 100%;
	max-width: 400px;
	background: white;
	border: 1px solid var(--border);
	border-radius: var(--radius);
	cursor: crosshair;
	touch-action: none;
}
</style>
