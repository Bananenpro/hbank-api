import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { auth } from '@/api'
import api from '@/api'
import { ref } from 'vue'
import i18n from '@/i18n'

export const pageTitle = ref("")

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    name: 'home',
    component: () => import(/* webpackChunkName: "group-home" */ '@/views/Home.vue'),
    meta: {
      auth: "false"
    }
  },
  {
    path: '/privacy_policy',
    name: 'privacyPolicy',
    component: () => import(/* webpackChunkName: "group-home" */ '@/views/PrivacyPolicy.vue'),
    meta: {
      pageTitle: "privacy-policy"
    }
  },

  {
    path: '/settings',
    name: 'settings',
    component: () => import(/* webpackChunkName: "group-settings" */ '@/views/Settings.vue'),
    meta: {
      auth: "true",
      pageTitle: "settings"
    }
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: () => import(/* webpackChunkName: "group-dashboard" */ '@/views/Dashboard.vue'),
    meta: {
      auth: "true",
      pageTitle: "dashboard"
    }
  },
  {
    path: '/cash',
    name: 'cash',
    component: () => import(/* webpackChunkName: "group-cash" */ '@/views/Cash.vue'),
    meta: {
      auth: "true",
      pageTitle: "cash"
    }
  },
  {
    path: '/cash/log',
    name: 'cashLog',
    component: () => import(/* webpackChunkName: "group-cash" */ '@/views/CashLog.vue'),
    meta: {
      auth: "true",
      pageTitle: "cash-log"
    }
  },
  {
    path: '/cash/:entryId',
    name: 'cashDetails',
    component: () => import(/* webpackChunkName: "group-cash" */ '@/views/CashLogEntryDetails.vue'),
    meta: {
      auth: "true",
      pageTitle: "cash-log-entry"
    }
  },
  {
    path: '/invitations',
    name: 'invitations',
    component: () => import(/* webpackChunkName: "group-dashboard" */ '@/views/Invitations.vue'),
    meta: {
      auth: "true",
      pageTitle: "invitations"
    }
  },

  {
    path: '/group/:id',
    name: 'group',
    component: () => import(/* webpackChunkName: "group-group" */ '@/views/Group.vue'),
    meta: {
      auth: "true",
    }
  },
  {
    path: '/group/:id/settings',
    name: 'groupSettings',
    component: () => import(/* webpackChunkName: "group-group" */ '@/views/GroupSettings.vue'),
    meta: {
      auth: "true",
    }
  },
  {
    path: '/group/:id/transfer',
    name: 'transfer',
    component: () => import(/* webpackChunkName: "group-transaction" */ '@/views/Transfer.vue'),
    meta: {
      auth: "true",
      pageTitle: "transfer"
    }
  },
  {
    path: '/group/:id/invite',
    name: 'invite',
    component: () => import(/* webpackChunkName: "group-member" */ '@/views/Invite.vue'),
    meta: {
      auth: "true",
      pageTitle: "invite"
    }
  },
  {
    path: '/group/:id/transaction',
    name: 'transactions',
    component: () => import(/* webpackChunkName: "group-transaction" */ '@/views/TransactionLog.vue'),
    meta: {
      auth: "true",
      pageTitle: "transactions"
    }
  },
  {
    path: '/group/:id/transaction/:transactionId',
    name: 'transactionDetails',
    component: () => import(/* webpackChunkName: "group-transaction" */ '@/views/TransactionDetails.vue'),
    meta: {
      auth: "true",
      pageTitle: "transaction-details"
    }
  },
  {
    path: '/group/:id/payment-plan/create',
    name: 'createPaymentPlan',
    component: () => import(/* webpackChunkName: "group-payment-plan" */ '@/views/CreatePaymentPlan.vue'),
    meta: {
      auth: "true",
      pageTitle: "payment-plan"
    }
  },
  {
    path: '/group/:id/payment-plan',
    name: 'paymentPlans',
    component: () => import(/* webpackChunkName: "group-payment-plan" */ '@/views/PaymentPlans.vue'),
    meta: {
      auth: "true",
      pageTitle: "payment-plans"
    }
  },
  {
    path: '/group/:id/payment-plan/:paymentPlanId',
    name: 'paymentPlanDetails',
    component: () => import(/* webpackChunkName: "group-payment-plan" */ '@/views/PaymentPlanDetails.vue'),
    meta: {
      auth: "true",
      pageTitle: "payment-plan-details"
    }
  },
  {
    path: '/group/:id/payment-plan/:paymentPlanId/update',
    name: 'updatePaymentPlan',
    component: () => import(/* webpackChunkName: "group-payment-plan" */ '@/views/UpdatePaymentPlan.vue'),
    meta: {
      auth: "true",
      pageTitle: "payment-plan"
    }
  },
  {
    path: '/group/:id/user',
    name: 'membersList',
    component: () => import(/* webpackChunkName: "group-member" */ '@/views/MembersList.vue'),
    meta: {
      auth: "true",
      pageTitle: "members"
    }
  },
  {
    path: '/group/create',
    name: 'createGroup',
    component: () => import(/* webpackChunkName: "group-create-group" */ '@/views/CreateGroup.vue'),
    meta: {
      auth: "true",
      pageTitle: "group-create"
    }
  },
  {
    path: '/error',
    name: "error",
    component: () => import(/* webpackChunkName: "group-error" */ '@/views/Error.vue')
  },
  {
    path: '/:pathMatch(.*)*',
    component: () => import(/* webpackChunkName: "group-error" */ '@/views/Error.vue')
  },
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  scrollBehavior() {
    const content = document.getElementById("content") as HTMLElement
    content.scrollTop = 0
  },
  routes
})

router.beforeEach(async (to) => {
  if (to.meta.auth) {
    const isAuth = await auth() !== ""
    if (to.meta.auth === "true" && !isAuth) {
      window.location.href = api.defaults.baseURL + "auth/login?redirect=" + encodeURIComponent(to.fullPath)
      return
    } else if (to.meta.auth === "false" && isAuth) {
      return { path: "/dashboard" }
    }
  }

  let title = to.meta.pageTitle ? "pageTitles." + to.meta.pageTitle as string : ""

  if (title) {
    title = i18n.global.t(title)
  }

  pageTitle.value = title
})

export default router
