<template>  
  <div class="wrapper">
    <aside class="col"> 
        <ul style="list-style-type:none;">            
            <li><button class="button-88" role="button" @click="generateNode('Number')"> Number</button></li>
            <li><button class="button-88" role="button" @click="generateNode('Var')"> Variable</button></li>
            <li><button class="button-88" role="button" @click="generateNode('Add')" >Basic Binary Operations</button></li>            
            <li><button class="button-88" role="button" @click="generateNode('Assing')"> Assing</button></li>
            <li><button class="button-88" role="button" @click="generateNode('If')"> If-Else</button></li>
            <li><button class="button-88" role="button" @click="generateNode('For')"> For</button></li>
            <li><button class="button-88" role="button" @click="generateNode('Code')"> Code</button></li>                        
            <li><input id="exportName" value="Program Name"/><button class="button-88" role="button" @click="exportDataWithExecute('false')"> Export</button></li>
            <li><select id="selectImport"/><button class="button-88" role="button" @click="importData()"> Import</button></li>            
            <li><button class="button-88" role="button" @click="exportDataWithExecute('true')"> Execute</button></li>
            <li><button class="button-88" role="button" @click="clear()"> Clear</button></li>            
            <li><label for="codetoexecute">Python Code</label><br/><textarea id="codetoexecute" rows="5" cols="30"></textarea></li>
            <li><label for="result">Result</label><br/><textarea id="result" rows="5" cols="30"></textarea></li>
        </ul>
    </aside>          
    <div class="col-right">        
        <div id="drawflow"></div>
    </div>
  </div> 
</template>

<script lang="ts">
/*eslint-disable */
import Drawflow from 'Drawflow'
import styleDrawflow from 'drawflow/dist/drawflow.min.css'
import { onMounted, shallowRef, h, getCurrentInstance, render, inject, ref } from 'vue'
import NodeAdd from './NodeAdd.vue'
import NodeNumber from './NodeNumber.vue'
import NodeAssing from './NodeAssing.vue'
import NodeIf from './NodeIf.vue'
import NodeFor from './NodeFor.vue'
import NodeCode from './NodeCode.vue'
import NodeVar from './NodeVar.vue'

