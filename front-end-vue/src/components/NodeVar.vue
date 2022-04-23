<template>
 <div ref="el">Variable</div>
 <input df-namevalue>
</template>

<script lang="ts">
import { onMounted, ref, getCurrentInstance } from 'vue'

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