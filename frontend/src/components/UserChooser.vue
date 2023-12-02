<template>
  <div class="user-chooser" ref="userChooser">
    <teleport to="#app">
    <div @click="selectorVisible = false" class="background" v-show="selectorVisible"></div>
    <div v-if="selectorVisible" class="selector">
      <h2>{{$t("choose")}}</h2>
      <img @click="hideSelector" class="dialog-close-btn clickable" src="@/assets/close.svg" alt="X">
      <input class="search" v-model="searchInput" type="text" :placeholder="$t('placeholders.search')">
      <div ref="list">
        <div v-if="showBank" class="user card">
          <img class="profile-picture" :src="darkTheme ? require('@/assets/bank-light.svg') : require('@/assets/bank-dark.svg')">
          <p class="name">{{$t("bank")}}</p>
          <img @click="select('bank', $t('bank'))" class="clickable select-btn" src="@/assets/select-arrow-light.svg" alt="->">
        </div>
        <div class="user card" v-for="user in users" :key="user.id">
          <ProfilePicture class="profile-picture" :user-id="user.id"/>
          <p class="name">{{user.name}}</p>
          <img @click="select(user.id, user.name)" class="clickable select-btn" src="@/assets/select-arrow-light.svg" alt="->">
        </div>
      </div>
    </div>
    </teleport>

    <span class="invalid-form-field-indicator">{{userId ? "" : "!"}}</span><label class="label-next-to-indicator">{{label}}</label>
    <span v-if="!userId" @click="showSelector" class="btn btn-sm">{{$t("choose")}}</span>
    <div v-if="userId" class="selected-user">
      <img v-if="userId == 'bank'" class="selected-profile-picture" :src="darkTheme ? require('@/assets/bank-light.svg') : require('@/assets/bank-dark.svg')">
      <ProfilePicture v-if="userId != 'bank'" class="selected-profile-picture" :user-id="userId"/>
      <p class="selected-name">{{userName}}</p>
      <img @click="showSelector" class="clickable edit-btn" :src="darkTheme ? require('@/assets/edit-light.svg') : require('@/assets/edit-dark.svg')" alt="edit">
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import api from '@/api'
import {auth} from '@/api'
import tc from 'tinycolor2'
import ProfilePicture from '@/components/ProfilePicture.vue'

interface User {
  id: string
  name: string
}

