<template>
  <div class="page">
    <div id="top">
      <h2 class="title">{{ $t("dashboard.hello") }}, {{ name }}!</h2>
      <div class="cash">
        <span class="cash-text"
        >{{ $t("dashboard.cash-lbl") }}: {{ cash }}{{ $t("currency") }}</span
        >
        <router-link to="/cash" id="cash-edit-desktop">
          <img
            class="edit-btn clickable"
            :src="
            darkTheme
              ? require('@/assets/edit-light.svg')
              : require('@/assets/edit-dark.svg')
            "
          >
        </router-link>
        <router-link to="/cash" class="btn btn-sm view-cash-btn">{{
          $t("edit")
          }}</router-link>
      </div>
    </div>
    <div
      class="separator"
      :class="invitationCount ? 'above-invitations' : ''"
    ></div>
    <div
      @click="$router.push('/invitations')"
      v-show="invitationCount"
      class="invitations clickable"
    >
      <img
        class="invitation-icon"
        :src="
          darkTheme
            ? require('@/assets/invitation-light.png')
            : require('@/assets/invitation-dark.png')
        "
      />
      <p class="invitation-text">
        {{ $t("dashboard.invitations", invitationCount) }}
      </p>
      <img class="right-arrow" src="@/assets/arrow-right.svg" alt=">" />
    </div>
    <div v-show="invitationCount" class="separator below-invitations"></div>
    <router-link to="/group/create" id="create-group-btn-desktop" class="btn clickable">+ {{ $t("dashboard.create-group") }}</router-link>
    <group-list />
    <teleport to="#app">
      <router-link to="/group/create" id="create-group-btn-mobile" class="floating-action-btn clickable"
        ><img src="@/assets/add.svg" alt="+"
      /></router-link>
    </teleport>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth } from "@/api";
import GroupList from "@/components/GroupsList.vue";
import tc from "tinycolor2";

export default defineComponent({
  name: "Dashboard",
  components: {
    GroupList,
  },
  data() {
    return {
      name: "",
      cash: "",
      invitationCount: 0,
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
  async mounted() {
    const userId = await auth();
    if (userId) {
      try {
        const userRes = await api.get("/user/" + userId);
        if (!userRes.data.success) {
          console.error(userRes.data.message);
          return;
        }

        this.name = userRes.data.name;

        const cashRes = await api.get("/user/cash/current");
        if (!userRes.data.success) {
          console.error(userRes.data.message);
          return;
        }

        this.cash = (cashRes.data.amount / 100.0)
          .toFixed(2)
          .replace(".", this.$t("decimal"));

        const invitationsRes = await api.get("/group/invitation?pageSize=1");
        if (!invitationsRes.data.success) {
          console.error(invitationsRes.data.message);
          return;
        }

        this.invitationCount = invitationsRes.data.count;
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
});
</script>

<style scoped>
.cash {
  display: flex;
  justify-content: space-between;
}
.view-cash-btn {
  float: right;
}
.title {
  margin: 0;
}
.cash-text {
  display: inline-block;
  font-size: 24px;
  margin-top: 2px;
}

.invitations {
  display: flex;
  justify-content: space-between;
  height: 24px;
  gap: 7px;
}

.invitation-text {
  margin: 0;
  line-height: 24px;
  flex-grow: 1;
  font-size: 14px;
}

.right-arrow {
  height: 24px;
}

.above-invitations {
  margin-bottom: 0.5vh;
}
.below-invitations {
  margin-top: 0.5vh;
}

#create-group-btn-desktop {
  margin-bottom: 1.5vh;
  display: none;
}

#cash-edit-desktop {
  display: none;
}

.edit-btn {
  height: 26px;
}

#top {
  margin-top: 1vh;
  display: flex;
  flex-direction: column;
  gap: 2vh;
}

@media screen and (min-width: 500px){
  #cash-edit-desktop {
    display: inline;
  }
  .view-cash-btn {
    display: none;
  }
  .cash {
    justify-content: flex-start;
    gap: 8px;
  }
  .cash-text {
    margin-top: 0;
  }
}

@media screen and (min-width: 700px){
  #create-group-btn-desktop {
    display: inline-block;
  }
  #create-group-btn-mobile {
    display: none;
  }
  #top {
    flex-direction: row;
    justify-content: space-between;
  }
}

@media screen and (min-width: 1000px) {
  #create-group-btn-desktop {
    margin-bottom: 3vh;
  }
}
</style>
