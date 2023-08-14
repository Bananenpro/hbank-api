<template>
  <div class="page form-page">
    <ConfirmDialog
      @yes="leave"
      @close="showLeaveDialog = false"
      :show="showLeaveDialog"
      :description="$t('group-settings.leave-description')"
    />
    <ConfirmDialog
      @yes="demote"
      @close="showDemoteDialog = false"
      :show="showDemoteDialog"
      :description="$t('group-settings.demote-description')"
    />
    <ConfirmDialog
      @yes="deleteGroup"
      @close="showDeleteDialog = false"
      :show="showDeleteDialog"
      :description="$t('group-settings.delete-description')"
    />
    <ConfirmDialog
      @yes="removePicture"
      @close="showRemovePictureDialog = false"
      :show="showRemovePictureDialog"
      :description="$t('group-settings.remove-picture-warning')"
    />

    <teleport to="#app">
      <div
        v-if="showEditDescriptionDialog"
        class="dialog-bg"
        @click="showEditDescriptionDialog = false"
      ></div>
      <div v-if="showEditDescriptionDialog" class="dialog">
        <img
          @click="showEditDescriptionDialog = false"
          class="dialog-close-btn clickable"
          src="@/assets/close.svg"
          alt="X"
        />
        <h3 class="dialog-title">{{ $t("edit") }}</h3>
        <form @submit.prevent="changeDescription">
          <span class="invalid-form-field-indicator">{{
            validNewDescription ? "" : "!"
          }}</span
          ><label class="label-next-to-indicator" for="new-description">{{
            $t("description")
          }}</label>
          <textarea
            type="text"
            name="new-description"
            v-model="newDescription"
            id="new-description"
            rows="7"
          ></textarea>

          <button type="submit" class="btn" :disabled="!validNewDescription">
            {{ $t("done") }}
          </button>
        </form>
      </div>
    </teleport>

    <h2 class="title">{{ $t("group-settings.title") }}</h2>

    <div class="picture-container">
      <input
        @change="uploadGroupPicture"
        type="file"
        name="picture-input"
        id="picture-input"
        ref="pictureInput"
      />
      <GroupPicture
        :group-id="id"
        :id="groupPictureId"
        @click="chooseGroupPicture"
        class="group-picture"
        :class="isAdmin && !uploadingGroupPicture ? 'clickable' : ''"
      />
      <img
        v-if="!uploadingGroupPicture && isAdmin"
        @click="showRemovePictureDialog = true"
        class="remove-picture-btn clickable"
        src="@/assets/delete.svg"
      />
    </div>

    <div class="form-error-container">
      <span v-if="uploadingGroupPicture" class="upload-progress"
        >{{ $t("uploading") }} {{ Math.round(uploadProgress * 100) }}%</span
      >
      <span v-if="groupPictureError" class="form-error"
        >! {{ groupPictureError }}</span
      >
    </div>

    <label>{{ $t("name") }}</label>
    <p class="box">{{ name }}</p>

    <div class="edit-lbl-container">
      <label>{{ $t("description") }}</label>
      <img
        v-if="isAdmin"
        @click="showEditDescriptionDialog = true"
        class="edit-btn clickable"
        :src="
          darkTheme
            ? require('@/assets/edit-light.svg')
            : require('@/assets/edit-dark.svg')
        "
      />
    </div>
    <div class="multiline-box-container">
      <p class="multiline-box-text">{{ description }}</p>
    </div>

    <div v-if="name" class="role-setting">
      <p>
        {{ $t("group-settings.member") }}:
        <span :class="isMember ? 'positive' : 'negative'">{{
          isMember ? $t("yes") : $t("no")
        }}</span>
      </p>
      <p
        v-if="isMember"
        @click="showLeaveDialog = true"
        class="clickable remove-role"
      >
        {{ $t("group-settings.leave") }}
      </p>
    </div>

    <div v-if="name" class="role-setting">
      <p>
        {{ $t("group-settings.admin") }}:
        <span :class="isAdmin ? 'positive' : 'negative'">{{
          isAdmin ? $t("yes") : $t("no")
        }}</span>
      </p>
      <p
        v-if="isAdmin && !soleAdmin"
        @click="showDemoteDialog = true"
        class="remove-role clickable"
      >
        {{ $t("group-settings.remove") }}
      </p>
    </div>

    <button
      @click="showDeleteDialog = true"
      v-if="enableDelete"
      class="btn bottom-btn btn-danger"
    >
      {{ $t("delete") }}
    </button>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import api from "@/api";
import { auth, config } from "@/api";
import { pageTitle } from "@/router";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import GroupPicture from "@/components/GroupPicture.vue";
import tc from "tinycolor2";

