<template>
  <div class="footer">
    <span>© 2021-2023 Julian Hofmann</span>
    <router-link class="clickable" to="/privacy_policy">{{$t("footer.privacy-policy")}}</router-link>
  </div>
</template>

<style scoped>
.footer {
  display: flex;
  justify-content: space-evenly;
  margin-top: 2vh;
  margin-bottom: 1vh;
}

span,
a {
  color: var(--fg-color);
  text-decoration: none;
  font-size: 10px;
}
</style>
