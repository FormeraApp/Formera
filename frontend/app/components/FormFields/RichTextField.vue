<script lang="ts" setup>
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Underline from "@tiptap/extension-underline";
import StarterKit from "@tiptap/starter-kit";
import { EditorContent, useEditor } from "@tiptap/vue-3";

const props = withDefaults(
	defineProps<{
		id?: string;
		modelValue?: string;
		placeholder?: string;
		required?: boolean;
	}>(),
	{
		id: "",
		modelValue: "",
		placeholder: "",
		required: false,
	}
);

const emit = defineEmits<{
	"update:modelValue": [value: string];
	blur: [];
}>();

const { t } = useI18n();

const editor = useEditor({
	content: props.modelValue,
	extensions: [
		StarterKit.configure({
			heading: {
				levels: [1, 2, 3],
			},
		}),
		Placeholder.configure({
			placeholder: props.placeholder || t("richText.placeholder"),
		}),
		Underline,
		Link.configure({
			openOnClick: false,
			HTMLAttributes: {
				class: "rte-link",
			},
		}),
	],
	onUpdate: ({ editor }) => {
		emit("update:modelValue", editor.getHTML());
	},
	onBlur: () => {
		emit("blur");
	},
});

// Watch for external value changes
watch(
	() => props.modelValue,
	(value) => {
		if (editor.value && editor.value.getHTML() !== value) {
			editor.value.commands.setContent(value);
		}
	}
);

onBeforeUnmount(() => {
	editor.value?.destroy();
});

const setLink = () => {
	const previousUrl = editor.value?.getAttributes("link").href;
	const url = window.prompt(t("richText.enterUrl"), previousUrl);

	if (url === null) return;

	if (url === "") {
		editor.value?.chain().focus().extendMarkRange("link").unsetLink().run();
		return;
	}

	editor.value?.chain().focus().extendMarkRange("link").setLink({ href: url }).run();
};
</script>

<template>
	<div class="rich-text-field" :class="{ 'has-required': required }">
		<!-- Toolbar -->
		<div v-if="editor" class="rte-toolbar">
			<div class="toolbar-group">
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('bold') }"
					:title="t('richText.bold')"
					@click="editor.chain().focus().toggleBold().run()"
				>
					<UISysIcon icon="fa-solid fa-bold" />
				</button>
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('italic') }"
					:title="t('richText.italic')"
					@click="editor.chain().focus().toggleItalic().run()"
				>
					<UISysIcon icon="fa-solid fa-italic" />
				</button>
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('underline') }"
					:title="t('richText.underline')"
					@click="editor.chain().focus().toggleUnderline().run()"
				>
					<UISysIcon icon="fa-solid fa-underline" />
				</button>
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('strike') }"
					:title="t('richText.strikethrough')"
					@click="editor.chain().focus().toggleStrike().run()"
				>
					<UISysIcon icon="fa-solid fa-strikethrough" />
				</button>
			</div>

			<div class="toolbar-divider" />

			<div class="toolbar-group">
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('heading', { level: 1 }) }"
					:title="t('richText.heading1')"
					@click="editor.chain().focus().toggleHeading({ level: 1 }).run()"
				>
					H1
				</button>
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('heading', { level: 2 }) }"
					:title="t('richText.heading2')"
					@click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
				>
					H2
				</button>
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('heading', { level: 3 }) }"
					:title="t('richText.heading3')"
					@click="editor.chain().focus().toggleHeading({ level: 3 }).run()"
				>
					H3
				</button>
			</div>

			<div class="toolbar-divider" />

			<div class="toolbar-group">
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('bulletList') }"
					:title="t('richText.bulletList')"
					@click="editor.chain().focus().toggleBulletList().run()"
				>
					<UISysIcon icon="fa-solid fa-list-ul" />
				</button>
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('orderedList') }"
					:title="t('richText.orderedList')"
					@click="editor.chain().focus().toggleOrderedList().run()"
				>
					<UISysIcon icon="fa-solid fa-list-ol" />
				</button>
			</div>

			<div class="toolbar-divider" />

			<div class="toolbar-group">
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('blockquote') }"
					:title="t('richText.quote')"
					@click="editor.chain().focus().toggleBlockquote().run()"
				>
					<UISysIcon icon="fa-solid fa-quote-left" />
				</button>
				<button
					type="button"
					class="toolbar-btn"
					:class="{ active: editor.isActive('link') }"
					:title="t('richText.link')"
					@click="setLink"
				>
					<UISysIcon icon="fa-solid fa-link" />
				</button>
			</div>

			<div class="toolbar-divider" />

			<div class="toolbar-group">
				<button
					type="button"
					class="toolbar-btn"
					:title="t('richText.undo')"
					:disabled="!editor.can().undo()"
					@click="editor.chain().focus().undo().run()"
				>
					<UISysIcon icon="fa-solid fa-rotate-left" />
				</button>
				<button
					type="button"
					class="toolbar-btn"
					:title="t('richText.redo')"
					:disabled="!editor.can().redo()"
					@click="editor.chain().focus().redo().run()"
				>
					<UISysIcon icon="fa-solid fa-rotate-right" />
				</button>
			</div>
		</div>

		<!-- Editor Content -->
		<EditorContent :id="id" :editor="editor" class="rte-content" />

		<!-- Hidden input for form validation -->
		<input v-if="required" type="hidden" :required="required && !modelValue" :value="modelValue" />
	</div>