export default defineComponent({
  name: "GroupSettings",
  components: {
    ConfirmDialog,
    GroupPicture,
  },
  data() {
    return {
      id: "",
      name: "",
      description: "",
      groupPictureId: "",
      isMember: false,
      isAdmin: false,
      showLeaveDialog: false,
      showDemoteDialog: false,
      showDeleteDialog: false,
      showEditDescriptionDialog: false,
      showRemovePictureDialog: false,
      soleAdmin: true,
      enableDelete: false,
      uploadingGroupPicture: false,
      uploadProgress: 0.0,
      groupPictureError: "",
      newDescription: "",
      minDescriptionLength: 0,
      maxDescriptionLength: 0
    };
  },
  async beforeCreate() {
    this.minDescriptionLength = (await config()).minDescriptionLength
    this.maxDescriptionLength = (await config()).maxDescriptionLength
  },
  computed: {
    darkTheme(): boolean {
      const bgColor = getComputedStyle(
        document.documentElement
      ).getPropertyValue("--bg-color");

      const color = tc(bgColor);

      return color.isDark();
    },
    validNewDescription(): boolean {
      return (
        this.newDescription.length >= this.minDescriptionLength &&
        this.newDescription.length <= this.maxDescriptionLength
      );
    },
  },
  mounted() {
    this.loadData();
  },
  methods: {
    async loadData() {
      const userId = await auth();
      if (userId) {
        if (this.$route.params.id) {
          try {
            const res = await api.get("/group/" + this.$route.params.id);
            if (!res.data.success) {
              console.error(res.data.message);
              return;
            }

            this.id = res.data.id;
            this.name = res.data.name;
            this.description = res.data.description;
            this.groupPictureId = res.data.groupPictureId;
            this.isMember = res.data.member;
            this.isAdmin = res.data.admin;

            pageTitle.value = res.data.name;

            if (this.isAdmin) {
              const res = await api.get(
                "/group/" + this.$route.params.id + "/admin"
              );
              if (!res.data.success) {
                console.error(res.data.message);
                return;
              }
              this.soleAdmin = res.data.count == 1;

              if (this.soleAdmin) {
                const res = await api.get(
                  "/group/" + this.$route.params.id + "/member"
                );
                if (!res.data.success) {
                  console.error(res.data.message);
                  return;
                }
                this.enableDelete = res.data.count == (this.isMember ? 1 : 0);
              }
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
      }
    },
    async leave() {
      if (this.isMember) {
        const userId = await auth();
        if (userId) {
          if (this.$route.params.id) {
            try {
              const res = await api.delete(
                "/group/" + this.$route.params.id + "/member"
              );
              if (!res.data.success) {
                console.error(res.data.message);
                return;
              }

              this.isMember = false;

              if (!this.isMember && !this.isAdmin) {
                this.$router.push("/dashboard");
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
      }
    },
    async demote() {
      if (this.isAdmin) {
        const userId = await auth();
        if (userId) {
          if (this.$route.params.id) {
            try {
              const res = await api.delete(
                "/group/" + this.$route.params.id + "/admin"
              );
              if (!res.data.success) {
                console.error(res.data.message);
                return;
              }

              this.isAdmin = false;

              if (!this.isMember && !this.isAdmin) {
                this.$router.push("/dashboard");
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
      }
    },
    async deleteGroup() {
      await this.leave();
      await this.demote();
    },
    async changeDescription() {
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.put("/group/" + this.$route.params.id, {
            description: this.newDescription,
          });
          if (!res.data.success) {
            console.error(res.data.message);
          }
          this.description = res.data.description;
          this.showEditDescriptionDialog = false;
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
    chooseGroupPicture() {
      if (this.isAdmin && !this.uploadingGroupPicture) {
        (this.$refs.pictureInput as HTMLElement).click();
      }
    },
    async uploadGroupPicture(event: any) {
      this.uploadingGroupPicture = true;
      this.groupPictureError = "";
      this.uploadProgress = 0;
      const userId = await auth();
      if (userId) {
        try {
          const formData = new FormData();
          formData.append("groupPicture", event.target.files[0]);
          const res = await api.post(
            "/group/" + this.$route.params.id + "/picture",
            formData,
            {
              headers: {
                "content-type": "multipart/form-data",
              },
              onUploadProgress: (progressEvent: any) => {
                this.uploadProgress =
                  progressEvent.loaded / progressEvent.total;
              },
            }
          );
          if (!res.data.success) {
            this.groupPictureError = res.data.message;
          } else {
            this.groupPictureId = res.data.id;
          }
        } catch (e: any) {
          if (e && e.response && e.response.data) {
            this.groupPictureError = e.response.data.message;
          }
        }
      }
      this.uploadingGroupPicture = false;
    },
    async removePicture() {
      this.groupPictureError = "";
      const userId = await auth();
      if (userId) {
        try {
          const res = await api.delete(
            "/group/" + this.$route.params.id + "/picture"
          );
          if (!res.data.success) {
            this.groupPictureError = res.data.message;
          } else {
            this.groupPictureId = res.data.id;
          }
        } catch (e: any) {
          if (e && e.response && e.response.data) {
            this.groupPictureError = e.response.data.message;
          }
        }
      }
    },
  },
  watch: {
    showEditDescriptionDialog() {
      this.newDescription = this.description;
    },
  },
});
</script>


<style scoped>
.title {
  margin-top: 1vh;
  margin-bottom: 3vh;
  text-align: center;
  font-size: 28px;
}
.box,
.multiline-box-container {
  margin-bottom: 2vh;
  min-height: 18px;
}
.picture-container {
  width: 50vw;
  max-width: 300px;
  height: 50vw;
  max-height: 300px;
  margin-left: auto;
  margin-right: auto;
  margin-bottom: 2vh;
  position: relative;
}
.group-picture {
  position: absolute;
  width: 100%;
  top: 0;
  left: 0;
  border-radius: 7px;
}
.remove-picture-btn {
  position: absolute;
  width: 20px;
  bottom: 8px;
  right: 8px;
}
.multiline-box-container {
  margin-bottom: 3vh;
}
.role-setting {
  display: flex;
  justify-content: space-between;
  margin-bottom: 2vh;
}
.remove-role {
  color: var(--color-red);
  text-decoration: underline;
}
.add-role {
  color: var(--color-green);
  text-decoration: underline;
}
p {
  margin: 0;
}
#picture-input {
  display: none;
}

.upload-progress {
  display: inline-block;
  font-size: 16px;
  margin-top: 0.5vh;
  margin-bottom: 1vh;
}

.form-error {
  margin-top: 0.5vh;
  margin-bottom: 1vh;
}
</style>