export default {
  name: 'drawflow',
  setup() {
    //const el = ref(null);
    const editor = shallowRef({})
    const Vue = { version: 3, h, render };
    const internalInstance = getCurrentInstance()
    internalInstance.appContext.app._context.config.globalProperties.$df = editor;
   
    function exportEditor() {
      alert(JSON.stringify(editor.value.export()));
    }    

    onMounted(() => {
        
      loadPrograms();

      const id = document.getElementById("drawflow");

      editor.value = new Drawflow(id, Vue, internalInstance.appContext.app._context);
      editor.reroute = true;
      editor.reroute_fix_curvature = true;
      editor.value.start();

      const props = { name: "hey" };
      const options = {};
      //editor.value.registerNode('NodeClick', NodeClick, props, options);
      editor.value.registerNode('NodeAdd', NodeAdd, {}, {});
      editor.value.registerNode('NodeNumber', NodeNumber, {}, {});
      editor.value.registerNode('NodeAssing', NodeAssing, {}, {});
      editor.value.registerNode('NodeIf', NodeIf, {}, {});
      editor.value.registerNode('NodeFor', NodeFor, {}, {});
      editor.value.registerNode('NodeCode', NodeCode, {}, {});
      editor.value.registerNode('NodeVar', NodeVar, {}, {});


      //CheckNumberZeros
      editor.value.on('nodeDataChanged', function(id) {
        let node_change = editor.value.getNodeFromId(id);
        let aux = node_change.data.namevalue
        
        if(node_change.name == "Number" && aux.match(/[0]\d*/)){
          editor.value.updateNodeDataFromId(id,{namevalue:parseInt(aux)})
        }
      })

      //Manage Connections
      editor.value.on('connectionCreated', function(id) {
        console.log("Connection created");
        let con_out = editor.value.getNodeFromId(id.output_id);
        let con_in = editor.value.getNodeFromId(id.input_id);
        if(con_in.inputs[id.input_class].connections.length > 1){
          editor.value.removeSingleConnection(id.output_id, id.input_id, id.output_class, id.input_class);
          return true;
        }else if(con_out.outputs[id.output_class].connections.length > 1){
          editor.value.removeSingleConnection(id.output_id, id.input_id, id.output_class, id.input_class);
          return true;
        }
        console.log(con_in)
        if(con_in.name === "Add"){ 
          if(con_in.inputs.input_1.connections.length>0 && con_in.inputs.input_2.connections.length>0){
            let input1_id = con_in.inputs.input_1.connections[0].node;
            let input2_id = con_in.inputs.input_2.connections[0].node;
            let op1 = parseInt(editor.value.getNodeFromId(input1_id).data.namevalue);
            let op2 = parseInt(editor.value.getNodeFromId(input2_id).data.namevalue);
            switch (con_in.data.opvalue){
              case "+":
                con_in.data.namevalue = op1 + op2;                
                break;
              case "-":
                con_in.data.namevalue = op1 - op2;                
                break;
              case "*":
                con_in.data.namevalue = op1 * op2; 
                break;
              case "/":
                if(op1 == 0){
                  con_in.data.namevalue = NaN;
                }else{
                  con_in.data.namevalue = op1 / op2; 
                }
                break;
              default:
            }
          }          
          editor.value.updateNodeDataFromId(id.input_id,{opvalue:con_in.data.opvalue,namevalue:con_in.data.namevalue});
        }else if(con_in.name === "Assing"){          
          con_in.data.namevalue = parseInt(con_out.data.namevalue);
          //console.log(con_in);
          editor.value.updateNodeDataFromId(id.input_id,{namevalue:con_in.data.namevalue});
        }
      }) 
    })  
    
    async function exportDataWithExecute(execute : string){

      const programName = document.getElementById("exportName").value;
      const code = editor.value.export();
      console.log(code)
      let uid = "_:program";
      const imports = document.getElementById("selectImport").options;
      for (let i=0; i<imports.length;i++){
          if(imports[i].innerText == programName){
              uid = imports[i].value;
              break;
          }          
      }
      let expotedData = {"uid":uid,"name": programName, "code":JSON.stringify(code), "exec":execute}; 

      const requestOptions = {
        method: "POST",
        headers: { "Content-Type": "application/json"},
        body: JSON.stringify(expotedData)
      };
      console.log(requestOptions.body);
      const response = await fetch("http://localhost:3333/program", requestOptions);
      const data = await response.json();
      if(execute=="true"){
         const codetoexecute = document.getElementById("codetoexecute");
         codetoexecute.value = data.stringcode
         const result = document.getElementById("result");
         result.value = data.result
      }
      console.log(data);

    }

    async function loadPrograms(){
        const response = await fetch("http://localhost:3333/program");
        const data = await response.json();
        console.log(data); 
        
        const sel = document.getElementById("selectImport");
        for (var i in data){
            sel.options.add(new Option(data[i]["Program.name"],data[i].uid));
        }
        
    }

    async function importData(){
/*
      let imp = {
    "drawflow": {
        "Home": {
            "data": {
                "1": {
                    "id": 1,
                    "name": "Add",
                    "data": {
                        "opvalue": "+",
                        "namevalue": 3
                    },
                    "class": "ClassAdd",
                    "html": "NodeAdd",
                    "typenode": "vue",
                    "inputs": {
                        "input_1": {
                            "connections": [
                                {
                                    "node": "3",
                                    "input": "output_1"
                                }
                            ]
                        },
                        "input_2": {
                            "connections": [
                                {
                                    "node": "2",
                                    "input": "output_1"
                                }
                            ]
                        }
                    },
                    "outputs": {
                        "output_1": {
                            "connections": [
                                {
                                    "node": "4",
                                    "output": "input_1"
                                }
                            ]
                        }
                    },
                    "pos_x": 228,
                    "pos_y": 23
                },
                "2": {
                    "id": 2,
                    "name": "Number",
                    "data": {
                        "namevalue": "1"
                    },
                    "class": "ClassNumber",
                    "html": "NodeNumber",
                    "typenode": "vue",
                    "inputs": {},
                    "outputs": {
                        "output_1": {
                            "connections": [
                                {
                                    "node": "1",
                                    "output": "input_2"
                                }
                            ]
                        }
                    },
                    "pos_x": -5,
                    "pos_y": 93
                },
                "3": {
                    "id": 3,
                    "name": "Number",
                    "data": {
                        "namevalue": "2"
                    },
                    "class": "ClassNumber",
                    "html": "NodeNumber",
                    "typenode": "vue",
                    "inputs": {},
                    "outputs": {
                        "output_1": {
                            "connections": [
                                {
                                    "node": "1",
                                    "output": "input_1"
                                }
                            ]
                        }
                    },
                    "pos_x": 3,
                    "pos_y": 2
                },
                "4": {
                    "id": 4,
                    "name": "Assing",
                    "data": {
                        "namevalue": 3
                    },
                    "class": "ClassAssing",
                    "html": "NodeAssing",
                    "typenode": "vue",
                    "inputs": {
                        "input_1": {
                            "connections": [
                                {
                                    "node": "1",
                                    "input": "output_1"
                                }
                            ]
                        }
                    },
                    "outputs": {
                        "output_1": {
                            "connections": [
                                {
                                    "node": "5",
                                    "output": "input_1"
                                }
                            ]
                        }
                    },
                    "pos_x": 495,
                    "pos_y": 29
                },
                "5": {
                    "id": 5,
                    "name": "Code",
                    "data": {},
                    "class": "ClassCode",
                    "html": "NodeCode",
                    "typenode": "vue",
                    "inputs": {
                        "input_1": {
                            "connections": [
                                {
                                    "node": "4",
                                    "input": "output_1"
                                }
                            ]
                        }
                    },
                    "outputs": {
                        "output_1": {
                            "connections": [
                                {
                                    "node": "6",
                                    "output": "input_1"
                                }
                            ]
                        }
                    },
                    "pos_x": 44,
                    "pos_y": 248
                },
                "6": {
                    "id": 6,
                    "name": "If",
                    "data": {
                        "condition": true
                    },
                    "class": "ClassIf",
                    "html": "NodeIf",
                    "typenode": "vue",
                    "inputs": {
                        "input_1": {
                            "connections": [
                                {
                                    "node": "5",
                                    "input": "output_1"
                                }
                            ]
                        },
                        "input_2": {
                            "connections": []
                        }
                    },
                    "outputs": {
                        "output_1": {
                            "connections": [
                                {
                                    "node": "7",
                                    "output": "input_1"
                                }
                            ]
                        }
                    },
                    "pos_x": 281,
                    "pos_y": 246
                },
                "7": {
                    "id": 7,
                    "name": "For",
                    "data": {
                        "from": 0,
                        "until": 10
                    },
                    "class": "ClassFor",
                    "html": "NodeFor",
                    "typenode": "vue",
                    "inputs": {
                        "input_1": {
                            "connections": [
                                {
                                    "node": "6",
                                    "input": "output_1"
                                }
                            ]
                        }
                    },
                    "outputs": {
                        "output_1": {
                            "connections": []
                        }
                    },
                    "pos_x": 527,
                    "pos_y": 223
                }
            }
        }
    }
}
*/
        const sel = document.getElementById("selectImport").value;
        const response = await fetch("http://localhost:3333/program/"+sel);
        const data = await response.json();
        //console.log(data["Program.code"]);

        editor.value.import(JSON.parse(data["Program.code"]));
    }

    function clear(){
        editor.value.clear()
    }

    function generateNode(op:string){                
        console.log(editor.value);

        switch (op) {
        case 'Add':
          editor.value.addNode(op, 2, 1, 0, 0, 'Class'+op, { opvalue:"+",namevalue: 0 }, 'Node'+op, 'vue');
          break;
        case 'Number':
          editor.value.addNode(op, 0, 1, 0, 0, 'Class'+op, { namevalue: 0 }, 'Node'+op, 'vue');
          break;
        case 'Var':
          editor.value.addNode(op, 0, 1, 0, 0, 'Class'+op, { namevalue: "" }, 'Node'+op, 'vue');
          break;
        case 'Assing':
          editor.value.addNode(op, 1, 1, 0, 0, 'Class'+op, { namevalue:0 }, 'Node'+op, 'vue');
          break;
        case 'If':
          editor.value.addNode(op, 2, 1, 0, 0, 'Class'+op, { condition: "" }, 'Node'+op, 'vue');
          break;
        case 'For':
          editor.value.addNode(op, 1, 1, 0, 0, 'Class'+op, { from: "", until:"" }, 'Node'+op, 'vue');
          break;
        case 'Code':
          editor.value.addNode(op, 1, 1, 0, 0, 'Class'+op, { }, 'Node'+op, 'vue');
          break;
        default:  
        }
    }

    return {
      exportEditor,generateNode,exportDataWithExecute,importData,clear
    }

  }
 
}
/*eslint-enable */
</script>
<style scoped>
#drawflow {
    width: calc(90vw - 301px);
    height: calc(100% - 50px);
    min-height: 400px;
    border: 1px solid red;
    text-align: initial;
    position: relative;
    top: 40px;
    background: var(--background-color);
    background-size: 25px 25px;
    background-image:
    linear-gradient(to right, #f1f1f1 1px, transparent 1px),
    linear-gradient(to bottom, #f1f1f1 1px, transparent 1px);
}

.wrapper {
  width: 100%;
  height: calc(100vh - 67px);
  display: flex;
}
.col {
  overflow: auto;
  width: 350px;
  height: 100%;
  border-right: 1px solid var(--border-color);
}

.button-88 {
  background-image: linear-gradient(#0dccea, #0d70ea);
  border: 0;
  border-radius: 4px;
  box-shadow: rgba(0, 0, 0, .3) 0 5px 15px;
  box-sizing: border-box;
  color: #fff;
  cursor: pointer;
  font-family: Montserrat,sans-serif;
  font-size: .9em;
  margin: 3px;
  padding: 7px 15px;
  text-align: center;
  user-select: none;
  -webkit-user-select: none;
  touch-action: manipulation;
}
</style>

<!-- HTML !-->


