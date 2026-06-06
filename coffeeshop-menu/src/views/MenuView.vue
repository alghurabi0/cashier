<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useMenu } from '../composables/useMenu'
import { useCart } from '../composables/useCart'
import CategoryBar from '../components/CategoryBar.vue'
import MenuItemCard from '../components/MenuItemCard.vue'

const props = defineProps<{
  tableNumber: string
}>()

const emit = defineEmits<{
  'open-cart': []
}>()

const { categories, menuItems, isLoading, loadAll } = useMenu()
const { itemCount, addItem } = useCart()

const selectedCategoryId = ref<string | null>(null)

const selectedCategoryName = computed(() => {
  if (!selectedCategoryId.value) return 'القائمة الكاملة'
  return categories.value.find(c => c.id === selectedCategoryId.value)?.name_ar ?? 'القائمة'
})

const filteredItems = computed(() => {
  if (!selectedCategoryId.value) return menuItems.value
  return menuItems.value.filter(i => i.category_id === selectedCategoryId.value)
})

onMounted(() => {
  loadAll()
})

function onAddItem(menuItemId: string, nameAr: string, price: number) {
  addItem(menuItemId, nameAr, price)
}
</script>

<template>
  <div class="menu-view" dir="rtl">

    <section class="hero">
      <img class="hero-bg" src="https://images.unsplash.com/photo-1501339847302-ac426a4a7cbb?w=800&q=80" alt="coffee shop" />
      <div class="hero-overlay" />
      <div class="hero-content">
        <div class="hero-logo">NJ</div>
        <h1 class="hero-title">NJ Coffee</h1>
        <p class="hero-tagline">تجربة قهوة لا تُنسى</p>
        <div class="hero-table-badge">
          <span>🪑</span>
          <span>طاولة {{ tableNumber }}</span>
        </div>
      </div>
    </section>

    <CategoryBar
      :categories="categories"
      :selected-id="selectedCategoryId"
      @select="selectedCategoryId = $event"
    />

    <main class="menu-content">
      <div class="section-header" v-if="!isLoading">
        <span class="section-deco">—</span>
        <span class="section-title">{{ selectedCategoryName }}</span>
        <span class="section-deco">—</span>
      </div>

      <div v-if="isLoading" class="loading-state">
        <div class="loading-ring" />
        <span>جاري التحميل...</span>
      </div>

      <div v-else-if="filteredItems.length > 0" class="items-list">
        <MenuItemCard
          v-for="item in filteredItems"
          :key="item.id"
          :item="item"
          @add="onAddItem"
        />
      </div>

      <div v-else class="empty-state">
        <span>🍃</span>
        <p>لا توجد منتجات في هذا القسم</p>
      </div>
    </main>

    <Transition name="fab">
      <button v-if="itemCount > 0" class="fab-cart" @click="emit('open-cart')">
        <span>🛒</span>
        <span class="fab-label">عرض السلة</span>
        <span class="fab-count">{{ itemCount }}</span>
      </button>
    </Transition>

  </div>
</template>

<style scoped>
.menu-view {
  min-height: 100dvh;
  background: var(--color-bg);
  color: var(--color-text);
  font-family: var(--font-family);
  display: flex;
  flex-direction: column;
  padding-bottom: 100px;
}

.hero {
  position: relative;
  height: 230px;
  overflow: hidden;
  flex-shrink: 0;
}

.hero-bg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 40%;
  filter: brightness(0.55);
}

.hero-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to bottom, rgba(13,13,13,0.3) 0%, rgba(13,13,13,0.55) 60%, rgba(13,13,13,0.95) 100%);
}

.hero-content {
  position: relative;
  z-index: 2;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  text-align: center;
  padding: 0 20px;
}

.hero-logo {
  width: 54px;
  height: 54px;
  border-radius: 50%;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  font-weight: 900;
  color: #0d0d0d;
  box-shadow: 0 0 24px rgba(201,168,76,0.55);
  letter-spacing: 1px;
  margin-bottom: 2px;
}

.hero-title {
  font-size: 1.75rem;
  font-weight: 900;
  color: #c9a84c;
  letter-spacing: 2px;
  margin: 0;
  line-height: 1.1;
}

.hero-tagline {
  font-size: 0.88rem;
  color: rgba(240,230,211,0.78);
  margin: 0;
}

.hero-table-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  padding: 6px 16px;
  background: rgba(201,168,76,0.15);
  border: 1px solid rgba(201,168,76,0.4);
  border-radius: 50px;
  font-size: 0.8rem;
  font-weight: 700;
  color: #c9a84c;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 16px 20px 6px;
}

.section-title {
  font-size: 1.05rem;
  font-weight: 800;
  color: var(--color-text);
}

.section-deco {
  color: var(--color-accent);
  opacity: 0.8;
}

.menu-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.items-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 8px 14px 16px;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  padding: 60px 0;
  color: var(--color-text-muted);
}

.loading-ring {
  width: 42px;
  height: 42px;
  border: 3px solid rgba(201,168,76,0.2);
  border-top-color: #c9a84c;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 60px 20px;
  color: var(--color-text-muted);
  text-align: center;
  font-size: 2rem;
}

.fab-cart {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 15px 28px;
  background: linear-gradient(135deg, #c9a84c, #e6c56a);
  color: #0d0d0d;
  border: none;
  border-radius: 50px;
  font-family: var(--font-family);
  font-size: 1rem;
  font-weight: 800;
  cursor: pointer;
  box-shadow: 0 8px 32px rgba(201,168,76,0.5);
  z-index: 100;
  white-space: nowrap;
}

.fab-cart:active { transform: translateX(-50%) scale(0.96); }
.fab-label { flex: 1; }

.fab-count {
  background: #0d0d0d;
  color: #c9a84c;
  min-width: 26px;
  height: 26px;
  border-radius: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 900;
  padding: 0 6px;
}

.fab-enter-active, .fab-leave-active { transition: opacity 0.25s ease, transform 0.3s ease; }
.fab-enter-from, .fab-leave-to { opacity: 0; transform: translateX(-50%) translateY(60px); }
</style>