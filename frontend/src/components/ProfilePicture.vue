<template>
<img ref="image">
</template>

<script lang="ts">
import {defineComponent} from 'vue'
import ResizeObserver from 'resize-observer-polyfill';
import { config } from '@/api';

export default defineComponent({
    name: "ProfilePicture",
    props: {
        userId: {
            type: String,
            required: true
        }
    },
    data() {
        return {
            ro: null as ResizeObserver | null,
        }
    },
    methods: {
        async onResize() {
            const img = this.$refs.image as HTMLImageElement
            img.src = (await config()).idProvider + "/user/" + this.userId + "/picture" + "?size=" + this.pictureSizeFromComponentSize(img.offsetWidth)
            matchMedia(`(resolution: ${window.devicePixelRatio}dppx)`).addEventListener("change", this.onResize, { once: true })
        },
        pictureSizeFromComponentSize(componentSize : number) : number {
            if (componentSize * window.devicePixelRatio <= 64) {
                return 64
            }
            if (componentSize * window.devicePixelRatio <= 128) {
                return 128
            }
            if (componentSize * window.devicePixelRatio <= 256) {
                return 256
            }
            return 512
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
