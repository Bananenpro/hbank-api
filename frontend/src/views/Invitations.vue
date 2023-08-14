<template>
  <div class="page">
    <h2 class="title">{{ $t("pageTitles.invitations") }}</h2>
    <div ref="invitationsList" class="invitations-list">
      <div
        v-for="invitation in invitations"
        :key="invitation.id"
        class="card invitation-card"
      >
        <div class="group-name-container">
          <h3 class="group-name">{{ invitation.groupName }}</h3>
          <img
            @click="deny(invitation.id)"
            class="decline-btn clickable"
            src="@/assets/decline.svg"
            alt="x"
          />
          <img
            @click="accept(invitation.id)"
            class="accept-btn clickable"
            src="@/assets/accept.svg"
            alt="✓"
          />
        </div>
        <div class="separator"></div>
        <div class="invitation-message-container">
          <p class="invitation-message">{{ invitation.message }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";

interface Invitation {
  id: string;
  groupName: string;
  message: string;
}

export default defineComponent({
  name: "Invitations",
  data() {
    return {
      invitations: [] as Invitation[],
      loadingInvitations: false,
      page: 0,
      pageSize: 5,
      onScrollInterval: 0,
    };
  },
  methods: {
    clear() {
      this.invitations = [];
      this.page = 0;
    },
    async accept(id: string) {
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.post("/group/invitation/" + id);
          if (!res.data.success) {
            console.error(res.data.message);
            return;
          }
          this.clear();
          await this.loadNextPage();
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
    async deny(id: string) {
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.delete("/group/invitation/" + id);
          if (!res.data.success) {
            console.error(res.data.message);
            return;
          }
          this.clear();
          await this.loadNextPage();
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
    async loadNextPage() {
      if (
        !this.loadingInvitations &&
        this.invitations.length >= this.page * this.pageSize
      ) {
        this.loadingInvitations = true;
        const userId = await auth();
        if (userId) {
          try {
            const res = await api.get(
              `/group/invitation?page=${this.page}&pageSize=${this.pageSize}`
            );
            if (!res.data.success) {
              this.loadingInvitations = false;
              console.error(res.data.message);
              return;
            }

            for (let i = 0; i < res.data.invitations.length; i++) {
              this.invitations.push({
                id: res.data.invitations[i].id,
                groupName: res.data.invitations[i].groupName,
                message: res.data.invitations[i].invitationMessage,
              });
            }

            this.page++;
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
        this.loadingInvitations = false;
      }
    },
    async onScroll(): Promise<boolean> {
      const contentElement = document.getElementById("content");
      const invitationList = this.$refs.invitationsList as HTMLElement;

      if (contentElement) {
        const nearBottom =
          contentElement.scrollTop + window.innerHeight >=
          invitationList.offsetHeight * 0.8;
        if (nearBottom) {
          await this.loadNextPage();
        }
        return nearBottom;
      }

      return false;
    },
  },
  async mounted() {
    this.onScrollInterval = setInterval(this.onScroll, 200);

    await this.loadNextPage();

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

simus. Quidem debitis non quo rem hic minus pro
<style scoped>
.title {
  text-align: center;
  margin-bottom: 5vh;
  font-size: 28px;
}
.invitations-list {
  margin-top: 3vh;
  display: flex;
  flex-direction: column;
  gap: 2vh;
}
.invitation-card {
  flex-grow: 1;
  flex-basis: 100%;
}
.group-name-container {
  display: flex;
  gap: 5px;
}
.group-name {
  margin: 0;
  line-height: 29px;
  flex-grow: 1;
  font-size: 20px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.decline-btn {
  background-color: var(--color-red);
  width: 15px;
  height: 15px;
  padding: 7px;
  border-radius: 100%;
}
.accept-btn {
  background-color: var(--color-green);
  width: 19px;
  height: 19px;
  padding: 5px;
  border-radius: 100%;
}
.separator {
  margin: 0;
  margin-top: 5px;
  margin-bottom: 10px;
}
.invitation-message {
  margin: 0px;
  font-size: 13px;
  line-height: 17px;
  height: 34px;
  overflow: hidden;
  overflow-wrap: anywhere;
}

@media screen and (min-width: 700px) {
  .invitations-list {
    flex-direction: row;
    justify-content: flex-start;
  }
  .invitation-card {
    max-width: 350px;
    padding: 23px;
  }
}
</style>
