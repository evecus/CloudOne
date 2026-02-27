<template>
  <div class="cm-editor-wrap" :class="{ 'cm-focused': isFocused }">
    <div class="cm-lang-badge" v-if="langLabel">{{ langLabel }}</div>
    <div ref="editorEl" class="cm-editor-container"></div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount, shallowRef } from 'vue'
import { EditorView, keymap, lineNumbers, highlightActiveLineGutter, highlightSpecialChars, drawSelection, dropCursor, rectangularSelection, crosshairCursor, highlightActiveLine } from '@codemirror/view'
import { EditorState, Compartment } from '@codemirror/state'
import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands'
import { syntaxHighlighting, defaultHighlightStyle, bracketMatching, foldGutter, indentOnInput } from '@codemirror/language'
import { javascript } from '@codemirror/lang-javascript'
import { html } from '@codemirror/lang-html'
import { css } from '@codemirror/lang-css'
import { json } from '@codemirror/lang-json'
import { python } from '@codemirror/lang-python'
import { markdown } from '@codemirror/lang-markdown'
import { xml } from '@codemirror/lang-xml'
import { sql } from '@codemirror/lang-sql'
import { java } from '@codemirror/lang-java'
import { rust } from '@codemirror/lang-rust'
import { cpp } from '@codemirror/lang-cpp'
import { php } from '@codemirror/lang-php'
import { go } from '@codemirror/lang-go'

const props = defineProps({
  modelValue: { type: String, default: '' },
  filename:   { type: String, default: '' },
  readonly:   { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])

const editorEl  = ref(null)
const isFocused = ref(false)
const view      = shallowRef(null)
const langCompartment = new Compartment()

// ── 语言检测 ────────────────────────────────────────────
const LANG_MAP = {
  // JavaScript 家族
  js:    { lang: () => javascript(),                    label: 'JavaScript' },
  mjs:   { lang: () => javascript(),                    label: 'JavaScript' },
  cjs:   { lang: () => javascript(),                    label: 'JavaScript' },
  ts:    { lang: () => javascript({ typescript: true }), label: 'TypeScript' },
  tsx:   { lang: () => javascript({ typescript: true, jsx: true }), label: 'TSX' },
  jsx:   { lang: () => javascript({ jsx: true }),       label: 'JSX' },
  vue:   { lang: () => html(),                          label: 'Vue' },
  // 样式
  css:   { lang: () => css(),                           label: 'CSS' },
  scss:  { lang: () => css(),                           label: 'SCSS' },
  sass:  { lang: () => css(),                           label: 'Sass' },
  less:  { lang: () => css(),                           label: 'Less' },
  // 标记语言
  html:  { lang: () => html(),                          label: 'HTML' },
  htm:   { lang: () => html(),                          label: 'HTML' },
  xml:   { lang: () => xml(),                           label: 'XML' },
  svg:   { lang: () => xml(),                           label: 'SVG' },
  // 数据格式
  json:  { lang: () => json(),                          label: 'JSON' },
  jsonc: { lang: () => json(),                          label: 'JSONC' },
  json5: { lang: () => json(),                          label: 'JSON5' },
  yaml:  { lang: () => [],                              label: 'YAML' },
  yml:   { lang: () => [],                              label: 'YAML' },
  toml:  { lang: () => [],                              label: 'TOML' },
  // 文档
  md:        { lang: () => markdown(),                  label: 'Markdown' },
  markdown:  { lang: () => markdown(),                  label: 'Markdown' },
  // 编程语言
  py:    { lang: () => python(),                        label: 'Python' },
  go:    { lang: () => go(),                            label: 'Go' },
  java:  { lang: () => java(),                          label: 'Java' },
  rs:    { lang: () => rust(),                          label: 'Rust' },
  c:     { lang: () => cpp(),                           label: 'C' },
  cpp:   { lang: () => cpp(),                           label: 'C++' },
  cc:    { lang: () => cpp(),                           label: 'C++' },
  h:     { lang: () => cpp(),                           label: 'C/C++ Header' },
  hpp:   { lang: () => cpp(),                           label: 'C++ Header' },
  php:   { lang: () => php(),                           label: 'PHP' },
  // Shell
  sh:    { lang: () => [],                              label: 'Shell' },
  bash:  { lang: () => [],                              label: 'Bash' },
  zsh:   { lang: () => [],                              label: 'Zsh' },
  fish:  { lang: () => [],                              label: 'Fish' },
  // 数据库
  sql:   { lang: () => sql(),                           label: 'SQL' },
}

function getExtInfo(filename) {
  if (!filename) return null
  // 特殊无扩展名文件
  const base = filename.split('/').pop().toLowerCase()
  if (['dockerfile'].includes(base)) return { lang: () => [], label: 'Dockerfile' }
  if (['makefile', 'gnumakefile'].includes(base)) return { lang: () => [], label: 'Makefile' }
  if (!filename.includes('.')) return null
  const ext = filename.split('.').pop().toLowerCase()
  return LANG_MAP[ext] || null
}

const langLabel = ref('')

function getLangExtension(filename) {
  const info = getExtInfo(filename)
  langLabel.value = info?.label || 'Plain Text'
  if (!info) return []
  const ext = info.lang()
  return Array.isArray(ext) ? ext : [ext]
}

// ── 主题：亮色，风格和项目 UI 统一 ─────────────────────
const lightTheme = EditorView.theme({
  '&': {
    fontSize: '13px',
    fontFamily: "'JetBrains Mono', 'Courier New', monospace",
    background: '#FFFFFF',
    color: '#1E293B',
    height: '100%',
    borderRadius: '0 0 10px 10px',
  },
  '.cm-content': {
    padding: '14px 0',
    caretColor: '#3B82F6',
    minHeight: '200px',
  },
  '.cm-scroller': {
    overflow: 'auto',
    lineHeight: '1.65',
  },
  '.cm-line': { padding: '0 16px 0 4px' },
  '.cm-gutters': {
    background: '#F8FAFC',
    color: '#94A3B8',
    border: 'none',
    borderRight: '1px solid #E2E8F0',
    minWidth: '44px',
  },
  '.cm-lineNumbers .cm-gutterElement': {
    padding: '0 10px 0 6px',
    minWidth: '32px',
    textAlign: 'right',
    fontSize: '12px',
  },
  '.cm-activeLineGutter': { background: '#EFF6FF' },
  '.cm-activeLine': { background: '#F0F9FF' },
  '.cm-selectionBackground, ::selection': { background: '#BFDBFE !important' },
  '.cm-cursor': { borderLeftColor: '#3B82F6', borderLeftWidth: '2px' },
  '.cm-matchingBracket': { background: '#BFDBFE', outline: 'none', borderRadius: '2px' },
  '.cm-foldGutter': { width: '16px' },
  '.cm-foldGutter .cm-gutterElement': { padding: '0 2px', cursor: 'pointer' },
}, { dark: false })

// ── 初始化编辑器 ────────────────────────────────────────
onMounted(() => {
  const langExts = getLangExtension(props.filename)

  const state = EditorState.create({
    doc: props.modelValue,
    extensions: [
      lineNumbers(),
      highlightActiveLineGutter(),
      highlightSpecialChars(),
      history(),
      foldGutter(),
      drawSelection(),
      dropCursor(),
      EditorState.allowMultipleSelections.of(true),
      indentOnInput(),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      bracketMatching(),
      rectangularSelection(),
      crosshairCursor(),
      highlightActiveLine(),
      keymap.of([...defaultKeymap, ...historyKeymap, indentWithTab]),
      lightTheme,
      langCompartment.of(langExts),
      EditorView.updateListener.of(update => {
        if (update.docChanged) {
          emit('update:modelValue', update.state.doc.toString())
        }
      }),
      EditorView.focusChangeEffect.of((state, focusing) => {
        isFocused.value = focusing
        return null
      }),
      EditorView.editable.of(!props.readonly),
      EditorView.lineWrapping,
    ],
  })

  view.value = new EditorView({ state, parent: editorEl.value })
})

onBeforeUnmount(() => {
  view.value?.destroy()
})

// 外部 v-model 变更时同步内容（避免循环）
watch(() => props.modelValue, (val) => {
  if (!view.value) return
  const current = view.value.state.doc.toString()
  if (current !== val) {
    view.value.dispatch({
      changes: { from: 0, to: current.length, insert: val }
    })
  }
})

// 文件名变更时切换语言
watch(() => props.filename, (name) => {
  if (!view.value) return
  const langExts = getLangExtension(name)
  view.value.dispatch({
    effects: langCompartment.reconfigure(langExts)
  })
})
</script>

<style scoped>
.cm-editor-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: 1.5px solid var(--blue-100);
  border-radius: 10px;
  overflow: hidden;
  background: white;
  box-shadow: 0 1px 4px rgba(15, 23, 42, 0.05);
  transition: border-color .2s, box-shadow .2s;
  min-height: 0;
  position: relative;
  /* 给 CodeMirror 一个明确的高度锚点 */
  height: 100%;
}
.cm-editor-wrap.cm-focused {
  border-color: var(--blue-400);
  box-shadow: 0 0 0 2px var(--blue-100);
}

