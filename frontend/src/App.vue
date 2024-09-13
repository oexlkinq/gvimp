<script setup lang="ts">
import { ref } from 'vue'
import { AspectRatio, Coordinates, Cropper, Size, SizeRestrictions } from 'vue-advanced-cropper'
import DropZone from './DropZone.vue';

const cropperInstance = ref<InstanceType<typeof Cropper>>();
const src = ref('')
const file = ref<File>()
const linkEl = ref<HTMLLinkElement>()

const stencilPropsVariants = {
    forceAspect: {
        aspectRatio: 400 / 300,
    },
    freeAspect: {},
}

const stencilProps = ref<Object>(stencilPropsVariants.forceAspect)

const forceDefaultAspect = ref(true)
function updateAspect(force: boolean) {
    stencilProps.value = (force) ? stencilPropsVariants.forceAspect : stencilPropsVariants.freeAspect
    cropperInstance.value?.refresh()
}

function calcSize({ imageSize }: defaultSizeFunctionParam) {
    return {
        width: imageSize.width,
        height: imageSize.height,
    }
}

type defaultSizeFunctionParam = {
    visibleArea: Coordinates,
    imageSize: Size,
    stencilRatio: AspectRatio,
    sizeRestrictions: SizeRestrictions,
}

const processing = ref(false)
async function doit() {
    try {
        processing.value = true

        if (!cropperInstance.value) {
            throw new Error('no cropper instance')
        }
        if (!linkEl.value) {
            throw new Error('no download link element')
        }
        if (!file.value) {
            return
        }

        const { coordinates } = cropperInstance.value.getResult()
        const form = new FormData()

        for (const [key, value] of Object.entries(coordinates)) {
            form.set(key, String(value))
        }
        form.append('img', file.value)

        const resp = await fetch(import.meta.env.BASE_URL + 'api/thumbnail', {
            method: 'post',
            body: form,
        })

        if (!resp.ok) {
            console.error(resp)
            throw new Error('api bad')
        }

        const url = await resp.text()

        const link = linkEl.value
        link.href = url
        link.setAttribute('download', file.value.name.slice(0, 16) + '_thumbnail.jpg')
        link.click()
    } catch (e) {
        throw e
    } finally {
        processing.value = false
    }
}

function changeFile(files: FileList | undefined) {
    if (!files) {
        return
    }

    if (src.value) {
        URL.revokeObjectURL(src.value)
    }

    src.value = URL.createObjectURL(files[0])
    file.value = files[0]
}

function maximize(){
    forceDefaultAspect.value = false
    updateAspect(false)
    cropperInstance.value?.setCoordinates(({imageSize}) => imageSize)
}
</script>

<template>
    <main class="container my-2">
        <div class="row gy-2">
            <div class="col-sm-12 col-lg-9">
                <Cropper ref="cropperInstance" :src="src" :stencil-props="stencilProps" :resize-image="false"
                    :default-size="calcSize" :canvas="false" />
            </div>
            <div class="col">
                <div class="form-check my-1">
                    <input class="form-check-input" type="checkbox" value="" id="forceDefaultAspectCheck"
                        v-model="forceDefaultAspect" @change="updateAspect(forceDefaultAspect)">
                    <label class="form-check-label" for="forceDefaultAspectCheck">
                        force 4:3
                    </label>
                </div>
                <button class="btn btn-primary my-1" @click="maximize">maximize</button>

                <hr>

                <button @click="doit" class="btn btn-primary w-100">
                    <div class="spinner-border spinner-border-sm" role="status" v-show="processing"></div>
                    process
                </button>
                <a ref="linkEl" class="d-none">download link</a>
            </div>
        </div>
    </main>
    <DropZone @file="changeFile" :force="!file"/>
</template>
