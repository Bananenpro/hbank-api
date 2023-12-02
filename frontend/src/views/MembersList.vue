<template>
  <div class="page form-page">
    <teleport to="#app">
      <ConfirmDialog
        @yes="promoteUser()"
        @close="showPromotionDialog = false"
        :show="showPromotionDialog"
        :description="$t('members-list.promotion-dialog-description')"
      />
    </teleport>

    <input
      class="search"
      v-model="searchInput"
      type="text"
      :placeholder="$t('placeholders.search')"
    />
    <router-link v-if="isAdmin" :to="'/group/' + groupId + '/invite'" id="invite-btn-desktop" class="btn clickable">+ {{ $t("members.invite") }}</router-link>
    <div ref="list">
      <div class="user card" v-for="user in users" :key="user.id">
        <ProfilePicture
          class="profile-picture"
          :user-id="user.id"
        />
        <p class="name" :class="user.isAdmin ? 'admin-name' : ''">
          {{ user.name }}
        </p>
        <img
          v-if="isAdmin && !user.isAdmin"
          @click="
            promoteUserId = user.id;
            showPromotionDialog = true;
          "
          class="clickable promotion-btn"
          src="@/assets/promote-light.svg"
          alt="+"
        />
      </div>
    </div>
    <teleport to="#app">
      <router-link
        v-if="isAdmin"
        :to="'/group/' + groupId + '/invite'"
        id="invite-btn-mobile"
        class="floating-action-btn clickable"
        ><img src="@/assets/add.svg" alt="+"
      /></router-link>
    </teleport>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import ProfilePicture from "@/components/ProfilePicture.vue";

interface User {
  id: string;
  name: string;
  isAdmin: boolean;
}

export default defineComponent({
  name: "MembersList",
  components: {
    ConfirmDialog,
    ProfilePicture,
  },
  data() {
    return {
      users: [] as User[],
      searchInput: "",
      searchTimeout: 0,
      page: 0,
      pageSize: 10,
      groupId: this.$route.params.id,
      message: "",
      memberIds: [] as string[],
      onScrollInterval: 0,
      loadingUsers: false,
      loadingInvitation: false,
      id: "",
      showPromotionDialog: false,
      promoteUserId: "",
      isAdmin: false,
    };
  },
  computed: {
    validMessage(): boolean {
      return this.message.length <= 256;
    },
  },
  methods: {
    async loadUsers() {
      if (
        !this.loadingUsers &&
        this.users.length >= this.page * this.pageSize
      ) {
        this.loadingUsers = true;
        const userId = await auth();
        if (userId) {
          const res = await api.get(
            `/group/${this.groupId}/user?page=${this.page}&pageSize=${this.pageSize}&includeSelf=true&search=${this.searchInput}`
          );
          if (!res.data.success) {
            console.error(res.data.message);
            this.loadingUsers = false;
            return;
          }
          for (let i = 0; i < res.data.users.length; i++) {
            this.users.push({
              id: res.data.users[i].id,
              name: res.data.users[i].id == userId ? this.$t("you") : res.data.users[i].name,
              isAdmin: res.data.users[i].admin,
            });
          }
          this.page++;
        }
        this.loadingUsers = false;
      }
    },
    async promoteUser() {
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.post(`/group/${this.groupId}/admin`, {
            id: this.promoteUserId,
          });
          if (res.data.success) {
            this.users = [];
            this.page = 0;
            this.loadUsers();
          } else {
            console.error(res.data.message);
          }
        } catch (e: any) {
          if (e.response) {
            this.$router.push({
              name: "error",
              query: {
                code: e.response.status,
                message: e.response.data.message,
              },
            });
          } else {
            this.$router.push({ name: "error", query: { code: "offline" } });
          }
        }
      }
    },
    async onScroll(): Promise<boolean> {
      const contentElement = document.getElementById("content");
      const list = this.$refs.list as HTMLElement;

      if (contentElement) {
        const nearBottom =
          contentElement.scrollTop + window.innerHeight >=
          list.offsetHeight * 0.8;
        if (nearBottom) {
          await this.loadUsers();
        }
        return nearBottom;
      }

      return false;
    },
  },
  watch: {
    searchInput: function () {
      clearTimeout(this.searchTimeout);
      this.searchTimeout = setTimeout(() => {
        this.users = [];
        this.page = 0;
        this.loadUsers();
      }, 500);
    },
  },
  async mounted() {
    this.onScrollInterval = setInterval(this.onScroll, 200);

    await this.loadUsers();

    const contentElement = document.getElementById("content");
    if (contentElement) {
      contentElement.addEventListener("scroll", this.onScroll);
    }

    const userId = await auth();
    if (userId) {
      try {
        const res = await api.get(`/group/${this.groupId}`);
        if (!res.data.success) {
          console.error(res.data.message);
        } else {
          this.isAdmin = res.data.admin;
        }
      } catch (e: any) {
        if (e.response) {
          this.$router.push({
            name: "error",
            query: {
              code: e.response.status,
              message: e.response.data.message,
            },
          });
        } else {
          this.$router.push({ name: "error", query: { code: "offline" } });
        }
      }
    }
  },
  unmounted() {
    clearInterval(this.onScrollInterval);
    const contentElement = document.getElementById("content");
    if (contentElement) {
      contentElement.removeEventListener("scroll", this.onScroll);
    }
  },
});
</script>


<style scoped>
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
.admin-name {
  color: var(--color-green);
}
.profile-picture {
  border-radius: 100%;
  width: 32px;
  height: 32px;
}
.promotion-btn {
  height: 20px;
  background-color: var(--button-bg-color);
  padding: 6px 6px;
  border-radius: 100%;
}

#invite-btn-desktop {
  margin-bottom: 1.5vh;
  display: none;
}

@media screen and (min-width: 470px) {
  .user {
    padding: 10px 2%;
  }
}

@media screen and (min-width: 700px){
  #invite-btn-desktop {
    display: inline-block;
  }
  #invite-btn-mobile {
    display: none;
  }
}
</style>