/* 语言标签 */
.cm-lang-badge {
  position: absolute;
  top: 8px;
  right: 10px;
  z-index: 10;
  font-family: 'JetBrains Mono', monospace;
  font-size: 10px;
  font-weight: 600;
  color: var(--gray-400);
  background: var(--gray-50);
  border: 1px solid var(--gray-200);
  padding: 2px 7px;
  border-radius: 5px;
  pointer-events: none;
  user-select: none;
}

.cm-editor-container {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
}

/* CodeMirror 的根元素需要填满容器并允许滚动 */
.cm-editor-container :deep(.cm-editor) {
  height: 100%;
  width: 100%;
  flex: 1;
  min-height: 0;
}
.cm-editor-container :deep(.cm-scroller) {
  overflow: auto !important;
  -webkit-overflow-scrolling: touch;
}
/* 文字加粗，提高可读性 */
.cm-editor-container :deep(.cm-content) {
  font-weight: 500;
  -webkit-font-smoothing: auto;
  font-smooth: auto;
}
/* 行号不加粗 */
.cm-editor-container :deep(.cm-gutterElement) {
  font-weight: 400;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .cm-editor-wrap {
    border-radius: 6px;
  }
  .cm-editor-container :deep(.cm-editor) {
    font-size: 14px !important;
  }
  .cm-editor-container :deep(.cm-content) {
    font-size: 14px;
    font-weight: 500;
  }
  /* 移动端行号缩窄 */
  .cm-editor-container :deep(.cm-lineNumbers .cm-gutterElement) {
    min-width: 26px;
    padding: 0 6px 0 4px;
    font-size: 11px;
  }
  .cm-lang-badge {
    font-size: 9px;
    padding: 1px 5px;
    top: 6px;
    right: 8px;
  }
}

@media (max-width: 480px) {
  /* 极小屏隐藏行号，节省横向空间 */
  .cm-editor-container :deep(.cm-gutters) {
    display: none;
  }
  .cm-editor-container :deep(.cm-content) {
    padding-left: 12px;
  }
}
</style>