</template>

<style scoped>
.rich-text-field {
	display: flex;
	flex-direction: column;
	border: 1px solid var(--border);
	border-radius: var(--radius);
	background: var(--surface);
	overflow: hidden;
	transition: border-color 0.2s;
}

.rich-text-field:focus-within {
	border-color: var(--primary);
	box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

/* Toolbar */
.rte-toolbar {
	display: flex;
	flex-wrap: wrap;
	gap: 0.25rem;
	padding: 0.5rem;
	background: var(--background);
	border-bottom: 1px solid var(--border);
}

.toolbar-group {
	display: flex;
	gap: 0.125rem;
}

.toolbar-divider {
	width: 1px;
	margin: 0 0.25rem;
	background: var(--border);
}

.toolbar-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	min-width: 2rem;
	height: 2rem;
	padding: 0 0.375rem;
	font-size: 0.75rem;
	font-weight: 600;
	color: var(--text-secondary);
	cursor: pointer;
	background: transparent;
	border: none;
	border-radius: var(--radius-sm);
	transition: all 0.15s;
}

.toolbar-btn:hover:not(:disabled) {
	color: var(--text);
	background: var(--surface-hover);
}

.toolbar-btn.active {
	color: var(--primary);
	background: rgba(99, 102, 241, 0.1);
}

.toolbar-btn:disabled {
	opacity: 0.4;
	cursor: not-allowed;
}

/* Editor Content */
.rte-content {
	min-height: 150px;
	max-height: 400px;
	overflow-y: auto;
}

.rte-content :deep(.tiptap) {
	padding: 0.75rem 1rem;
	min-height: 150px;
	outline: none;
}

.rte-content :deep(.tiptap p) {
	margin: 0 0 0.5em 0;
}

.rte-content :deep(.tiptap p:last-child) {
	margin-bottom: 0;
}

.rte-content :deep(.tiptap h1),
.rte-content :deep(.tiptap h2),
.rte-content :deep(.tiptap h3) {
	margin: 1em 0 0.5em 0;
	font-weight: 600;
	line-height: 1.3;
}

.rte-content :deep(.tiptap h1) {
	font-size: 1.5rem;
}

.rte-content :deep(.tiptap h2) {
	font-size: 1.25rem;
}

.rte-content :deep(.tiptap h3) {
	font-size: 1.1rem;
}

.rte-content :deep(.tiptap ul),
.rte-content :deep(.tiptap ol) {
	margin: 0.5em 0;
	padding-left: 1.5em;
	color: var(--text);
}

.rte-content :deep(.tiptap ul) {
	list-style-type: disc;
}

.rte-content :deep(.tiptap ol) {
	list-style-type: decimal;
}

.rte-content :deep(.tiptap li) {
	margin: 0.25em 0;
}

.rte-content :deep(.tiptap li::marker) {
	color: var(--text);
}

.rte-content :deep(.tiptap blockquote) {
	margin: 0.5em 0;
	padding: 0.5em 1em;
	border-left: 3px solid var(--primary);
	background: var(--background);
	font-style: italic;
}

.rte-content :deep(.tiptap a),
.rte-content :deep(.rte-link) {
	color: var(--primary);
	text-decoration: underline;
	cursor: pointer;
}

.rte-content :deep(.tiptap a:hover) {
	text-decoration: none;
}

/* Placeholder */
.rte-content :deep(.tiptap p.is-editor-empty:first-child::before) {
	content: attr(data-placeholder);
	float: left;
	height: 0;
	color: var(--text-secondary);
	pointer-events: none;
}

/* Responsive */
@media (max-width: 640px) {
	.rte-toolbar {
		gap: 0.125rem;
		padding: 0.375rem;
	}

	.toolbar-btn {
		min-width: 1.75rem;
		height: 1.75rem;
		font-size: 0.7rem;
	}

	.toolbar-divider {
		margin: 0 0.125rem;
	}
}
</style>
