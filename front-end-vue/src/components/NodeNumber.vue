<template>
 <div ref="el">Number</div>
 <input df-namevalue type="number">
</template>

<script lang="ts">
import { onMounted, ref, getCurrentInstance, onUpdated } from 'vue'

export default {

  setup() {
   
   const el = ref(null);
    const app = getCurrentInstance()
    const df = app.appContext.config.globalProperties.$df

    onMounted(() => {
      
    });
    
    return {
      el
    }
  }
}
</script>