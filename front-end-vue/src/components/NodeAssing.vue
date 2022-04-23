<template>
 <div ref="el">Assing</div>
 <input df-varname type="text" placeholder="set varname" pattern="[a-zA-Z]*">
 <input df-namevalue readonly>
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