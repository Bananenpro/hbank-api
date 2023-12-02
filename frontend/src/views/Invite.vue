<template>
  <div class="page form-page">
    <teleport to="#app">
      <div
        v-if="showInviteDialog"
        class="dialog-bg"
        @click="showInviteDialog = false"
      ></div>
      <div v-if="showInviteDialog" class="dialog">
        <img
          @click="showInviteDialog = false"
          class="dialog-close-btn clickable"
          src="@/assets/close.svg"
          alt="X"
        />
        <h3 class="dialog-title">{{ $t("invite") }}</h3>
        <form @submit.prevent="invite">
          <span class="invalid-form-field-indicator">{{
            validMessage ? "" : "!"
          }}</span
          ><label class="label-next-to-indicator" for="message">{{
            $t("message")
          }}</label>
          <textarea
            type="text"
            name="message"
            v-model="message"
            id="message"
            rows="7"
          ></textarea>

          <button
            type="submit"
            class="btn"
            :disabled="!validMessage || loadingInvitation"
          >
            {{ loadingInvitation ? $t("loading") : $t("invite") }}
          </button>
        </form>
      </div>
    </teleport>

    <input
      class="search"
      v-model="searchInput"
      type="text"
      :placeholder="$t('placeholders.search')"
    />
    <div ref="list">
      <div class="user card" v-for="user in users" :key="user.id">
        <ProfilePicture
          class="profile-picture"
          :user-id="user.id"
        />
        <p class="name">{{ user.name }}</p>
        <img
          @click="inviteBtnPressed(user.id)"
          class="clickable invite-btn"
          :id="'invite-btn-' + user.id"
          src="@/assets/invitation-light.png"
          alt="+"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth, config } from "@/api";
import ProfilePicture from "@/components/ProfilePicture.vue";

interface User {
  id: string;
  name: string;
}

export default defineComponent({
  name: "Invite",
  components: {
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
      showInviteDialog: false,
      minDescLength: 0,
      maxDescLength: 0,
    };
  },
  async beforeCreate() {
    this.minDescLength = (await config()).minDescriptionLength
    this.maxDescLength = (await config()).maxDescriptionLength
  },
  computed: {
    validMessage(): boolean {
      return (
        this.message.length <= this.maxDescLength &&
        this.message.length >= this.minDescLength
      );
    },
  },
  methods: {
    inviteBtnPressed(id: string) {
      this.id = id;
      this.showInviteDialog = true;
    },
    async invite() {
      if (!this.loadingInvitation) {
        this.loadingInvitation = true;
        const userId = await auth();
        if (userId) {
          try {
            const res = await api.post(`/group/${this.groupId}/invitation`, {
              message: this.message,
              userId: this.id,
            });
            if (!res.data.success) {
              console.error(res.data.message);
            }
            const btn = document.getElementById("invite-btn-" + this.id);
            if (btn) {
              btn.classList.add("invited");
              btn.classList.remove("clickable");
            }
            this.showInviteDialog = false;
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
        this.loadingInvitation = false;
      }
    },
    async getMemberIds(): Promise<string[]> {
      const userId = await auth();
      if (userId) {
        try {
          let count = 100;
          const memberIds = [] as string[];
          while (memberIds.length < count) {
            const res = await api.get(
              `/group/${this.groupId}/member?pageSize=100&includeSelf=true`
            );
            if (!res.data.success) {
              console.error(res.data.message);
              return [];
            }
            count = res.data.count;
            for (let i = 0; i < res.data.users.length; i++) {
              memberIds.push(res.data.users[i].id);
            }
          }
          return memberIds;
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
      return [];
    },
    async loadUsers() {
      if (
        !this.loadingUsers &&
        this.users.length >= this.page * this.pageSize
      ) {
        this.loadingUsers = true;
        const userId = await auth();
        if (userId) {
          if (this.memberIds.length === 0) {
            this.memberIds = await this.getMemberIds();
          }

          const res = await api.get(
            `/user?search=${this.searchInput}&page=${this.page}&pageSize=${
              this.pageSize
            }&exclude=${userId}${
              this.memberIds.length > 0 ? "," : ""
            }${this.memberIds.join()}`
          );
          if (!res.data.success) {
            console.error(res.data.message);
            this.loadingUsers = false;
            return;
          }
          for (let i = 0; i < res.data.users.length; i++) {
            this.users.push({
              id: res.data.users[i].id,
              name: res.data.users[i].name,
            });
          }
          this.page++;
        }
        this.loadingUsers = false;
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
.profile-picture {
  border-radius: 100%;
  width: 32px;
  height: 32px;
}
.invite-btn {
  height: 20px;
  background-color: var(--button-bg-color);
  padding: 6px 12px;
  border-radius: 10px;
}
.invited {
  background-color: var(--color-green);
}

@media screen and (min-width: 470px) {
  .user {
    padding: 10px 2%;
  }
}
</style>
