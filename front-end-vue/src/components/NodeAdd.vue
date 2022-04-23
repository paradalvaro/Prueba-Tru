<template>
 <div ref="el">Basic Operation</div>
 <select df-opvalue>
   <option value="+">+</option>
   <option value="-">-</option>
   <option value="*">*</option>
   <option value="/">/</option>
 </select>
 <input df-namevalue readonly type="number">
</template>

<script lang="ts">
import { onMounted, ref, getCurrentInstance } from 'vue'

export default {

  setup() {
   
    const el = ref(null);
    const app = getCurrentInstance()
    const df = app.appContext.config.globalProperties.$df

    function getdata() {
      console.log(df._value.getNodeFromId(el.value.parentElement.parentElement.id.slice(5)).inputs);
      let data = df._value.getNodeFromId(el.value.parentElement.parentElement.id.slice(5));      
    }

    onMounted(() => {
      //console.log(el.value.parentElement.parentElement.id);
    });
    
    return {
      el, getdata
    }
  }
}
</script>