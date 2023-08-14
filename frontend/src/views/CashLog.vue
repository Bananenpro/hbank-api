<template>
  <div class="page form-page">
    <input class="search" v-model="searchInput" type="text" :placeholder="$t('placeholders.search')">
    <div ref="list">
      <div class="entry card clickable" @click="$router.push('/cash/' + entry.id)" v-for="entry in log" :key="entry.id">
        <p class="time">{{entry.time}}</p>
        <p class="title">{{entry.title}}</p>
        <p class="amount" :class="entry.amount.startsWith('-') ? 'negative' : 'positive'">{{entry.amount + $t("currency")}}</p>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import api from  '@/api'
import {auth} from '@/api'

interface Entry {
  id: string
  time: string
  title: string
  amount: string
}

export default defineComponent({
  name: "CashLog",
  data() {
    return {
      log: [] as Entry[],
      searchInput: "",
      searchTimeout: 0,
      page: 0,
      pageSize: 20,
      onScrollInterval: 0,
      loading: false,
    }
  },
  methods: {
    async load() {
      if (!this.loading && this.log.length >= this.page * this.pageSize) {
        this.loading = true
        const userId = await auth()
        if (userId) {
          try {
            const res = await api.get(`/user/cash?page=${this.page}&pageSize=${this.pageSize}&search=${this.searchInput}`)
            if (!res.data.success) {
              console.error(res.data.message)
              this.loading = false
              return
            }

            for (let i = 0; i < res.data.log.length; i++) {
              const time = new Date(res.data.log[i].time * 1000).toLocaleDateString([], {
                day: '2-digit',
                month: '2-digit',
                year: "2-digit"
              })

              let amount = (res.data.log[i].difference / 100).toFixed(2).replace(".", this.$t("decimal"))
              if (!amount.startsWith('-')) {
                amount = "+" + amount
              }

              this.log.push({
                id: res.data.log[i].id,
                time: time,
                title: res.data.log[i].title,
                amount: amount
              })
            }
            this.page++
          } catch (e: any) {
            if (e.response) {
              this.$router.push({name: "error", query: {code: e.response.status, message: e.response.data.message}})
            } else {
              this.$router.push({name: "error", query: {code: "offline"}})
            }
          }
        }
        this.loading = false
      }
    },
    async onScroll() : Promise<boolean> {
      const contentElement = document.getElementById("content")
      const list = this.$refs.list as HTMLElement


      if (contentElement) {
        const nearBottom = contentElement.scrollTop + window.innerHeight >= list.offsetHeight * 0.8
        if (nearBottom) {
          await this.load()
        }
        return nearBottom
      }

      return false
    }
  },
  watch: {
    searchInput: function() {
      clearTimeout(this.searchTimeout)
      this.searchTimeout = setTimeout(() => {
        this.log = []
        this.page = 0
        this.load()
      }, 500)
    },
    bank() {
      this.log = []
      this.page = 0
      this.load()
    }
  },
  async mounted() {
    this.onScrollInterval = setInterval(this.onScroll, 200)

    await this.load()

    const contentElement = document.getElementById("content")
    if (contentElement) {
      contentElement.addEventListener("scroll", this.onScroll)
    }
  },
  unmounted() {
    clearInterval(this.onScrollInterval)
    const contentElement = document.getElementById("content")
    if (contentElement) {
      contentElement.removeEventListener("scroll", this.onScroll)
    }
  }
})
</script>


<style scoped>
.search {
  margin-top: 1vh;
}
.entry {
  display: flex;
  padding: 2.5% 2%;
  gap: 7px;
  margin-bottom: 1.5vh;
}
.title {
  margin: 0;
  flex-grow: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.time {
  margin: 0;
  color: var(--date-in-card-color);
}
.amount {
  margin: 0;
  text-align: right;
}

@media screen and (min-width: 470px) {
  .entry {
    padding: 12px 2%;
  }
}
</style>