export default defineComponent({
  name: "UserChooser",
  components: {
    ProfilePicture
  },
  props: {
    label: String,
    url: {
      required: true,
      type: String
    },
    showBank: Boolean,
    includeSelf: Boolean
  },
  data() {
    return {
      userId: "",
      userName: "",
      selectorVisible: false,

      users: [] as User[],
      searchInput: "",
      searchTimeout: 0,
      page: 0,
      pageSize: 10,
      onScrollInterval: 0,
      loading: false,
    }
  },
  computed: {
    darkTheme() : boolean {
      const bgColor = getComputedStyle(document.documentElement).getPropertyValue('--bg-color')

      const color = tc(bgColor);

      return color.isDark()
    },
  },
  methods: {
    select(id: string, name: string) {
      this.userId = id
      this.userName = name
      this.$emit("selected", id)
      this.hideSelector()
    },
    async showSelector() {
      this.selectorVisible = true
      this.onScrollInterval = setInterval(this.onScroll, 200)

      await this.loadUsers()

      const userChooser = this.$refs.userChooser as HTMLElement
      if (userChooser) {
        userChooser.addEventListener("scroll", this.onScroll)
      }
    },
    hideSelector() {
      clearInterval(this.onScrollInterval)
      const userChooser = this.$refs.userChooser as HTMLElement
      if (userChooser) {
        userChooser.removeEventListener("scroll", this.onScroll)
      }
      this.selectorVisible = false
    },
    async loadUsers() {
      if (!this.loading && this.users.length >= this.page * this.pageSize) {
        this.loading = true
        const userId = await auth()
        if (userId) {
          const res = await api.get(this.url + `?includeSelf=${this.includeSelf}&exclude=${!this.includeSelf ? userId : ''}&search=${this.searchInput}&page=${this.page}&pageSize=${this.pageSize}`)
          if (!res.data.success) {
            console.error(res.data.message)
            this.loading = false
            return
          }
          for (let i = 0; i < res.data.users.length; i++) {
            this.users.push({
              id: res.data.users[i].id,
              name: res.data.users[i].name,
            })
          }
          this.page++
        }
        this.loading = false
      }
    },
    async onScroll() : Promise<boolean> {
      if (this.selectorVisible) {
        const userChooser = this.$refs.userChooser as HTMLElement
        const list = this.$refs.list as HTMLElement


        const nearBottom = userChooser.scrollTop + window.innerHeight >= list.offsetHeight * 0.8
        if (nearBottom) {
          await this.loadUsers()
        }
        return nearBottom
      }
      return false
    }
  },
  watch: {
    searchInput() {
      clearTimeout(this.searchTimeout)
      this.searchTimeout = setTimeout(() => {
        this.users = []
        this.page = 0
        this.loadUsers()
      }, 500)
    },
    showBank() {
      if (!this.showBank && this.userId == 'bank') {
        this.userId = ''
        this.userName = ''
        this.$emit("selected", '')
      }
    },
    async includeSelf() {
      const userId = await auth()
      if (this.userId == userId) {
        this.userId = ''
        this.userName = ''
        this.$emit("selected", '')
      }
      this.users = []
      this.page = 0
      this.loadUsers()
    }
  },
  unmounted() {
    this.hideSelector()
  },
  emits: ["selected"]
})
</script>


<style scoped>
h2 {
  text-align: center;
  font-size: 28px;
  margin-top: 4vh;
}
label {
  line-height: 28px;
}
.selected-user {
  display: flex;
  flex-grow: 1;
  flex-shrink: 1;
  min-width: 0;
}
.selected-name {
  line-height: 28px;
  font-size: 18px;
  margin: 0;
  margin-left: 7px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}
.selected-profile-picture {
  border-radius: 100%;
  width: 28px;
  height: 28px;
  margin-left: 8px;
}
.edit-btn {
  width: 26px;
  height: 26px;
  margin-top: 1px;
  margin-left: 10px;
}
.selector {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: var(--bg-color);
  padding: 0 2vw;
  overflow-y: auto;
  overflow-x: hidden;
}
.user-chooser {
  margin-bottom: 3vh;
  display: flex;
}
.btn {
  margin-left: 8px;
}

.search {
  margin-top: 1vh;
}
.user {
  display: flex;
  padding: 2%;
  gap: 7px;
  margin-bottom: 1.5vh;
}
.name {
  line-height: 32px;
  margin: 0;
  flex-grow: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.profile-picture {
  border-radius: 100%;
  width: 32px;
  height: 32px;
}
.select-btn {
  height: 20px;
  background-color: var(--button-bg-color);
  padding: 6px 12px;
  border-radius: 10px;
}

.background {
  display: none;
  position: absolute;
  top: 50px;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--dialog-bg-color);
  z-index: 99;
}

@media screen and (max-height: 760px) {
  .user-chooser{
    margin-bottom: 2vh;
  }
}

@media screen and (max-height: 760px) {
  .user-chooser{
    margin-bottom: 1vh;
  }
}

@media screen and (min-width: 850px) {
  .selector {
    position: absolute;
    background: var(--card-bg-color);
    color: var(--card-fg-color);
    border: 1px solid var(--separator-color);
    border-radius: 10px;
    top: 10vh;
    bottom: 10vh;
    z-index: 100;
    overflow-y: auto;
    padding-left: 20px;
    padding-right: 20px;
    left: calc(50% - 350px);
    right: calc(50% - 350px);
  }
  .background {
    display: block;
  }
}
</style>
