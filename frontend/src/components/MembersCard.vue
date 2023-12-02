<template>
  <div class="members-card card">
    <div class="card-header">
      <h3 class="title">{{ $t("group.members") }}</h3>
      <img
        v-if="showAddBtn"
        @click="$router.push('/group/' + groupId + '/invite')"
        class="clickable"
        :src="
          darkTheme
            ? require('@/assets/add-in-card-light.svg')
            : require('@/assets/add-in-card-dark.svg')
        "
        alt="+"
      />
    </div>
    <div class="separator"></div>
    <div class="list" @click="$router.push('/group/' + groupId + '/user')">
      <div class="member" v-for="member in members" :key="member.id">
        <ProfilePicture
          class="profile-picture"
          :user-id="member.id"
        />
        <p class="name">{{ member.name }}</p>
      </div>
      <div class="gradient"></div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";
import tc from "tinycolor2";
import ProfilePicture from "@/components/ProfilePicture.vue";

interface User {
  id: string;
  name: string;
}

export default defineComponent({
  name: "MembersCard",
  components: {
    ProfilePicture,
  },
  props: {
    groupId: {
      type: String,
      required: true,
    },
    showAddBtn: Boolean,
  },
  data() {
    return {
      members: [] as User[],
      memberCount: 0,
    };
  },
  computed: {
    darkTheme(): boolean {
      const bgColor = getComputedStyle(
        document.documentElement
      ).getPropertyValue("--bg-color");

      const color = tc(bgColor);

      return color.isDark();
    },
  },
  methods: {
    async loadMembers() {
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.get(
            "/group/" +
              this.groupId +
              "/member?includeSelf=true&pageSize=" +
              this.memberCount
          );
          if (!res.data.success) {
            console.error(res.data.message);
            return;
          }

          this.members = []
          for (let i = 0; i < res.data.users.length; i++) {
            if (res.data.users[i].id == userId) {
              res.data.users[i].name = this.$t("you");
            }
            this.members.push(res.data.users[i]);
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
    onResize() {
      if (Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0) >= 1150) {
        this.memberCount = 20;
      } else {
        this.memberCount = 4;
      }
    }
  },
  async mounted() {
    window.addEventListener("resize", this.onResize);
    this.onResize()
  },
  unmounted() {
    window.removeEventListener("resize", this.onResize);
  },
  watch: {
    async memberCount(newVal: number, oldVal: number) {
      if (newVal > oldVal)
        await this.loadMembers()
    }
  }
});
</script>


<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
}
.title {
  margin: 0;
}
.separator {
  margin-top: 5px;
  margin-bottom: 10px;
}
.list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 72px;
  max-height: 152px;
  overflow: hidden;
  position: relative;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  -webkit-touch-callout: none;
  user-select: none;
  outline: none !important;
}
.gradient {
  position: absolute;
  top: 15%;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(
    0deg,
    var(--card-bg-color) 0%,
    var(--card-bg-color-transparent) 100%
  );
}
.member {
  display: flex;
  gap: 7px;
}
.profile-picture {
  width: 32px;
  height: 32px;
  border-radius: 100%;
}
.name {
  margin: 0;
  line-height: 32px;
  font-size: 18px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

@media screen and (min-width: 1150px) {
  .list {
    min-height: 90%;
    max-height: 50vh;
  }
  .members-card {
    min-height: 25vh;
  }
}
</style>
