<script setup lang="ts">
import { ref } from 'vue';

const props = withDefaults(defineProps<{
        force?: boolean,
    }>(),
    {
        force: false,
    }
)

const emit = defineEmits<{
    file: [files: FileList | undefined],
}>()

const insideness = ref(0)
const show = ref(false)

document.addEventListener('dragenter', () => {
    insideness.value++
    show.value = true
})
document.addEventListener('dragleave', () => {
    insideness.value--

    if (insideness.value === 0) {
        show.value = false
    }
})
document.addEventListener('dragover', (event) => {
    event.preventDefault()
})
document.addEventListener('drop', (event) => {
    event.preventDefault()

    insideness.value = 0
    show.value = false

    emit('file', event.dataTransfer?.files)
})
</script>

<template>
    <div class="dropzone" v-show="props.force || show">
        <h1>Drop file here</h1>
    </div>
</template>

<style scoped>
.dropzone {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    background: #5f55;
    align-items: center;
    justify-content: center;
    border: 5px dashed green;
}
</style>