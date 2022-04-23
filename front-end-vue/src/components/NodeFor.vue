<template>
 <div ref="el">For</div>
 <input df-from type="number" placeholder="From">
 <input df-until type="number" placeholder="Until">
</template>

<script lang="ts">
import { onMounted, ref, getCurrentInstance } from 'vue'

export default {

  setup() {
   
   const el = ref(null);
    const app = getCurrentInstance()
    const df = app.appContext.config.globalProperties.$df

    onMounted(() => {
      //console.log(el.value.parentElement.parentElement.id);
    });
    
    return {
      el
    }
  }
}
</script>