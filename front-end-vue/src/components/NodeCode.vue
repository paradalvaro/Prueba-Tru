<template>
 <button @click="addInput()">+</button>
 <button @click="removeInput()">-</button>
 <div ref="el">Code Block</div>
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
    
    function addInput(){
        df._value.addNodeInput(el.value.parentElement.parentElement.id.slice(5));
    }
    function removeInput(){
        let id = el.value.parentElement.parentElement.id.slice(5);
        let index = Object.keys(df._value.getNodeFromId(id).inputs).length;
        df._value.removeNodeInput(id,"input_"+index);
    }
    return {
      el,addInput,removeInput
    }
  }
}
</script>