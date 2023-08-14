<template>
  <div ref="groupList" class="group-list">
    <div
      v-for="group in groups"
      :key="group.id"
      class="card clickable group-card"
      @click="$router.push('/group/' + group.id)"
    >
      <GroupPicture
        class="group-picture"
        :group-id="group.id"
        :id="group.groupPictureId"
      />
      <div class="group-info">
        <h3 class="group-name">{{ group.name }}</h3>
        <div class="separator"></div>
        <div class="group-description-container">
          <p class="group-description">{{ group.description }}</p>
        </div>
        <div class="separator"></div>
        <div class="member-names-container">
          <p class="member-names">{{ group.memberNames }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";
import GroupPicture from "@/components/GroupPicture.vue";

interface Group {
  id: string;
  name: string;
  description: string;
  groupPictureId: string;
  memberNames: string;
}

export default defineComponent({
  name: "GroupList",
  components: {
    GroupPicture,
  },
  data() {
    return {
      groups: [] as Group[],
      page: 0,
      membersPage: 0,
      pageSize: 5,
      loadingGroups: false,
      baseUrl: api.defaults.baseURL,
      onScrollInterval: 0,
    };
  },
  methods: {
    async loadMemberNames() {
      if (this.page > this.membersPage) {
        const userId = await auth();
        if (userId) {
          for (
            let i = this.membersPage * this.pageSize;
            i < this.groups.length;
            i++
          ) {
            try {
              const res = await api.get(
                `/group/${this.groups[i].id}/member?includeSelf=true&pageSize=10`
              );
              if (!res.data.success) {
                console.error(res.data.message);
                continue;
              }

              for (let j = 0; j < res.data.users.length; j++) {
                let memberName = res.data.users[j].name;
                if (res.data.users[j].id === userId) {
                  memberName = this.$t("you");
                }
                this.groups[i].memberNames += (j > 0 ? ", " : "") + memberName;
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
                this.$router.push({
                  name: "error",
                  query: { code: "offline" },
                });
              }
            }
          }
        }

        this.membersPage = this.page;
      }
    },
    async loadNextPage() {
      if (
        !this.loadingGroups &&
        this.groups.length >= this.page * this.pageSize
      ) {
        this.loadingGroups = true;
        const userId = await auth();
        if (userId) {
          try {
            const res = await api.get(
              `/group?page=${this.page}&pageSize=${this.pageSize}`
            );
            if (!res.data.success) {
              this.loadingGroups = false;
              console.error(res.data.message);
              return;
            }

            for (let i = 0; i < res.data.groups.length; i++) {
              res.data.groups[i].memberNames = "";
              this.groups.push(res.data.groups[i]);
            }

            this.page++;

            await this.loadMemberNames();
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
        this.loadingGroups = false;
      }
    },
    async onScroll(): Promise<boolean> {
      const contentElement = document.getElementById("content");
      const groupList = this.$refs.groupList as HTMLElement;

      if (contentElement) {
        const nearBottom =
          contentElement.scrollTop + window.innerHeight >=
          groupList.offsetHeight * 0.8;
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

<style scoped>
.group-card {
  display: flex;
  height: 13vh;
  margin-bottom: 2vh;
}
.separator {
  margin: 3px 0;
}
.group-info {
  margin-left: 3%;
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-evenly;
  min-width: 0;
}
.group-picture {
  border-radius: 7px;
}
.group-name {
  margin: 0;
  font-size: 24px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.group-description-container {
  flex-grow: 100;
}
.group-description {
  margin: 0;
  overflow: hidden;
  overflow-wrap: anywhere;
  font-size: 12px;
  line-height: 14px;
  max-height: 42px;
}

.member-names {
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
@media screen and (max-height: 800px) {
  .group-description {
    max-height: 28px;
  }
}
@media screen and (max-height: 690px) {
  .group-description {
    max-height: 14px;
  }
}
@media screen and (min-width: 1000px){
  .group-list {
    display: flex;
    flex-direction: row;
    justify-content: flex-start;
    gap: 30px;
    row-gap: 10px;
    column-gap: 30px;
    flex-wrap: wrap;
  }
  .group-card {
    padding: 25px;
    width: 42%;
  }
}

@media screen and (min-width: 1135px){
  .group-card {
    padding: 25px;
    width: 450px;
  }
}
</style>
