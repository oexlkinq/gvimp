<script setup lang="ts">
import { ref } from 'vue'
import { AspectRatio, Coordinates, Cropper, Size, SizeRestrictions } from 'vue-advanced-cropper'

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
function updateAspect() {
    stencilProps.value = (forceDefaultAspect.value) ? stencilPropsVariants.forceAspect : stencilPropsVariants.freeAspect
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

function onFileChange(event: Event) {
    const { files } = event.target as HTMLInputElement
    if (!files) {
        return
    }

    if (src.value) {
        URL.revokeObjectURL(src.value)
    }

    src.value = URL.createObjectURL(files[0])
    file.value = files[0]
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
                <input type="file" class="form-control" @change="onFileChange">

                <hr>

                <div class="form-check">
                    <input class="form-check-input" type="checkbox" value="" id="forceDefaultAspectCheck"
                        v-model="forceDefaultAspect" @change="updateAspect">
                    <label class="form-check-label" for="forceDefaultAspectCheck">
                        force 4:3
                    </label>
                </div>

                <hr>

                <button @click="doit" class="btn btn-primary w-100">
                    <div class="spinner-border spinner-border-sm" role="status" v-show="processing"></div>
                    process
                </button>
                <a ref="linkEl" class="d-none">download link</a>
            </div>
        </div>
    </main>
</template>
