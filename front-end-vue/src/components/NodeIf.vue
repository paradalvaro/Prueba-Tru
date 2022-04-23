<template>
 <div ref="el">If-Else</div>
 <input df-condition placeholder="set condition">
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
      //console.log(df._value.getNodeFromId(el.value.parentElement.parentElement.id.slice(5)));
    });
    
    return {
      el
    }
  }
}
</script>