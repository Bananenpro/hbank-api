<template>
<img ref="image">
</template>

<script lang="ts">
import {defineComponent} from 'vue'
import ResizeObserver from 'resize-observer-polyfill';
import api from '@/api'

export default defineComponent({
    name: "GroupPicture",
    props: {
        groupId: {
            type: String,
            required: true
        },
        id: {
            type: String,
            required: true
        },
    },
    data() {
        return {
            ro: null as ResizeObserver | null,
        }
    },
    methods: {
        onResize() {
            const img = this.$refs.image as HTMLImageElement
            img.src = api.defaults.baseURL + "group/" + this.groupId + "/picture" + "?id=" + this.id + "&size=" + this.pictureSizeFromComponentSize(img.offsetWidth)
            matchMedia(`(resolution: ${window.devicePixelRatio}dppx)`).addEventListener("change", this.onResize, { once: true })
        },
        pictureSizeFromComponentSize(componentSize : number) : string {
            if (componentSize * window.devicePixelRatio <= 64) {
                return "tiny"
            }
            if (componentSize * window.devicePixelRatio <= 128) {
                return "small"
            }
            if (componentSize * window.devicePixelRatio <= 256) {
                return "medium"
            }
            if (componentSize * window.devicePixelRatio <= 512) {
                return "large"
            }

            return "huge"
        }
    },
    mounted() {
        if (this.$refs.image) {
            this.ro = new ResizeObserver(this.onResize)
            this.ro.observe(this.$refs.image as Element)
        }
        this.onResize()
    },
    beforeUnmount() {
        if (this.ro && this.$refs.image) {
            this.ro.unobserve(this.$refs.image as Element)
        }
    },
    watch: {
        url() {
            this.onResize()
        },
        id() {
            this.onResize()
        }
    }
})
</script>

<style scoped>

</style>
